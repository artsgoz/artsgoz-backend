package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/artsgoz/artsgoz-backend/internal/auth"
	"github.com/artsgoz/artsgoz-backend/internal/user"
)

type fullMockRepo struct {
	users map[string]*user.User
}

func newFullMockRepo() *fullMockRepo {
	return &fullMockRepo{users: make(map[string]*user.User)}
}

func (m *fullMockRepo) FindByID(ctx context.Context, id string) (*user.User, error) {
	return m.users[id], nil
}

func (m *fullMockRepo) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}

func (m *fullMockRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *fullMockRepo) CreateUser(ctx context.Context, u *user.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *fullMockRepo) GetAllUsers() ([]*user.User, error) {
	var list []*user.User
	for _, u := range m.users {
		list = append(list, u)
	}
	return list, nil
}

func (m *fullMockRepo) UpdateRole(id string, role string) error {
	if u, ok := m.users[id]; ok {
		u.Role = role
		u.UpdatedAt = time.Now()
	}
	return nil
}

func (m *fullMockRepo) UpdateProfile(id string, username string, email string, avatarURL string) (*user.User, error) {
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

func (m *fullMockRepo) UpdatePassword(id string, passwordHash string) error {
	if u, ok := m.users[id]; ok {
		u.PasswordHash = passwordHash
		u.UpdatedAt = time.Now()
	}
	return nil
}

func setupTestServer() (*gin.Engine, *auth.Service, *user.AdminHandler) {
	gin.SetMode(gin.TestMode)
	app := gin.New()

	repo := newFullMockRepo()
	jwtManager := auth.NewJWTManager("test-secret-key-12345", 24*time.Hour)
	authSvc := auth.NewService(repo, jwtManager)
	adminHandler := user.NewAdminHandler(user.NewAdminService(repo))

	api := app.Group("/api/v1")
	auth.Register(api, authSvc)
	user.RegisterRoutes(api, adminHandler, auth.Middleware(authSvc), nil)

	return app, authSvc, adminHandler
}

func TestE2E_AuthAndAdminFlow(t *testing.T) {
	app, _, _ := setupTestServer()

	// 1. Register a new user
	regBody, _ := json.Marshal(map[string]string{
		"username": "clubmember1",
		"password": "password123",
		"email":    "member1@arts.chula.ac.th",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 on register, got %d: %s", w.Code, w.Body.String())
	}

	var regResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &regResp)
	if regResp["success"] != true {
		t.Errorf("expected success=true, got %v", regResp["success"])
	}
	token, ok := regResp["token"].(string)
	if !ok || token == "" {
		t.Fatalf("expected token in register response")
	}

	// 2. Login
	loginBody, _ := json.Marshal(map[string]string{
		"username": "clubmember1",
		"password": "password123",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 on login, got %d: %s", w.Code, w.Body.String())
	}

	// 3. Access admin dashboard (Role check is removed: ANY authenticated member can access)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 on admin dashboard, got %d: %s", w.Code, w.Body.String())
	}

	// 4. Access admin/me
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 on /admin/me, got %d: %s", w.Code, w.Body.String())
	}
	var meResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &meResp)
	if meResp["username"] != "clubmember1" {
		t.Errorf("expected username clubmember1, got %v", meResp["username"])
	}

	// 5. Update profile via PUT /api/v1/admin/me
	updateBody, _ := json.Marshal(map[string]string{
		"username":   "updated_member",
		"avatar_url": "https://example.com/avatar.png",
	})
	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/me", bytes.NewBuffer(updateBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 on update profile, got %d: %s", w.Code, w.Body.String())
	}

	// 6. Change password via PUT /api/v1/admin/me/password
	pwBody, _ := json.Marshal(map[string]string{
		"current_password": "password123",
		"new_password":     "newsecret456",
	})
	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/me/password", bytes.NewBuffer(pwBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 on change password, got %d: %s", w.Code, w.Body.String())
	}

	// 7. Login with old password should fail, new password should succeed
	badLogin, _ := json.Marshal(map[string]string{
		"username": "updated_member",
		"password": "password123",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(badLogin))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on login with old password, got %d", w.Code)
	}

	goodLogin, _ := json.Marshal(map[string]string{
		"username": "updated_member",
		"password": "newsecret456",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(goodLogin))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on login with new password, got %d: %s", w.Code, w.Body.String())
	}

	// 8. Logout
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on logout, got %d", w.Code)
	}
}
