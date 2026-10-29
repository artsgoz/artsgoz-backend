package auth

import (
	"context"
	"errors"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/user"
)

// UserRepository contains the profile operations needed by authentication.
type UserRepository interface {
	FindByUID(ctx context.Context, uid string) (*user.User, error)
	CreateUser(ctx context.Context, profile *user.User) error
}

type Service struct {
	users    UserRepository
	firebase FirebaseClient
}

func NewService(users UserRepository, firebase FirebaseClient) *Service {
	return &Service{users: users, firebase: firebase}
}

func (s *Service) VerifyToken(ctx context.Context, idToken string) (Identity, error) {
	token, err := s.firebase.VerifyIDTokenAndCheckRevoked(ctx, idToken)
	if err != nil {
		return Identity{}, err
	}
	if token == nil || token.UID == "" {
		return Identity{}, errors.New("Firebase token has no user ID")
	}
	email, _ := token.Claims["email"].(string)
	return Identity{UID: token.UID, Email: email}, nil
}

// Login checks application registration after VerifyToken has authenticated the identity.
func (s *Service) Login(ctx context.Context, identity Identity) (*LoginResponse, error) {
	profile, err := s.users.FindByUID(ctx, identity.UID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return &LoginResponse{UID: identity.UID, OnboardingRequired: true}, nil
	}
	return &LoginResponse{UID: profile.FirebaseUID, Role: profile.Role}, nil
}

func (s *Service) Register(ctx context.Context, identity Identity) (*user.User, error) {
	if identity.UID == "" {
		return nil, errors.New("user ID is required")
	}
	now := time.Now()
	profile := &user.User{
		FirebaseUID: identity.UID,
		Email:       identity.Email,
		Role:        "student",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.users.CreateUser(ctx, profile); err != nil {
		return nil, err
	}

	// Return the stored profile so a retry preserves the user's existing role.
	profile, err := s.users.FindByUID(ctx, identity.UID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, errors.New("profile not found after registration")
	}
	return profile, nil
}

func (s *Service) GetProfile(ctx context.Context, uid string) (*user.User, error) {
	return s.users.FindByUID(ctx, uid)
}

func (s *Service) Logout(ctx context.Context, uid string) error {
	return s.firebase.RevokeRefreshTokens(ctx, uid)
}
