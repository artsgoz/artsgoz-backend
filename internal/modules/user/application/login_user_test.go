package application_test

import (
	"errors"
	"testing"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/application"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"
)

func TestLoginUsecase_GetUserProfile_EmptyUID(t *testing.T) {
	repo := &MockUserRepository{}
	usecase := application.NewLoginUsecase(repo, nil)

	_, err := usecase.GetUserProfile("")
	if err == nil {
		t.Error("expected error for empty uid, got nil")
	}
	if err != nil && err.Error() != "กรุณาระบุ uid" {
		t.Errorf("expected error 'กรุณาระบุ uid', got %q", err.Error())
	}

	_, err = usecase.GetUserProfile("   ")
	if err == nil {
		t.Error("expected error for whitespace uid, got nil")
	}
}

func TestLoginUsecase_GetUserProfile_Success(t *testing.T) {
	expectedUser := &domain.User{
		FirebaseUID: "user-abc",
		Email:       "test@example.com",
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

func TestLoginUsecase_Authenticate_EmptyToken(t *testing.T) {
	repo := &MockUserRepository{}
	usecase := application.NewLoginUsecase(repo, nil)

	_, _, err := usecase.Authenticate(application.LoginRequest{IDToken: ""})
	if err == nil {
		t.Error("expected error for empty id_token, got nil")
	}
	if err != nil && err.Error() != "กรุณาระบุ id_token" {
		t.Errorf("expected 'กรุณาระบุ id_token', got %q", err.Error())
	}

	_, _, err = usecase.Authenticate(application.LoginRequest{IDToken: "   "})
	if err == nil {
		t.Error("expected error for whitespace id_token, got nil")
	}
}

func TestLoginUsecase_Authenticate_NilFirebaseAuth(t *testing.T) {
	repo := &MockUserRepository{}
	usecase := application.NewLoginUsecase(repo, nil)

	_, _, err := usecase.Authenticate(application.LoginRequest{IDToken: "valid-token-format"})
	if err == nil {
		t.Error("expected error for nil firebase auth, got nil")
	}
	if err != nil && err.Error() != "firebase auth is not initialized" {
		t.Errorf("expected 'firebase auth is not initialized', got %q", err.Error())
	}
}

func TestDomain_IsChulaEmail(t *testing.T) {
	testCases := []struct {
		email    string
		expected bool
	}{
		{"student@student.chula.ac.th", true},
		{"somchai.p@chula.ac.th", true},
		{"lecturer@arts.chula.ac.th", true},
		{"alumni@alumni.chula.ac.th", true},
		{"TEST@STUDENT.CHULA.AC.TH", true},
		{"user@gmail.com", false},
		{"user@hotmail.com", false},
		{"user@fakechula.ac.th", false},
		{"chula.ac.th@gmail.com", false},
		{"user@notachula.ac.th", false},
		{"", false},
		{"   ", false},
		{"invalid-email", false},
	}

	for _, tc := range testCases {
		result := domain.IsChulaEmail(tc.email)
		if result != tc.expected {
			t.Errorf("IsChulaEmail(%q) = %v; want %v", tc.email, result, tc.expected)
		}
	}
}

func TestRegisterUsecase_EmptyToken(t *testing.T) {
	repo := &MockUserRepository{}
	usecase := application.NewRegisterUsecase(repo, nil)

	err := usecase.Register(application.RegisterRequest{
		IDToken: "",
	})
	if err == nil {
		t.Fatal("expected error for empty id_token, got nil")
	}
	expected := "กรุณาสมัครสมาชิกผ่าน Google เท่านั้น"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestRegisterUsecase_NilFirebaseAuth(t *testing.T) {
	repo := &MockUserRepository{}
	usecase := application.NewRegisterUsecase(repo, nil)

	err := usecase.Register(application.RegisterRequest{
		IDToken: "sample-token",
	})
	if err == nil {
		t.Fatal("expected error for nil firebase auth, got nil")
	}
	expected := "firebase auth is not initialized"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}
