package application

import (
	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"
)

type AdminUsecase interface {
	GetAllUsers() ([]*domain.User, error)
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
