package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"

	"firebase.google.com/go/v4/auth"
)

type RegisterUsecase interface {
	Register(req RegisterRequest) error
}

type registerUsecase struct {
	userRepo     domain.UserRepository
	firebaseAuth *auth.Client
}

func NewRegisterUsecase(repo domain.UserRepository, firebaseAuth *auth.Client) RegisterUsecase {
	return &registerUsecase{
		userRepo:     repo,
		firebaseAuth: firebaseAuth,
	}
}

func (u *registerUsecase) Register(req RegisterRequest) error {
	if strings.TrimSpace(req.IDToken) == "" {
		return errors.New("กรุณาสมัครสมาชิกผ่าน Google เท่านั้น")
	}

	if u.firebaseAuth == nil {
		return errors.New("firebase auth is not initialized")
	}

	ctx := context.Background()

	// 1. ตรวจสอบ Firebase ID Token
	token, err := u.firebaseAuth.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		return errors.New("token ไม่ถูกต้อง")
	}

	// 2. ตรวจสอบว่าต้องเป็น Google เท่านั้น (sign_in_provider ต้องเป็น google.com)
	if token.Firebase.SignInProvider != "google.com" {
		return errors.New("ให้เข้าสู่ระบบด้วยอีเมล์จุฬาเท่านั้น")
	}

	// 3. ตรวจสอบว่าต้องเป็นอีเมลของจุฬาฯ (@chula.ac.th หรือลงท้ายด้วย .chula.ac.th)
	email := ""
	if em, ok := token.Claims["email"].(string); ok {
		email = strings.TrimSpace(strings.ToLower(em))
	}
	if !domain.IsChulaEmail(email) {
		return errors.New("ให้เข้าสู่ระบบด้วยอีเมล์จุฬาเท่านั้น")
	}

	// 4. ตรวจสอบว่ามีผู้ใช้อยู่แล้วหรือไม่
	existingUser, err := u.userRepo.GetByUID(token.UID)
	if err == nil && existingUser != nil {
		return errors.New("บัญชีนี้มีอยู่ในระบบแล้ว")
	}

	if email != "" {
		existingByEmail, emailErr := u.userRepo.GetByEmail(email)
		if emailErr == nil && existingByEmail != nil {
			return errors.New("บัญชีนี้มีอยู่ในระบบแล้ว")
		}
	}

	// 5. บันทึกลง Firestore (บทบาทเริ่มต้นคือ student เสมอ)
	user := &domain.User{
		FirebaseUID: token.UID,
		Email:       email,
		Role:        "student",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := u.userRepo.CreateUser(user); err != nil {
		return errors.New("บันทึกข้อมูลไม่สำเร็จ")
	}

	return nil
}
