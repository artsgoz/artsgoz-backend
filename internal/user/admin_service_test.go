package user_test

import (
	"context"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/artsgoz/artsgoz-backend/internal/user"
)

type mockUserRepository struct {
	users map[string]*user.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*user.User),
	}
}

func (m *mockUserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	return m.users[id], nil
}

func (m *mockUserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepository) CreateUser(ctx context.Context, u *user.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepository) GetAllUsers() ([]*user.User, error) {
	var list []*user.User
	for _, u := range m.users {
		list = append(list, u)
	}
	return list, nil
}

func (m *mockUserRepository) UpdateRole(id string, role string) error {
	if u, ok := m.users[id]; ok {
		u.Role = role
		u.UpdatedAt = time.Now()
		return nil
	}
	return nil
}

func (m *mockUserRepository) UpdateProfile(id string, updates map[string]any) (*user.User, error) {
	if u, ok := m.users[id]; ok {
		if email, ok := updates["email"].(string); ok {
			u.Email = email
		}
		if username, ok := updates["username"].(string); ok {
			u.Username = username
		}
		if avatarURL, ok := updates["avatar_url"].(string); ok {
			u.AvatarURL = avatarURL
		}
		u.UpdatedAt = time.Now()
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepository) UpdatePassword(id string, passwordHash string) error {
	if u, ok := m.users[id]; ok {
		u.PasswordHash = passwordHash
		u.UpdatedAt = time.Now()
		return nil
	}
	return nil
}

func TestAdminService_UpdateUserRole(t *testing.T) {
	repo := newMockUserRepository()
	repo.users["u1"] = &user.User{
		ID:       "u1",
		Username: "member1",
		Role:     "student",
	}

	adminSvc := user.NewAdminService(repo)

	// Valid roles
	validRoles := []string{"student", "teacher", "club-member", "admin"}
	for _, r := range validRoles {
		err := adminSvc.UpdateUserRole("u1", r)
		if err != nil {
			t.Errorf("expected role %s to be valid, got error: %v", r, err)
		}
		if repo.users["u1"].Role != r {
			t.Errorf("expected user role to be %s, got %s", r, repo.users["u1"].Role)
		}
	}

	// Invalid role
	err := adminSvc.UpdateUserRole("u1", "superadmin")
	if err == nil {
		t.Fatal("expected error for invalid role 'superadmin', got nil")
	}
}

func TestAdminService_UpdateProfile(t *testing.T) {
	repo := newMockUserRepository()
	repo.users["u1"] = &user.User{
		ID:        "u1",
		Username:  "meow1",
		Email:     "meow1@chula.ac.th",
		AvatarURL: "/uploads/old.png",
	}
	repo.users["u2"] = &user.User{
		ID:       "u2",
		Username: "existing_user",
		Email:    "existing@chula.ac.th",
	}

	adminSvc := user.NewAdminService(repo)

	// 1. Success update username, email, avatar
	updated, err := adminSvc.UpdateProfile("u1", user.UpdateProfileRequest{
		Username:  "meow_new",
		Email:     "meow_new@chula.ac.th",
		AvatarURL: "/uploads/new.png",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Username != "meow_new" || updated.Email != "meow_new@chula.ac.th" || updated.AvatarURL != "/uploads/new.png" {
		t.Errorf("expected profile updated, got %+v", updated)
	}

	// 2. Duplicate username error
	_, err = adminSvc.UpdateProfile("u1", user.UpdateProfileRequest{
		Username: "existing_user",
	})
	if err == nil || err.Error() != "username is already taken" {
		t.Errorf("expected 'username is already taken', got %v", err)
	}

	// 3. Duplicate email error
	_, err = adminSvc.UpdateProfile("u1", user.UpdateProfileRequest{
		Email: "existing@chula.ac.th",
	})
	if err == nil || err.Error() != "email is already registered" {
		t.Errorf("expected 'email is already registered', got %v", err)
	}

	// 4. Invalid email format
	_, err = adminSvc.UpdateProfile("u1", user.UpdateProfileRequest{
		Email: "invalid-email",
	})
	if err == nil || err.Error() != "invalid email format" {
		t.Errorf("expected 'invalid email format', got %v", err)
	}

	// 5. Short username
	_, err = adminSvc.UpdateProfile("u1", user.UpdateProfileRequest{
		Username: "ab",
	})
	if err == nil || err.Error() != "username must be at least 3 characters" {
		t.Errorf("expected 'username must be at least 3 characters', got %v", err)
	}
}

func TestAdminService_ChangePassword(t *testing.T) {
	repo := newMockUserRepository()
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("oldpassword123"), bcrypt.DefaultCost)
	repo.users["u1"] = &user.User{
		ID:           "u1",
		Username:     "meow1",
		PasswordHash: string(oldHash),
	}

	adminSvc := user.NewAdminService(repo)

	// 1. Incorrect current password
	err := adminSvc.ChangePassword("u1", user.ChangePasswordRequest{
		CurrentPassword: "wrongpassword",
		NewPassword:     "newpassword123",
	})
	if err == nil || err.Error() != "current password is incorrect" {
		t.Errorf("expected 'current password is incorrect', got %v", err)
	}

	// 2. Short new password
	err = adminSvc.ChangePassword("u1", user.ChangePasswordRequest{
		CurrentPassword: "oldpassword123",
		NewPassword:     "12345",
	})
	if err == nil || err.Error() != "new password must be at least 6 characters" {
		t.Errorf("expected 'new password must be at least 6 characters', got %v", err)
	}

	// 3. Success change password
	err = adminSvc.ChangePassword("u1", user.ChangePasswordRequest{
		CurrentPassword: "oldpassword123",
		NewPassword:     "newpassword123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify new password matches stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(repo.users["u1"].PasswordHash), []byte("newpassword123")); err != nil {
		t.Errorf("stored hash does not match new password: %v", err)
	}
}
