package config

import (
	"testing"
)

func TestHasDatabaseParts(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"DB_USER":     "app user",
		"DB_PASSWORD": "p@ssword",
		"DB_NAME":     "artsgoz",
	}
	if !hasDatabaseParts(values) {
		t.Fatal("hasDatabaseParts() = false, want true")
	}
}

func TestHasDatabasePartsRejectsIncompleteSettings(t *testing.T) {
	t.Parallel()

	if hasDatabaseParts(map[string]string{"DB_HOST": "localhost"}) {
		t.Fatal("hasDatabaseParts() = true, want false")
	}
}

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
