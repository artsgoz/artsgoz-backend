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

	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

// Thai label constants matching the Backoffice UI
const (
	StatusDraftThai     = "ร่างบทความ"
	StatusPublishedThai = "เผยแพร่แล้ว"
	StatusPendingThai   = "รออนุมัติ"

	VisibilityPublicThai  = "สาธารณะ"
	VisibilityPrivateThai = "ส่วนตัว"
)

type Article struct {
	ID                 string         `gorm:"type:uuid;primaryKey" json:"id"`
	Slug               string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Title              string         `gorm:"type:varchar(500);not null" json:"title"`
	Subtitle           string         `gorm:"type:text" json:"subtitle,omitempty"`
	Excerpt            string         `gorm:"type:text" json:"excerpt"`
	Content            string         `gorm:"type:text;not null" json:"content"`
	CoverImageURL      string         `gorm:"type:text" json:"cover_image_url"`
	AuthorID           string         `gorm:"type:varchar(255);index" json:"author_id"`
	AuthorName         string         `gorm:"type:varchar(255)" json:"author_name"`
	CategoryIDs        []string       `gorm:"serializer:json" json:"category_ids"`
	Category           string         `gorm:"type:varchar(100);index" json:"category"`
	Tags               []string       `gorm:"serializer:json" json:"tags"`
	Status             Status         `gorm:"type:varchar(50);index;not null;default:'draft'" json:"status"`
	StatusLabel        string         `gorm:"-" json:"status_label"`
	Visibility         Visibility     `gorm:"type:varchar(50);not null;default:'public'" json:"visibility"`
	VisibilityLabel    string         `gorm:"-" json:"visibility_label"`
	PublishedAt        *time.Time     `gorm:"index" json:"published_at"`
	SEOMetaDescription string         `gorm:"type:text" json:"seo_meta_description"`
	ViewsCount         int64          `gorm:"not null;default:0" json:"views_count"`
	References         []string       `gorm:"serializer:json" json:"references"`
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

	if a.CategoryIDs == nil {
		a.CategoryIDs = []string{}
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
	case "published", "เผยแพร่แล้ว":
		return StatusPublished
	case "pending", "รออนุมัติ":
		return StatusPending
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
