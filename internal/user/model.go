package user

import "time"

type User struct {
	FirebaseUID string    `json:"firebase_uid" gorm:"column:firebase_uid;primaryKey"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
