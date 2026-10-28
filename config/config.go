package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

func Load() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("config: load env: %w", err)
	}
	return nil
}

// DatabaseURL builds the PostgreSQL connection URL from environment variables.
func DatabaseURL() string {
	dbName := getEnvOrDefault("DB_NAME", "artsgoz")
	u := url.URL{
		Scheme: "postgres",
		User: url.UserPassword(
			getEnvOrDefault("DB_USER", "user"),
			getEnvOrDefault("DB_PASSWORD", "password"),
		),
		Host: net.JoinHostPort(
			getEnvOrDefault("DB_HOST", "localhost"),
			getEnvOrDefault("DB_PORT", "5432"),
		),
		Path:    "/" + dbName,
		RawPath: "/" + url.PathEscape(dbName),
	}
	query := url.Values{"sslmode": {getEnvOrDefault("DB_SSL_MODE", "disable")}}
	u.RawQuery = query.Encode()
	return u.String()
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
