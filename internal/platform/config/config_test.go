package config

import (
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
