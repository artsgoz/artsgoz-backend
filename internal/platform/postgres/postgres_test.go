package postgres_test

import (
	"testing"

	"github.com/artsgoz/artsgoz-backend/internal/platform/postgres"
)

func TestConnectionURL(t *testing.T) {
	t.Parallel()

	got, err := postgres.ConnectionURL(postgres.Config{
		Host: "localhost", Port: "5432", User: "app user", Password: "p@ssword",
		Database: "artsgoz", SSLMode: "disable",
	})
	if err != nil {
		t.Fatalf("ConnectionURL() error = %v", err)
	}
	const want = "postgres://app%20user:p%40ssword@localhost:5432/artsgoz?sslmode=disable"
	if got != want {
		t.Fatalf("ConnectionURL() = %q, want %q", got, want)
	}
}

func TestConnectionURLRejectsIncompleteConfig(t *testing.T) {
	t.Parallel()

	if _, err := postgres.ConnectionURL(postgres.Config{Host: "localhost"}); err == nil {
		t.Fatal("ConnectionURL() error = nil, want an error")
	}
}
