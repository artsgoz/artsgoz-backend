package auth_test

import (
	"github.com/artsgoz/artsgoz-backend/internal/auth"
	"strings"
	"testing"
)

func TestInitFirebaseRequiresProjectID(t *testing.T) {
	_, err := auth.InitFirebase("", "")
	if err == nil || !strings.Contains(err.Error(), "FIREBASE_PROJECT_ID") {
		t.Fatalf("error=%v", err)
	}
}
