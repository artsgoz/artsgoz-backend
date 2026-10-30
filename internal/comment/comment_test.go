package comment

import (
	"context"
	"testing"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/activity"
)

type mockCommentRepo struct {
	items []Comment
}

func (m *mockCommentRepo) Create(ctx context.Context, item *Comment) error {
	m.items = append(m.items, *item)
	return nil
}

func (m *mockCommentRepo) Delete(ctx context.Context, id string) error {
	for i, c := range m.items {
		if c.ID == id {
			m.items = append(m.items[:i], m.items[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockCommentRepo) List(ctx context.Context, limit, offset int, search, articleID string) ([]Comment, int64, error) {
	total := int64(len(m.items))
	if offset >= len(m.items) {
		return []Comment{}, total, nil
	}
	end := offset + limit
	if end > len(m.items) {
		end = len(m.items)
	}
	return m.items[offset:end], total, nil
}

func (m *mockCommentRepo) CountSince(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	for _, it := range m.items {
		if !it.CreatedAt.Before(since) {
			count++
		}
	}
	return count, nil
}

func (m *mockCommentRepo) CountBetween(ctx context.Context, start, end time.Time) (int64, error) {
	var count int64
	for _, it := range m.items {
		if !it.CreatedAt.Before(start) && it.CreatedAt.Before(end) {
			count++
		}
	}
	return count, nil
}

type mockActivityLogger struct {
	logged []activity.LogInput
}

func (m *mockActivityLogger) Log(ctx context.Context, input activity.LogInput) (*activity.ActivityLog, error) {
	m.logged = append(m.logged, input)
	return &activity.ActivityLog{ID: "act-1"}, nil
}

func TestCommentService(t *testing.T) {
	repo := &mockCommentRepo{}
	actLogger := &mockActivityLogger{}
	svc := NewService(repo, actLogger)

	// Test Create
	created, err := svc.Create(context.Background(), CreateCommentInput{
		ArticleID:    "art-123",
		ArticleTitle: "10 Ways to Graduate in 2025",
		AuthorName:   "อรัญญา รักษาดี",
		Content:      "เนื้อหามีประโยชน์มาก ๆ เลยค่ะ",
	})
	if err != nil {
		t.Fatalf("unexpected error creating comment: %v", err)
	}
	if created.ID == "" {
		t.Errorf("expected generated ID")
	}

	// Verify Activity was logged
	if len(actLogger.logged) != 1 {
		t.Fatalf("expected 1 logged activity, got %d", len(actLogger.logged))
	}
	if actLogger.logged[0].Type != activity.TypeComment {
		t.Errorf("expected TypeComment, got %s", actLogger.logged[0].Type)
	}

	// Test List
	res, err := svc.List(context.Background(), 1, 10, "", "")
	if err != nil {
		t.Fatalf("unexpected error listing comments: %v", err)
	}
	if res.TotalItems != 1 {
		t.Errorf("expected 1 comment, got %d", res.TotalItems)
	}
	if res.Items[0].Author != "อรัญญา รักษาดี" {
		t.Errorf("expected author อรัญญา รักษาดี, got %s", res.Items[0].Author)
	}

	// Test Delete
	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("unexpected error deleting comment: %v", err)
	}
	resAfter, _ := svc.List(context.Background(), 1, 10, "", "")
	if resAfter.TotalItems != 0 {
		t.Errorf("expected 0 comments after delete, got %d", resAfter.TotalItems)
	}
}

