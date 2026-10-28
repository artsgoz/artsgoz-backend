package config

import (
	"os"
	"testing"
)

func TestLoadMissingFile(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := Load(); err != nil {
		t.Fatalf("missing .env.local should be optional: %v", err)
	}
}

func TestLoadLocalFilePreservesEnvironment(t *testing.T) {
	t.Chdir(t.TempDir())
	const loadedKey = "ARTSGOZ_CONFIG_TEST_LOADED"
	const overrideKey = "ARTSGOZ_CONFIG_TEST_OVERRIDE"
	t.Setenv(loadedKey, "")
	if err := os.Unsetenv(loadedKey); err != nil {
		t.Fatal(err)
	}
	t.Setenv(overrideKey, "from-shell")
	if err := os.WriteFile(".env.local", []byte(loadedKey+"=from-file\n"+overrideKey+"=from-file\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv(loadedKey); got != "from-file" {
		t.Errorf("file value = %q, want from-file", got)
	}
	if got := os.Getenv(overrideKey); got != "from-shell" {
		t.Errorf("environment override = %q, want from-shell", got)
	}
}

func TestLoadInvalidFile(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".env.local", []byte("INVALID='unterminated\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Load(); err == nil {
		t.Fatal("expected an error for malformed .env.local")
	}
}
