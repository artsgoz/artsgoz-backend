package comment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/activity"
	"github.com/google/uuid"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrInvalidContent  = errors.New("content is required")
	ErrInvalidAuthor   = errors.New("author name is required")
	ErrInvalidArticle  = errors.New("article id is required")
)

type ActivityLogger interface {
	Log(ctx context.Context, input activity.LogInput) (*activity.ActivityLog, error)
}

type CreateCommentInput struct {
	ArticleID    string `json:"article_id"`
	ArticleTitle string `json:"article_title"`
	AuthorID     string `json:"author_id"`
	AuthorName   string `json:"author_name"`
	AuthorAvatar string `json:"author_avatar"`
	Content      string `json:"content"`
}

type ListResult struct {
	Items      []CommentItem `json:"items"`
	TotalItems int64         `json:"total_items"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	HasMore    bool          `json:"has_more"`
}

type Service struct {
	repo           Repository
	activityLogger ActivityLogger
}

func NewService(repo Repository, activityLogger ActivityLogger) *Service {
	return &Service{
		repo:           repo,
		activityLogger: activityLogger,
	}
}

func (s *Service) Create(ctx context.Context, input CreateCommentInput) (*Comment, error) {
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, ErrInvalidContent
	}

	authorName := strings.TrimSpace(input.AuthorName)
	if authorName == "" {
		return nil, ErrInvalidAuthor
	}

	articleID := strings.TrimSpace(input.ArticleID)
	if articleID == "" {
		return nil, ErrInvalidArticle
	}

	articleTitle := strings.TrimSpace(input.ArticleTitle)
	if articleTitle == "" {
		articleTitle = "บทความ"
	}

	now := time.Now().UTC()
	item := &Comment{
		ID:           uuid.New().String(),
		ArticleID:    articleID,
		ArticleTitle: articleTitle,
		AuthorID:     strings.TrimSpace(input.AuthorID),
		AuthorName:   authorName,
		AuthorAvatar: strings.TrimSpace(input.AuthorAvatar),
		Content:      content,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}

	// Record activity feed event for new comment
	if s.activityLogger != nil {
		action := fmt.Sprintf("ความคิดเห็นใหม่ในบทความ %s", articleTitle)
		detail := fmt.Sprintf("%q โดย %s", content, authorName)
		_, _ = s.activityLogger.Log(ctx, activity.LogInput{
			Type:         activity.TypeComment,
			Action:       action,
			Detail:       detail,
			ArticleID:    articleID,
			ArticleTitle: articleTitle,
			ActorID:      item.AuthorID,
			ActorName:    authorName,
			CreatedAt:    now,
		})
	}

	return item, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrCommentNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, page, limit int, search, articleID string) (*ListResult, error) {
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
	items, total, err := s.repo.List(ctx, limit, offset, strings.TrimSpace(search), strings.TrimSpace(articleID))
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	resultItems := make([]CommentItem, 0, len(items))
	for _, it := range items {
		resultItems = append(resultItems, CommentItem{
			ID:           it.ID,
			Author:       it.AuthorName,
			AuthorAvatar: it.AuthorAvatar,
			Content:      it.Content,
			ArticleTitle: it.ArticleTitle,
			ArticleID:    it.ArticleID,
			TimeAgo:      activity.FormatThaiTimeAgo(it.CreatedAt, now),
			CreatedAt:    it.CreatedAt,
		})
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

