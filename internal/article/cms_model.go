package article

import "time"

// CMSRole represents a role in the CMS permission system.
type CMSRole struct {
	ID          int64  `gorm:"type:bigserial;primaryKey" json:"id"`
	RoleName    string `gorm:"type:varchar(50);not null;uniqueIndex" json:"role_name"` // super_admin, editor, writer
	Description string `gorm:"type:text" json:"description,omitempty"`
}

func (CMSRole) TableName() string {
	return "cms_roles"
}

// CMSUserRole represents user assignment to CMS roles.
type CMSUserRole struct {
	UserID     string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	RoleID     int64     `gorm:"type:bigint;primaryKey" json:"role_id"`
	AssignedAt time.Time `gorm:"type:timestamptz;not null;default:now()" json:"assigned_at"`
}

func (CMSUserRole) TableName() string {
	return "cms_user_roles"
}

// Tag represents a taxonomy tag for articles.
type Tag struct {
	ID        int64     `gorm:"type:bigserial;primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(80);not null" json:"name"`
	Slug      string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"slug"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (Tag) TableName() string {
	return "tags"
}

// ArticleCategory represents the many-to-many relationship between articles and categories.
type ArticleCategory struct {
	ArticleID  string `gorm:"type:uuid;primaryKey" json:"article_id"`
	CategoryID int64  `gorm:"type:bigint;primaryKey" json:"category_id"`
	IsPrimary  bool   `gorm:"not null;default:false" json:"is_primary"`
}

func (ArticleCategory) TableName() string {
	return "article_categories"
}

// ArticleTag represents the many-to-many relationship between articles and tags.
type ArticleTag struct {
	ArticleID string `gorm:"type:uuid;primaryKey" json:"article_id"`
	TagID     int64  `gorm:"type:bigint;primaryKey" json:"tag_id"`
}

func (ArticleTag) TableName() string {
	return "article_tags"
}

