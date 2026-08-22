package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/artsgoz/artsgoz-backend/internal/user/domain"
)

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("password must contain between 8 and 64 characters and fit the hashing limit")
	ErrEmailAlreadyExists = errors.New("email already registered")
)

type RegisterUserInput struct {
	Email    string
	Password string
}

type RegisterUserOutput struct {
	ID    string
	Email string
}

type RegisterUser struct {
	users  UserRepository
	hasher PasswordHasher
	ids    IDGenerator
	clock  Clock
}

func NewRegisterUser(users UserRepository, hasher PasswordHasher, ids IDGenerator, clock Clock) *RegisterUser {
	return &RegisterUser{users: users, hasher: hasher, ids: ids, clock: clock}
}

func (uc *RegisterUser) Execute(ctx context.Context, input RegisterUserInput) (RegisterUserOutput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return RegisterUserOutput{}, ErrInvalidEmail
	}
	passwordLength := utf8.RuneCountInString(input.Password)
	if passwordLength < 8 || passwordLength > 64 || len(input.Password) > 72 {
		return RegisterUserOutput{}, ErrInvalidPassword
	}

	passwordHash, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return RegisterUserOutput{}, fmt.Errorf("hash password: %w", err)
	}
	user, err := domain.NewUser(uc.ids.NewID(), email, passwordHash, uc.clock.Now())
	if err != nil {
		return RegisterUserOutput{}, fmt.Errorf("create user: %w", err)
	}
	if err := uc.users.Create(ctx, user); err != nil {
		return RegisterUserOutput{}, err
	}
	return RegisterUserOutput{ID: user.ID(), Email: user.Email()}, nil
}
