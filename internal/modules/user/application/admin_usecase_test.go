package application_test

import (
	"errors"
	"testing"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/application"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"
)

// MockUserRepository implements domain.UserRepository for testing
type MockUserRepository struct {
	GetByEmailFunc  func(email string) (*domain.User, error)
	GetByUIDFunc    func(uid string) (*domain.User, error)
	CreateUserFunc  func(user *domain.User) error
	GetAllUsersFunc func() ([]*domain.User, error)
	UpdateRoleFunc  func(uid string, role string) error
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(email)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByUID(uid string) (*domain.User, error) {
	if m.GetByUIDFunc != nil {
		return m.GetByUIDFunc(uid)
	}
	return nil, nil
}

func (m *MockUserRepository) CreateUser(user *domain.User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(user)
	}
	return nil
}

func (m *MockUserRepository) GetAllUsers() ([]*domain.User, error) {
	if m.GetAllUsersFunc != nil {
		return m.GetAllUsersFunc()
	}
	return nil, nil
}

func (m *MockUserRepository) UpdateRole(uid string, role string) error {
	if m.UpdateRoleFunc != nil {
		return m.UpdateRoleFunc(uid, role)
	}
	return nil
}

func TestAdminUsecase_UpdateUserRole_InvalidRole(t *testing.T) {
	repo := &MockUserRepository{}
	usecase := application.NewAdminUsecase(repo)

	invalidRoles := []string{"invalid", "superadmin", "", "123", "STUDENT"}
	for _, role := range invalidRoles {
		err := usecase.UpdateUserRole("user-123", role)
		if err == nil {
			t.Errorf("expected error for invalid role %q, got nil", role)
		}
		expectedMsg := "บทบาทผู้ใช้งานไม่ถูกต้อง"
		if err != nil && err.Error() != expectedMsg {
			t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
		}
	}
}

func TestAdminUsecase_UpdateUserRole_ValidRoles(t *testing.T) {
	validRoles := []string{"student", "teacher", "club-member", "admin"}
	for _, role := range validRoles {
		updated := false
		repo := &MockUserRepository{
			UpdateRoleFunc: func(uid string, r string) error {
				if uid == "user-123" && r == role {
					updated = true
					return nil
				}
				return errors.New("unexpected params")
			},
		}
		usecase := application.NewAdminUsecase(repo)
		err := usecase.UpdateUserRole("user-123", role)
		if err != nil {
			t.Errorf("expected nil error for valid role %q, got %v", role, err)
		}
		if !updated {
			t.Errorf("expected UpdateRole to be called for role %q", role)
		}
	}
}

func TestAdminUsecase_UpdateUserRole_RepoError(t *testing.T) {
	repo := &MockUserRepository{
		UpdateRoleFunc: func(uid string, role string) error {
			return errors.New("database connection failure")
		},
	}
	usecase := application.NewAdminUsecase(repo)
	err := usecase.UpdateUserRole("user-123", "student")
	if err == nil {
		t.Error("expected error from repo failure, got nil")
	}
}

func TestAdminUsecase_GetAllUsers_RepoError(t *testing.T) {
	repo := &MockUserRepository{
		GetAllUsersFunc: func() ([]*domain.User, error) {
			return nil, errors.New("firestore unavailable")
		},
	}
	usecase := application.NewAdminUsecase(repo)
	users, err := usecase.GetAllUsers()
	if err == nil {
		t.Error("expected error, got nil")
	}
	if users != nil {
		t.Errorf("expected nil users, got %v", users)
	}
}
