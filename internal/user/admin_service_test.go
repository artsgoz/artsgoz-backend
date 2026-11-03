package user_test

import (
	"context"
	"testing"
	"time"

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

func (m *mockUserRepository) UpdateProfile(id string, email string) (*user.User, error) {
	if u, ok := m.users[id]; ok {
		u.Email = email
		u.UpdatedAt = time.Now()
		return u, nil
	}
	return nil, nil
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
