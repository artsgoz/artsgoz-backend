package activity

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, log *ActivityLog) error
	List(ctx context.Context, limit, offset int, search string) ([]ActivityLog, int64, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, item *ActivityLog) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *gormRepository) List(ctx context.Context, limit, offset int, search string) ([]ActivityLog, int64, error) {
	var items []ActivityLog
	var total int64

	query := r.db.WithContext(ctx).Model(&ActivityLog{})

	if search != "" {
		s := "%" + search + "%"
		query = query.Where("action ILIKE ? OR detail ILIKE ? OR article_title ILIKE ? OR actor_name ILIKE ?", s, s, s, s)
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

