package application

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"
)

type AdminUsecase interface {
	GetAllUsers() ([]*domain.User, error)
	UpdateUserRole(uid string, role string) error
	GetProfile(uid string) (*domain.User, error)
	UpdateProfile(uid string, req UpdateProfileRequest) (*domain.User, error)
}

type adminUsecase struct {
	userRepo domain.UserRepository
}

func NewAdminUsecase(userRepo domain.UserRepository) AdminUsecase {
	return &adminUsecase{
		userRepo: userRepo,
	}
}

func (u *adminUsecase) GetAllUsers() ([]*domain.User, error) {
	return u.userRepo.GetAllUsers()
}

func (u *adminUsecase) UpdateUserRole(uid string, role string) error {
	if role != "student" && role != "teacher" && role != "club-member" && role != "admin" {
		return errors.New("บทบาทไม่ถูกต้อง (ต้องเป็น student, teacher, club-member หรือ admin)")
	}
	return u.userRepo.UpdateRole(uid, role)
}

func (u *adminUsecase) GetProfile(uid string) (*domain.User, error) {
	if uid == "" {
		return nil, errors.New("กรุณาระบุ uid")
	}
	return u.userRepo.GetByUID(uid)
}

func (u *adminUsecase) UpdateProfile(uid string, req UpdateProfileRequest) (*domain.User, error) {
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
