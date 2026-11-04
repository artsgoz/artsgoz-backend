package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryCount struct {
	Category string
	Count    int64
	Views    int64
}

type SourceCount struct {
	Source string
	Count  int64
}

type Repository interface {
	RecordTraffic(ctx context.Context, item *TrafficRecord) error
	RecordArticleView(ctx context.Context, articleID, ipHash, userAgent string) (int64, error)
	GetArticlesCountByStatus(ctx context.Context) (total int64, draft int64, published int64, archived int64, err error)
	GetTotalArticleViews(ctx context.Context) (int64, error)
	GetPublishedArticlesCount(ctx context.Context) (int64, error)
	GetTopCategories(ctx context.Context, limit int) ([]string, error)
	GetCategoryCounts(ctx context.Context) ([]CategoryCount, error)
	GetDistinctCategoryCount(ctx context.Context) (int, error)
	GetTopArticles(ctx context.Context, limit int) ([]ArticleSummary, error)
	GetPublishedArticlesForExport(ctx context.Context) ([]ArticleSummary, error)
	GetTrafficCountBetween(ctx context.Context, start, end time.Time) (int64, error)
	GetTrafficCountBySourceBetween(ctx context.Context, start, end time.Time) ([]SourceCount, error)
	GetDailyTraffic(ctx context.Context, start, end time.Time) (map[string]int64, error)
	GetMonthlyTraffic(ctx context.Context, monthsBack int) ([]MonthlyData, error)
	GetMonthlyTrafficByYear(ctx context.Context, year int) ([]MonthlyTrendItem, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) RecordTraffic(ctx context.Context, item *TrafficRecord) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *gormRepository) RecordArticleView(ctx context.Context, articleID, ipHash, userAgent string) (int64, error) {
	view := ArticleView{
		ArticleID: articleID,
		IPHash:    ipHash,
		UserAgent: userAgent,
		ViewedAt:  time.Now().UTC(),
	}
	_ = r.db.WithContext(ctx).Create(&view).Error

	_ = r.db.WithContext(ctx).Create(&TrafficRecord{
		ID:        uuid.New().String(),
		ArticleID: articleID,
		Source:    "organic",
		CreatedAt: time.Now().UTC(),
	}).Error

	if err := r.db.WithContext(ctx).Table("articles").
		Where("id = ?", articleID).
		Update("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
		return 0, err
	}

	var count int64
	_ = r.db.WithContext(ctx).Table("articles").
		Where("id = ?", articleID).
		Select("view_count").
		Scan(&count).Error

	return count, nil
}

func (r *gormRepository) GetArticlesCountByStatus(ctx context.Context) (total int64, draft int64, published int64, archived int64, err error) {
	type StatusCount struct {
		Status string
		Count  int64
	}
	var results []StatusCount
	err = r.db.WithContext(ctx).Table("articles").
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&results).Error
	if err != nil {
		return 0, 0, 0, 0, err
	}
	for _, sc := range results {
		total += sc.Count
		switch sc.Status {
		case "published":
			published += sc.Count
		case "draft":
			draft += sc.Count
		case "archived":
			archived += sc.Count
		}
	}
	return total, draft, published, archived, nil
}

func (r *gormRepository) GetTotalArticleViews(ctx context.Context) (int64, error) {
	var total *int64
	err := r.db.WithContext(ctx).Table("articles").
		Select("COALESCE(SUM(view_count), SUM(views_count), 0)").
		Scan(&total).Error
	if err != nil || total == nil {
		return 0, err
	}
	return *total, nil
}

func (r *gormRepository) GetPublishedArticlesCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("articles").
		Where("status = ?", "published").
		Count(&count).Error
	return count, err
}

func (r *gormRepository) GetTopCategories(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 3
	}
	var results []string
	err := r.db.WithContext(ctx).Table("articles").
		Select("category").
		Where("category != '' AND category IS NOT NULL").
		Group("category").
		Order("COALESCE(SUM(views_count), 0) DESC, COUNT(*) DESC").
		Limit(limit).
		Pluck("category", &results).Error
	return results, err
}

func (r *gormRepository) GetCategoryCounts(ctx context.Context) ([]CategoryCount, error) {
	var results []CategoryCount
	err := r.db.WithContext(ctx).Table("articles").
		Select("category, COUNT(*) as count, COALESCE(SUM(views_count), 0) as views").
		Where("category != '' AND category IS NOT NULL").
		Group("category").
		Order("views DESC, count DESC").
		Scan(&results).Error
	return results, err
}

