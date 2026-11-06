package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/auth"
	"github.com/artsgoz/artsgoz-backend/internal/user"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	usersByID       map[string]*user.User
	usersByUsername map[string]*user.User
	usersByEmail    map[string]*user.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		usersByID:       make(map[string]*user.User),
		usersByUsername: make(map[string]*user.User),
		usersByEmail:    make(map[string]*user.User),
	}
}

func (m *mockUserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	return m.usersByID[id], nil
}

func (m *mockUserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	return m.usersByUsername[username], nil
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	return m.usersByEmail[email], nil
}

func (m *mockUserRepository) CreateUser(ctx context.Context, u *user.User) error {
	m.usersByID[u.ID] = u
	m.usersByUsername[u.Username] = u
	if u.Email != "" {
		m.usersByEmail[u.Email] = u
	}
	return nil
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	repo := newMockUserRepository()
	jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
	svc := auth.NewService(repo, jwtManager)
	ctx := context.Background()

	// 1. Test Register
	regReq := auth.RegisterRequest{
		Username: "artsgoz_user",
		Password: "securepassword123",
		Email:    "test@example.com",
	}

	newUser, token, err := svc.Register(ctx, regReq)
	if err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}
	if newUser.Username != "artsgoz_user" {
		t.Errorf("expected username 'artsgoz_user', got %s", newUser.Username)
	}
	if newUser.Role != "member" {
		t.Errorf("expected default role 'member', got %s", newUser.Role)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// 2. Test Duplicate Register
	_, _, err = svc.Register(ctx, regReq)
	if err == nil {
		t.Fatal("expected error for duplicate username, got nil")
	}

	// 3. Test Login Success
	loginRes, err := svc.Login(ctx, auth.LoginRequest{
		Username: "artsgoz_user",
		Password: "securepassword123",
	})
	if err != nil {
		t.Fatalf("unexpected login error: %v", err)
	}
	if loginRes.Token == "" {
		t.Fatal("expected non-empty login token")
	}
	if loginRes.User.Username != "artsgoz_user" {
		t.Errorf("expected user username 'artsgoz_user', got %s", loginRes.User.Username)
	}

	// 4. Test Login Wrong Password
	_, err = svc.Login(ctx, auth.LoginRequest{
		Username: "artsgoz_user",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}

	// 5. Test Login Non-existent User
	_, err = svc.Login(ctx, auth.LoginRequest{
		Username: "unknown_user",
		Password: "somepassword",
	})
	if err == nil {
		t.Fatal("expected error for unknown user, got nil")
	}
}

func TestAuthService_PasswordIsHashed(t *testing.T) {
	repo := newMockUserRepository()
	jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
	svc := auth.NewService(repo, jwtManager)
	ctx := context.Background()

	rawPassword := "mypassword123"
	newUser, _, err := svc.Register(ctx, auth.RegisterRequest{
		Username: "hashed_user",
		Password: rawPassword,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if newUser.PasswordHash == rawPassword {
		t.Fatal("password was stored in plain text, expected bcrypt hash")
	}

	err = bcrypt.CompareHashAndPassword([]byte(newUser.PasswordHash), []byte(rawPassword))
	if err != nil {
		t.Fatalf("bcrypt compare failed: %v", err)
	}
}
