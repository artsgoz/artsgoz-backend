package article

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type Status string
type Visibility string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusPending   Status = "pending"
	StatusArchived  Status = "archived"

	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

// Thai label constants matching the Backoffice UI
const (
	StatusDraftThai     = "ร่างบทความ"
	StatusPublishedThai = "เผยแพร่แล้ว"
	StatusPendingThai   = "รออนุมัติ"
	StatusArchivedThai  = "เก็บถาวร"

	VisibilityPublicThai  = "สาธารณะ"
	VisibilityPrivateThai = "ส่วนตัว"
)

type UserContext struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"` // Writer, Editor, SuperAdmin
}

type Category struct {
	ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Slug         string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"slug"`
	Description  string    `gorm:"type:text" json:"description,omitempty"`
	Color        string    `gorm:"type:varchar(50)" json:"color,omitempty"`
	Order        int       `gorm:"column:sort_order;default:0" json:"order"`
	ArticleCount int64     `gorm:"-" json:"article_count"`
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`
}

func (Category) TableName() string {
	return "categories"
}

type ActivityLog struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	ArticleID string    `gorm:"type:uuid;index;not null" json:"article_id"`
	UserID    string    `gorm:"type:varchar(255)" json:"user_id"`
	UserName  string    `gorm:"type:varchar(255)" json:"user_name"`
	UserRole  string    `gorm:"type:varchar(50)" json:"user_role"`
	Action    string    `gorm:"type:varchar(50);not null" json:"action"` // CREATE, UPDATE, PUBLISH, ARCHIVE, STATUS_CHANGE, DELETE
	Details   string    `gorm:"type:text" json:"details,omitempty"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (ActivityLog) TableName() string {
	return "article_activity_logs"
}

type Article struct {
	ID                 string         `gorm:"type:uuid;primaryKey" json:"id"`
	Slug               string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Title              string         `gorm:"type:varchar(500);not null" json:"title"`
	Subtitle           string         `gorm:"type:text" json:"subtitle,omitempty"`
	Excerpt            string         `gorm:"type:text" json:"excerpt,omitempty"`
	Content            string         `gorm:"type:text;not null" json:"content"`
	CoverImageURL      string         `gorm:"type:text;column:cover_image_url" json:"cover_image_url"`
	ImageURL           string         `gorm:"type:text;column:image_url" json:"image_url,omitempty"`
	AuthorID           string         `gorm:"type:varchar(255);index" json:"author_id,omitempty"`
	AuthorName         string         `gorm:"type:varchar(255)" json:"author_name,omitempty"`
	Category           string         `gorm:"type:varchar(100);index" json:"category"`
	Categories         []string       `gorm:"serializer:json" json:"categories"`
	CategoryIDs        []string       `gorm:"serializer:json" json:"category_ids,omitempty"`
	Tags               []string       `gorm:"serializer:json" json:"tags"`
	Status             Status         `gorm:"type:varchar(50);index;not null;default:'draft'" json:"status"`
	StatusLabel        string         `gorm:"-" json:"status_label"`
	Visibility         Visibility     `gorm:"type:varchar(50);not null;default:'public'" json:"visibility"`
	VisibilityLabel    string         `gorm:"-" json:"visibility_label"`
	PublishedAt        *time.Time     `gorm:"index" json:"published_at"`
	MetaDescription    string         `gorm:"type:text;column:meta_description" json:"meta_description"`
	SEOMetaDescription string         `gorm:"type:text;column:seo_meta_description" json:"seo_meta_description,omitempty"`
	ViewsCount         int64          `gorm:"not null;default:0" json:"views_count"`
	References         []string       `gorm:"serializer:json" json:"references,omitempty"`
	CreatedAt          time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Article) TableName() string {
	return "articles"
}

func (a *Article) AfterFind(_ *gorm.DB) error {
	a.PopulateLabels()
	return nil
}

func (a *Article) PopulateLabels() {
	switch a.Status {
	case StatusPublished:
		a.StatusLabel = StatusPublishedThai
	case StatusPending:
		a.StatusLabel = StatusPendingThai
	case StatusArchived:
		a.StatusLabel = StatusArchivedThai
	default:
		a.Status = StatusDraft
		a.StatusLabel = StatusDraftThai
	}

	switch a.Visibility {
	case VisibilityPrivate:
		a.VisibilityLabel = VisibilityPrivateThai
	default:
		a.Visibility = VisibilityPublic
		a.VisibilityLabel = VisibilityPublicThai
	}

	// Sync CoverImageURL and ImageURL
	if a.CoverImageURL == "" && a.ImageURL != "" {
		a.CoverImageURL = a.ImageURL
	} else if a.ImageURL == "" && a.CoverImageURL != "" {
		a.ImageURL = a.CoverImageURL
	}

	// Sync MetaDescription and SEOMetaDescription
	if a.MetaDescription == "" && a.SEOMetaDescription != "" {
		a.MetaDescription = a.SEOMetaDescription
	} else if a.SEOMetaDescription == "" && a.MetaDescription != "" {
		a.SEOMetaDescription = a.MetaDescription
	}

	// Sync Category and Categories
	if a.Categories == nil {
		a.Categories = []string{}
	}
	if len(a.Categories) == 0 && a.Category != "" {
		a.Categories = []string{a.Category}
	} else if a.Category == "" && len(a.Categories) > 0 {
		a.Category = a.Categories[0]
	}

	if a.CategoryIDs == nil {
		a.CategoryIDs = a.Categories
	}
	if a.Tags == nil {
		a.Tags = []string{}
	}
	if a.References == nil {
		a.References = []string{}
	}
}

func ParseStatus(s string) Status {
	clean := strings.ToLower(strings.TrimSpace(s))
	switch clean {
	case "published", "เผยแพร่", "เผยแพร่แล้ว":
		return StatusPublished
	case "pending", "รออนุมัติ":
		return StatusPending
	case "archived", "เก็บถาวร":
		return StatusArchived
	default:
		return StatusDraft
	}
}

func ParseVisibility(v string) Visibility {
	clean := strings.ToLower(strings.TrimSpace(v))
	switch clean {
	case "private", "ส่วนตัว":
		return VisibilityPrivate
	default:
		return VisibilityPublic
	}
}
