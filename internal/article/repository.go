package article

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrArticleNotFound = errors.New("article not found")
	ErrSlugConflict    = errors.New("article slug already in use")
)

type ListParams struct {
	Search         string
	Status         string
	Category       string
	Author         string
	SortBy         string // title, published_at, date, views, created_at
	Order          string // asc, desc
	Page           int
	PageSize       int
	IncludeDeleted bool
}

type ListResult struct {
	Articles   []Article `json:"articles"`
	TotalItems int64     `json:"total_items"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}

type Repository interface {
	Create(ctx context.Context, item *Article) error
	GetByID(ctx context.Context, id string) (*Article, error)
	GetBySlug(ctx context.Context, slug string) (*Article, error)
	Update(ctx context.Context, item *Article) error
	UpdateStatus(ctx context.Context, id string, status Status, publishedAt *time.Time) (*Article, error)
	Delete(ctx context.Context, id string, hard bool) error
	List(ctx context.Context, params ListParams) (*ListResult, error)
	IncrementViews(ctx context.Context, id string) error
	SlugExists(ctx context.Context, slug string, excludeID string) (bool, error)
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, item *Article) error {
	if item == nil {
		return errors.New("article is nil")
	}
	err := r.db.WithContext(ctx).Create(item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrSlugConflict
		}
		return fmt.Errorf("create article: %w", err)
	}
	item.PopulateLabels()
	return nil
}

func (r *GormRepository) GetByID(ctx context.Context, id string) (*Article, error) {
	var item Article
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, fmt.Errorf("get article by id: %w", err)
	}
	item.PopulateLabels()
	return &item, nil
}

func (r *GormRepository) GetBySlug(ctx context.Context, slug string) (*Article, error) {
	var item Article
	err := r.db.WithContext(ctx).First(&item, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, fmt.Errorf("get article by slug: %w", err)
	}
	item.PopulateLabels()
	return &item, nil
}

func (r *GormRepository) Update(ctx context.Context, item *Article) error {
	if item == nil {
		return errors.New("article is nil")
	}
	err := r.db.WithContext(ctx).Save(item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrSlugConflict
		}
		return fmt.Errorf("update article: %w", err)
	}
	item.PopulateLabels()
	return nil
}

func (r *GormRepository) UpdateStatus(ctx context.Context, id string, status Status, publishedAt *time.Time) (*Article, error) {
	if _, err := r.GetByID(ctx, id); err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now().UTC(),
	}
	if publishedAt != nil {
		updates["published_at"] = publishedAt
	}

	if err := r.db.WithContext(ctx).Model(&Article{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update article status: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *GormRepository) Delete(ctx context.Context, id string, hard bool) error {
	var res *gorm.DB
	if hard {
		res = r.db.WithContext(ctx).Unscoped().Delete(&Article{}, "id = ?", id)
	} else {
		res = r.db.WithContext(ctx).Delete(&Article{}, "id = ?", id)
	}
	if res.Error != nil {
		return fmt.Errorf("delete article: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrArticleNotFound
	}
	return nil
}

func (r *GormRepository) List(ctx context.Context, params ListParams) (*ListResult, error) {
	query := r.db.WithContext(ctx).Model(&Article{})

	if params.IncludeDeleted {
		query = query.Unscoped()
	}

	// 1. Search filter
	if search := strings.TrimSpace(params.Search); search != "" {
		likePattern := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(title) LIKE ? OR LOWER(excerpt) LIKE ? OR LOWER(author_name) LIKE ? OR LOWER(category) LIKE ?",
			likePattern, likePattern, likePattern, likePattern,
		)
	}

	// 2. Status filter
	if status := strings.TrimSpace(params.Status); status != "" {
		normalized := strings.ToLower(status)
		if normalized != "all" && normalized != "ทั้งหมด" {
			parsed := ParseStatus(status)
			query = query.Where("status = ?", parsed)
		}
	}

	// 3. Category filter
	if cat := strings.TrimSpace(params.Category); cat != "" && cat != "ทั้งหมด" && strings.ToLower(cat) != "all" {
		query = query.Where("category = ? OR category ILIKE ?", cat, "%"+cat+"%")
	}

	// 4. Author filter
	if author := strings.TrimSpace(params.Author); author != "" {
		query = query.Where("author_name = ? OR author_id = ? OR author_name ILIKE ?", author, author, "%"+author+"%")
	}

	// Total count before pagination
	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, fmt.Errorf("count articles: %w", err)
	}

	// 5. Sorting
	sortClause := buildSortClause(params.SortBy, params.Order)
	query = query.Order(sortClause)

	// 6. Pagination
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	query = query.Limit(pageSize).Offset(offset)

	var items []Article
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list articles: %w", err)
	}

	for i := range items {
		items[i].PopulateLabels()
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(pageSize)))
	}

	return &ListResult{
		Articles:   items,
		TotalItems: totalItems,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *GormRepository) IncrementViews(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&Article{}).Where("id = ?", id).
		UpdateColumn("views_count", gorm.Expr("views_count + 1")).Error
}

func (r *GormRepository) SlugExists(ctx context.Context, slug string, excludeID string) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&Article{}).Where("slug = ?", slug)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, fmt.Errorf("check slug exists: %w", err)
	}
	return count > 0, nil
}

func buildSortClause(sortBy, order string) string {
	direction := "DESC"
	if strings.ToLower(order) == "asc" {
		direction = "ASC"
	}

	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "title":
		return fmt.Sprintf("title %s", direction)
	case "date", "published_at":
		return fmt.Sprintf("published_at %s NULLS LAST, created_at %s", direction, direction)
	case "views", "views_count":
		return fmt.Sprintf("views_count %s", direction)
	case "created_at":
		return fmt.Sprintf("created_at %s", direction)
	default:
		return "created_at DESC"
	}
}
