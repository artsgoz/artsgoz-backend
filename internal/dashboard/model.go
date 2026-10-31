package dashboard

import (
	"fmt"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/activity"
	"gorm.io/gorm"
)

// ArticleSummary is a lightweight projection of the articles table for dashboard queries.
// This allows the dashboard module to query article metrics without a direct dependency on the article package.
type ArticleSummary struct {
	ID            string         `gorm:"type:uuid;primaryKey" json:"id"`
	Slug          string         `json:"slug"`
	Title         string         `json:"title"`
	CoverImageURL string         `gorm:"column:cover_image_url" json:"cover_image_url"`
	ImageURL      string         `gorm:"column:image_url" json:"image_url,omitempty"`
	AuthorName    string         `gorm:"column:author_name" json:"author_name,omitempty"`
	Category      string         `json:"category"`
	Status        string         `json:"status"`
	StatusLabel   string         `gorm:"-" json:"status_label"`
	PublishedAt   *time.Time     `json:"published_at"`
	ViewsCount    int64          `gorm:"column:views_count" json:"views_count"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (ArticleSummary) TableName() string {
	return "articles"
}

func (a *ArticleSummary) AfterFind(_ *gorm.DB) error {
	switch a.Status {
	case "published":
		a.StatusLabel = "เผยแพร่แล้ว"
	case "pending":
		a.StatusLabel = "รออนุมัติ"
	default:
		a.Status = "draft"
		a.StatusLabel = "ร่างบทความ"
	}
	return nil
}

// TrafficRecord records page views and traffic source channels.
type TrafficRecord struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	ArticleID string    `gorm:"type:varchar(255);index" json:"article_id,omitempty"`
	Source    string    `gorm:"type:varchar(50);index;not null;default:'direct'" json:"source"` // organic, direct, social, email
	CreatedAt time.Time `gorm:"not null;index" json:"created_at"`
}

func (TrafficRecord) TableName() string {
	return "traffic_records"
}

// OverviewMetrics represents the 4 key overview metrics cards.
type OverviewMetrics struct {
	TotalViews             int64    `json:"total_views"`
	TotalViewsFormatted    string   `json:"total_views_formatted"`
	TotalViewsGrowthRate   float64  `json:"total_views_growth_rate"`
	TopCategories          []string `json:"top_categories"`
	PublishedArticlesCount int64    `json:"published_articles_count"`
	NewCommentsCount       int64    `json:"new_comments_count"`
	NewCommentsGrowthRate  float64  `json:"new_comments_growth_rate"`
}

// MonthlyData represents a bar on the overview traffic chart.
type MonthlyData struct {
	Month  string `json:"month"` // e.g. "JAN", "FEB"
	Visits int64  `json:"visits"`
	Pct    int    `json:"pct"` // percentage relative to max month (0-100)
}

// TopArticleItem represents an article in the monthly top performing articles list.
type TopArticleItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle"`
	Category    string `json:"category"`
	Status      string `json:"status"`
	Views       int64  `json:"views"`
	Thumb       string `json:"thumb"`
	PublishedAt string `json:"published_at,omitempty"`
}

// OverviewResponse is the complete payload for AksornsaraPage (Dashboard Overview).
type OverviewResponse struct {
	Metrics          OverviewMetrics         `json:"metrics"`
	MonthlyData      []MonthlyData           `json:"monthly_data"`
	TopArticles      []TopArticleItem        `json:"top_articles"`
	RecentActivities []activity.ActivityItem `json:"recent_activities"`
}

// TrafficSourceStats represents one of the 4 traffic acquisition channels.
type TrafficSourceStats struct {
	ID     int     `json:"id"`
	Label  string  `json:"label"` // Organic Search, Direct, Social Media, Email Campaigns
	Visits int64   `json:"visits"`
	Change float64 `json:"change"` // Growth rate (+% / -%)
}

// ChartPoint represents a single data point in the traffic trend line/area chart.
type ChartPoint struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

