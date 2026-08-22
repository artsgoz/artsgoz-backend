package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"time"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

type Client struct {
	DB    *gorm.DB
	sqlDB *sql.DB
}

// New opens a GORM-backed PostgreSQL connection pool and verifies it.
func New(ctx context.Context, cfg Config) (*Client, error) {
	dsn, err := ConnectionURL(cfg)
	if err != nil {
		return nil, err
	}
	database, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("postgres: connect: %w", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("postgres: access connection pool: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return &Client{DB: database, sqlDB: sqlDB}, nil
}

func (c *Client) Close() error {
	return c.sqlDB.Close()
}

// ConnectionURL safely constructs the driver URL from split database settings.
func ConnectionURL(cfg Config) (string, error) {
	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" || cfg.Password == "" || cfg.Database == "" {
		return "", fmt.Errorf("postgres: incomplete configuration")
	}
	dsn := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, cfg.Port),
		Path:   "/" + cfg.Database,
	}
	if cfg.SSLMode != "" {
		query := dsn.Query()
		query.Set("sslmode", cfg.SSLMode)
		dsn.RawQuery = query.Encode()
	}
	return dsn.String(), nil
}
