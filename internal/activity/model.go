package activity

import (
	"fmt"
	"time"
)

type Type string

const (
	TypePublish Type = "publish"
	TypeComment Type = "comment"
)

// ActivityLog represents an event logged in the system.
type ActivityLog struct {
	ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
	Type         Type      `gorm:"type:varchar(50);index;not null" json:"type"` // "publish" or "comment"
	Action       string    `gorm:"type:varchar(500);not null" json:"action"`
	Detail       string    `gorm:"type:text" json:"detail"`
	ArticleID    string    `gorm:"type:varchar(255);index" json:"article_id,omitempty"`
	ArticleTitle string    `gorm:"type:varchar(500)" json:"article_title,omitempty"`
	ActorID      string    `gorm:"type:varchar(255);index" json:"actor_id,omitempty"`
	ActorName    string    `gorm:"type:varchar(255)" json:"actor_name,omitempty"`
	CreatedAt    time.Time `gorm:"not null;index" json:"created_at"`
}

func (ActivityLog) TableName() string {
	return "activity_logs"
}

// ActivityItem represents the formatted response for the frontend Timeline.
type ActivityItem struct {
	ID           string    `json:"id"`
	Type         Type      `json:"type"`
	Action       string    `json:"action"`
	Detail       string    `json:"detail"`
	TimeAgo      string    `json:"time_ago"`
	ArticleTitle string    `json:"article_title,omitempty"`
	ArticleID    string    `json:"article_id,omitempty"`
	ActorName    string    `json:"actor_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// FormatThaiTimeAgo converts a timestamp into a human-friendly Thai relative time string.
func FormatThaiTimeAgo(t time.Time, now time.Time) string {
	diff := now.Sub(t)
	if diff < 0 {
		return "เมื่อสักครู่"
	}

	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := hours / 24

	switch {
	case seconds < 60:
		return "เมื่อสักครู่"
	case minutes < 60:
		return fmt.Sprintf("%d นาทีที่แล้ว", minutes)
	case hours < 24:
		return fmt.Sprintf("%d ชั่วโมงที่แล้ว", hours)
	case days == 1:
		return "เมื่อวานนี้"
	case days < 7:
		return fmt.Sprintf("%d วันที่แล้ว", days)
	case days < 30:
		weeks := days / 7
		return fmt.Sprintf("%d สัปดาห์ที่แล้ว", weeks)
	default:
		// Thai Buddhist year formatting
		thMonths := []string{"", "ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."}
		month := t.Month()
		yearBE := (t.Year() + 543) % 100
		return fmt.Sprintf("%d %s %02d", t.Day(), thMonths[month], yearBE)
	}
}

func (l *ActivityLog) ToItem(now time.Time) ActivityItem {
	return ActivityItem{
		ID:           l.ID,
		Type:         l.Type,
		Action:       l.Action,
		Detail:       l.Detail,
		TimeAgo:      FormatThaiTimeAgo(l.CreatedAt, now),
		ArticleTitle: l.ArticleTitle,
		ArticleID:    l.ArticleID,
		ActorName:    l.ActorName,
		CreatedAt:    l.CreatedAt,
	}
}

