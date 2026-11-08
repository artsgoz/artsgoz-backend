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

func DatabaseURL() string {
	dbName := os.Getenv("DB_NAME")
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")),
		Host:     net.JoinHostPort(os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		Path:     "/" + dbName,
		RawPath:  "/" + url.PathEscape(dbName),
		RawQuery: url.Values{"sslmode": {os.Getenv("DB_SSL_MODE")}}.Encode(),
	}
	return u.String()
}

func FirebaseCredentialsPath() string {
	path := os.Getenv("FIREBASE_CREDENTIALS")
	if path == "" {
		return "firebase-service-account.json"
	}
	return path
}

func FirebaseProjectID() string {
	if project := os.Getenv("FIREBASE_PROJECT_ID"); project != "" { return project }
    return os.Getenv("GCP_PROJECT_ID")
}
