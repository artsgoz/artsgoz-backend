package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/artsgoz/artsgoz-backend/internal/user"
)

// UserRepository contains the operations needed by authentication.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*user.User, error)
	FindByUsername(ctx context.Context, username string) (*user.User, error)
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindByFirebaseUID(ctx context.Context, uid string) (*user.User, error)
	LinkFirebaseUID(ctx context.Context, id string, firebaseUID string) error
	CreateUser(ctx context.Context, profile *user.User) error
}

type Service struct {
	users          UserRepository
	jwtManager     *JWTManager
	firebaseClient FirebaseClient
}

func NewService(users UserRepository, jwtManager *JWTManager, fb ...FirebaseClient) *Service {
	s := &Service{users: users, jwtManager: jwtManager}
	if len(fb) > 0 && fb[0] != nil {
		s.firebaseClient = fb[0]
	}
	return s
}

func (s *Service) HasFirebase() bool {
	return s.firebaseClient != nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*user.User, string, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return nil, "", errors.New("username is required")
	}

	// 1. Check duplicate username
	existingUser, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		return nil, "", err
	}
	if existingUser != nil {
		return nil, "", errors.New("username is already taken")
	}

	// 2. Check duplicate email (if provided)
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email != "" {
		existingEmail, err := s.users.FindByEmail(ctx, email)
		if err != nil {
			return nil, "", err
		}
		if existingEmail != nil {
			return nil, "", errors.New("email is already registered")
		}
	}

	// 3. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", errors.New("failed to encrypt password")
	}

	now := time.Now()
	newUser := &user.User{
		ID:           uuid.NewString(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         "member",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.CreateUser(ctx, newUser); err != nil {
		return nil, "", err
	}

	// 4. Generate JWT token
	token, err := s.jwtManager.GenerateToken(newUser.ID, newUser.Username, newUser.Role)
	if err != nil {
		return nil, "", errors.New("failed to generate access token")
	}

	return newUser, token, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		return nil, errors.New("username and password are required")
	}

	profile, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, errors.New("invalid username or password")
	}

	// Compare password hash
	if err := bcrypt.CompareHashAndPassword([]byte(profile.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	token, err := s.jwtManager.GenerateToken(profile.ID, profile.Username, profile.Role)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	return &LoginResponse{
		Token: token,
		User:  profile,
	}, nil
}

// FirebaseLogin authenticates or provisions a user using a verified Firebase ID token.
func (s *Service) FirebaseLogin(ctx context.Context, req FirebaseLoginRequest) (*LoginResponse, error) {
	if s.firebaseClient == nil {
		return nil, errors.New("firebase authentication is not configured on this server")
	}

	idToken := strings.TrimSpace(req.IDToken)
	if idToken == "" {
		return nil, errors.New("id_token is required")
	}

	decodedToken, err := s.firebaseClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, errors.New("invalid or expired firebase token: " + err.Error())
	}

	firebaseUID := decodedToken.UID
	email := ""
	if emailVal, ok := decodedToken.Claims["email"].(string); ok {
		email = strings.TrimSpace(strings.ToLower(emailVal))
	}

	// 1. Try finding user by FirebaseUID
	profile, err := s.users.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	// 2. If not found by FirebaseUID, try finding by email
	if profile == nil && email != "" {
		profile, err = s.users.FindByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if profile != nil {
			// Link the existing user account to this Firebase UID
			_ = s.users.LinkFirebaseUID(ctx, profile.ID, firebaseUID)
			profile.FirebaseUID = firebaseUID
		}
	}

	// 3. If still not found, auto-create user (first-time login / register via Firebase)
	if profile == nil {
		username := ""
		if req.Username != "" {
			username = strings.TrimSpace(req.Username)
		} else if email != "" {
			parts := strings.Split(email, "@")
			username = parts[0]
		} else {
			username = "user_" + firebaseUID[:8]
		}

		// Ensure unique username
		existing, err := s.users.FindByUsername(ctx, username)
		if err == nil && existing != nil {
			username = username + "_" + uuid.NewString()[:4]
		}

		now := time.Now()
		newUser := &user.User{
			ID:          uuid.NewString(),
			FirebaseUID: firebaseUID,
			Username:    username,
			Email:       email,
			Role:        "member",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := s.users.CreateUser(ctx, newUser); err != nil {
			return nil, errors.New("failed to provision user profile: " + err.Error())
		}
		profile = newUser
	}

	// Generate backend session JWT
	token, err := s.jwtManager.GenerateToken(profile.ID, profile.Username, profile.Role)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	return &LoginResponse{
		Token: token,
		User:  profile,
	}, nil
}

// VerifyAndGetUserFromFirebase verifies a Firebase ID token and resolves the associated User.
func (s *Service) VerifyAndGetUserFromFirebase(ctx context.Context, idToken string) (*user.User, error) {
	if s.firebaseClient == nil {
		return nil, errors.New("firebase authentication is not configured")
	}

	decodedToken, err := s.firebaseClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}

	firebaseUID := decodedToken.UID
	profile, err := s.users.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	if profile == nil {
		email := ""
		if emailVal, ok := decodedToken.Claims["email"].(string); ok {
			email = strings.TrimSpace(strings.ToLower(emailVal))
		}
		if email != "" {
			profile, err = s.users.FindByEmail(ctx, email)
			if err == nil && profile != nil {
				_ = s.users.LinkFirebaseUID(ctx, profile.ID, firebaseUID)
				profile.FirebaseUID = firebaseUID
			}
		}
	}

	return profile, nil
}

func (s *Service) GetProfile(ctx context.Context, id string) (*user.User, error) {
	return s.users.FindByID(ctx, id)
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	return s.jwtManager.ValidateToken(tokenString)
}
