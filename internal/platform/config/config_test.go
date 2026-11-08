package config

import (
	"os"
	"strings"
	"testing"
)

func TestValidatePort(t *testing.T) {
	t.Parallel()

	for _, port := range []string{"1", "3000", "65535"} {
		if err := validatePort(port); err != nil {
			t.Errorf("validatePort(%q) error = %v", port, err)
		}
	}
	for _, port := range []string{"", "0", "65536", "invalid", ":3000"} {
		if err := validatePort(port); err == nil {
			t.Errorf("validatePort(%q) error = nil, want an error", port)
		}
	}
}

func TestValidateGinMode(t *testing.T) {
	t.Parallel()

	for _, mode := range []string{"debug", "release", "test"} {
		if err := validateGinMode(mode); err != nil {
			t.Errorf("validateGinMode(%q) error = %v", mode, err)
		}
	}
	if err := validateGinMode("production"); err == nil {
		t.Fatal("validateGinMode() error = nil, want an error")
	}
}

func TestFirebaseConfiguration(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_USER", "test")
	t.Setenv("DB_NAME", "test")
	t.Setenv("PORT", "3000")
	t.Setenv("GIN_MODE", "test")
	t.Setenv("FIREBASE_PROJECT_ID", "")
	t.Setenv("GCP_PROJECT_ID", "fallback-project")
	t.Setenv("FIREBASE_CREDENTIALS", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FirebaseProjectID != "fallback-project" || cfg.FirebaseCredentials != "firebase-service-account.json" {
		t.Fatal("Firebase defaults were not loaded")
	}
	t.Setenv("FIREBASE_PROJECT_ID", "explicit-project")
	t.Setenv("FIREBASE_CREDENTIALS", "/test/credentials.json")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FirebaseProjectID != "explicit-project" || cfg.FirebaseCredentials != "/test/credentials.json" {
		t.Fatal("explicit Firebase configuration was not loaded")
	}
}

func TestDotEnvFallbackAndPrecedence(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_USER", "test")
	t.Setenv("DB_NAME", "test")
	t.Setenv("GIN_MODE", "test")
	t.Setenv("PORT", "")
	if err := os.Unsetenv("PORT"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".env", []byte("PORT=4010\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil || cfg.Port != "4010" {
		t.Fatalf(".env fallback: port=%s error=%v", cfg.Port, err)
	}
	if err := os.WriteFile(".env.local", []byte("PORT=4020\n"), 0600); err != nil {
		t.Fatal(err)
	}
	os.Unsetenv("PORT")
	cfg, err = Load()
	if err != nil || cfg.Port != "4020" {
		t.Fatalf(".env.local precedence: port=%s error=%v", cfg.Port, err)
	}
	t.Setenv("PORT", "4030")
	cfg, err = Load()
	if err != nil || cfg.Port != "4030" {
		t.Fatalf("process precedence: port=%s error=%v", cfg.Port, err)
	}
}

func TestDotEnvErrorsDoNotExposeSecrets(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".env", []byte("FIREBASE_CREDENTIALS='SECRET_MARKER"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := Load()
	if err == nil {
		t.Fatal("invalid dotenv accepted")
	}
	if strings.Contains(err.Error(), "SECRET_MARKER") {
		t.Fatal("dotenv error exposed secret")
	}
}
