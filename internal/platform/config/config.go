package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
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

// Load reads an optional .env.local file and then resolves configuration from
// the process environment. Existing environment variables take precedence.
func Load() (Config, error) {
	if err := godotenv.Load(".env.local"); err != nil {
		if !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("config: load .env.local: %w", err)
		}
	}

	cfg := Config{
		Port:           envOrDefault("PORT", "3000"),
		GinMode:        envOrDefault("GIN_MODE", "release"),
		DBHost:         os.Getenv("DB_HOST"),
		DBPort:         envOrDefault("DB_PORT", "5432"),
		DBUser:         os.Getenv("DB_USER"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         os.Getenv("DB_NAME"),
		DBSSLMode:      envOrDefault("DB_SSL_MODE", "disable"),
		StartupTimeout: 10 * time.Second,
	}
	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" {
		return Config{}, fmt.Errorf("config: DB_HOST, DB_USER, and DB_NAME are required")
	}
	if err := validatePort(cfg.Port); err != nil {
		return Config{}, err
	}
	if err := validateGinMode(cfg.GinMode); err != nil {
		return Config{}, err
	}
	return cfg, nil
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

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
