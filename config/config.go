package config

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/artsgoz/artsgoz-backend/pkg/utils"
)

// Bootstrap environment variable names used to locate the secret blob in GCP.
const (
	envGCPProject    = "GCP_PROJECT_ID"
	envSecretID      = "GCP_SECRET_ID"
	envSecretVersion = "GCP_SECRET_VERSION"
)

// LoadFromSecretManager loads .env.local (if present) for the GCP identifiers,
// then fetches a .env-formatted secret payload from GCP Secret Manager and sets
// each key/value pair into the process environment. Existing env vars are left
// untouched so shell overrides and prod-injected env keep working.
//
// Local-mode shortcut: if DATABASE_URL is already set after loading .env.local
// (i.e. the developer put it there directly), GCP Secret Manager is skipped
// entirely — no GCP credentials required.
func LoadFromSecretManager(ctx context.Context) error {
	if err := loadDotenvIfPresent(".env.local"); err != nil {
		return err
	}

	// Local-mode: skip GCP if DATABASE_URL is already available.
	if os.Getenv("DATABASE_URL") != "" {
		return nil
	}

	projectID := os.Getenv(envGCPProject)
	secretID := os.Getenv(envSecretID)
	if projectID == "" || secretID == "" {
		return fmt.Errorf("config: %s and %s must be set (or set DATABASE_URL directly for local dev)", envGCPProject, envSecretID)
	}

	version := os.Getenv(envSecretVersion)
	if version == "" {
		version = "latest"
	}

	payload, err := utils.AccessSecretVersion(ctx, projectID, secretID, version)
	if err != nil {
		return err
	}

	parsed, err := godotenv.Unmarshal(payload)
	if err != nil {
		return fmt.Errorf("config: parse secret payload: %w", err)
	}

	return setEnvFromMap(parsed)
}

// loadDotenvIfPresent reads a dotenv file and merges its values into the
// process environment, skipping keys already set. A missing file is not an
// error so prod (where vars are injected) works without a file present.
func loadDotenvIfPresent(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("config: stat %s: %w", path, err)
	}
	parsed, err := godotenv.Read(path)
	if err != nil {
		return fmt.Errorf("config: read %s: %w", path, err)
	}
	return setEnvFromMap(parsed)
}

func setEnvFromMap(m map[string]string) error {
	for k, v := range m {
		if _, exists := os.LookupEnv(k); exists {
			continue
		}
		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("config: setenv %s: %w", k, err)
		}
	}
	return nil
}
