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
	if role != "student" && role != "teacher" && role != "club-member" && role != "admin" {
		return errors.New("invalid role (must be student, teacher, club-member, or admin)")
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

	currentUser, err := u.userRepo.FindByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, errors.New("user not found")
	}

	updates := make(map[string]any)

	// Validate & check email if provided
	req.Email = strings.TrimSpace(req.Email)
	if req.Email != "" && req.Email != currentUser.Email {
		if _, err := mail.ParseAddress(req.Email); err != nil {
			return nil, errors.New("invalid email format")
		}
		existing, err := u.userRepo.FindByEmail(context.Background(), req.Email)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, errors.New("email is already registered")
		}
		updates["email"] = req.Email
	}

	// Validate & check username if provided
	req.Username = strings.TrimSpace(req.Username)
	if req.Username != "" && req.Username != currentUser.Username {
		if len(req.Username) < 3 {
			return nil, errors.New("username must be at least 3 characters")
		}
		existing, err := u.userRepo.FindByUsername(context.Background(), req.Username)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, errors.New("username is already taken")
		}
		updates["username"] = req.Username
	}

	// AvatarURL if provided
	req.AvatarURL = strings.TrimSpace(req.AvatarURL)
	if req.AvatarURL != "" && req.AvatarURL != currentUser.AvatarURL {
		updates["avatar_url"] = req.AvatarURL
	}

	if len(updates) == 0 {
		return currentUser, nil
	}

	return u.userRepo.UpdateProfile(id, updates)
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

	currentUser, err := u.userRepo.FindByID(context.Background(), id)
	if err != nil {
		return err
	}
	if currentUser == nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentUser.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	return u.userRepo.UpdatePassword(id, string(hashedPassword))
}
