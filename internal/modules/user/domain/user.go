package domain

import "time"

type User struct {
	FirebaseUID string    `json:"firebase_uid" db:"firebase_uid"`
	Email       string    `json:"email" db:"email"`
	Role        string    `json:"role" db:"role"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
