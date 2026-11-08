package postgres_test

import (
	"context"
	"errors"
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

func TestNewRespectsCanceledStartupContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := postgres.New(ctx, postgres.Config{
		Host: "unreachable.invalid", Port: "5432", User: "test", Password: "test",
		Database: "test", SSLMode: "disable",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want canceled startup", err)
	}
}
