package auth_test

import (
	"bytes"
	"context"
	"errors"
	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/artsgoz/artsgoz-backend/internal/server"
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

func (m *fullMockRepo) FindByFirebaseUID(ctx context.Context, uid string) (*user.User, error) {
	for _, u := range m.users {
		if u.FirebaseUID == uid {
			return u, nil
		}
	}
	return nil, nil
}

func (m *fullMockRepo) LinkFirebaseUID(ctx context.Context, id string, firebaseUID string) error {
	if u, ok := m.users[id]; ok {
		u.FirebaseUID = firebaseUID
		u.UpdatedAt = time.Now()
	}
	return nil
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

func setupFirebaseServer(provider string) (*gin.Engine, *mockFirebaseClient) {
	gin.SetMode(gin.TestMode)
	app := gin.New()
	repo := newFullMockRepo()
	fb := &mockFirebaseClient{}
	fb.verifyFunc = func(_ context.Context, raw string) (*firebaseauth.Token, error) {
		if raw != "firebase-token" || fb.revoked != "" {
			return nil, errors.New("invalid token")
		}
		return firebaseToken("firebase-uid", "person@example.com", provider, true), nil
	}
	svc := auth.NewService(repo, fb)
	api := app.Group("/api/v1")
	auth.Register(api, svc)
	user.RegisterAdminRoutes(api, user.NewAdminHandler(user.NewAdminService(repo)), auth.Middleware(svc), nil)
	admin := api.Group("/secure-admin", server.AuthRequired(auth.Middleware(svc)), server.RequireRoles(server.RoleEditor))
	admin.GET("", func(c *gin.Context) { c.Status(http.StatusOK) })
	return app, fb
}

func request(app *gin.Engine, method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("X-User-Role", "editor")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, req)
	return response
}

func TestFirebaseLoginAndProtectedRoutes(t *testing.T) {
	for _, provider := range []string{"password", "google.com"} {
		t.Run(provider, func(t *testing.T) {
			app, _ := setupFirebaseServer(provider)
			for _, path := range []string{"/auth/login", "/login", "/auth/firebase", "/auth/firebase/login", "/auth/register", "/register", "/users/register"} {
				response := request(app, http.MethodPost, "/api/v1"+path, `{"id_token":"firebase-token"}`, "")
				if response.Code != http.StatusOK {
					t.Fatalf("%s: status=%d, body=%s", path, response.Code, response.Body.String())
				}
				if !bytes.Contains(response.Body.Bytes(), []byte(`"token":"firebase-token"`)) {
					t.Fatalf("missing Firebase token: %s", response.Body.String())
				}
			}
			response := request(app, http.MethodGet, "/api/v1/user/me", "", "firebase-token")
			if response.Code != http.StatusOK {
				t.Fatalf("profile status=%d", response.Code)
			}
			for _, token := range []string{"", "legacy-backend-jwt", "expired"} {
				if response := request(app, http.MethodGet, "/api/v1/auth/me", "", token); response.Code != http.StatusUnauthorized {
					t.Fatalf("token %q: status=%d", token, response.Code)
				}
			}
			// A client-supplied role cannot promote a Firebase member to editor.
			if response := request(app, http.MethodGet, "/api/v1/secure-admin", "", "firebase-token"); response.Code != http.StatusForbidden {
				t.Fatalf("spoofed role status=%d", response.Code)
			}
			response = request(app, http.MethodPut, "/api/v1/user/password", `{"current_password":"old","new_password":"newpassword"}`, "firebase-token")
			if response.Code != http.StatusGone {
				t.Fatalf("password status=%d", response.Code)
			}
		})
	}
}

func TestLegacyPasswordRequestsRejected(t *testing.T) {
	app, _ := setupFirebaseServer("password")
	for _, path := range []string{"/auth/login", "/auth/register", "/users/register"} {
		response := request(app, http.MethodPost, "/api/v1"+path, `{"email":"person@example.com","username":"person","password":"password123"}`, "")
		if response.Code != http.StatusBadRequest {
			t.Fatalf("%s: status=%d", path, response.Code)
		}
	}
}

func TestFirebaseLogoutRevokesSessions(t *testing.T) {
	app, fb := setupFirebaseServer("google.com")
	request(app, http.MethodPost, "/api/v1/auth/login", `{"id_token":"firebase-token"}`, "")
	response := request(app, http.MethodPost, "/api/v1/auth/logout", "", "firebase-token")
	if response.Code != http.StatusOK || fb.revoked != "firebase-uid" {
		t.Fatalf("logout status=%d, UID=%s", response.Code, fb.revoked)
	}
	if response := request(app, http.MethodGet, "/api/v1/auth/me", "", "firebase-token"); response.Code != http.StatusUnauthorized {
		t.Fatalf("revoked status=%d", response.Code)
	}
}

func TestFirebaseLogoutReportsRevocationFailure(t *testing.T) {
	app, fb := setupFirebaseServer("password")
	request(app, http.MethodPost, "/api/v1/auth/login", `{"id_token":"firebase-token"}`, "")
	fb.revokeErr = errors.New("Firebase unavailable")
	if response := request(app, http.MethodPost, "/api/v1/auth/logout", "", "firebase-token"); response.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", response.Code)
	}
}
