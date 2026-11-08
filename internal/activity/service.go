package activity

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type LogInput struct {
	Type         Type      `json:"type"`
	Action       string    `json:"action"`
	Detail       string    `json:"detail"`
	ArticleID    string    `json:"article_id"`
	ArticleTitle string    `json:"article_title"`
	ActorID      string    `json:"actor_id"`
	ActorName    string    `json:"actor_name"`
	CreatedAt    time.Time `json:"created_at"`
}

type ListResult struct {
	Items      []ActivityItem `json:"items"`
	TotalItems int64          `json:"total_items"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	HasMore    bool           `json:"has_more"`
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Log(ctx context.Context, input LogInput) (*ActivityLog, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	actType := input.Type
	if actType == "" {
		actType = TypePublish
	}

	item := &ActivityLog{
		ID:           uuid.New().String(),
		Type:         actType,
		Action:       strings.TrimSpace(input.Action),
		Detail:       strings.TrimSpace(input.Detail),
		ArticleID:    strings.TrimSpace(input.ArticleID),
		ArticleTitle: strings.TrimSpace(input.ArticleTitle),
		ActorID:      strings.TrimSpace(input.ActorID),
		ActorName:    strings.TrimSpace(input.ActorName),
		CreatedAt:    createdAt,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) List(ctx context.Context, page, limit int, search string) (*ListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	items, total, err := s.repo.List(ctx, limit, offset, strings.TrimSpace(search))
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	resultItems := make([]ActivityItem, 0, len(items))
	for _, it := range items {
		resultItems = append(resultItems, it.ToItem(now))
	}

	hasMore := int64(offset+len(items)) < total

	return &ListResult{
		Items:      resultItems,
		TotalItems: total,
		Page:       page,
		Limit:      limit,
		HasMore:    hasMore,
	}, nil
}

