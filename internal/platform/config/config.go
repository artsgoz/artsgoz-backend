package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	GinMode             string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBSSLMode           string
	StartupTimeout      time.Duration
	UploadDir           string
	UploadBaseURL       string
	StorageBucket       string
	FirebaseCredentials string
	FirebaseProjectID   string
	CORSAllowedOrigins  []string
}

// Load reads an optional .env.local file and then resolves configuration from
// the process environment. Existing environment variables take precedence.
func Load() (Config, error) {
	for _, filename := range []string{".env.local", ".env"} {
		if err := godotenv.Load(filename); err != nil && !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("config: cannot load %s; check file syntax and permissions", filename)
		}
	}

	cfg := Config{
		Port:                envOrDefault("PORT", "3000"),
		GinMode:             envOrDefault("GIN_MODE", "release"),
		DBHost:              os.Getenv("DB_HOST"),
		DBPort:              envOrDefault("DB_PORT", "5432"),
		DBUser:              os.Getenv("DB_USER"),
		DBPassword:          os.Getenv("DB_PASSWORD"),
		DBName:              os.Getenv("DB_NAME"),
		DBSSLMode:           envOrDefault("DB_SSL_MODE", "disable"),
		StartupTimeout:      10 * time.Second,
		UploadDir:           envOrDefault("UPLOAD_DIR", "./uploads"),
		UploadBaseURL:       envOrDefault("UPLOAD_BASE_URL", "/uploads"),
		StorageBucket:       os.Getenv("STORAGE_BUCKET"),
		FirebaseCredentials: envOrDefault("FIREBASE_CREDENTIALS", "firebase-service-account.json"),
		FirebaseProjectID:   envOrDefault("FIREBASE_PROJECT_ID", os.Getenv("GCP_PROJECT_ID")),
	}
	origins, err := parseCORSOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if err != nil {
		return Config{}, err
	}
	cfg.CORSAllowedOrigins = origins
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

func parseCORSOrigins(raw string) ([]string, error) {
	origins := []string{}
	seen := make(map[string]bool)
	for _, value := range strings.Split(raw, ",") {
		origin := strings.TrimSpace(value)
		if origin == "" {
			continue
		}
		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" ||
			strings.Contains(parsed.Host, "*") || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
			return nil, fmt.Errorf("config: CORS_ALLOWED_ORIGINS must contain comma-separated http/https origins without paths or wildcards")
		}
		origin = parsed.Scheme + "://" + strings.ToLower(parsed.Host)
		if !seen[origin] {
			origins = append(origins, origin)
			seen[origin] = true
		}
	}
	return origins, nil
}
