package comment

import (
	"time"

	"gorm.io/gorm"
)

// Comment represents a user comment on an article.
type Comment struct {
	ID           string         `gorm:"type:uuid;primaryKey" json:"id"`
	ArticleID    string         `gorm:"type:varchar(255);index;not null" json:"article_id"`
	ArticleTitle string         `gorm:"type:varchar(500)" json:"article_title"`
	AuthorID     string         `gorm:"type:varchar(255);index" json:"author_id,omitempty"`
	AuthorName   string         `gorm:"type:varchar(255);not null" json:"author_name"`
	AuthorAvatar string         `gorm:"type:text" json:"author_avatar,omitempty"`
	Content      string         `gorm:"type:text;not null" json:"content"`
	CreatedAt    time.Time      `gorm:"not null;index" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Comment) TableName() string {
	return "comments"
}

// CommentItem represents the formatted response for the frontend CommentsPage.
type CommentItem struct {
	ID           string    `json:"id"`
	Author       string    `json:"author"`
	AuthorAvatar string    `json:"author_avatar,omitempty"`
	Content      string    `json:"content"`
	ArticleTitle string    `json:"article_title"`
	ArticleID    string    `json:"article_id"`
	TimeAgo      string    `json:"time_ago"`
	CreatedAt    time.Time `json:"created_at"`
}

