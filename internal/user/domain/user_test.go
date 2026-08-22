package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/user/domain"
)

func TestNewUserRequiresCompleteState(t *testing.T) {
	t.Parallel()

	_, err := domain.NewUser("", "person@example.com", "hash", time.Now())
	if !errors.Is(err, domain.ErrInvalidUser) {
		t.Fatalf("NewUser() error = %v, want ErrInvalidUser", err)
	}
}
