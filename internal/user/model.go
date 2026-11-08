package user

import "time"

type User struct {
	ID           string    `json:"id" gorm:"column:id;primaryKey"`
	Username     string    `json:"username" gorm:"column:username;type:varchar(100);uniqueIndex;not null"`
	Email        string    `json:"email" gorm:"column:email;type:varchar(255);uniqueIndex"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;not null"`
	Role         string    `json:"role" gorm:"column:role;default:'student'"`
	AvatarURL    string    `json:"avatar_url" gorm:"column:avatar_url;type:varchar(512)"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`
}
