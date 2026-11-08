package user

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("password must contain between 8 and 64 characters and fit the hashing limit")
	ErrEmailAlreadyExists = errors.New("email already registered")
)

type Service struct {
	repository Repository
}

type RegisterInput struct {
	Email    string
	Password string
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (User, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return User{}, ErrInvalidEmail
	}
	passwordLength := utf8.RuneCountInString(input.Password)
	if passwordLength < 8 || passwordLength > 64 || len(input.Password) > 72 {
		return User{}, ErrInvalidPassword
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}
	item := User{
		ID: uuid.NewString(), Email: email, PasswordHash: string(passwordHash), CreatedAt: time.Now().UTC(),
	}
	if err := s.repository.CreateUser(ctx, &item); err != nil {
		return User{}, err
	}
	return item, nil
}
