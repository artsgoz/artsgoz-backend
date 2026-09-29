package rest_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/application"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user/interface/rest"

	"github.com/gofiber/fiber/v3"
)

// MockLoginUsecase mocks application.LoginUsecase
type MockLoginUsecase struct {
	AuthenticateFunc   func(req application.LoginRequest) (string, string, error)
	GetUserProfileFunc func(uid string) (*domain.User, error)
}

func (m *MockLoginUsecase) Authenticate(req application.LoginRequest) (string, string, error) {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return "", "", nil
}

func (m *MockLoginUsecase) GetUserProfile(uid string) (*domain.User, error) {
	if m.GetUserProfileFunc != nil {
		return m.GetUserProfileFunc(uid)
	}
	return nil, nil
}

// MockRegisterUsecase mocks application.RegisterUsecase
type MockRegisterUsecase struct {
	RegisterFunc func(req application.RegisterRequest) error
}

func (m *MockRegisterUsecase) Register(req application.RegisterRequest) error {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(req)
	}
	return nil
}

// MockAdminUsecase mocks application.AdminUsecase
type MockAdminUsecase struct {
	GetAllUsersFunc    func() ([]*domain.User, error)
	UpdateUserRoleFunc func(uid string, role string) error
}

func (m *MockAdminUsecase) GetAllUsers() ([]*domain.User, error) {
	if m.GetAllUsersFunc != nil {
		return m.GetAllUsersFunc()
	}
	return nil, nil
}

func (m *MockAdminUsecase) UpdateUserRole(uid string, role string) error {
	if m.UpdateUserRoleFunc != nil {
		return m.UpdateUserRoleFunc(uid, role)
	}
	return nil
}

func TestAuthHandler_Login_EmptyToken(t *testing.T) {
	app := fiber.New()
	loginUsecase := &MockLoginUsecase{}
	registerUsecase := &MockRegisterUsecase{}
	handler := rest.NewAuthHandler(loginUsecase, registerUsecase)

	app.Post("/login", handler.Login)

	body, _ := json.Marshal(map[string]string{"id_token": "   "})
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty id_token, got %d", resp.StatusCode)
	}
}

func TestAuthHandler_Me_NotFound(t *testing.T) {
	app := fiber.New()
	loginUsecase := &MockLoginUsecase{
		GetUserProfileFunc: func(uid string) (*domain.User, error) {
			return nil, errors.New("ไม่พบผู้ใช้ในระบบ")
		},
	}
	registerUsecase := &MockRegisterUsecase{}
	handler := rest.NewAuthHandler(loginUsecase, registerUsecase)

	app.Get("/me", func(c fiber.Ctx) error {
		c.Locals("uid", "user-not-found")
		return handler.Me(c)
	})

	req := httptest.NewRequest("GET", "/me", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404 for not found user, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_UpdateUserRole_InvalidRole_Returns400(t *testing.T) {
	app := fiber.New()
	adminUsecase := &MockAdminUsecase{
		UpdateUserRoleFunc: func(uid string, role string) error {
			if role != "student" && role != "teacher" && role != "club-member" && role != "admin" {
				return errors.New("บทบาทผู้ใช้งานไม่ถูกต้อง")
			}
			return nil
		},
	}
	handler := rest.NewAdminHandler(adminUsecase)

	app.Put("/admin/users/:uid/role", handler.UpdateUserRole)

	body, _ := json.Marshal(map[string]string{"role": "invalid-role"})
	req := httptest.NewRequest("PUT", "/admin/users/123/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		respBody, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status 400 for invalid role, got %d. Body: %s", resp.StatusCode, string(respBody))
	}
}

func TestAdminHandler_UpdateUserRole_EmptyRole_Returns400(t *testing.T) {
	app := fiber.New()
	adminUsecase := &MockAdminUsecase{}
	handler := rest.NewAdminHandler(adminUsecase)

	app.Put("/admin/users/:uid/role", handler.UpdateUserRole)

	body, _ := json.Marshal(map[string]string{"role": "  "})
	req := httptest.NewRequest("PUT", "/admin/users/123/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty role, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_UpdateUserRole_DatabaseError_Returns500(t *testing.T) {
	app := fiber.New()
	adminUsecase := &MockAdminUsecase{
		UpdateUserRoleFunc: func(uid string, role string) error {
			return errors.New("database connection failed")
		},
	}
	handler := rest.NewAdminHandler(adminUsecase)

	app.Put("/admin/users/:uid/role", handler.UpdateUserRole)

	body, _ := json.Marshal(map[string]string{"role": "admin"})
	req := httptest.NewRequest("PUT", "/admin/users/123/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500 for db failure, got %d", resp.StatusCode)
	}
}
