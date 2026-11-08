package user

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type AdminService interface {
	GetAllUsers() ([]*User, error)
	UpdateUserRole(id string, role string) error
	GetProfile(id string) (*User, error)
	UpdateProfile(id string, req UpdateProfileRequest) (*User, error)
	ChangePassword(id string, req ChangePasswordRequest) error
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

func (u *adminService) UpdateUserRole(id string, role string) error {
	role = strings.TrimSpace(role)
	if role == "" {
		return errors.New("role is required")
	}
	return u.userRepo.UpdateRole(id, role)
}

func (u *adminService) GetProfile(id string) (*User, error) {
	if id == "" {
		return nil, errors.New("user ID is required")
	}
	profile, err := u.userRepo.FindByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, errors.New("user not found")
	}
	return profile, nil
}

func (u *adminService) UpdateProfile(id string, req UpdateProfileRequest) (*User, error) {
	if id == "" {
		return nil, errors.New("user ID is required")
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.AvatarURL = strings.TrimSpace(req.AvatarURL)

	if req.Username == "" && req.Email == "" && req.AvatarURL == "" {
		return nil, errors.New("at least one field (username, email, or avatar_url) must be provided")
	}

	if req.Email != "" {
		if _, err := mail.ParseAddress(req.Email); err != nil {
			return nil, errors.New("invalid email format")
		}
	}
	return u.userRepo.UpdateProfile(id, req.Username, req.Email, req.AvatarURL)
}

func (u *adminService) ChangePassword(id string, req ChangePasswordRequest) error {
	if id == "" {
		return errors.New("user ID is required")
	}
	if req.CurrentPassword == "" {
		return errors.New("current password is required")
	}
	if len(req.NewPassword) < 6 {
		return errors.New("new password must be at least 6 characters")
	}

	user, err := u.userRepo.FindByID(context.Background(), id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.New("incorrect current password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to encrypt password")
	}

	return u.userRepo.UpdatePassword(id, string(hashedPassword))
}
