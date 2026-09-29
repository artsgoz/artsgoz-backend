package domain

import (
	"strings"
	"time"
)

type User struct {
	FirebaseUID string    `firestore:"firebase_uid" json:"firebase_uid"`
	Email       string    `firestore:"email" json:"email"`
	Role        string    `firestore:"role" json:"role"`
	CreatedAt   time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at" json:"updated_at"`
}

// IsChulaEmail ตรวจสอบว่าอีเมลเป็นของจุฬาลงกรณ์มหาวิทยาลัยหรือไม่ (@chula.ac.th หรือลงท้ายด้วย .chula.ac.th)
func IsChulaEmail(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	domain := parts[1]
	return domain == "chula.ac.th" || strings.HasSuffix(domain, ".chula.ac.th")
}
