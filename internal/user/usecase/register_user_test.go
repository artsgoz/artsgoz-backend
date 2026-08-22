package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/user/domain"
	"github.com/artsgoz/artsgoz-backend/internal/user/usecase"
)

type userRepositoryStub struct {
	created domain.User
	err     error
}

func (r *userRepositoryStub) Create(_ context.Context, user domain.User) error {
	r.created = user
	return r.err
}

func TestRegisterUser(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 22, 10, 0, 0, 0, time.UTC)
	repository := &userRepositoryStub{}
	register := usecase.NewRegisterUser(
		repository,
		usecase.PasswordHasherFunc(func(string) (string, error) { return "password-hash", nil }),
		usecase.IDGeneratorFunc(func() string { return "user-id" }),
		usecase.ClockFunc(func() time.Time { return now }),
	)

	output, err := register.Execute(context.Background(), usecase.RegisterUserInput{
		Email: "  Person@Example.com ", Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.ID != "user-id" || output.Email != "person@example.com" {
		t.Fatalf("Execute() output = %#v", output)
	}
	if repository.created.PasswordHash() != "password-hash" || !repository.created.CreatedAt().Equal(now) {
		t.Fatalf("created user = %#v", repository.created)
	}
}

func TestRegisterUserRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	register := usecase.NewRegisterUser(
		&userRepositoryStub{},
		usecase.PasswordHasherFunc(func(string) (string, error) { t.Fatal("hasher called"); return "", nil }),
		usecase.IDGeneratorFunc(func() string { t.Fatal("ID generator called"); return "" }),
		usecase.ClockFunc(func() time.Time { t.Fatal("clock called"); return time.Time{} }),
	)

	tests := []struct {
		name  string
		input usecase.RegisterUserInput
		want  error
	}{
		{name: "email", input: usecase.RegisterUserInput{Email: "not-an-email", Password: "correct-password"}, want: usecase.ErrInvalidEmail},
		{name: "password", input: usecase.RegisterUserInput{Email: "person@example.com", Password: "short"}, want: usecase.ErrInvalidPassword},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := register.Execute(context.Background(), test.input)
			if !errors.Is(err, test.want) {
				t.Fatalf("Execute() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestRegisterUserPropagatesConflict(t *testing.T) {
	t.Parallel()

	repository := &userRepositoryStub{err: usecase.ErrEmailAlreadyExists}
	register := usecase.NewRegisterUser(
		repository,
		usecase.PasswordHasherFunc(func(string) (string, error) { return "password-hash", nil }),
		usecase.IDGeneratorFunc(func() string { return "user-id" }),
		usecase.ClockFunc(time.Now),
	)

	_, err := register.Execute(context.Background(), usecase.RegisterUserInput{
		Email: "person@example.com", Password: "correct-password",
	})
	if !errors.Is(err, usecase.ErrEmailAlreadyExists) {
		t.Fatalf("Execute() error = %v, want conflict", err)
	}
}
