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
	ErrArticleNotFound  = errors.New("article not found")
	ErrSlugConflict     = errors.New("article slug already in use")
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryConflict = errors.New("category name or slug already in use")
)

type ListParams struct {
	Search         string
	Status         string
	Category       string
	CategoryID     string
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

	// Categories
	CreateCategory(ctx context.Context, cat *Category) error
	GetCategoryByID(ctx context.Context, id string) (*Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*Category, error)
	UpdateCategory(ctx context.Context, cat *Category) error
	DeleteCategory(ctx context.Context, id string) error
	ListCategories(ctx context.Context) ([]Category, error)
	CategorySlugExists(ctx context.Context, slug string, excludeID string) (bool, error)

	// Activity Logs
	CreateActivityLog(ctx context.Context, log *ActivityLog) error
	ListActivityLogs(ctx context.Context, articleID string) ([]ActivityLog, error)
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

	// 3. Category filter (supports name or category_id)
	cat := strings.TrimSpace(params.CategoryID)
	if cat == "" {
		cat = strings.TrimSpace(params.Category)
	}
	if cat != "" && cat != "ทั้งหมด" && strings.ToLower(cat) != "all" {
		query = query.Where(
			"category = ? OR category ILIKE ? OR categories LIKE ? OR category_ids LIKE ?",
			cat, "%"+cat+"%", "%\""+cat+"\"%", "%\""+cat+"\"%",
		)
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

// Category methods

func (r *GormRepository) CreateCategory(ctx context.Context, cat *Category) error {
	if cat == nil {
		return errors.New("category is nil")
	}
	err := r.db.WithContext(ctx).Create(cat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrCategoryConflict
		}
		return fmt.Errorf("create category: %w", err)
	}
	return nil
}

func (r *GormRepository) GetCategoryByID(ctx context.Context, id string) (*Category, error) {
	var cat Category
	err := r.db.WithContext(ctx).First(&cat, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category by id: %w", err)
	}
	return &cat, nil
}

func (r *GormRepository) GetCategoryBySlug(ctx context.Context, slug string) (*Category, error) {
	var cat Category
	err := r.db.WithContext(ctx).First(&cat, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category by slug: %w", err)
	}
	return &cat, nil
}

func (r *GormRepository) UpdateCategory(ctx context.Context, cat *Category) error {
	if cat == nil {
		return errors.New("category is nil")
	}
	err := r.db.WithContext(ctx).Save(cat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrCategoryConflict
		}
		return fmt.Errorf("update category: %w", err)
	}
	return nil
}

func (r *GormRepository) DeleteCategory(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&Category{}, "id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("delete category: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *GormRepository) ListCategories(ctx context.Context) ([]Category, error) {
	var items []Category
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	// Calculate article counts per category
	type catCount struct {
		Category string
		Count    int64
	}
	var counts []catCount
	_ = r.db.WithContext(ctx).Model(&Article{}).
		Select("category, count(*) as count").
		Where("deleted_at IS NULL").
		Group("category").
		Scan(&counts).Error

	countMap := make(map[string]int64, len(counts))
	for _, c := range counts {
		countMap[strings.ToLower(strings.TrimSpace(c.Category))] = c.Count
	}

	for i := range items {
		keyName := strings.ToLower(strings.TrimSpace(items[i].Name))
		keySlug := strings.ToLower(strings.TrimSpace(items[i].Slug))
		keyID := strings.ToLower(strings.TrimSpace(items[i].ID))
		if count, ok := countMap[keyName]; ok {
			items[i].ArticleCount += count
		} else if count, ok := countMap[keySlug]; ok {
			items[i].ArticleCount += count
		} else if count, ok := countMap[keyID]; ok {
			items[i].ArticleCount += count
		}
	}

	return items, nil
}

func (r *GormRepository) CategorySlugExists(ctx context.Context, slug string, excludeID string) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&Category{}).Where("slug = ?", slug)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, fmt.Errorf("check category slug exists: %w", err)
	}
	return count > 0, nil
}

// Activity Log methods

func (r *GormRepository) CreateActivityLog(ctx context.Context, log *ActivityLog) error {
	if log == nil {
		return errors.New("activity log is nil")
	}
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *GormRepository) ListActivityLogs(ctx context.Context, articleID string) ([]ActivityLog, error) {
	var logs []ActivityLog
	err := r.db.WithContext(ctx).
		Where("article_id = ?", articleID).
		Order("created_at DESC").
		Find(&logs).Error
	if err != nil {
		return nil, fmt.Errorf("list activity logs: %w", err)
	}
	return logs, nil
}

func buildSortClause(sortBy, order string) string {
	direction := "DESC"
	if strings.ToLower(strings.TrimSpace(order)) == "asc" {
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
