package auth_test

import (
	"testing"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/auth"
)

func TestJWTManager_GenerateAndValidateToken(t *testing.T) {
	manager := auth.NewJWTManager("test-secret-key", 1*time.Hour)

	token, err := manager.GenerateToken("user-123", "johndoe", "student")
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("expected UserID 'user-123', got '%s'", claims.UserID)
	}
	if claims.Username != "johndoe" {
		t.Errorf("expected Username 'johndoe', got '%s'", claims.Username)
	}
	if claims.Role != "student" {
		t.Errorf("expected Role 'student', got '%s'", claims.Role)
	}
}

func TestJWTManager_ValidateExpiredToken(t *testing.T) {
	manager := auth.NewJWTManager("test-secret-key", -1*time.Minute)

	token, err := manager.GenerateToken("user-123", "johndoe", "student")
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	_, err = manager.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestJWTManager_ValidateInvalidSecret(t *testing.T) {
	manager1 := auth.NewJWTManager("secret-key-1", 1*time.Hour)
	manager2 := auth.NewJWTManager("secret-key-2", 1*time.Hour)

	token, err := manager1.GenerateToken("user-123", "johndoe", "student")
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	_, err = manager2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected signature verification error, got nil")
	}
}
