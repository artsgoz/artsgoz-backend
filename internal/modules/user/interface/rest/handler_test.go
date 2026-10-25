package rest_test

import (
	"bytes"
	"context"
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
	LogoutFunc         func(ctx context.Context, uid string) error
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

func (m *MockLoginUsecase) Logout(ctx context.Context, uid string) error {
	if m.LogoutFunc != nil {
		return m.LogoutFunc(ctx, uid)
	}
	return nil
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
	GetProfileFunc     func(uid string) (*domain.User, error)
	UpdateProfileFunc  func(uid string, req application.UpdateProfileRequest) (*domain.User, error)
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

func (m *MockAdminUsecase) GetProfile(uid string) (*domain.User, error) {
	if m.GetProfileFunc != nil {
		return m.GetProfileFunc(uid)
	}
	return nil, nil
}

func (m *MockAdminUsecase) UpdateProfile(uid string, req application.UpdateProfileRequest) (*domain.User, error) {
	if m.UpdateProfileFunc != nil {
		return m.UpdateProfileFunc(uid, req)
	}
	return nil, nil
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

// ----------------- AuthHandler Logout Tests -----------------

func TestAuthHandler_Logout_WithoutUID(t *testing.T) {
	loginUsecase := &MockLoginUsecase{}
	registerUsecase := &MockRegisterUsecase{}
	handler := rest.NewAuthHandler(loginUsecase, registerUsecase)

	router := gin.New()
	router.POST("/auth/logout", handler.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for logout, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["message"] != "ออกจากระบบสำเร็จ" {
		t.Errorf("expected message 'ออกจากระบบสำเร็จ', got %q", resp["message"])
	}
}

func TestAuthHandler_Logout_WithUID(t *testing.T) {
	logoutCalled := false
	loginUsecase := &MockLoginUsecase{
		LogoutFunc: func(ctx context.Context, uid string) error {
			if uid == "user-123" {
				logoutCalled = true
				return nil
			}
			return errors.New("unexpected uid")
		},
	}
	registerUsecase := &MockRegisterUsecase{}
	handler := rest.NewAuthHandler(loginUsecase, registerUsecase)

	router := gin.New()
	router.POST("/auth/logout", func(c *gin.Context) {
		c.Set("uid", "user-123")
		handler.Logout(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for logout, got %d", w.Code)
	}
	if !logoutCalled {
		t.Error("expected LogoutFunc to be called with user-123")
	}
}

// ----------------- AdminHandler Profile Tests -----------------

func TestAdminHandler_GetProfile_MissingUID(t *testing.T) {
	adminUsecase := &MockAdminUsecase{}
	handler := rest.NewAdminHandler(adminUsecase)

	router := gin.New()
	router.GET("/admin/me", handler.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/admin/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 when uid is missing, got %d", w.Code)
	}
}

func TestAdminHandler_GetProfile_EmptyUID(t *testing.T) {
	adminUsecase := &MockAdminUsecase{}
	handler := rest.NewAdminHandler(adminUsecase)

	router := gin.New()
	router.GET("/admin/me", func(c *gin.Context) {
		c.Set("uid", "")
		handler.GetProfile(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 when uid is empty string, got %d", w.Code)
	}
}

func TestAdminHandler_GetProfile_Error(t *testing.T) {
	adminUsecase := &MockAdminUsecase{
		GetProfileFunc: func(uid string) (*domain.User, error) {
			return nil, errors.New("db error")
		},
	}
	handler := rest.NewAdminHandler(adminUsecase)

	router := gin.New()
	router.GET("/admin/me", func(c *gin.Context) {
		c.Set("uid", "admin-123")
		handler.GetProfile(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 when get profile fails, got %d", w.Code)
	}
}

func TestAdminHandler_GetProfile_Success(t *testing.T) {
	mockAdmin := &domain.User{
		FirebaseUID: "admin-123",
		Email:       "admin@arts.chula.ac.th",
		Role:        "admin",
	}
	adminUsecase := &MockAdminUsecase{
		GetProfileFunc: func(uid string) (*domain.User, error) {
			if uid == "admin-123" {
				return mockAdmin, nil
			}
			return nil, errors.New("not found")
		},
	}
	handler := rest.NewAdminHandler(adminUsecase)

	router := gin.New()
	router.GET("/admin/me", func(c *gin.Context) {
		c.Set("uid", "admin-123")
		handler.GetProfile(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for GetProfile, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["firebase_uid"] != "admin-123" || resp["email"] != "admin@arts.chula.ac.th" || resp["role"] != "admin" {
		t.Errorf("unexpected profile response: %+v", resp)
	}
}

func TestAdminHandler_UpdateProfile_MissingUID(t *testing.T) {
	adminUsecase := &MockAdminUsecase{}
	handler := rest.NewAdminHandler(adminUsecase)

	router := gin.New()
	router.PUT("/admin/me", handler.UpdateProfile)

	req := httptest.NewRequest(http.MethodPut, "/admin/me", bytes.NewBufferString(`{"email":"test@arts.chula.ac.th"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 when uid missing, got %d", w.Code)
	}
}

func TestAdminHandler_UpdateProfile_InvalidJSON(t *testing.T) {
	adminUsecase := &MockAdminUsecase{}
	handler := rest.NewAdminHandler(adminUsecase)

	router := gin.New()
	router.PUT("/admin/me", func(c *gin.Context) {
		c.Set("uid", "admin-123")
		handler.UpdateProfile(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/admin/me", bytes.NewBufferString(`{bad-json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for bad json, got %d", w.Code)
	}
}

func TestAdminHandler_UpdateProfile_Error(t *testing.T) {
	adminUsecase := &MockAdminUsecase{
		UpdateProfileFunc: func(uid string, req application.UpdateProfileRequest) (*domain.User, error) {
			return nil, errors.New("รูปแบบ email ไม่ถูกต้อง")
		},
	}
	handler := rest.NewAdminHandler(adminUsecase)

	router := gin.New()
	router.PUT("/admin/me", func(c *gin.Context) {
		c.Set("uid", "admin-123")
		handler.UpdateProfile(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/admin/me", bytes.NewBufferString(`{"email":"invalid"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 when update profile fails, got %d", w.Code)
	}
}

func TestAdminHandler_UpdateProfile_Success(t *testing.T) {
	updatedAdmin := &domain.User{
		FirebaseUID: "admin-123",
		Email:       "newadmin@arts.chula.ac.th",
		Role:        "admin",
	}
	adminUsecase := &MockAdminUsecase{
		UpdateProfileFunc: func(uid string, req application.UpdateProfileRequest) (*domain.User, error) {
			if uid == "admin-123" && req.Email == "newadmin@arts.chula.ac.th" {
				return updatedAdmin, nil
			}
			return nil, errors.New("unexpected params")
		},
	}
	handler := rest.NewAdminHandler(adminUsecase)

	router := gin.New()
	router.PUT("/admin/me", func(c *gin.Context) {
		c.Set("uid", "admin-123")
		handler.UpdateProfile(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/admin/me", bytes.NewBufferString(`{"email":"newadmin@arts.chula.ac.th"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for update profile success, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["message"] != "แก้ไขข้อมูลส่วนตัวสำเร็จ" || resp["email"] != "newadmin@arts.chula.ac.th" {
		t.Errorf("unexpected update profile response: %+v", resp)
	}
}

// ----------------- Route Registration Tests -----------------

func TestRegisterRoutes_PathRegistration(t *testing.T) {
	loginUsecase := &MockLoginUsecase{
		AuthenticateFunc: func(req application.LoginRequest) (string, string, error) {
			return "user-1", "student", nil
		},
	}
	registerUsecase := &MockRegisterUsecase{}
	adminUsecase := &MockAdminUsecase{}
	authHandler := rest.NewAuthHandler(loginUsecase, registerUsecase)
	adminHandler := rest.NewAdminHandler(adminUsecase)

	router := gin.New()
	// Register without firebase auth & pool for routing check
	rest.RegisterRoutes(router.Group(""), authHandler, adminHandler, nil, nil)

	routes := router.Routes()
	hasPostAuthLogin := false
	hasPostAuthLogout := false
	hasGetAdminMe := false
	hasPutAdminMe := false

	for _, r := range routes {
		if r.Method == http.MethodPost && r.Path == "/auth/login" {
			hasPostAuthLogin = true
		}
		if r.Method == http.MethodPost && r.Path == "/auth/logout" {
			hasPostAuthLogout = true
		}
		if r.Method == http.MethodGet && r.Path == "/admin/me" {
			hasGetAdminMe = true
		}
		if r.Method == http.MethodPut && r.Path == "/admin/me" {
			hasPutAdminMe = true
		}
	}

	if !hasPostAuthLogin {
		t.Error("route POST /auth/login is missing")
	}
	if !hasPostAuthLogout {
		t.Error("route POST /auth/logout is missing")
	}
	if !hasGetAdminMe {
		t.Error("route GET /admin/me is missing")
	}
	if !hasPutAdminMe {
		t.Error("route PUT /admin/me is missing")
	}
}
