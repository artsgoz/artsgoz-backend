package dashboard

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/activity"
	"github.com/artsgoz/artsgoz-backend/internal/server"
	"github.com/gin-gonic/gin"
)

// mockDashboardRepo is an in-memory stub supporting full concurrency and invariant checks.
type mockDashboardRepo struct {
	mu           sync.RWMutex
	articles     map[string]*ArticleSummary
	views        []ArticleView
	traffic      []TrafficRecord
	articleViews map[string]*int64 // atomic counters for each article
}

func newMockDashboardRepo() *mockDashboardRepo {
	return &mockDashboardRepo{
		articles:     make(map[string]*ArticleSummary),
		views:        make([]ArticleView, 0),
		traffic:      make([]TrafficRecord, 0),
		articleViews: make(map[string]*int64),
	}
}

func (m *mockDashboardRepo) addArticle(a ArticleSummary) {
	m.mu.Lock()
	defer m.mu.Unlock()
	copy := a
	m.articles[a.ID] = &copy
	var c int64 = a.ViewsCount
	m.articleViews[a.ID] = &c
}

func (m *mockDashboardRepo) RecordTraffic(_ context.Context, item *TrafficRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.traffic = append(m.traffic, *item)
	return nil
}

func (m *mockDashboardRepo) RecordArticleView(_ context.Context, articleID, ipHash, userAgent string) (int64, error) {
	m.mu.Lock()
	m.views = append(m.views, ArticleView{
		ID:        int64(len(m.views) + 1),
		ArticleID: articleID,
		IPHash:    ipHash,
		UserAgent: userAgent,
		ViewedAt:  time.Now().UTC(),
	})
	m.traffic = append(m.traffic, TrafficRecord{
		ID:        fmt.Sprintf("tr-%d", len(m.traffic)+1),
		ArticleID: articleID,
		Source:    "organic",
		CreatedAt: time.Now().UTC(),
	})
	counterPtr, exists := m.articleViews[articleID]
	if !exists {
		var zero int64 = 0
		m.articleViews[articleID] = &zero
		counterPtr = &zero
	}
	if art, ok := m.articles[articleID]; ok {
		art.ViewsCount++
	}
	m.mu.Unlock()

	// Atomic increment to simulate Postgres UPDATE view_count = view_count + 1
	newVal := atomic.AddInt64(counterPtr, 1)
	return newVal, nil
}

func (m *mockDashboardRepo) GetArticlesCountByStatus(_ context.Context) (total int64, draft int64, published int64, archived int64, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, a := range m.articles {
		total++
		switch a.Status {
		case "published":
			published++
		case "draft":
			draft++
		case "archived":
			archived++
		}
	}
	return total, draft, published, archived, nil
}

func (m *mockDashboardRepo) GetTotalArticleViews(_ context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var total int64
	for _, ptr := range m.articleViews {
		total += atomic.LoadInt64(ptr)
	}
	return total, nil
}

func (m *mockDashboardRepo) GetPublishedArticlesCount(_ context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var count int64
	for _, a := range m.articles {
		if a.Status == "published" {
			count++
		}
	}
	return count, nil
}

func (m *mockDashboardRepo) GetTopCategories(_ context.Context, limit int) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cats := []string{"ศิลปะและการออกแบบ", "ปรัชญา", "วัฒนธรรม"}
	if limit < len(cats) {
		return cats[:limit], nil
	}
	return cats, nil
}

func (m *mockDashboardRepo) GetCategoryCounts(_ context.Context) ([]CategoryCount, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return []CategoryCount{
		{Category: "ปรัชญา", Count: 16, Views: 4000},
		{Category: "ศิลปะและการออกแบบ", Count: 12, Views: 3000},
		{Category: "วัฒนธรรม", Count: 8, Views: 2000},
		{Category: "ภาษา", Count: 4, Views: 1000},
	}, nil
}

func (m *mockDashboardRepo) GetDistinctCategoryCount(_ context.Context) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	set := make(map[string]bool)
	for _, a := range m.articles {
		if a.Category != "" {
			set[a.Category] = true
		}
	}
	return len(set), nil
}

