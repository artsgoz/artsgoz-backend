package activity

import (
	"context"
	"testing"
	"time"
)

type mockActivityRepo struct {
	items []ActivityLog
}

func (m *mockActivityRepo) Create(ctx context.Context, item *ActivityLog) error {
	m.items = append(m.items, *item)
	return nil
}

func (m *mockActivityRepo) List(ctx context.Context, limit, offset int, search string) ([]ActivityLog, int64, error) {
	total := int64(len(m.items))
	if offset >= len(m.items) {
		return []ActivityLog{}, total, nil
	}
	end := offset + limit
	if end > len(m.items) {
		end = len(m.items)
	}
	return m.items[offset:end], total, nil
}

func TestActivityService(t *testing.T) {
	repo := &mockActivityRepo{}
	svc := NewService(repo)

	// Log activity
	item, err := svc.Log(context.Background(), LogInput{
		Type:         TypePublish,
		Action:       "สมพร ศรีแสง เผยแพร่บทความใหม่ ปรัชญาสโตอิก",
		Detail:       "ปรัชญาสโตอิกกับการใช้ชีวิต",
		ArticleTitle: "ปรัชญาสโตอิก",
		ActorName:    "สมพร ศรีแสง",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID == "" {
		t.Errorf("expected generated id, got empty")
	}

	// List activities
	list, err := svc.List(context.Background(), 1, 10, "")
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}
	if list.TotalItems != 1 {
		t.Errorf("expected 1 item, got %d", list.TotalItems)
	}
	if len(list.Items) != 1 {
		t.Errorf("expected 1 item in slice, got %d", len(list.Items))
	}
	if list.Items[0].TimeAgo == "" {
		t.Errorf("expected formatted time_ago, got empty")
	}
}

func TestFormatThaiTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		diff     time.Duration
		expected string
	}{
		{diff: 10 * time.Second, expected: "เมื่อสักครู่"},
		{diff: 10 * time.Minute, expected: "10 นาทีที่แล้ว"},
		{diff: 2 * time.Hour, expected: "2 ชั่วโมงที่แล้ว"},
		{diff: 25 * time.Hour, expected: "เมื่อวานนี้"},
		{diff: 3 * 24 * time.Hour, expected: "3 วันที่แล้ว"},
	}

	for _, tt := range tests {
		past := now.Add(-tt.diff)
		got := FormatThaiTimeAgo(past, now)
		if got != tt.expected {
			t.Errorf("for diff %v: expected %q, got %q", tt.diff, tt.expected, got)
		}
	}
}

