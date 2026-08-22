package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	envGCPProject    = "GCP_PROJECT_ID"
	envSecretID      = "GCP_SECRET_ID"
	envSecretVersion = "GCP_SECRET_VERSION"
)

type Config struct {
	Port           string
	GinMode        string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	StartupTimeout time.Duration
}

// Load resolves split database configuration from the process environment, an
// optional local dotenv file, and finally Secret Manager when DB_* is incomplete.
func Load(ctx context.Context) (Config, error) {
	values := environment()
	if err := mergeDotenv(values, ".env.local"); err != nil {
		return Config{}, err
	}

	if !hasDatabaseParts(values) {
		if err := mergeSecret(ctx, values); err != nil {
			return Config{}, err
		}
	}

	cfg := Config{
		Port:           valueOr(values, "PORT", "3000"),
		GinMode:        valueOr(values, "GIN_MODE", "release"),
		DBHost:         values["DB_HOST"],
		DBPort:         values["DB_PORT"],
		DBUser:         values["DB_USER"],
		DBPassword:     values["DB_PASSWORD"],
		DBName:         values["DB_NAME"],
		DBSSLMode:      values["DB_SSL_MODE"],
		StartupTimeout: 10 * time.Second,
	}
	if !hasDatabaseParts(values) {
		return Config{}, fmt.Errorf("config: complete DB_* settings must be set")
	}
	if err := validatePort(cfg.Port); err != nil {
		return Config{}, err
	}
	if err := validateGinMode(cfg.GinMode); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func hasDatabaseParts(values map[string]string) bool {
	return values["DB_HOST"] != "" && values["DB_PORT"] != "" && values["DB_USER"] != "" &&
		values["DB_PASSWORD"] != "" && values["DB_NAME"] != ""
}

func validatePort(port string) error {
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return fmt.Errorf("config: PORT must be a number between 1 and 65535")
	}
	return nil
}

func validateGinMode(mode string) error {
	switch mode {
	case "debug", "release", "test":
		return nil
	default:
		return fmt.Errorf("config: GIN_MODE must be debug, release, or test")
	}
}

func mergeSecret(ctx context.Context, values map[string]string) error {
	projectID := values[envGCPProject]
	secretID := values[envSecretID]
	if projectID == "" || secretID == "" {
		return fmt.Errorf("config: complete DB_* settings or both %s and %s must be set", envGCPProject, envSecretID)
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
