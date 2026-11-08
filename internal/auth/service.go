package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/artsgoz/artsgoz-backend/internal/user"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken        = errors.New("invalid or expired Firebase token")
	ErrFirebaseUnavailable = errors.New("Firebase authentication is not configured")
	ErrAccountConflict     = errors.New("account cannot be linked; verify your email and use the Firebase account already linked to this profile")
	ErrGoogleEmailDomain   = errors.New("Google sign-in requires a verified @student.chula.ac.th email")
	ErrInvalidUsername     = errors.New("username must contain between 3 and 50 characters")
)

type UserRepository interface {
	FindByID(context.Context, string) (*user.User, error)
	FindByUsername(context.Context, string) (*user.User, error)
	FindByEmail(context.Context, string) (*user.User, error)
	FindByFirebaseUID(context.Context, string) (*user.User, error)
	LinkFirebaseUID(context.Context, string, string) error
	CreateUser(context.Context, *user.User) error
}

type Service struct {
	users          UserRepository
	firebaseClient FirebaseClient
}

func NewService(users UserRepository, firebaseClient FirebaseClient) *Service {
	return &Service{users: users, firebaseClient: firebaseClient}
}

func (s *Service) verify(ctx context.Context, idToken string) (*firebaseauth.Token, error) {
	if s.firebaseClient == nil {
		return nil, ErrFirebaseUnavailable
	}
	if strings.TrimSpace(idToken) == "" {
		return nil, ErrInvalidToken
	}
	token, err := s.firebaseClient.VerifyIDToken(ctx, strings.TrimSpace(idToken))
	if err != nil || token == nil || token.UID == "" {
		return nil, ErrInvalidToken
	}
	switch token.Firebase.SignInProvider {
	case "password":
	case "google.com":
		email, _ := token.Claims["email"].(string)
		verified, _ := token.Claims["email_verified"].(bool)
		address, parseErr := mail.ParseAddress(email)
		if !verified || parseErr != nil || address.Address != email || !strings.HasSuffix(strings.ToLower(email), "@student.chula.ac.th") {
			return nil, ErrGoogleEmailDomain
		}
	default:
		return nil, ErrInvalidToken
	}
	return token, nil
}

// FirebaseLogin synchronizes a PostgreSQL profile after either Firebase sign-in provider.
// The client keeps using the Firebase ID token; no separate backend session is issued.
func (s *Service) FirebaseLogin(ctx context.Context, req FirebaseLoginRequest) (*LoginResponse, error) {
	token, err := s.verify(ctx, req.IDToken)
	if err != nil {
		return nil, err
	}
	profile, err := s.users.FindByFirebaseUID(ctx, token.UID)
	if err != nil {
		return nil, fmt.Errorf("find Firebase profile: %w", err)
	}
	if profile != nil {
		return &LoginResponse{Token: strings.TrimSpace(req.IDToken), User: profile}, nil
	}

	rawEmail, _ := token.Claims["email"].(string)
	email := strings.ToLower(strings.TrimSpace(rawEmail))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 255 {
		return nil, ErrInvalidToken
	}
	profile, err = s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("find legacy profile: %w", err)
	}
	if profile != nil {
		verified, _ := token.Claims["email_verified"].(bool)
		if !verified || (profile.FirebaseUID != "" && profile.FirebaseUID != token.UID) {
			return nil, ErrAccountConflict
		}
		if err := s.users.LinkFirebaseUID(ctx, profile.ID, token.UID); err != nil {
			if errors.Is(err, user.ErrFirebaseUIDAlreadyLinked) {
				return nil, ErrAccountConflict
			}
			return nil, fmt.Errorf("link Firebase profile: %w", err)
		}
		profile.FirebaseUID = token.UID
	} else {
		id := uuid.NewString()
		username := strings.TrimSpace(req.Username)
		if username != "" && (len([]rune(username)) < 3 || len([]rune(username)) > 50) {
			return nil, ErrInvalidUsername
		}
		if username == "" {
			// The UUID suffix avoids clashes between email local parts and concurrent sign-ins.
			local := strings.SplitN(email, "@", 2)[0]
			localRunes := []rune(local)
			if len(localRunes) > 50 {
				local = string(localRunes[:50])
			}
			username = local + "_" + id
		} else {
			existing, err := s.users.FindByUsername(ctx, username)
			if err != nil {
				return nil, fmt.Errorf("find username: %w", err)
			}
			if existing != nil {
				return nil, ErrAccountConflict
			}
		}
		now := time.Now().UTC()
		profile = &user.User{ID: id, FirebaseUID: token.UID, Username: username, Email: email,
			Role: "member", CreatedAt: now, UpdatedAt: now}
		if err := s.users.CreateUser(ctx, profile); err != nil {
			// Concurrent first sign-ins can both reach the insert. Reuse the winning UID.
			existing, lookupErr := s.users.FindByFirebaseUID(ctx, token.UID)
			if lookupErr != nil {
				return nil, fmt.Errorf("find concurrent Firebase profile: %w", lookupErr)
			}
			if existing == nil {
				if errors.Is(err, user.ErrEmailAlreadyExists) {
					return nil, ErrAccountConflict
				}
				return nil, fmt.Errorf("create Firebase profile: %w", err)
			}
			profile = existing
		}
	}
	return &LoginResponse{Token: strings.TrimSpace(req.IDToken), User: profile}, nil
}

func (s *Service) VerifyAndGetUserFromFirebase(ctx context.Context, idToken string) (*user.User, error) {
	token, err := s.verify(ctx, idToken)
	if err != nil {
		return nil, err
	}
	profile, err := s.users.FindByFirebaseUID(ctx, token.UID)
	if err != nil {
		return nil, fmt.Errorf("find Firebase profile: %w", err)
	}
	if profile == nil {
		return nil, ErrInvalidToken
	}
	return profile, nil
}

func (s *Service) GetProfile(ctx context.Context, id string) (*user.User, error) {
	return s.users.FindByID(ctx, id)
}

func (s *Service) Logout(ctx context.Context, firebaseUID string) error {
	if s.firebaseClient == nil {
		return ErrFirebaseUnavailable
	}
	if firebaseUID == "" {
		return ErrInvalidToken
	}
	return s.firebaseClient.RevokeRefreshTokens(ctx, firebaseUID)
}
