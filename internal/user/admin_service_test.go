package user_test

import (
	"context"
	"testing"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/user"
	"golang.org/x/crypto/bcrypt"
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

func (m *mockUserRepository) UpdateProfile(id string, username string, email string, avatarURL string) (*user.User, error) {
	if u, ok := m.users[id]; ok {
		if username != "" {
			u.Username = username
		}
		if email != "" {
			u.Email = email
		}
		if avatarURL != "" {
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

func (m *mockUserRepository) FindByFirebaseUID(ctx context.Context, firebaseUID string) (*user.User, error) {
	for _, u := range m.users {
		if u.FirebaseUID == firebaseUID {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepository) LinkFirebaseUID(ctx context.Context, id string, firebaseUID string) error {
	if u, ok := m.users[id]; ok {
		u.FirebaseUID = firebaseUID
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
		Role:     "member",
	}

	adminSvc := user.NewAdminService(repo)

	// Valid role update
	err := adminSvc.UpdateUserRole("u1", "editor")
	if err != nil {
		t.Errorf("expected role update to succeed, got error: %v", err)
	}
	if repo.users["u1"].Role != "editor" {
		t.Errorf("expected user role to be editor, got %s", repo.users["u1"].Role)
	}

	// Invalid empty role
	err = adminSvc.UpdateUserRole("u1", "")
	if err == nil {
		t.Fatal("expected error for empty role, got nil")
	}
}

func TestAdminService_UpdateProfile(t *testing.T) {
	repo := newMockUserRepository()
	repo.users["u1"] = &user.User{
		ID:        "u1",
		Username:  "oldname",
		Email:     "old@example.com",
		AvatarURL: "http://example.com/old.png",
	}

	adminSvc := user.NewAdminService(repo)

	// 1. Valid update username and avatar
	updated, err := adminSvc.UpdateProfile("u1", user.UpdateProfileRequest{
		Username:  "newname",
		AvatarURL: "http://example.com/new.png",
	})
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}
	if updated.Username != "newname" {
		t.Errorf("expected username 'newname', got %s", updated.Username)
	}
	if updated.AvatarURL != "http://example.com/new.png" {
		t.Errorf("expected avatar 'http://example.com/new.png', got %s", updated.AvatarURL)
	}

	// 2. Empty request error
	_, err = adminSvc.UpdateProfile("u1", user.UpdateProfileRequest{})
	if err == nil {
		t.Fatal("expected error when all fields are empty, got nil")
	}

	// 3. Invalid email error
	_, err = adminSvc.UpdateProfile("u1", user.UpdateProfileRequest{
		Email: "not-an-email",
	})
	if err == nil {
		t.Fatal("expected error for invalid email format, got nil")
	}
}

func TestAdminService_ChangePassword(t *testing.T) {
	repo := newMockUserRepository()
	adminSvc := user.NewAdminService(repo)

	// Setup user with hashed password
	hashed, _ := bcrypt.GenerateFromPassword([]byte("oldpassword123"), bcrypt.DefaultCost)
	repo.users["u1"] = &user.User{
		ID:           "u1",
		Username:     "member1",
		PasswordHash: string(hashed),
	}

	// 1. Success change password
	err := adminSvc.ChangePassword("u1", user.ChangePasswordRequest{
		CurrentPassword: "oldpassword123",
		NewPassword:     "newpassword456",
	})
	if err != nil {
		t.Fatalf("unexpected change password error: %v", err)
	}
	// Verify new password works with bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(repo.users["u1"].PasswordHash), []byte("newpassword456")); err != nil {
		t.Errorf("failed to verify new password: %v", err)
	}

	// 2. Incorrect current password
	err = adminSvc.ChangePassword("u1", user.ChangePasswordRequest{
		CurrentPassword: "wrongpassword",
		NewPassword:     "anothernew123",
	})
	if err == nil {
		t.Fatal("expected error for wrong current password, got nil")
	}

	// 3. New password too short
	err = adminSvc.ChangePassword("u1", user.ChangePasswordRequest{
		CurrentPassword: "newpassword456",
		NewPassword:     "123",
	})
	if err == nil {
		t.Fatal("expected error for short new password, got nil")
	}
}
