package application

import (
	"context"
	"errors"

	"firebase.google.com/go/v4/auth"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"
)

type LoginUsecase interface {
	Authenticate(req LoginRequest) (string, string, error)
	GetUserProfile(uid string) (*domain.User, error)
}

type loginUsecase struct {
	userRepo     domain.UserRepository
	firebaseAuth *auth.Client
}

func NewLoginUsecase(repo domain.UserRepository, firebaseAuth *auth.Client) LoginUsecase {
	return &loginUsecase{
		userRepo:     repo,
		firebaseAuth: firebaseAuth,
	}
}

func (u *loginUsecase) Authenticate(req LoginRequest) (string, string, error) {
	ctx := context.Background()

	// 1. Verify Firebase ID Token
	token, err := u.firebaseAuth.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		return "", "", errors.New("token ไม่ถูกต้อง")
	}

	// 2. ดึง user จาก Firestore ด้วย UID (เร็วกว่า query by email)
	user, err := u.userRepo.GetByUID(token.UID)
	if err != nil {
		return "", "", errors.New("ไม่พบผู้ใช้ในระบบ")
	}

	// 3. return uid + role
	return token.UID, user.Role, nil
}

func (u *loginUsecase) GetUserProfile(uid string) (*domain.User, error) {
	return u.userRepo.GetByUID(uid)
}
