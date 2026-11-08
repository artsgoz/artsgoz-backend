package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	firebaseauth "firebase.google.com/go/v4/auth"
	"golang.org/x/crypto/bcrypt"

	"github.com/artsgoz/artsgoz-backend/internal/auth"
	"github.com/artsgoz/artsgoz-backend/internal/user"
)

type mockUserRepository struct {
	usersByID          map[string]*user.User
	usersByUsername    map[string]*user.User
	usersByEmail       map[string]*user.User
	usersByFirebaseUID map[string]*user.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		usersByID:          make(map[string]*user.User),
		usersByUsername:    make(map[string]*user.User),
		usersByEmail:       make(map[string]*user.User),
		usersByFirebaseUID: make(map[string]*user.User),
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

func (m *mockUserRepository) FindByFirebaseUID(ctx context.Context, uid string) (*user.User, error) {
	return m.usersByFirebaseUID[uid], nil
}

func (m *mockUserRepository) LinkFirebaseUID(ctx context.Context, id string, firebaseUID string) error {
	if u, ok := m.usersByID[id]; ok {
		u.FirebaseUID = firebaseUID
		m.usersByFirebaseUID[firebaseUID] = u
	}
	return nil
}

func (m *mockUserRepository) CreateUser(ctx context.Context, u *user.User) error {
	m.usersByID[u.ID] = u
	m.usersByUsername[u.Username] = u
	if u.Email != "" {
		m.usersByEmail[u.Email] = u
	}
	if u.FirebaseUID != "" {
		m.usersByFirebaseUID[u.FirebaseUID] = u
	}
	return nil
}

type mockFirebaseClient struct {
	verifyFunc func(ctx context.Context, idToken string) (*firebaseauth.Token, error)
}

func (m *mockFirebaseClient) VerifyIDToken(ctx context.Context, idToken string) (*firebaseauth.Token, error) {
	if m.verifyFunc != nil {
		return m.verifyFunc(ctx, idToken)
	}
	return nil, errors.New("invalid token")
}

func (m *mockFirebaseClient) RevokeRefreshTokens(ctx context.Context, uid string) error {
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
		t.Fatalf("unexpected error during register: %v", err)
	}
	if newUser == nil || newUser.Username != "artsgoz_user" {
		t.Errorf("expected user 'artsgoz_user', got %v", newUser)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// 2. Test Register Duplicate Username
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
		t.Fatalf("unexpected error during login: %v", err)
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

func TestAuthService_FirebaseLogin(t *testing.T) {
	repo := newMockUserRepository()
	jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)

	mockFB := &mockFirebaseClient{
		verifyFunc: func(ctx context.Context, idToken string) (*firebaseauth.Token, error) {
			if idToken == "valid-token" {
				return &firebaseauth.Token{
					UID: "fb-uid-12345",
					Claims: map[string]interface{}{
						"email": "student@arts.chula.ac.th",
					},
				}, nil
			}
			return nil, errors.New("invalid firebase token")
		},
	}

	svc := auth.NewService(repo, jwtManager, mockFB)
	ctx := context.Background()

	// 1. First-time login: provisions new user
	res, err := svc.FirebaseLogin(ctx, auth.FirebaseLoginRequest{
		IDToken:  "valid-token",
		Username: "chula_student",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.User.FirebaseUID != "fb-uid-12345" {
		t.Errorf("expected FirebaseUID 'fb-uid-12345', got %s", res.User.FirebaseUID)
	}
	if res.User.Email != "student@arts.chula.ac.th" {
		t.Errorf("expected Email 'student@arts.chula.ac.th', got %s", res.User.Email)
	}
	if res.User.Username != "chula_student" {
		t.Errorf("expected Username 'chula_student', got %s", res.User.Username)
	}
	if res.Token == "" {
		t.Error("expected non-empty JWT token")
	}

	// 2. Second-time login: retrieves same user
	res2, err := svc.FirebaseLogin(ctx, auth.FirebaseLoginRequest{
		IDToken: "valid-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res2.User.ID != res.User.ID {
		t.Errorf("expected same user ID %s, got %s", res.User.ID, res2.User.ID)
	}

	// 3. Invalid token returns error
	_, err = svc.FirebaseLogin(ctx, auth.FirebaseLoginRequest{
		IDToken: "bad-token",
	})
	if err == nil {
		t.Fatal("expected error for bad token, got nil")
	}
}