func (m *mockDashboardRepo) GetTopArticles(_ context.Context, limit int) ([]ArticleSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []ArticleSummary
	for _, a := range m.articles {
		if a.Status == "published" {
			copy := *a
			if ptr, ok := m.articleViews[a.ID]; ok {
				copy.ViewsCount = atomic.LoadInt64(ptr)
			}
			list = append(list, copy)
		}
	}
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (m *mockDashboardRepo) GetPublishedArticlesForExport(_ context.Context) ([]ArticleSummary, error) {
	return m.GetTopArticles(context.Background(), 100)
}

func (m *mockDashboardRepo) GetTrafficCountBetween(_ context.Context, _, _ time.Time) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.traffic)), nil
}

func (m *mockDashboardRepo) GetTrafficCountBySourceBetween(_ context.Context, _, _ time.Time) ([]SourceCount, error) {
	return []SourceCount{
		{Source: "organic", Count: 12450},
		{Source: "direct", Count: 5820},
		{Source: "social", Count: 4105},
		{Source: "email", Count: 2217},
	}, nil
}

func (m *mockDashboardRepo) GetDailyTraffic(_ context.Context, _, _ time.Time) (map[string]int64, error) {
	return map[string]int64{
		"2026-11-01": 800,
	}, nil
}

func (m *mockDashboardRepo) GetMonthlyTraffic(_ context.Context, monthsBack int) ([]MonthlyData, error) {
	return []MonthlyData{
		{Month: "JUN", Visits: 12000, Pct: 100},
		{Month: "JUL", Visits: 9500, Pct: 79},
		{Month: "AUG", Visits: 11000, Pct: 91},
		{Month: "SEP", Visits: 10200, Pct: 85},
		{Month: "OCT", Visits: 11500, Pct: 95},
		{Month: "NOV", Visits: 12500, Pct: 100},
	}, nil
}

func (m *mockDashboardRepo) GetMonthlyTrafficByYear(_ context.Context, year int) ([]MonthlyTrendItem, error) {
	monthNames := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	items := make([]MonthlyTrendItem, 12)
	for i, m := range monthNames {
		items[i] = MonthlyTrendItem{
			Month:  m,
			Visits: int64((i + 1) * 1250),
		}
	}
	return items, nil
}

// setupE2ETestEngine sets up Gin engine with real auth middleware and dashboard routes.
func setupE2ETestEngine(repo *mockDashboardRepo, actSvc *activity.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	svc := NewService(repo, nil, actSvc)
	h := NewHandler(svc)

	apiV1 := engine.Group("/api/v1")
	RegisterRoutes(apiV1, h)

	adminV1 := apiV1.Group("/admin",
		server.AuthRequired(),
		server.RequireRoles(server.RoleEditor, server.RoleSuperAdmin),
	)
	RegisterAdminRoutes(adminV1, h)

	return engine
}

// ==============================================================================
// TEST CASES
// ==============================================================================