// TrafficResponse represents the response for TrafficPage.
type TrafficResponse struct {
	Filter      string               `json:"filter"` // วันนี้, 7 วันล่าสุด, 30 วันล่าสุด
	TotalVisits int64                `json:"total_visits"`
	GrowthRate  float64              `json:"growth_rate"`
	ChartPoints []ChartPoint         `json:"chart_points"`
	Sources     []TrafficSourceStats `json:"sources"`
}

// CategoryDistribution represents category percentage and UI color palette.
type CategoryDistribution struct {
	Name     string `json:"name"`
	Pct      int    `json:"pct"`
	Count    int64  `json:"count"`
	BarColor string `json:"bar_color"`
	DotColor string `json:"dot_color"`
}

// TrendingArticleItem represents a top popular article on the Categories page.
type TrendingArticleItem struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Views  string `json:"views"`
	Author string `json:"author"`
}

// CategoriesResponse represents the payload for CategoriesPage.
type CategoriesResponse struct {
	TotalCategoriesCount int                    `json:"total_categories_count"`
	Categories           []CategoryDistribution `json:"categories"`
	TrendingArticles     []TrendingArticleItem  `json:"trending_articles"`
}

// ---------------------------------------------------------------------------
// Admin Dashboard Analytics response types (spec-compliant)
// ---------------------------------------------------------------------------

// StatsResponse is the payload for GET /api/v1/admin/dashboard/stats
// Returns 4 key metric cards: total views, published articles, pending comments, total categories.
type StatsResponse struct {
	TotalViews             int64   `json:"total_views"`
	TotalViewsFormatted    string  `json:"total_views_formatted"`
	TotalViewsGrowthRate   float64 `json:"total_views_growth_rate"`
	PublishedArticlesCount int64   `json:"published_articles_count"`
	PendingCommentsCount   int64   `json:"pending_comments_count"`
	PendingCommentsGrowth  float64 `json:"pending_comments_growth_rate"`
	TotalCategoriesCount   int     `json:"total_categories_count"`
}

// MonthlyTrendItem represents a single month data point for the trend chart.
type MonthlyTrendItem struct {
	Month  string `json:"month"`  // e.g. "JAN", "FEB", ... "DEC"
	Visits int64  `json:"visits"` // total visits in that month
}

// MonthlyTrendsResponse is the payload for GET /api/v1/admin/dashboard/monthly-trends
type MonthlyTrendsResponse struct {
	Year    int                `json:"year"`
	Months  []MonthlyTrendItem `json:"months"`
}

// TopArticlesResponse is the payload for GET /api/v1/admin/dashboard/top-articles
type TopArticlesResponse struct {
	Articles []TopArticleItem `json:"articles"`
	Total    int              `json:"total"`
}

// TopCategoriesResponse is the payload for GET /api/v1/admin/dashboard/top-categories
type TopCategoriesResponse struct {
	TotalCategoriesCount int                    `json:"total_categories_count"`
	Categories           []CategoryDistribution `json:"categories"`
}

// RecentActivitiesResponse is the payload for GET /api/v1/admin/dashboard/recent-activities
type RecentActivitiesResponse struct {
	Activities []activity.ActivityItem `json:"activities"`
	Total      int                     `json:"total"`
}

// FormatViews formats an integer into a condensed readable string (e.g. 125400 -> "125.4k").
func FormatViews(views int64) string {
	if views >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(views)/1_000_000.0)
	}
	if views >= 1_000 {
		return fmt.Sprintf("%.1fk", float64(views)/1_000.0)
	}
	return fmt.Sprintf("%d", views)
}

// CalculateGrowthRate computes the percentage change between current and previous period.
func CalculateGrowthRate(current, previous int64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100.0
		}
		return 0.0
	}
	rate := (float64(current-previous) / float64(previous)) * 100.0
	return float64(int(rate*10)) / 10.0 // 1 decimal place
}
