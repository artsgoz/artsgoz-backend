package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/application"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"
)

// MockUserRepository implements domain.UserRepository for unit testing
type MockUserRepository struct {
	GetByEmailFunc  func(email string) (*domain.User, error)
	GetByUIDFunc    func(uid string) (*domain.User, error)
	CreateUserFunc  func(user *domain.User) error
	GetAllUsersFunc func() ([]*domain.User, error)
	UpdateRoleFunc  func(uid string, role string) error
	UpdateProfileFunc func(uid string, email string) (*domain.User, error)
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

func (m *MockUserRepository) UpdateProfile(uid string, email string) (*domain.User, error) {
	if m.UpdateProfileFunc != nil {
		return m.UpdateProfileFunc(uid, email)
	}
	return nil, nil
}

// ---------------- Admin Usecase Tests ----------------

func TestAdminUsecase_UpdateUserRole_InvalidRole(t *testing.T) {
	repo := &MockUserRepository{}
	usecase := application.NewAdminUsecase(repo)

	invalidRoles := []string{"invalid", "superadmin", "", "123", "STUDENT"}
	for _, role := range invalidRoles {
		err := usecase.UpdateUserRole("user-123", role)
		if err == nil {
			t.Errorf("expected error for invalid role %q, got nil", role)
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

func TestAdminUsecase_GetAllUsers(t *testing.T) {
	expectedUsers := []*domain.User{
		{FirebaseUID: "u1", Email: "test1@chula.ac.th", Role: "student"},
		{FirebaseUID: "u2", Email: "test2@chula.ac.th", Role: "admin"},
	}
	repo := &MockUserRepository{
		GetAllUsersFunc: func() ([]*domain.User, error) {
			return expectedUsers, nil
		},
	}
	usecase := application.NewAdminUsecase(repo)
	users, err := usecase.GetAllUsers()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != len(expectedUsers) {
		t.Fatalf("expected %d users, got %d", len(expectedUsers), len(users))
	}
}

// ---------------- Login Usecase Tests ----------------

func TestLoginUsecase_GetUserProfile_Success(t *testing.T) {
	expectedUser := &domain.User{
		FirebaseUID: "user-abc",
		Email:       "test@chula.ac.th",
		Role:        "student",
	}
	repo := &MockUserRepository{
		GetByUIDFunc: func(uid string) (*domain.User, error) {
			if uid == "user-abc" {
				return expectedUser, nil
			}
			return nil, errors.New("not found")
		},
	}
	usecase := application.NewLoginUsecase(repo, nil)

	user, err := usecase.GetUserProfile("user-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.FirebaseUID != expectedUser.FirebaseUID {
		t.Errorf("expected uid %q, got %q", expectedUser.FirebaseUID, user.FirebaseUID)
	}
}

func TestLoginUsecase_GetUserProfile_NotFound(t *testing.T) {
	repo := &MockUserRepository{
		GetByUIDFunc: func(uid string) (*domain.User, error) {
			return nil, errors.New("user not found")
		},
	}
	usecase := application.NewLoginUsecase(repo, nil)

	_, err := usecase.GetUserProfile("not-exist")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------- Register Usecase Validation Tests ----------------

func TestRegisterUsecase_ValidationErrors(t *testing.T) {
	repo := &MockUserRepository{}
	usecase := application.NewRegisterUsecase(repo, nil)

	// Empty email/password
	err := usecase.Register(application.RegisterRequest{
		Email:    "",
		Password: "",
	})
	if err == nil || err.Error() != "กรุณากรอก email และ password" {
		t.Fatalf("expected error 'กรุณากรอก email และ password', got %v", err)
	}

	// Invalid email format
	err = usecase.Register(application.RegisterRequest{
		Email:    "not-an-email",
		Password: "password123",
	})
	if err == nil || err.Error() != "รูปแบบ email ไม่ถูกต้อง" {
		t.Fatalf("expected error 'รูปแบบ email ไม่ถูกต้อง', got %v", err)
	}

	// Password too short (<6 characters)
	err = usecase.Register(application.RegisterRequest{
		Email:    "user@chula.ac.th",
		Password: "123",
	})
	if err == nil || err.Error() != "password ต้องมีอย่างน้อย 6 ตัวอักษร" {
		t.Fatalf("expected error 'password ต้องมีอย่างน้อย 6 ตัวอักษร', got %v", err)
	}
}

// ---------------- Admin Usecase Profile Tests ----------------

func TestAdminUsecase_GetProfile_Success(t *testing.T) {
	expectedUser := &domain.User{
		FirebaseUID: "admin-123",
		Email:       "admin@arts.chula.ac.th",
		Role:        "admin",
	}
	repo := &MockUserRepository{
		GetByUIDFunc: func(uid string) (*domain.User, error) {
			if uid == "admin-123" {
				return expectedUser, nil
			}
			return nil, errors.New("user not found")
		},
	}
	usecase := application.NewAdminUsecase(repo)

	user, err := usecase.GetProfile("admin-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.FirebaseUID != expectedUser.FirebaseUID || user.Role != expectedUser.Role {
		t.Errorf("expected user %+v, got %+v", expectedUser, user)
	}
}

func TestAdminUsecase_GetProfile_EmptyUID(t *testing.T) {
	repo := &MockUserRepository{}
	usecase := application.NewAdminUsecase(repo)

	_, err := usecase.GetProfile("")
	if err == nil || err.Error() != "กรุณาระบุ uid" {
		t.Fatalf("expected error 'กรุณาระบุ uid', got %v", err)
	}
}

func TestAdminUsecase_UpdateProfile_Success(t *testing.T) {
	expectedUser := &domain.User{
		FirebaseUID: "admin-123",
		Email:       "newemail@arts.chula.ac.th",
		Role:        "admin",
	}
	updated := false
	repo := &MockUserRepository{
		UpdateProfileFunc: func(uid string, email string) (*domain.User, error) {
			if uid == "admin-123" && email == "newemail@arts.chula.ac.th" {
				updated = true
				return expectedUser, nil
			}
			return nil, errors.New("unexpected params")
		},
	}
	usecase := application.NewAdminUsecase(repo)

	user, err := usecase.UpdateProfile("admin-123", application.UpdateProfileRequest{
		Email: "newemail@arts.chula.ac.th",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated {
		t.Error("expected UpdateProfile to be called on repo")
	}
	if user.Email != "newemail@arts.chula.ac.th" {
		t.Errorf("expected email newemail@arts.chula.ac.th, got %q", user.Email)
	}
}

func TestAdminUsecase_UpdateProfile_ValidationErrors(t *testing.T) {
	repo := &MockUserRepository{}
	usecase := application.NewAdminUsecase(repo)

	// Missing UID
	_, err := usecase.UpdateProfile("", application.UpdateProfileRequest{Email: "admin@arts.chula.ac.th"})
	if err == nil || err.Error() != "กรุณาระบุ uid" {
		t.Fatalf("expected error 'กรุณาระบุ uid', got %v", err)
	}

	// Empty Email
	_, err = usecase.UpdateProfile("admin-123", application.UpdateProfileRequest{Email: "   "})
	if err == nil || err.Error() != "กรุณากรอก email" {
		t.Fatalf("expected error 'กรุณากรอก email', got %v", err)
	}

	// Invalid Email format
	_, err = usecase.UpdateProfile("admin-123", application.UpdateProfileRequest{Email: "not-an-email"})
	if err == nil || err.Error() != "รูปแบบ email ไม่ถูกต้อง" {
		t.Fatalf("expected error 'รูปแบบ email ไม่ถูกต้อง', got %v", err)
	}
}

// ---------------- Login Usecase Logout Tests ----------------

func TestLoginUsecase_Logout_NilClient(t *testing.T) {
	repo := &MockUserRepository{}
	usecase := application.NewLoginUsecase(repo, nil)

	err := usecase.Logout(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("expected nil error for Logout with nil firebase auth client, got %v", err)
	}

	err = usecase.Logout(context.Background(), "")
	if err != nil {
		t.Fatalf("expected nil error for Logout with empty uid, got %v", err)
	}
}
