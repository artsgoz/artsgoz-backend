package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Load reads .env.local if present. Existing environment variables take precedence.
func Load() error {
	return loadDotenvIfPresent(".env.local")
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
