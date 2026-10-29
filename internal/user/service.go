package user

import (
	"context"
	"errors"
	"log"
	"net/mail"
	"strings"
	"time"

	"firebase.google.com/go/v4/auth"
)

type LoginService interface {
	Authenticate(req LoginRequest) (string, string, error)
	GetUserProfile(uid string) (*User, error)
	Logout(ctx context.Context, uid string) error
}

type loginService struct {
	userRepo     UserRepository
	firebaseAuth FirebaseAuthClient
}

func NewLoginService(repo UserRepository, firebaseAuth FirebaseAuthClient) LoginService {
	return &loginService{
		userRepo:     repo,
		firebaseAuth: firebaseAuth,
	}
}

func (u *loginService) Authenticate(req LoginRequest) (string, string, error) {
	ctx := context.Background()

	// 1. Verify Firebase ID token, including revocation and account disablement.
	token, err := u.firebaseAuth.VerifyIDTokenAndCheckRevoked(ctx, req.IDToken)
	if err != nil {
		return "", "", errors.New("invalid, revoked, or disabled token")
	}

	// 2. Retrieve user from PostgreSQL using UID (faster than querying by email)
	user, err := u.userRepo.GetByUID(token.UID)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	// 3. Return uid + role
	return token.UID, user.Role, nil
}

func (u *loginService) GetUserProfile(uid string) (*User, error) {
	return u.userRepo.GetByUID(uid)
}

func (u *loginService) Logout(ctx context.Context, uid string) error {
	if uid == "" {
		return errors.New("user ID is required")
	}
	if u.firebaseAuth == nil {
		return errors.New("Firebase Auth is not configured")
	}
	// Firebase revocation signs the user out across all sessions.
	return u.firebaseAuth.RevokeRefreshTokens(ctx, uid)
}

type RegisterService interface {
	Register(req RegisterRequest) error
}

type registerService struct {
	userRepo     UserRepository
	firebaseAuth *auth.Client
}

func NewRegisterService(repo UserRepository, firebaseAuth *auth.Client) RegisterService {
	return &registerService{
		userRepo:     repo,
		firebaseAuth: firebaseAuth,
	}
}

func (u *registerService) Register(req RegisterRequest) error {
	ctx := context.Background()

	// 0. Input Validation
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return errors.New("email and password are required")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return errors.New("invalid email format")
	}
	if len(req.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	// 1. Create user in Firebase Auth
	params := (&auth.UserToCreate{}).
		Email(req.Email).
		Password(req.Password)

	firebaseUser, err := u.firebaseAuth.CreateUser(ctx, params)
	if err != nil {
		if auth.IsEmailAlreadyExists(err) {
			return errors.New("an account with this email already exists")
		}
		log.Printf("Firebase create user error: %v", err)
		return errors.New("registration failed")
	}

	// 2. Save to PostgreSQL (force role to "student" for public registration)
	user := &User{
		FirebaseUID: firebaseUser.UID,
		Email:       req.Email,
		Role:        "student",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := u.userRepo.CreateUser(user); err != nil {
		// Rollback: Delete user from Firebase Auth to prevent orphan user
		if delErr := u.firebaseAuth.DeleteUser(ctx, firebaseUser.UID); delErr != nil {
			log.Printf("Rollback failed - could not delete Firebase user %s: %v", firebaseUser.UID, delErr)
		}
		log.Printf("PostgreSQL create user error: %v", err)
		return errors.New("failed to save user data")
	}

	return nil
}

type AdminService interface {
	GetAllUsers() ([]*User, error)
	UpdateUserRole(uid string, role string) error
	GetProfile(uid string) (*User, error)
	UpdateProfile(uid string, req UpdateProfileRequest) (*User, error)
}

type adminService struct {
	userRepo UserRepository
}

func NewAdminService(userRepo UserRepository) AdminService {
	return &adminService{
		userRepo: userRepo,
	}
}

func (u *adminService) GetAllUsers() ([]*User, error) {
	return u.userRepo.GetAllUsers()
}

func (u *adminService) UpdateUserRole(uid string, role string) error {
	if role != "student" && role != "teacher" && role != "club-member" && role != "admin" {
		return errors.New("invalid role (must be student, teacher, club-member, or admin)")
	}
	return u.userRepo.UpdateRole(uid, role)
}

func (u *adminService) GetProfile(uid string) (*User, error) {
	if uid == "" {
		return nil, errors.New("user ID is required")
	}
	return u.userRepo.GetByUID(uid)
}

func (u *adminService) UpdateProfile(uid string, req UpdateProfileRequest) (*User, error) {
	if uid == "" {
		return nil, errors.New("user ID is required")
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		return nil, errors.New("email is required")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, errors.New("invalid email format")
	}
	return u.userRepo.UpdateProfile(uid, req.Email)
}
