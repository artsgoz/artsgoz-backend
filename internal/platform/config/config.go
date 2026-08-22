package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	envGCPProject    = "GCP_PROJECT_ID"
	envSecretID      = "GCP_SECRET_ID"
	envSecretVersion = "GCP_SECRET_VERSION"
)

type Config struct {
	HTTPAddress    string
	DatabaseURL    string
	LogLevel       string
	StartupTimeout time.Duration
}

// Load resolves configuration from the process environment, an optional local
// dotenv file, and finally Secret Manager when DATABASE_URL is not already set.
func Load(ctx context.Context) (Config, error) {
	values := environment()
	if err := mergeDotenv(values, ".env.local"); err != nil {
		return Config{}, err
	}

	if values["DATABASE_URL"] == "" {
		if err := mergeSecret(ctx, values); err != nil {
			return Config{}, err
		}
	}

	cfg := Config{
		HTTPAddress:    valueOr(values, "HTTP_ADDRESS", ":3000"),
		DatabaseURL:    values["DATABASE_URL"],
		LogLevel:       valueOr(values, "LOG_LEVEL", "info"),
		StartupTimeout: 10 * time.Second,
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL must be set")
	}
	return cfg, nil
}

func mergeSecret(ctx context.Context, values map[string]string) error {
	projectID := values[envGCPProject]
	secretID := values[envSecretID]
	if projectID == "" || secretID == "" {
		return fmt.Errorf("config: DATABASE_URL or both %s and %s must be set", envGCPProject, envSecretID)
	}
	version := valueOr(values, envSecretVersion, "latest")
	payload, err := accessSecretVersion(ctx, projectID, secretID, version)
	if err != nil {
		return err
	}
	secretValues, err := godotenv.Unmarshal(payload)
	if err != nil {
		return fmt.Errorf("config: parse secret payload: %w", err)
	}
	mergeMissing(values, secretValues)
	return nil
}

func mergeDotenv(values map[string]string, path string) error {
	parsed, err := godotenv.Read(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("config: read %s: %w", path, err)
	}
	mergeMissing(values, parsed)
	return nil
}

func environment() map[string]string {
	values := make(map[string]string)
	for _, entry := range os.Environ() {
		for i := 0; i < len(entry); i++ {
			if entry[i] == '=' {
				values[entry[:i]] = entry[i+1:]
				break
			}
		}
	}
	return values
}

func mergeMissing(destination, source map[string]string) {
	for key, value := range source {
		if destination[key] == "" {
			destination[key] = value
		}
	}
}

func valueOr(values map[string]string, key, fallback string) string {
	if values[key] != "" {
		return values[key]
	}
	return fallback
}
