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
		"author_name": "วู กานดา",
		"category":    "ทั่วไป",
		"status":      "ร่างบทความ",
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

func TestHandler_CreateArticle_BadRequest(t *testing.T) {
	t.Parallel()

	router, _ := setupTestRouter()

	// Missing title
	body := map[string]interface{}{
		"content": "เนื้อหาไม่มีชื่อเรื่อง",
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
		"status": "เผยแพร่แล้ว",
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
