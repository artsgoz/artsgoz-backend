package usecase

import (
	"context"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/user/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
}

type IDGenerator interface {
	NewID() string
}

type Clock interface {
	Now() time.Time
}

type PasswordHasherFunc func(string) (string, error)

func (f PasswordHasherFunc) Hash(password string) (string, error) { return f(password) }

type IDGeneratorFunc func() string

func (f IDGeneratorFunc) NewID() string { return f() }

type ClockFunc func() time.Time

func (f ClockFunc) Now() time.Time { return f() }
