package user

import (
	"context"
	"errors"
	"testing"
)

type repositoryStub struct {
	created *User
	err     error
}

func (r *repositoryStub) CreateUser(_ context.Context, item *User) error {
	if r.err != nil {
		return r.err
	}
	copy := *item
	r.created = &copy
	return nil
}

func TestRegister(t *testing.T) {
	t.Parallel()

	repository := &repositoryStub{}
	service := NewService(repository)
	item, err := service.Register(context.Background(), RegisterInput{
		Email: " Person@Example.com ", Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if item.ID == "" || item.Email != "person@example.com" || item.PasswordHash == "" || item.CreatedAt.IsZero() {
		t.Fatalf("Register() user = %#v", item)
	}
	if repository.created == nil || repository.created.ID != item.ID {
		t.Fatalf("created user = %#v", repository.created)
	}
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	service := NewService(&repositoryStub{})
	tests := []struct {
		name  string
		input RegisterInput
		want  error
	}{
		{name: "email", input: RegisterInput{Email: "invalid", Password: "correct-password"}, want: ErrInvalidEmail},
		{name: "password", input: RegisterInput{Email: "person@example.com", Password: "short"}, want: ErrInvalidPassword},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.Register(context.Background(), test.input)
			if !errors.Is(err, test.want) {
				t.Fatalf("Register() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestRegisterPropagatesConflict(t *testing.T) {
	t.Parallel()

	service := NewService(&repositoryStub{err: ErrEmailAlreadyExists})
	_, err := service.Register(context.Background(), RegisterInput{
		Email: "person@example.com", Password: "correct-password",
	})
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("Register() error = %v, want conflict", err)
	}
}
