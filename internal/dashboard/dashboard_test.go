package dashboard

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/article"
)

type mockDashboardRepo struct {
	trafficRecords []TrafficRecord
	articles       []article.Article
}

func (m *mockDashboardRepo) RecordTraffic(ctx context.Context, item *TrafficRecord) error {
	m.trafficRecords = append(m.trafficRecords, *item)
	return nil
}

func (m *mockDashboardRepo) GetTotalArticleViews(ctx context.Context) (int64, error) {
	var total int64
	for _, a := range m.articles {
		total += a.ViewsCount
	}
	return total, nil
}

func (m *mockDashboardRepo) GetPublishedArticlesCount(ctx context.Context) (int64, error) {
	var count int64
	for _, a := range m.articles {
		if a.Status == article.StatusPublished {
			count++
		}
	}
	return count, nil
}

func (m *mockDashboardRepo) GetTopCategories(ctx context.Context, limit int) ([]string, error) {
	return []string{"ปรัชญา", "วัฒนธรรม", "ภาษา"}, nil
}

func (m *mockDashboardRepo) GetCategoryCounts(ctx context.Context) ([]CategoryCount, error) {
	return []CategoryCount{
		{Category: "วรรณกรรม", Count: 10, Views: 3500},
		{Category: "ภาษา", Count: 8, Views: 2500},
		{Category: "วัฒนธรรม", Count: 5, Views: 1500},
		{Category: "ปรัชญา", Count: 4, Views: 1200},
		{Category: "วิจารณ์", Count: 3, Views: 800},
		{Category: "ทั่วไป", Count: 2, Views: 400},
		{Category: "อื่น ๆ", Count: 1, Views: 100},
	}, nil
}

func (m *mockDashboardRepo) GetTopArticles(ctx context.Context, limit int) ([]article.Article, error) {
	if len(m.articles) > limit {
		return m.articles[:limit], nil
	}
	return m.articles, nil
}

func (m *mockDashboardRepo) GetPublishedArticlesForExport(ctx context.Context) ([]article.Article, error) {
	var pub []article.Article
	for _, a := range m.articles {
		if a.Status == article.StatusPublished {
			pub = append(pub, a)
		}
	}
	return pub, nil
}

func (m *mockDashboardRepo) GetTrafficCountBetween(ctx context.Context, start, end time.Time) (int64, error) {
	var count int64
	for _, r := range m.trafficRecords {
		if !r.CreatedAt.Before(start) && r.CreatedAt.Before(end) {
			count++
		}
	}
	return count, nil
}

func (m *mockDashboardRepo) GetTrafficCountBySourceBetween(ctx context.Context, start, end time.Time) ([]SourceCount, error) {
	resMap := make(map[string]int64)
	for _, r := range m.trafficRecords {
		if !r.CreatedAt.Before(start) && r.CreatedAt.Before(end) {
			resMap[r.Source]++
		}
	}
	var list []SourceCount
	for k, v := range resMap {
		list = append(list, SourceCount{Source: k, Count: v})
	}
	return list, nil
}

func (m *mockDashboardRepo) GetDailyTraffic(ctx context.Context, start, end time.Time) (map[string]int64, error) {
	return map[string]int64{}, nil
}

func (m *mockDashboardRepo) GetMonthlyTraffic(ctx context.Context, monthsBack int) ([]MonthlyData, error) {
	return []MonthlyData{
		{Month: "JAN", Visits: 380, Pct: 38},
		{Month: "FEB", Visits: 490, Pct: 49},
		{Month: "MAR", Visits: 930, Pct: 93},
		{Month: "APR", Visits: 660, Pct: 66},
		{Month: "MAY", Visits: 750, Pct: 75},
		{Month: "JUN", Visits: 1000, Pct: 100},
	}, nil
}

