package article

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() (*gin.Engine, *Service) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	repo := newArticleRepoStub()
	svc := NewService(repo)
	handler := NewHandler(svc)

	api := engine.Group("/api")
	RegisterRoutes(api, handler)

	return engine, svc
}

func TestHandler_CreateArticle(t *testing.T) {
	t.Parallel()

	router, _ := setupTestRouter()

	body := map[string]interface{}{
		"title":       "ทดสอบสร้างบทความ",
		"content":     "เนื้อหาบทความทดสอบ",
		"excerpt":     "คำอธิบายสั้น",
		"author_name": "วิภาวรรณ แสงสุข",
		"category":    "ศิลปะ",
		"status":      "draft",
		"tags":        []string{"test", "demo"},
	}
	raw, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/api/articles", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	var res Article
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if res.ID == "" || res.Title != "ทดสอบสร้างบทความ" || res.Status != StatusDraft {
		t.Fatalf("unexpected article created: %#v", res)
	}
}

func TestHandler_CreateArticle_NineFields(t *testing.T) {
	t.Parallel()

	router, _ := setupTestRouter()

	// Testing all 9 fields from user request:
	// 1. ชื่อบทความ (title)
	// 2. เนื้อหา (content)
	// 3. รูปภาพ (cover_image_url)
	// 4. สถานะการมองเห็น (status: เผยแพร่)
	// 5. วันที่เผยแพร่ (publish_date: 01/01/2569 , 10:30 AM)
	// 6. หมวดหมู่ (categories: ['ภาษา', 'ปรัชญา'])
	// 7. แท็ก (tags: ['อาหรับ', 'AI'])
	// 8. URL Slug (slug: 10-ways-minimalist-desk-productivity)
	// 9. Meta Description (meta_description: คำอธิบายสั้นๆ สำหรับ Google)
	body := map[string]interface{}{
		"title":            "10 วิธีจัดโต๊ะทำงานแบบ Minimalist เพิ่ม Productivity",
		"content":          "สภาพแวดล้อมในการทำงานมีผลอย่างมากต่อประสิทธิภาพการทำงานของเรา...",
		"cover_image_url":  "/uploads/articles/2026/10/desk-cover.webp",
		"status":           "เผยแพร่",
		"publish_date":     "01/01/2569 , 10:30 AM",
		"categories":       []string{"ภาษา", "ปรัชญา"},
		"tags":             []string{"อาหรับ", "AI"},
		"slug":             "10-ways-minimalist-desk-productivity",
		"meta_description": "คำอธิบายสั้นๆ สำหรับ Google",
	}
	raw, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/api/articles", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	var res Article
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if res.Title != "10 วิธีจัดโต๊ะทำงานแบบ Minimalist เพิ่ม Productivity" {
		t.Errorf("expected Title %q, got %q", body["title"], res.Title)
	}
	if res.CoverImageURL != "/uploads/articles/2026/10/desk-cover.webp" {
		t.Errorf("expected CoverImageURL %q, got %q", body["cover_image_url"], res.CoverImageURL)
	}
	if res.Status != StatusPublished {
		t.Errorf("expected Status %q, got %q", StatusPublished, res.Status)
	}
	if res.StatusLabel != StatusPublishedThai {
		t.Errorf("expected StatusLabel %q, got %q", StatusPublishedThai, res.StatusLabel)
	}
	if res.PublishedAt == nil || res.PublishedAt.Year() != 2026 {
		t.Errorf("expected PublishedAt year 2026, got %v", res.PublishedAt)
	}
	if len(res.Categories) != 2 || res.Categories[0] != "ภาษา" {
		t.Errorf("expected categories, got %v", res.Categories)
	}
	if len(res.Tags) != 2 || res.Tags[0] != "อาหรับ" {
		t.Errorf("expected tags, got %v", res.Tags)
	}
	if res.Slug != "10-ways-minimalist-desk-productivity" {
		t.Errorf("expected Slug %q, got %q", body["slug"], res.Slug)
	}
	if res.MetaDescription != "คำอธิบายสั้นๆ สำหรับ Google" {
		t.Errorf("expected MetaDescription %q, got %q", body["meta_description"], res.MetaDescription)
	}
}

func TestHandler_CreateArticle_BadRequest(t *testing.T) {
	t.Parallel()

	router, _ := setupTestRouter()

	// Missing title
	body := map[string]interface{}{
		"content": "เนื้อหาไม่มีชื่อ",
	}
	raw, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/api/articles", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d", rec.Code)
	}
}

