package postgres

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// New builds a pgx connection pool and verifies it with a Ping.
func New(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	dsn, err := ConnectionURL(cfg)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return pool, nil
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
