package application

import (
	"context"
	"errors"
	"log"
	"net/mail"
	"strings"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"
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

	// 1. สร้าง user ใน Firebase Auth
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

	// 2. บันทึกลง Firestore
	user := &domain.User{
		FirebaseUID: firebaseUser.UID,
		Email:       req.Email,
		Role:        "student",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := u.userRepo.CreateUser(user); err != nil {
		// Rollback: ลบ user ออกจาก Firebase Auth เพื่อป้องกัน orphan user
		if delErr := u.firebaseAuth.DeleteUser(ctx, firebaseUser.UID); delErr != nil {
			log.Printf("Rollback failed - could not delete Firebase user %s: %v", firebaseUser.UID, delErr)
		}
		log.Printf("Firestore create user error: %v", err)
		return errors.New("บันทึกข้อมูลไม่สำเร็จ")
	}

	return nil
}