// TC-VIEW-001: [Privacy & Audit] บันทึก View Event พร้อมตรวจสอบ SHA-256 IP Hashing
func TestE2E_TC_VIEW_001_RecordViewEvent_SHA256_Privacy(t *testing.T) {
	repo := newMockDashboardRepo()
	repo.addArticle(ArticleSummary{
		ID:         "art-001",
		Title:      "ศิลปะฟอนต์ไทย",
		Status:     "published",
		ViewsCount: 0,
	})

	engine := setupE2ETestEngine(repo, nil)

	clientIP := "203.0.113.195"
	expectedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(clientIP)))

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/articles/art-001/views", strings.NewReader(`{"source":"organic"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", clientIP)
	req.Header.Set("User-Agent", "Mozilla/5.0 QA-Automation-Test")

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("[TC-VIEW-001] expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res struct {
		Status    string `json:"status"`
		ViewCount int64  `json:"view_count"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("[TC-VIEW-001] unmarshal error: %v", err)
	}

	if res.Status != "ok" || res.ViewCount != 1 {
		t.Errorf("[TC-VIEW-001] expected status 'ok' and view_count 1, got status=%q, view_count=%d", res.Status, res.ViewCount)
	}

	// Verify database record
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	if len(repo.views) != 1 {
		t.Fatalf("[TC-VIEW-001] expected 1 view recorded in article_views, got %d", len(repo.views))
	}

	savedView := repo.views[0]
	if savedView.IPHash != expectedHash {
		t.Errorf("[TC-VIEW-001] expected ip_hash=%q, got %q", expectedHash, savedView.IPHash)
	}
	if len(savedView.IPHash) != 64 {
		t.Errorf("[TC-VIEW-001] expected 64-char hex SHA-256 hash, got length %d", len(savedView.IPHash))
	}
	// Verify raw IP is NOT stored
	if strings.Contains(savedView.IPHash, clientIP) {
		t.Errorf("[TC-VIEW-001] privacy violation: raw IP leaked in ip_hash field!")
	}
}

// TC-VIEW-002: [Concurrency] ทดสอบการยิงเปิดอ่านพร้อมกัน 50 Concurrent Requests (Atomic Increment)
func TestE2E_TC_VIEW_002_Concurrency_CounterCacheAtomic(t *testing.T) {
	repo := newMockDashboardRepo()
	repo.addArticle(ArticleSummary{
		ID:         "art-concurrent-002",
		Title:      "การทดสอบความพร้อมเพรียง Concurrency",
		Status:     "published",
		ViewsCount: 0,
	})

	engine := setupE2ETestEngine(repo, nil)

	concurrentRequests := 50
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func(idx int) {
			defer wg.Done()
			ip := fmt.Sprintf("198.51.100.%d", idx)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/articles/art-concurrent-002/views", strings.NewReader(`{"source":"direct"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Forwarded-For", ip)

			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("[TC-VIEW-002] request %d failed with code %d", idx, rec.Code)
			}
		}(i)
	}

	wg.Wait()

	totalViews, _ := repo.GetTotalArticleViews(context.Background())
	if totalViews != int64(concurrentRequests) {
		t.Fatalf("[TC-VIEW-002] Race condition detected: expected %d total views, got %d", concurrentRequests, totalViews)
	}
}

// TC-DASH-001: [Happy Path] Overview Metrics Endpoint (การ์ดสรุปยอดรวมและสถานะ)
func TestE2E_TC_DASH_001_OverviewMetrics_Invariants(t *testing.T) {
	repo := newMockDashboardRepo()
	repo.addArticle(ArticleSummary{ID: "art-1", Status: "published", Category: "ศิลปะ", ViewsCount: 150})
	repo.addArticle(ArticleSummary{ID: "art-2", Status: "published", Category: "ปรัชญา", ViewsCount: 250})
	repo.addArticle(ArticleSummary{ID: "art-3", Status: "draft", Category: "วัฒนธรรม", ViewsCount: 0})
	repo.addArticle(ArticleSummary{ID: "art-4", Status: "archived", Category: "ภาษา", ViewsCount: 50})

	engine := setupE2ETestEngine(repo, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	req.Header.Set("Authorization", "Bearer mock.jwt.admin")
	req.Header.Set("X-User-Role", "editor")

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("[TC-DASH-001] expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var res StatsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("[TC-DASH-001] JSON unmarshal error: %v", err)
	}

	// Verify Mathematical Invariants
	// 1. Total Articles = Published (2) + Draft (1) + Archived (1) = 4
	if res.TotalArticlesCount != 4 {
		t.Errorf("[TC-DASH-001] expected TotalArticlesCount=4, got %d", res.TotalArticlesCount)
	}
	if res.PublishedArticlesCount != 2 {
		t.Errorf("[TC-DASH-001] expected PublishedArticlesCount=2, got %d", res.PublishedArticlesCount)
	}
	if res.DraftArticlesCount != 1 {
		t.Errorf("[TC-DASH-001] expected DraftArticlesCount=1, got %d", res.DraftArticlesCount)
	}
	if res.ArchivedArticlesCount != 1 {
		t.Errorf("[TC-DASH-001] expected ArchivedArticlesCount=1, got %d", res.ArchivedArticlesCount)
	}
	if res.TotalArticlesCount != (res.PublishedArticlesCount + res.DraftArticlesCount + res.ArchivedArticlesCount) {
		t.Errorf("[TC-DASH-001] invariant failed: Total != Published + Draft + Archived")
	}

	// 2. Total Views = 150 + 250 + 0 + 50 = 450
	if res.TotalViews != 450 {
		t.Errorf("[TC-DASH-001] expected TotalViews=450, got %d", res.TotalViews)
	}
	if res.TotalViewsFormatted != "450" {
		t.Errorf("[TC-DASH-001] expected formatted '450', got %q", res.TotalViewsFormatted)
	}
}

// TC-DASH-002: [Time-Series] Analytics by Time Range Endpoint (?period=7d)
func TestE2E_TC_DASH_002_Analytics_7Days(t *testing.T) {
	repo := newMockDashboardRepo()
	engine := setupE2ETestEngine(repo, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/traffic?period=7d", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("[TC-DASH-002] expected 200 OK, got %d", rec.Code)
	}

	var res TrafficResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("[TC-DASH-002] unmarshal error: %v", err)
	}

	if len(res.ChartPoints) != 7 {
		t.Errorf("[TC-DASH-002] expected 7 chart points for 7d period, got %d", len(res.ChartPoints))
	}
	for i, pt := range res.ChartPoints {
		if pt.Label == "" {
			t.Errorf("[TC-DASH-002] chart point %d missing label", i)
		}
		if pt.Value < 0 {
			t.Errorf("[TC-DASH-002] chart point %d has negative value %d", i, pt.Value)
		}
	}
	if len(res.Sources) != 4 {
		t.Errorf("[TC-DASH-002] expected 4 traffic acquisition sources, got %d", len(res.Sources))
	}
}