func TestDashboardOverview(t *testing.T) {
	pubTime := time.Now()
	repo := &mockDashboardRepo{
		articles: []article.Article{
			{
				ID:          "art-1",
				Title:       "ปรัชญาสโตอิก",
				AuthorName:  "สมพร ศรีแสง",
				Category:    "ปรัชญา",
				Status:      article.StatusPublished,
				StatusLabel: "เผยแพร่แล้ว",
				ViewsCount:  12402,
				PublishedAt: &pubTime,
			},
			{
				ID:          "art-2",
				Title:       "ประวัติศาสตร์ภาษาไทย",
				AuthorName:  "กานดา วู",
				Category:    "ภาษา",
				Status:      article.StatusPublished,
				StatusLabel: "เผยแพร่แล้ว",
				ViewsCount:  8950,
				PublishedAt: &pubTime,
			},
		},
	}

	svc := NewService(repo, nil, nil)
	overview, err := svc.GetOverview(context.Background())
	if err != nil {
		t.Fatalf("unexpected error getting overview: %v", err)
	}

	if overview.Metrics.TotalViews != 21352 {
		t.Errorf("expected 21352 views, got %d", overview.Metrics.TotalViews)
	}
	if overview.Metrics.PublishedArticlesCount != 2 {
		t.Errorf("expected 2 published articles, got %d", overview.Metrics.PublishedArticlesCount)
	}
	if len(overview.Metrics.TopCategories) != 3 {
		t.Errorf("expected 3 top categories, got %d", len(overview.Metrics.TopCategories))
	}
	if len(overview.MonthlyData) != 6 {
		t.Errorf("expected 6 monthly data points, got %d", len(overview.MonthlyData))
	}
	if len(overview.TopArticles) != 2 {
		t.Errorf("expected 2 top articles, got %d", len(overview.TopArticles))
	}
}

func TestDashboardTraffic(t *testing.T) {
	repo := &mockDashboardRepo{}
	svc := NewService(repo, nil, nil)

	res, err := svc.GetTraffic(context.Background(), "7d")
	if err != nil {
		t.Fatalf("unexpected traffic error: %v", err)
	}

	if res.Filter != "7 วันล่าสุด" {
		t.Errorf("expected 7 วันล่าสุด, got %s", res.Filter)
	}
	if len(res.Sources) != 4 {
		t.Errorf("expected 4 sources, got %d", len(res.Sources))
	}
	if len(res.ChartPoints) != 7 {
		t.Errorf("expected 7 chart points, got %d", len(res.ChartPoints))
	}
}

func TestDashboardCategories(t *testing.T) {
	repo := &mockDashboardRepo{}
	svc := NewService(repo, nil, nil)

	res, err := svc.GetCategories(context.Background())
	if err != nil {
		t.Fatalf("unexpected categories error: %v", err)
	}

	if res.TotalCategoriesCount != 7 {
		t.Errorf("expected 7 categories, got %d", res.TotalCategoriesCount)
	}
	if len(res.TrendingArticles) != 4 {
		t.Errorf("expected 4 trending articles, got %d", len(res.TrendingArticles))
	}
}

func TestDashboardExportCSV(t *testing.T) {
	pubTime := time.Now()
	repo := &mockDashboardRepo{
		articles: []article.Article{
			{
				ID:          "art-1",
				Title:       "ปรัชญาสโตอิก",
				AuthorName:  "สมพร",
				Category:    "ปรัชญา",
				Status:      article.StatusPublished,
				StatusLabel: "เผยแพร่แล้ว",
				ViewsCount:  100,
				PublishedAt: &pubTime,
			},
		},
	}
	svc := NewService(repo, nil, nil)

	data, err := svc.ExportCSV(context.Background())
	if err != nil {
		t.Fatalf("unexpected export error: %v", err)
	}

	str := string(data)
	if !strings.HasPrefix(str, "\xef\xbb\xbf") {
		t.Errorf("expected UTF-8 BOM prefix")
	}
	if !strings.Contains(str, "ปรัชญาสโตอิก") {
		t.Errorf("expected CSV to contain article title")
	}
	if !strings.Contains(str, "ยอดเข้าชมทั้งหมด") {
		t.Errorf("expected CSV to contain overview metrics")
	}
}

