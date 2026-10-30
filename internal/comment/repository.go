package comment

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, comment *Comment) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int, search, articleID string) ([]Comment, int64, error)
	CountSince(ctx context.Context, since time.Time) (int64, error)
	CountBetween(ctx context.Context, start, end time.Time) (int64, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, comment *Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *gormRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Comment{}).Error
}

func (r *gormRepository) List(ctx context.Context, limit, offset int, search, articleID string) ([]Comment, int64, error) {
	var items []Comment
	var total int64

	query := r.db.WithContext(ctx).Model(&Comment{})

	if articleID != "" {
		query = query.Where("article_id = ?", articleID)
	}

	if search != "" {
		s := "%" + search + "%"
		query = query.Where("author_name ILIKE ? OR content ILIKE ? OR article_title ILIKE ?", s, s, s)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *gormRepository) CountSince(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Comment{}).Where("created_at >= ?", since).Count(&count).Error
	return count, err
}

func (r *gormRepository) CountBetween(ctx context.Context, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Comment{}).Where("created_at >= ? AND created_at < ?", start, end).Count(&count).Error
	return count, err
}

