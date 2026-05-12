package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

// NewUser is the aggregate-root constructor: it enforces invariants
// (non-empty email, password length, hashing) before a User can exist.
func NewUser(email, plaintextPassword string) (*User, error) {
	if email == "" {
		return nil, apperr.Validation("email is required", nil)
	}
	if len(plaintextPassword) < 8 {
		return nil, apperr.Validation("password must be at least 8 characters", nil)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperr.Internal("hash password", err)
	}
	return &User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}, nil
}
