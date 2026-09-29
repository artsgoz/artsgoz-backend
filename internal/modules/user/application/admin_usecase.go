package application

import (
	"errors"
	"strings"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"
)

type AdminUsecase interface {
	GetAllUsers() ([]*domain.User, error)
	UpdateUserRole(uid string, role string) error
}

type adminUsecase struct {
	userRepo domain.UserRepository
}

func NewAdminUsecase(repo domain.UserRepository) AdminUsecase {
	return &adminUsecase{
		userRepo: repo,
	}
}

func (u *adminUsecase) GetAllUsers() ([]*domain.User, error) {
	return u.userRepo.GetAllUsers()
}

func (u *adminUsecase) UpdateUserRole(uid string, role string) error {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return errors.New("กรุณาระบุ uid")
	}

	role = strings.TrimSpace(role)
	if role != "student" && role != "teacher" && role != "club-member" && role != "admin" {
		return errors.New("บทบาทผู้ใช้งานไม่ถูกต้อง")
	}

	return u.userRepo.UpdateRole(uid, role)
}
