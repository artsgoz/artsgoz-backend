package rest_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/application"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user/interface/rest"
)

func init() {
	gin.SetMode(gin.TestMode)
}

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

// ----------------- AuthHandler Tests -----------------

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	loginUsecase := &MockLoginUsecase{}
	registerUsecase := &MockRegisterUsecase{}
	handler := rest.NewAuthHandler(loginUsecase, registerUsecase)

	router := gin.New()
	router.POST("/login", handler.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("{invalid-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid json, got %d", w.Code)
	}
}

func TestAuthHandler_Login_AuthenticateError(t *testing.T) {
	loginUsecase := &MockLoginUsecase{
		AuthenticateFunc: func(req application.LoginRequest) (string, string, error) {
			return "", "", errors.New("token ไม่ถูกต้อง")
		},
	}
	registerUsecase := &MockRegisterUsecase{}
	handler := rest.NewAuthHandler(loginUsecase, registerUsecase)

	router := gin.New()
	router.POST("/login", handler.Login)

	body, _ := json.Marshal(application.LoginRequest{IDToken: "bad-token"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for auth failure, got %d", w.Code)
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	loginUsecase := &MockLoginUsecase{
		AuthenticateFunc: func(req application.LoginRequest) (string, string, error) {
			return "user-123", "student", nil
		},
	}
	registerUsecase := &MockRegisterUsecase{}
	handler := rest.NewAuthHandler(loginUsecase, registerUsecase)

	router := gin.New()
	router.POST("/login", handler.Login)

	body, _ := json.Marshal(application.LoginRequest{IDToken: "valid-token"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for login success, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["uid"] != "user-123" || resp["role"] != "student" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	loginUsecase := &MockLoginUsecase{}
	registerUsecase := &MockRegisterUsecase{}
	handler := rest.NewAuthHandler(loginUsecase, registerUsecase)

	router := gin.New()
	router.POST("/register", handler.Register)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("{invalid-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid json, got %d", w.Code)
	}
}

func TestAuthHandler_Register_UsecaseError(t *testing.T) {
	loginUsecase := &MockLoginUsecase{}
	registerUsecase := &MockRegisterUsecase{
		RegisterFunc: func(req application.RegisterRequest) error {
			return errors.New("อีเมล์นี้มีบัญชีในระบบแล้ว")
		},
	}
	handler := rest.NewAuthHandler(loginUsecase, registerUsecase)

	router := gin.New()
	router.POST("/register", handler.Register)

	body, _ := json.Marshal(application.RegisterRequest{
		Email:    "test@chula.ac.th",
		Password: "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for register error, got %d", w.Code)
	}
}

func TestAuthHandler_Register_Success(t *testing.T) {
	loginUsecase := &MockLoginUsecase{}
	registerUsecase := &MockRegisterUsecase{
		RegisterFunc: func(req application.RegisterRequest) error {
			return nil
		},
	}
	handler := rest.NewAuthHandler(loginUsecase, registerUsecase)

	router := gin.New()
	router.POST("/register", handler.Register)

	body, _ := json.Marshal(application.RegisterRequest{
		Email:    "newuser@chula.ac.th",
		Password: "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for register success, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["message"] != "สมัครสมาชิกสำเร็จ" {
		t.Errorf("expected message 'สมัครสมาชิกสำเร็จ', got %q", resp["message"])
	}
}

func TestAuthHandler_Me_MissingUID(t *testing.T) {
	loginUsecase := &MockLoginUsecase{}
	registerUsecase := &MockRegisterUsecase{}
	handler := rest.NewAuthHandler(loginUsecase, registerUsecase)

	router := gin.New()
	router.GET("/me", handler.Me)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 when uid is missing, got %d", w.Code)
	}
}

func TestAuthHandler_Me_EmptyUID(t *testing.T) {
	loginUsecase := &MockLoginUsecase{}
	registerUsecase := &MockRegisterUsecase{}
	handler := rest.NewAuthHandler(loginUsecase, registerUsecase)

	router := gin.New()
	router.GET("/me", func(c *gin.Context) {
		c.Set("uid", "")
		handler.Me(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 when uid is empty, got %d", w.Code)
	}
}

func TestAuthHandler_Me_NotFound(t *testing.T) {
	loginUsecase := &MockLoginUsecase{
		GetUserProfileFunc: func(uid string) (*domain.User, error) {
			return nil, errors.New("ไม่สามารถดึงข้อมูลผู้ใช้ได้")
		},
	}
	registerUsecase := &MockRegisterUsecase{}
	handler := rest.NewAuthHandler(loginUsecase, registerUsecase)

	router := gin.New()
	router.GET("/me", func(c *gin.Context) {
		c.Set("uid", "user-not-found")
		handler.Me(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 when user fetch fails, got %d", w.Code)
	}
}

func TestAuthHandler_Me_Success(t *testing.T) {
	loginUsecase := &MockLoginUsecase{
		GetUserProfileFunc: func(uid string) (*domain.User, error) {
			return &domain.User{
				FirebaseUID: "user-123",
				Role:        "student",
				Email:       "test@chula.ac.th",
			}, nil
		},
	}
	registerUsecase := &MockRegisterUsecase{}
	handler := rest.NewAuthHandler(loginUsecase, registerUsecase)

	router := gin.New()
	router.GET("/me", func(c *gin.Context) {
		c.Set("uid", "user-123")
		handler.Me(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for Me success, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["uid"] != "user-123" || resp["role"] != "student" {
		t.Errorf("unexpected Me response: %+v", resp)
	}
}

// ----------------- AdminHandler Tests -----------------

func TestAdminHandler_GetAllUsers_Error(t *testing.T) {
	adminUsecase := &MockAdminUsecase{
		GetAllUsersFunc: func() ([]*domain.User, error) {
			return nil, errors.New("db error")
		},
	}
	handler := rest.NewAdminHandler(adminUsecase)

	router := gin.New()
	router.GET("/admin/users", handler.GetAllUsers)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 on db error, got %d", w.Code)
	}
}

func TestAdminHandler_GetAllUsers_Success(t *testing.T) {
	mockUsers := []*domain.User{
		{FirebaseUID: "u1", Email: "u1@chula.ac.th", Role: "student"},
		{FirebaseUID: "u2", Email: "u2@chula.ac.th", Role: "admin"},
	}
	adminUsecase := &MockAdminUsecase{
		GetAllUsersFunc: func() ([]*domain.User, error) {
			return mockUsers, nil
		},
	}
	handler := rest.NewAdminHandler(adminUsecase)

	router := gin.New()
	router.GET("/admin/users", handler.GetAllUsers)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 on success, got %d", w.Code)
	}

	var users []*domain.User
	if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}

func TestAdminHandler_UpdateUserRole_InvalidJSON(t *testing.T) {
	adminUsecase := &MockAdminUsecase{}
	handler := rest.NewAdminHandler(adminUsecase)

	router := gin.New()
	router.PUT("/admin/users/:uid/role", handler.UpdateUserRole)

	req := httptest.NewRequest(http.MethodPut, "/admin/users/123/role", bytes.NewBufferString("{bad-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for bad json, got %d", w.Code)
	}
}

func TestAdminHandler_UpdateUserRole_InvalidRole(t *testing.T) {
	adminUsecase := &MockAdminUsecase{
		UpdateUserRoleFunc: func(uid string, role string) error {
			return errors.New("บทบาทไม่ถูกต้อง")
		},
	}
	handler := rest.NewAdminHandler(adminUsecase)

	router := gin.New()
	router.PUT("/admin/users/:uid/role", handler.UpdateUserRole)

	body, _ := json.Marshal(map[string]string{"role": "super-hacker"})
	req := httptest.NewRequest(http.MethodPut, "/admin/users/123/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid role error, got %d", w.Code)
	}
}

func TestAdminHandler_UpdateUserRole_Success(t *testing.T) {
	called := false
	adminUsecase := &MockAdminUsecase{
		UpdateUserRoleFunc: func(uid string, role string) error {
			if uid == "user-456" && role == "teacher" {
				called = true
				return nil
			}
			return errors.New("unexpected params")
		},
	}
	handler := rest.NewAdminHandler(adminUsecase)

	router := gin.New()
	router.PUT("/admin/users/:uid/role", handler.UpdateUserRole)

	body, _ := json.Marshal(map[string]string{"role": "teacher"})
	req := httptest.NewRequest(http.MethodPut, "/admin/users/user-456/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for update role success, got %d", w.Code)
	}
	if !called {
		t.Error("expected UpdateUserRoleFunc to be called with user-456 and teacher")
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["message"] != "อัปเดตบทบาทผู้ใช้สำเร็จ" {
		t.Errorf("unexpected message: %q", resp["message"])
	}
}