// GetDistinctCategoryCount returns the total number of distinct non-empty categories.
func (r *gormRepository) GetDistinctCategoryCount(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("articles").
		Where("category != '' AND category IS NOT NULL").
		Distinct("category").
		Count(&count).Error
	return int(count), err
}

func (r *gormRepository) GetTopArticles(ctx context.Context, limit int) ([]ArticleSummary, error) {
	if limit <= 0 {
		limit = 5
	}
	var articles []ArticleSummary
	err := r.db.WithContext(ctx).
		Where("status = ?", "published").
		Order("views_count DESC, created_at DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func (r *gormRepository) GetPublishedArticlesForExport(ctx context.Context) ([]ArticleSummary, error) {
	var articles []ArticleSummary
	err := r.db.WithContext(ctx).
		Where("status = ?", "published").
		Order("published_at DESC, created_at DESC").
		Find(&articles).Error
	return articles, err
}

func (r *gormRepository) GetTrafficCountBetween(ctx context.Context, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TrafficRecord{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Count(&count).Error
	return count, err
}

func (r *gormRepository) GetTrafficCountBySourceBetween(ctx context.Context, start, end time.Time) ([]SourceCount, error) {
	var results []SourceCount
	err := r.db.WithContext(ctx).Model(&TrafficRecord{}).
		Select("source, COUNT(*) as count").
		Where("created_at >= ? AND created_at < ?", start, end).
		Group("source").
		Scan(&results).Error
	return results, err
}

func (r *gormRepository) GetDailyTraffic(ctx context.Context, start, end time.Time) (map[string]int64, error) {
	type DayCount struct {
		Day   string
		Count int64
	}
	var results []DayCount
	err := r.db.WithContext(ctx).Model(&TrafficRecord{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') as day, COUNT(*) as count").
		Where("created_at >= ? AND created_at < ?", start, end).
		Group("day").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	res := make(map[string]int64)
	for _, r := range results {
		res[r.Day] = r.Count
	}
	return res, nil
}

func (r *gormRepository) GetMonthlyTraffic(ctx context.Context, monthsBack int) ([]MonthlyData, error) {
	if monthsBack <= 0 {
		monthsBack = 6
	}

	monthNames := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	now := time.Now().UTC()

	var data []MonthlyData
	var maxVisits int64

	for i := monthsBack - 1; i >= 0; i-- {
		// Calculate target month start & end
		targetYear := now.Year()
		targetMonth := int(now.Month()) - i
		for targetMonth <= 0 {
			targetMonth += 12
			targetYear--
		}

		startOfMonth := time.Date(targetYear, time.Month(targetMonth), 1, 0, 0, 0, 0, time.UTC)
		endOfMonth := startOfMonth.AddDate(0, 1, 0)

		var count int64
		_ = r.db.WithContext(ctx).Model(&TrafficRecord{}).
			Where("created_at >= ? AND created_at < ?", startOfMonth, endOfMonth).
			Count(&count).Error

		if count > maxVisits {
			maxVisits = count
		}

		data = append(data, MonthlyData{
			Month:  monthNames[targetMonth-1],
			Visits: count,
		})
	}

	for i := range data {
		if maxVisits > 0 {
			data[i].Pct = int((float64(data[i].Visits) / float64(maxVisits)) * 100)
		} else {
			data[i].Pct = 0
		}
	}

	return data, nil
}

// GetMonthlyTrafficByYear returns traffic data grouped by month for a specific year (12 months).
func (r *gormRepository) GetMonthlyTrafficByYear(ctx context.Context, year int) ([]MonthlyTrendItem, error) {
	thaiShortMonths := []string{"ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."}

	data := make([]MonthlyTrendItem, 12)
	for m := 1; m <= 12; m++ {
		startOfMonth := time.Date(year, time.Month(m), 1, 0, 0, 0, 0, time.UTC)
		endOfMonth := startOfMonth.AddDate(0, 1, 0)

		var count int64
		_ = r.db.WithContext(ctx).Model(&TrafficRecord{}).
			Where("created_at >= ? AND created_at < ?", startOfMonth, endOfMonth).
			Count(&count).Error

		monthCode := fmt.Sprintf("%d-%02d", year, m)
		label := fmt.Sprintf("%s %d", thaiShortMonths[m-1], year)

		data[m-1] = MonthlyTrendItem{
			Month:  monthCode,
			Label:  label,
			Visits: count,
			Views:  count,
		}
	}

	return data, nil
}
