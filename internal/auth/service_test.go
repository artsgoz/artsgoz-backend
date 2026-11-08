package auth_test

import (
	"context"
	"errors"
	"testing"

	firebaseauth "firebase.google.com/go/v4/auth"

	"github.com/artsgoz/artsgoz-backend/internal/auth"
	"github.com/artsgoz/artsgoz-backend/internal/user"
)

type mockUserRepository struct {
	usersByID          map[string]*user.User
	usersByUsername    map[string]*user.User
	usersByEmail       map[string]*user.User
	usersByFirebaseUID map[string]*user.User
	linkErr            error
	findErr            error
	createHook         func(*user.User) error
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
	return m.usersByUsername[username], m.findErr
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	return m.usersByEmail[email], m.findErr
}

func (m *mockUserRepository) FindByFirebaseUID(ctx context.Context, uid string) (*user.User, error) {
	return m.usersByFirebaseUID[uid], m.findErr
}

func (m *mockUserRepository) LinkFirebaseUID(ctx context.Context, id string, firebaseUID string) error {
	if m.linkErr != nil {
		return m.linkErr
	}
	if u, ok := m.usersByID[id]; ok {
		u.FirebaseUID = firebaseUID
		m.usersByFirebaseUID[firebaseUID] = u
	}
	return nil
}

func (m *mockUserRepository) CreateUser(ctx context.Context, u *user.User) error {
	if m.createHook != nil {
		return m.createHook(u)
	}
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
	revoked    string
	revokeErr  error
}

func (m *mockFirebaseClient) VerifyIDToken(ctx context.Context, idToken string) (*firebaseauth.Token, error) {
	if m.verifyFunc != nil {
		return m.verifyFunc(ctx, idToken)
	}
	return nil, errors.New("invalid token")
}

func (m *mockFirebaseClient) RevokeRefreshTokens(ctx context.Context, uid string) error {
	m.revoked = uid
	return m.revokeErr
}

func firebaseToken(uid, email, provider string, verified bool) *firebaseauth.Token {
	return &firebaseauth.Token{UID: uid, Firebase: firebaseauth.FirebaseInfo{SignInProvider: provider},
		Claims: map[string]interface{}{"email": email, "email_verified": verified}}
}

func clientFor(token *firebaseauth.Token) *mockFirebaseClient {
	return &mockFirebaseClient{verifyFunc: func(context.Context, string) (*firebaseauth.Token, error) { return token, nil }}
}

func TestFirebaseLoginBothProvidersShareProfile(t *testing.T) {
	repo := newMockUserRepository()
	fb := clientFor(firebaseToken("uid", "person@example.com", "password", false))
	svc := auth.NewService(repo, fb)
	first, err := svc.FirebaseLogin(context.Background(), auth.FirebaseLoginRequest{IDToken: "password-token"})
	if err != nil {
		t.Fatal(err)
	}
	if first.Token != "password-token" || first.User.PasswordHash != "" || first.User.Role != "member" {
		t.Fatalf("response = %#v", first)
	}
	fb.verifyFunc = clientFor(firebaseToken("uid", "person@example.com", "google.com", true)).verifyFunc
	second, err := svc.FirebaseLogin(context.Background(), auth.FirebaseLoginRequest{IDToken: "google-token"})
	if err != nil {
		t.Fatal(err)
	}
	if second.User.ID != first.User.ID || len(repo.usersByID) != 1 || second.Token != "google-token" {
		t.Fatalf("profile duplicated: %#v", second)
	}
}

func TestFirebaseLoginLinksOnlyVerifiedUnlinkedEmail(t *testing.T) {
	for _, tc := range []struct {
		name        string
		verified    bool
		existingUID string
		wantError   bool
	}{
		{"verified legacy", true, "", false},
		{"unverified legacy", false, "", true},
		{"different Firebase identity", true, "other-uid", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := newMockUserRepository()
			legacy := &user.User{ID: "legacy-id", Email: "person@example.com", Username: "legacy", Role: "editor", FirebaseUID: tc.existingUID, PasswordHash: "old-hash"}
			repo.CreateUser(context.Background(), legacy)
			svc := auth.NewService(repo, clientFor(firebaseToken("new-uid", legacy.Email, "password", tc.verified)))
			result, err := svc.FirebaseLogin(context.Background(), auth.FirebaseLoginRequest{IDToken: "token"})
			if tc.wantError {
				if !errors.Is(err, auth.ErrAccountConflict) || legacy.FirebaseUID != tc.existingUID {
					t.Fatalf("link error = %v, UID = %s", err, legacy.FirebaseUID)
				}
				return
			}
			if err != nil || result.User.ID != "legacy-id" || result.User.Role != "editor" || legacy.FirebaseUID != "new-uid" || legacy.PasswordHash != "old-hash" {
				t.Fatalf("legacy link: result=%#v, err=%v", result, err)
			}
		})
	}
}

