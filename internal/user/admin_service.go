package user

import (
	"context"
	"errors"
	"net/mail"
	"strings"
)

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
	profile, err := u.userRepo.FindByUID(context.Background(), uid)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, errors.New("user not found")
	}
	return profile, nil
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
