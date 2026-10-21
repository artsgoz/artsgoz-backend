package article

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

type articleRepoStub struct {
	mu       sync.Mutex
	articles map[string]*Article
	slugMap  map[string]string // slug -> id
}

func newArticleRepoStub() *articleRepoStub {
	return &articleRepoStub{
		articles: make(map[string]*Article),
		slugMap:  make(map[string]string),
	}
}

func (r *articleRepoStub) Create(_ context.Context, item *Article) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.slugMap[item.Slug]; exists {
		return ErrSlugConflict
	}
	item.PopulateLabels()
	copy := *item
	r.articles[item.ID] = &copy
	r.slugMap[item.Slug] = item.ID
	return nil
}

func (r *articleRepoStub) GetByID(_ context.Context, id string) (*Article, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.articles[id]
	if !exists {
		return nil, ErrArticleNotFound
	}
	item.PopulateLabels()
	copy := *item
	return &copy, nil
}

func (r *articleRepoStub) GetBySlug(_ context.Context, slug string) (*Article, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id, exists := r.slugMap[slug]
	if !exists {
		return nil, ErrArticleNotFound
	}
	item := r.articles[id]
	item.PopulateLabels()
	copy := *item
	return &copy, nil
}

func (r *articleRepoStub) Update(_ context.Context, item *Article) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.articles[item.ID]
	if !exists {
		return ErrArticleNotFound
	}

	if existing.Slug != item.Slug {
		if ownerID, slugExists := r.slugMap[item.Slug]; slugExists && ownerID != item.ID {
			return ErrSlugConflict
		}
		delete(r.slugMap, existing.Slug)
		r.slugMap[item.Slug] = item.ID
	}

	copy := *item
	r.articles[item.ID] = &copy
	return nil
}

func (r *articleRepoStub) UpdateStatus(_ context.Context, id string, status Status, publishedAt *time.Time) (*Article, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.articles[id]
	if !exists {
		return nil, ErrArticleNotFound
	}
	item.Status = status
	if publishedAt != nil {
		item.PublishedAt = publishedAt
	}
	item.PopulateLabels()
	copy := *item
	return &copy, nil
}

func (r *articleRepoStub) Delete(_ context.Context, id string, hard bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.articles[id]
	if !exists {
		return ErrArticleNotFound
	}

	if hard {
		delete(r.slugMap, item.Slug)
		delete(r.articles, id)
	} else {
		delete(r.slugMap, item.Slug)
		delete(r.articles, id)
	}
	return nil
}

func (r *articleRepoStub) List(_ context.Context, params ListParams) (*ListResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var matched []Article
	for _, a := range r.articles {
		if params.Search != "" {
			s := strings.ToLower(params.Search)
			if !strings.Contains(strings.ToLower(a.Title), s) &&
				!strings.Contains(strings.ToLower(a.Excerpt), s) &&
				!strings.Contains(strings.ToLower(a.AuthorName), s) &&
				!strings.Contains(strings.ToLower(a.Category), s) {
				continue
			}
		}

		if params.Status != "" && params.Status != "all" && params.Status != "ทั้งหมด" {
			if a.Status != ParseStatus(params.Status) {
				continue
			}
		}

		if params.Category != "" && params.Category != "all" && params.Category != "ทั้งหมด" {
			if !strings.EqualFold(a.Category, params.Category) {
				continue
			}
		}

		if params.Author != "" {
			if !strings.Contains(strings.ToLower(a.AuthorName), strings.ToLower(params.Author)) {
				continue
			}
		}

		matched = append(matched, *a)
	}

	return &ListResult{
		Articles:   matched,
		TotalItems: int64(len(matched)),
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: 1,
	}, nil
}

func (r *articleRepoStub) IncrementViews(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.articles[id]
	if !exists {
		return ErrArticleNotFound
	}
	item.ViewsCount++
	return nil
}

func (r *articleRepoStub) SlugExists(_ context.Context, slug string, excludeID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ownerID, exists := r.slugMap[slug]
	if !exists {
		return false, nil
	}
	if excludeID != "" && ownerID == excludeID {
		return false, nil
	}
	return true, nil
}

func TestCreateArticle_Draft(t *testing.T) {
	t.Parallel()

	repo := newArticleRepoStub()
	svc := NewService(repo)

	item, err := svc.Create(context.Background(), CreateArticleInput{
		Title:      "10 วิธีจัดโต๊ะทำงานแบบ Minimalist",
		Content:    "สภาพแวดล้อมในการทำงานมีผลอย่างมาก...",
		Excerpt:    "คำอธิบายสั้นๆ",
		AuthorName: "วู กานดา",
		Category:   "ทั่วไป",
		Status:     "ร่างบทความ",
		Visibility: "สาธารณะ",
		Tags:       []string{"Minimalist", "Productivity"},
	})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	if item.ID == "" {
		t.Errorf("expected generated UUID, got empty")
	}
	if item.Slug != "10-วิธีจัดโต๊ะทำงานแบบ-minimalist" {
		t.Errorf("expected slug %q, got %q", "10-วิธีจัดโต๊ะทำงานแบบ-minimalist", item.Slug)
	}
	if item.Status != StatusDraft {
		t.Errorf("expected status %q, got %q", StatusDraft, item.Status)
	}
	if item.StatusLabel != StatusDraftThai {
		t.Errorf("expected status label %q, got %q", StatusDraftThai, item.StatusLabel)
	}
	if item.Visibility != VisibilityPublic {
		t.Errorf("expected visibility %q, got %q", VisibilityPublic, item.Visibility)
	}
	if item.PublishedAt != nil {
		t.Errorf("draft article should not have published_at timestamp")
	}
}