// TC-DASH-003: [Time-Series] Analytics by Time Range Endpoint (?period=30d)
func TestE2E_TC_DASH_003_Analytics_30Days(t *testing.T) {
	repo := newMockDashboardRepo()
	engine := setupE2ETestEngine(repo, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/traffic?period=30d", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("[TC-DASH-003] expected 200 OK, got %d", rec.Code)
	}

	var res TrafficResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	if res.Filter != "30 วันล่าสุด" {
		t.Errorf("[TC-DASH-003] expected filter '30 วันล่าสุด', got %q", res.Filter)
	}
	if len(res.ChartPoints) == 0 {
		t.Errorf("[TC-DASH-003] expected chart points for 30d, got 0")
	}
}

// TC-DASH-004: [Category Analytics] Top Categories Endpoint (Donut Chart Ready)
func TestE2E_TC_DASH_004_TopCategoriesDistribution(t *testing.T) {
	repo := newMockDashboardRepo()
	engine := setupE2ETestEngine(repo, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/top-categories", nil)
	req.Header.Set("Authorization", "Bearer mock.jwt.admin")
	req.Header.Set("X-User-Role", "editor")

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("[TC-DASH-004] expected 200 OK, got %d", rec.Code)
	}

	var res TopCategoriesResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &res)

	if len(res.Categories) == 0 {
		t.Fatalf("[TC-DASH-004] expected categories list, got 0")
	}

	var sumPct int
	for _, c := range res.Categories {
		sumPct += c.Pct
		if c.BarColor == "" || c.DotColor == "" {
			t.Errorf("[TC-DASH-004] category %s missing chart color palette", c.Name)
		}
	}
	// Invariant: Percentage sum should be approx 100% (allowing integer rounding tolerance)
	if sumPct < 95 || sumPct > 105 {
		t.Errorf("[TC-DASH-004] expected sum of percentages ~100%%, got %d%%", sumPct)
	}
}