func TestHandler_GetArticle(t *testing.T) {
	t.Parallel()

	router, svc := setupTestRouter()

	created, err := svc.Create(t.Context(), CreateArticleInput{
		Title:   "Article To Fetch",
		Slug:    "fetch-me",
		Content: "Content for fetching",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	// 1. Fetch by ID
	req, _ := http.NewRequest(http.MethodGet, "/api/articles/"+created.ID, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// 2. Fetch by Slug
	reqSlug, _ := http.NewRequest(http.MethodGet, "/api/articles/fetch-me", nil)
	recSlug := httptest.NewRecorder()
	router.ServeHTTP(recSlug, reqSlug)

	if recSlug.Code != http.StatusOK {
		t.Fatalf("expected status 200 for slug, got %d", recSlug.Code)
	}

	// 3. Not Found
	req404, _ := http.NewRequest(http.MethodGet, "/api/articles/does-not-exist", nil)
	rec404 := httptest.NewRecorder()
	router.ServeHTTP(rec404, req404)

	if rec404.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec404.Code)
	}
}

func TestHandler_UpdateArticle(t *testing.T) {
	t.Parallel()

	router, svc := setupTestRouter()

	created, _ := svc.Create(t.Context(), CreateArticleInput{
		Title:   "Before Update",
		Content: "Before Content",
	})

	updateBody := map[string]interface{}{
		"title":   "After Update",
		"content": "After Content",
	}
	raw, _ := json.Marshal(updateBody)

	req, _ := http.NewRequest(http.MethodPut, "/api/articles/"+created.ID, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res Article
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	if res.Title != "After Update" {
		t.Errorf("expected title 'After Update', got %q", res.Title)
	}
}

func TestHandler_UpdateStatus(t *testing.T) {
	t.Parallel()

	router, svc := setupTestRouter()

	created, _ := svc.Create(t.Context(), CreateArticleInput{
		Title:   "Status Update Article",
		Content: "Article Content",
		Status:  "draft",
	})

	statusBody := map[string]interface{}{
		"status": "published",
	}
	raw, _ := json.Marshal(statusBody)

	req, _ := http.NewRequest(http.MethodPatch, "/api/articles/"+created.ID+"/status", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res Article
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	if res.Status != StatusPublished {
		t.Errorf("expected status %q, got %q", StatusPublished, res.Status)
	}
}

func TestHandler_DeleteArticle(t *testing.T) {
	t.Parallel()

	router, svc := setupTestRouter()

	created, _ := svc.Create(t.Context(), CreateArticleInput{
		Title:   "To Delete Via API",
		Content: "Content...",
	})

	req, _ := http.NewRequest(http.MethodDelete, "/api/articles/"+created.ID, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// Verify it's gone
	reqGet, _ := http.NewRequest(http.MethodGet, "/api/articles/"+created.ID, nil)
	recGet := httptest.NewRecorder()
	router.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 after deletion, got %d", recGet.Code)
	}
}

func TestHandler_ListArticles(t *testing.T) {
	t.Parallel()

	router, svc := setupTestRouter()

	_, _ = svc.Create(t.Context(), CreateArticleInput{
		Title: "Article 1", Content: "Content 1", Category: "ศิลปะ", Status: "published",
	})
	_, _ = svc.Create(t.Context(), CreateArticleInput{
		Title: "Article 2", Content: "Content 2", Category: "ปรัชญา", Status: "draft",
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/articles?search=Article&status=all&page=1&page_size=10", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res struct {
		Data []Article `json:"data"`
		Meta struct {
			Page       int   `json:"page"`
			PageSize   int   `json:"page_size"`
			TotalItems int64 `json:"total_items"`
			TotalPages int   `json:"total_pages"`
		} `json:"meta"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(res.Data) != 2 || res.Meta.TotalItems != 2 {
		t.Fatalf("expected 2 items, got %d (total: %d)", len(res.Data), res.Meta.TotalItems)
	}
}

func TestHandler_IncrementViews(t *testing.T) {
	t.Parallel()

	router, svc := setupTestRouter()

	created, _ := svc.Create(t.Context(), CreateArticleInput{
		Title:   "Views Test API",
		Content: "Content...",
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/articles/"+created.ID+"/views", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