func TestCreateArticle_Published(t *testing.T) {
	t.Parallel()

	repo := newArticleRepoStub()
	svc := NewService(repo)

	item, err := svc.Create(context.Background(), CreateArticleInput{
		Title:      "วรรณกรรมคลาสสิก",
		Content:    "เนื้อหาบทความวรรณกรรม...",
		AuthorName: "สมชาย ใจดี",
		Category:   "วรรณกรรม",
		Status:     "เผยแพร่แล้ว",
	})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	if item.Status != StatusPublished {
		t.Errorf("expected status %q, got %q", StatusPublished, item.Status)
	}
	if item.StatusLabel != StatusPublishedThai {
		t.Errorf("expected status label %q, got %q", StatusPublishedThai, item.StatusLabel)
	}
	if item.PublishedAt == nil {
		t.Errorf("published article should automatically have published_at timestamp set")
	}
}

func TestCreateArticle_Validation(t *testing.T) {
	t.Parallel()

	repo := newArticleRepoStub()
	svc := NewService(repo)

	tests := []struct {
		name    string
		input   CreateArticleInput
		wantErr error
	}{
		{
			name:    "empty title",
			input:   CreateArticleInput{Title: "   ", Content: "valid content"},
			wantErr: ErrInvalidTitle,
		},
		{
			name:    "title exceeds 500 chars",
			input:   CreateArticleInput{Title: strings.Repeat("A", 501), Content: "valid content"},
			wantErr: ErrInvalidTitle,
		},
		{
			name:    "empty content",
			input:   CreateArticleInput{Title: "Valid Title", Content: "   "},
			wantErr: ErrInvalidContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateArticle_SlugCollisionHandling(t *testing.T) {
	t.Parallel()

	repo := newArticleRepoStub()
	svc := NewService(repo)

	item1, err := svc.Create(context.Background(), CreateArticleInput{
		Title:   "Unique Title",
		Content: "Content 1",
	})
	if err != nil {
		t.Fatalf("Create 1 error: %v", err)
	}

	item2, err := svc.Create(context.Background(), CreateArticleInput{
		Title:   "Unique Title",
		Content: "Content 2",
	})
	if err != nil {
		t.Fatalf("Create 2 error: %v", err)
	}

	if item1.Slug == item2.Slug {
		t.Fatalf("expected unique slugs, both got %q", item1.Slug)
	}
	if item2.Slug != "unique-title-1" {
		t.Errorf("expected slug %q, got %q", "unique-title-1", item2.Slug)
	}
}

func TestGetArticle(t *testing.T) {
	t.Parallel()

	repo := newArticleRepoStub()
	svc := NewService(repo)

	created, err := svc.Create(context.Background(), CreateArticleInput{
		Title:   "My Article",
		Slug:    "custom-slug-abc",
		Content: "Some content",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	// 1. By ID
	byID, err := svc.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Get by ID error: %v", err)
	}
	if byID.Title != created.Title {
		t.Errorf("expected title %q, got %q", created.Title, byID.Title)
	}

	// 2. By Slug
	bySlug, err := svc.Get(context.Background(), "custom-slug-abc")
	if err != nil {
		t.Fatalf("Get by Slug error: %v", err)
	}
	if bySlug.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, bySlug.ID)
	}

	// 3. Not Found
	_, err = svc.Get(context.Background(), "non-existent")
	if !errors.Is(err, ErrArticleNotFound) {
		t.Errorf("expected ErrArticleNotFound, got %v", err)
	}
}

func TestUpdateArticle(t *testing.T) {
	t.Parallel()

	repo := newArticleRepoStub()
	svc := NewService(repo)

	created, err := svc.Create(context.Background(), CreateArticleInput{
		Title:   "Original Title",
		Content: "Original Content",
		Status:  "draft",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	newTitle := "Updated Title"
	newContent := "Updated Content"
	newStatus := "เผยแพร่แล้ว"

	updated, err := svc.Update(context.Background(), created.ID, UpdateArticleInput{
		Title:   &newTitle,
		Content: &newContent,
		Status:  &newStatus,
	})
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("expected updated title %q, got %q", "Updated Title", updated.Title)
	}
	if updated.Content != "Updated Content" {
		t.Errorf("expected updated content %q, got %q", "Updated Content", updated.Content)
	}
	if updated.Status != StatusPublished {
		t.Errorf("expected status %q, got %q", StatusPublished, updated.Status)
	}
	if updated.PublishedAt == nil {
		t.Errorf("expected PublishedAt to be set when updated to published")
	}
}

func TestUpdateStatus(t *testing.T) {
	t.Parallel()

	repo := newArticleRepoStub()
	svc := NewService(repo)

	created, err := svc.Create(context.Background(), CreateArticleInput{
		Title:   "Draft to Publish",
		Content: "Article content",
		Status:  "draft",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	// Change to published
	published, err := svc.UpdateStatus(context.Background(), created.ID, "published")
	if err != nil {
		t.Fatalf("UpdateStatus error: %v", err)
	}
	if published.Status != StatusPublished {
		t.Errorf("expected status %q, got %q", StatusPublished, published.Status)
	}
	if published.PublishedAt == nil {
		t.Errorf("expected published_at timestamp to be populated")
	}

	// Change back to draft
	draft, err := svc.UpdateStatus(context.Background(), created.ID, "ร่างบทความ")
	if err != nil {
		t.Fatalf("UpdateStatus error: %v", err)
	}
	if draft.Status != StatusDraft {
		t.Errorf("expected status %q, got %q", StatusDraft, draft.Status)
	}
}

func TestDeleteArticle(t *testing.T) {
	t.Parallel()

	repo := newArticleRepoStub()
	svc := NewService(repo)

	created, err := svc.Create(context.Background(), CreateArticleInput{
		Title:   "To Delete",
		Content: "To Delete content",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if err := svc.Delete(context.Background(), created.ID, false); err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	_, err = svc.Get(context.Background(), created.ID)
	if !errors.Is(err, ErrArticleNotFound) {
		t.Errorf("expected ErrArticleNotFound after delete, got %v", err)
	}
}

func TestListArticles_Filtering(t *testing.T) {
	t.Parallel()

	repo := newArticleRepoStub()
	svc := NewService(repo)

	// Create 3 test articles
	_, _ = svc.Create(context.Background(), CreateArticleInput{
		Title: "ปรัชญาเซนกับการใช้ชีวิต", Content: "Content A", Category: "ปรัชญา", Status: "เผยแพร่แล้ว", AuthorName: "อาจารย์กานดา",
	})
	_, _ = svc.Create(context.Background(), CreateArticleInput{
		Title: "วิวัฒนาการของภาษาไทย", Content: "Content B", Category: "ภาษา", Status: "เผยแพร่แล้ว", AuthorName: "อาจารย์สมชาย",
	})
	_, _ = svc.Create(context.Background(), CreateArticleInput{
		Title: "ร่างบทความวรรณกรรม", Content: "Content C", Category: "วรรณกรรม", Status: "ร่างบทความ", AuthorName: "อาจารย์กานดา",
	})

	// 1. Search by keyword
	res, err := svc.List(context.Background(), ListParams{Search: "เซน"})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(res.Articles) != 1 || res.Articles[0].Category != "ปรัชญา" {
		t.Errorf("expected 1 result for search 'เซน', got %d", len(res.Articles))
	}

	// 2. Filter by status: draft
	resDraft, err := svc.List(context.Background(), ListParams{Status: "ร่างบทความ"})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(resDraft.Articles) != 1 || resDraft.Articles[0].Title != "ร่างบทความวรรณกรรม" {
		t.Errorf("expected 1 draft article, got %d", len(resDraft.Articles))
	}

	// 3. Filter by author
	resAuthor, err := svc.List(context.Background(), ListParams{Author: "กานดา"})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(resAuthor.Articles) != 2 {
		t.Errorf("expected 2 articles by กานดา, got %d", len(resAuthor.Articles))
	}
}

func TestIncrementViews(t *testing.T) {
	t.Parallel()

	repo := newArticleRepoStub()
	svc := NewService(repo)

	created, err := svc.Create(context.Background(), CreateArticleInput{
		Title:   "View Counter Test",
		Content: "Content...",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if err := svc.IncrementViews(context.Background(), created.ID); err != nil {
		t.Fatalf("IncrementViews error: %v", err)
	}

	item, _ := svc.Get(context.Background(), created.ID)
	if item.ViewsCount != 1 {
		t.Errorf("expected views_count = 1, got %d", item.ViewsCount)
	}
}

func TestSlugify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"10 วิธีจัดโต๊ะทำงานแบบ Minimalist เพิ่ม Productivity", "10-วิธีจัดโต๊ะทำงานแบบ-minimalist-เพิ่ม-productivity"},
		{"   Multiple    Spaces   & Special * Characters!  ", "multiple-spaces-special-characters"},
		{"", "article"},
	}

	for _, tt := range tests {
		got := Slugify(tt.input)
		if got != tt.want {
			t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
