package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrInvalidUser = errors.New("invalid user")

// User is the business representation of a registered user. Its fields are
// private so instances can only be created through NewUser.
type User struct {
	id           string
	email        string
	passwordHash string
	createdAt    time.Time
}

func NewUser(id, email, passwordHash string, createdAt time.Time) (User, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(email) == "" || passwordHash == "" || createdAt.IsZero() {
		return User{}, ErrInvalidUser
	}
	return User{id: id, email: email, passwordHash: passwordHash, createdAt: createdAt.UTC()}, nil
}

func (u User) ID() string           { return u.id }
func (u User) Email() string        { return u.email }
func (u User) PasswordHash() string { return u.passwordHash }
func (u User) CreatedAt() time.Time { return u.createdAt }
