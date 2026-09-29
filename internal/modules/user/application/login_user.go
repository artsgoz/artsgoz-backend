package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"

	"firebase.google.com/go/v4/auth"
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
	if strings.TrimSpace(req.IDToken) == "" {
		return "", "", errors.New("กรุณาระบุ id_token")
	}

	if u.firebaseAuth == nil {
		return "", "", errors.New("firebase auth is not initialized")
	}

	ctx := context.Background()

	// 1. Verify Firebase ID Token
	token, err := u.firebaseAuth.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		return "", "", errors.New("token ไม่ถูกต้อง")
	}

	// 2. ตรวจสอบว่าต้อง login ผ่าน Google เท่านั้น (sign_in_provider ต้องเป็น google.com)
	if token.Firebase.SignInProvider != "google.com" {
		return "", "", errors.New("ให้เข้าสู่ระบบด้วยอีเมล์จุฬาเท่านั้น")
	}

	// 3. ตรวจสอบว่าต้องเป็นอีเมลของจุฬาฯ (@chula.ac.th หรือลงท้ายด้วย .chula.ac.th)
	email := ""
	if em, ok := token.Claims["email"].(string); ok {
		email = strings.TrimSpace(strings.ToLower(em))
	}
	if !domain.IsChulaEmail(email) {
		return "", "", errors.New("ให้เข้าสู่ระบบด้วยอีเมล์จุฬาเท่านั้น")
	}

	// 4. ดึง user จาก Firestore ด้วย UID (เร็วกว่า query by email)
	user, err := u.userRepo.GetByUID(token.UID)
	if err != nil {
		if err.Error() == "ไม่พบผู้ใช้" {
			// เมื่อผู้ใช้เข้าสู่ระบบด้วย Google เป็นครั้งแรก ให้สร้างข้อมูลผู้ใช้อัตโนมัติ (role เริ่มต้นเป็น student)
			// ตรวจสอบเผื่อมีบัญชีเดิมที่ผูกกับ email นี้อยู่แล้ว
			if email != "" {
				existingUser, emailErr := u.userRepo.GetByEmail(email)
				if emailErr == nil && existingUser != nil {
					return existingUser.FirebaseUID, existingUser.Role, nil
				}
			}

			newUser := &domain.User{
				FirebaseUID: token.UID,
				Email:       email,
				Role:        "student",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			if createErr := u.userRepo.CreateUser(newUser); createErr != nil {
				return "", "", errors.New("ไม่สามารถสร้างข้อมูลผู้ใช้ได้")
			}

			return token.UID, newUser.Role, nil
		}
		return "", "", err
	}

	// 5. return uid + role
	return token.UID, user.Role, nil
}

func (u *loginUsecase) GetUserProfile(uid string) (*domain.User, error) {
	if strings.TrimSpace(uid) == "" {
		return nil, errors.New("กรุณาระบุ uid")
	}
	return u.userRepo.GetByUID(uid)
}
