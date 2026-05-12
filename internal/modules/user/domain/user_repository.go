package domain

import "context"

// UserRepository is the persistence port for the user aggregate.
// Implementations live in the infrastructure layer.
type UserRepository interface {
	Save(ctx context.Context, u *User) error
	// FindByEmail returns the user or an *apperr.Error of kind not_found.
	FindByEmail(ctx context.Context, email string) (*User, error)
}
