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
	firebaseAuth *auth.Client
}

func NewLoginService(repo UserRepository, firebaseAuth *auth.Client) LoginService {
	return &loginService{
		userRepo:     repo,
		firebaseAuth: firebaseAuth,
	}
}

func (u *loginService) Authenticate(req LoginRequest) (string, string, error) {
	ctx := context.Background()

	// 1. Verify Firebase ID Token
	token, err := u.firebaseAuth.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		return "", "", errors.New("token ไม่ถูกต้อง")
	}

	// 2. Retrieve user from PostgreSQL using UID (faster than querying by email)
	user, err := u.userRepo.GetByUID(token.UID)
	if err != nil {
		return "", "", errors.New("ไม่พบผู้ใช้ในระบบ")
	}

	// 3. Return uid + role
	return token.UID, user.Role, nil
}

func (u *loginService) GetUserProfile(uid string) (*User, error) {
	return u.userRepo.GetByUID(uid)
}

func (u *loginService) Logout(ctx context.Context, uid string) error {
	if u.firebaseAuth != nil && uid != "" {
		return u.firebaseAuth.RevokeRefreshTokens(ctx, uid)
	}
	return nil
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
		return errors.New("กรุณากรอก email และ password")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return errors.New("รูปแบบ email ไม่ถูกต้อง")
	}
	if len(req.Password) < 6 {
		return errors.New("password ต้องมีอย่างน้อย 6 ตัวอักษร")
	}

	// 1. Create user in Firebase Auth
	params := (&auth.UserToCreate{}).
		Email(req.Email).
		Password(req.Password)

	firebaseUser, err := u.firebaseAuth.CreateUser(ctx, params)
	if err != nil {
		if auth.IsEmailAlreadyExists(err) {
			return errors.New("อีเมล์นี้มีบัญชีในระบบแล้ว")
		}
		log.Printf("Firebase create user error: %v", err)
		return errors.New("สมัครสมาชิกไม่สำเร็จ")
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
		return errors.New("บันทึกข้อมูลไม่สำเร็จ")
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
		return errors.New("บทบาทไม่ถูกต้อง (ต้องเป็น student, teacher, club-member หรือ admin)")
	}
	return u.userRepo.UpdateRole(uid, role)
}

func (u *adminService) GetProfile(uid string) (*User, error) {
	if uid == "" {
		return nil, errors.New("กรุณาระบุ uid")
	}
	return u.userRepo.GetByUID(uid)
}

func (u *adminService) UpdateProfile(uid string, req UpdateProfileRequest) (*User, error) {
	if uid == "" {
		return nil, errors.New("กรุณาระบุ uid")
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		return nil, errors.New("กรุณากรอก email")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, errors.New("รูปแบบ email ไม่ถูกต้อง")
	}
	return u.userRepo.UpdateProfile(uid, req.Email)
}
