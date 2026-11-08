package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadDoesNotExposeDotEnvSecrets(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".env", []byte("FIREBASE_CREDENTIALS='SECRET_MARKER"), 0600); err != nil {
		t.Fatal(err)
	}
	err := Load()
	if err == nil {
		t.Fatal("invalid dotenv accepted")
	}
	if strings.Contains(err.Error(), "SECRET_MARKER") {
		t.Fatal("dotenv error exposed secret")
	}
}
