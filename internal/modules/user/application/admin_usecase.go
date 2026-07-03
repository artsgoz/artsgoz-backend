package application

import (
	"errors"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"
)

type AdminUsecase interface {
	GetAllUsers() ([]*domain.User, error)
	UpdateUserRole(uid string, role string) error
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