func TestFirebaseLoginRejectsInvalidTokens(t *testing.T) {
	cases := []*firebaseauth.Token{nil, firebaseToken("", "person@example.com", "password", true), firebaseToken("uid", "person@example.com", "anonymous", true), firebaseToken("uid", "bad-email", "google.com", true)}
	for _, token := range cases {
		repo := newMockUserRepository()
		svc := auth.NewService(repo, clientFor(token))
		if _, err := svc.FirebaseLogin(context.Background(), auth.FirebaseLoginRequest{IDToken: "token"}); !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("error = %v", err)
		}
		if len(repo.usersByID) != 0 {
			t.Fatal("created profile for invalid token")
		}
	}
	svc := auth.NewService(newMockUserRepository(), &mockFirebaseClient{})
	for _, raw := range []string{"", "expired", "revoked", "legacy-backend-jwt"} {
		if _, err := svc.FirebaseLogin(context.Background(), auth.FirebaseLoginRequest{IDToken: raw}); !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("token %q: %v", raw, err)
		}
	}
	if _, err := auth.NewService(newMockUserRepository(), nil).FirebaseLogin(context.Background(), auth.FirebaseLoginRequest{IDToken: "token"}); !errors.Is(err, auth.ErrFirebaseUnavailable) {
		t.Fatalf("error = %v", err)
	}
}

func TestFirebaseLoginPropagatesRepositoryErrors(t *testing.T) {
	repo := newMockUserRepository()
	failure := errors.New("database unavailable")
	svc := auth.NewService(repo, clientFor(firebaseToken("uid", "person@example.com", "google.com", true)))
	repo.findErr = failure
	if _, err := svc.FirebaseLogin(context.Background(), auth.FirebaseLoginRequest{IDToken: "token"}); !errors.Is(err, failure) {
		t.Fatalf("lookup error = %v", err)
	}
	repo.findErr = nil
	legacy := &user.User{ID: "old", Email: "person@example.com"}
	repo.CreateUser(context.Background(), legacy)
	repo.linkErr = failure
	if _, err := svc.FirebaseLogin(context.Background(), auth.FirebaseLoginRequest{IDToken: "token"}); !errors.Is(err, failure) || legacy.FirebaseUID != "" {
		t.Fatalf("link error = %v", err)
	}
}

func TestFirebaseLoginReusesConcurrentProfile(t *testing.T) {
	repo := newMockUserRepository()
	repo.createHook = func(profile *user.User) error {
		profile.ID = "winner"
		repo.usersByFirebaseUID[profile.FirebaseUID] = profile
		return errors.New("unique constraint")
	}
	svc := auth.NewService(repo, clientFor(firebaseToken("uid", "person@example.com", "password", false)))
	result, err := svc.FirebaseLogin(context.Background(), auth.FirebaseLoginRequest{IDToken: "token"})
	if err != nil || result.User.ID != "winner" {
		t.Fatalf("result=%#v, error=%v", result, err)
	}
}

func TestFirebaseLoginHandlesShortUIDAndUsernameCollision(t *testing.T) {
	repo := newMockUserRepository()
	svc := auth.NewService(repo, clientFor(firebaseToken("x", "person@example.com", "password", true)))
	if _, err := svc.FirebaseLogin(context.Background(), auth.FirebaseLoginRequest{IDToken: "token", Username: "ab"}); !errors.Is(err, auth.ErrInvalidUsername) {
		t.Fatalf("error=%v", err)
	}
	repo.usersByUsername["taken"] = &user.User{ID: "other"}
	if _, err := svc.FirebaseLogin(context.Background(), auth.FirebaseLoginRequest{IDToken: "token", Username: "taken"}); !errors.Is(err, auth.ErrAccountConflict) {
		t.Fatalf("error=%v", err)
	}
	if _, err := svc.FirebaseLogin(context.Background(), auth.FirebaseLoginRequest{IDToken: "token"}); err != nil {
		t.Fatal(err)
	}
}