// TC-ACT-001: [Happy Path] Recent Activities Endpoint สำหรับ Admin (Ordered created_at DESC)
func TestE2E_TC_ACT_001_RecentActivities_Admin_Descending(t *testing.T) {
	repo := newMockDashboardRepo()

	// In-memory activity repo
	actRepo := &mockActivityRepo{
		logs: []activity.ActivityLog{
			{
				ID:           "act-1",
				Type:         activity.TypePublish,
				Action:       "เผยแพร่บทความใหม่",
				Detail:       "ศิลปะฟอนต์ไทยร่วมสมัย",
				ArticleID:    "art-001",
				ArticleTitle: "ศิลปะฟอนต์ไทยร่วมสมัย",
				EntityType:   "article",
				EntityID:     "art-001",
				ActorName:    "บรรณาธิการใหญ่",
				CreatedAt:    time.Now().UTC().Add(-10 * time.Minute),
			},
			{
				ID:           "act-2",
				Type:         activity.TypeComment,
				Action:       "อนุมัติความคิดเห็นใหม่",
				Detail:       "ความคิดเห็นใหม่จากผู้อ่าน",
				ArticleID:    "art-001",
				ArticleTitle: "ศิลปะฟอนต์ไทยร่วมสมัย",
				EntityType:   "comment",
				EntityID:     "com-001",
				ActorName:    "สมพร",
				CreatedAt:    time.Now().UTC().Add(-2 * time.Minute), // Later than act-1
			},
		},
	}
	actSvc := activity.NewService(actRepo)
	engine := setupE2ETestEngine(repo, actSvc)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/recent-activities?limit=10", nil)
	req.Header.Set("Authorization", "Bearer mock.jwt.admin")
	req.Header.Set("X-User-Role", "editor")

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("[TC-ACT-001] expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var res RecentActivitiesResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &res)

	if len(res.Activities) < 2 {
		t.Fatalf("[TC-ACT-001] expected at least 2 activities, got %d", len(res.Activities))
	}

	// Verify descending order by CreatedAt
	for i := 0; i < len(res.Activities)-1; i++ {
		curr := res.Activities[i].CreatedAt
		next := res.Activities[i+1].CreatedAt
		if curr.Before(next) {
			t.Errorf("[TC-ACT-001] activities not ordered descending: %v was before %v", curr, next)
		}
	}

	// Verify essential audit fields
	first := res.Activities[0]
	if first.Action == "" || first.EntityType == "" || first.EntityID == "" {
		t.Errorf("[TC-ACT-001] missing essential audit trail fields: action=%q, entity_type=%q, entity_id=%q",
			first.Action, first.EntityType, first.EntityID)
	}
}

// TC-ACT-002: [Security / RBAC] สมาชิกทั่วไป (Role: Member) พยายามเข้าถึง Recent Activities
func TestE2E_TC_ACT_002_RecentActivities_ForbiddenForMember(t *testing.T) {
	repo := newMockDashboardRepo()
	engine := setupE2ETestEngine(repo, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/recent-activities", nil)
	req.Header.Set("Authorization", "Bearer mock.jwt.member")
	req.Header.Set("X-User-Role", "member") // Non-admin role

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("[TC-ACT-002] expected 403 Forbidden for member, got %d: %s", rec.Code, rec.Body.String())
	}

	var errRes map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &errRes)
	if errObj, ok := errRes["error"].(map[string]interface{}); ok {
		if errObj["code"] != "forbidden" {
			t.Errorf("[TC-ACT-002] expected error code 'forbidden', got %v", errObj["code"])
		}
	}
}

// mockActivityRepo implements activity.Repository for testing
type mockActivityRepo struct {
	logs []activity.ActivityLog
}

func (m *mockActivityRepo) Create(_ context.Context, item *activity.ActivityLog) error {
	m.logs = append(m.logs, *item)
	return nil
}

func (m *mockActivityRepo) List(_ context.Context, limit, offset int, _ string) ([]activity.ActivityLog, int64, error) {
	// Sort descending
	sorted := make([]activity.ActivityLog, len(m.logs))
	copy(sorted, m.logs)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].CreatedAt.Before(sorted[j].CreatedAt) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	if offset >= len(sorted) {
		return []activity.ActivityLog{}, int64(len(sorted)), nil
	}
	end := offset + limit
	if end > len(sorted) {
		end = len(sorted)
	}
	return sorted[offset:end], int64(len(sorted)), nil
}

func (m *mockActivityRepo) CountSince(_ context.Context, _ time.Time) (int64, error) {
	return int64(len(m.logs)), nil
}
