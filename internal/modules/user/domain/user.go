package domain

import "time"

type User struct {
	FirebaseUID string    `firestore:"firebase_uid" json:"firebase_uid"`
	Email       string    `firestore:"email" json:"email"`
	Role        string    `firestore:"role" json:"role"`
	CreatedAt   time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at" json:"updated_at"`
}
