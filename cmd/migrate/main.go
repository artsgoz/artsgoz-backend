// Migration runner for artsgoz-backend.
//
// Usage:
//
//	go run ./cmd/migrate up                # apply all pending migrations
//	go run ./cmd/migrate down              # roll back the most recent migration
//	go run ./cmd/migrate version           # print current schema version
//	go run ./cmd/migrate force <version>   # mark a version as applied (recovery only)
//	go run ./cmd/migrate goto <version>    # migrate up or down to a specific version
//
// Database settings are read through the same configuration path as the main
// app, so migrations use the same environment and Secret Manager behavior.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/artsgoz/artsgoz-backend/internal/platform/config"
	platformpostgres "github.com/artsgoz/artsgoz-backend/internal/platform/postgres"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]

	ctx := context.Background()
	cfg, err := config.Load(ctx)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	dsn, err := platformpostgres.ConnectionURL(platformpostgres.Config{
		Host: cfg.DBHost, Port: cfg.DBPort, User: cfg.DBUser,
		Password: cfg.DBPassword, Database: cfg.DBName, SSLMode: cfg.DBSSLMode,
	})
	if err != nil {
		log.Fatal(err)
	}
	driverDSN, err := migrationURL(dsn)
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.New("file://migrations", driverDSN)
	if err != nil {
		log.Fatalf("init migrate: %v", err)
	}
	defer func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			log.Printf("migrate close: source=%v db=%v", srcErr, dbErr)
		}
	}()

	switch cmd {
	case "up":
		runAndReport("up", m.Up())
	case "down":
		runAndReport("down 1", m.Steps(-1))
	case "version":
		printVersion(m)
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("force requires a version: go run ./cmd/migrate force <version>")
		}
		v, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("invalid version: %v", err)
		}
		runAndReport(fmt.Sprintf("force %d", v), m.Force(v))
	case "goto":
		if len(os.Args) < 3 {
			log.Fatal("goto requires a version: go run ./cmd/migrate goto <version>")
		}
		v, err := strconv.ParseUint(os.Args[2], 10, 64)
		if err != nil {
			log.Fatalf("invalid version: %v", err)
		}
		runAndReport(fmt.Sprintf("goto %d", v), m.Migrate(uint(v)))
	default:
		usage()
		os.Exit(2)
	}
}

func migrationURL(dsn string) (string, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("parse database DSN: %w", err)
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return "", fmt.Errorf("database DSN must use postgres or postgresql scheme")
	}
	parsed.Scheme = "pgx5"
	return parsed.String(), nil
}

func runAndReport(action string, err error) {
	if err == nil {
		log.Printf("migrate %s: success", action)
		return
	}
	if errors.Is(err, migrate.ErrNoChange) {
		log.Printf("migrate %s: no change", action)
		return
	}
	log.Fatalf("migrate %s: %v", action, err)
}

func printVersion(m *migrate.Migrate) {
	v, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("no migrations applied yet")
			return
		}
		log.Fatalf("version: %v", err)
	}
	if dirty {
		fmt.Printf("version: %d (DIRTY — last migration failed mid-way; use `force` after manual cleanup)\n", v)
		os.Exit(1)
	}
	fmt.Printf("version: %d\n", v)
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: go run ./cmd/migrate <command> [args]

commands:
  up                 apply all pending migrations
  down               roll back the most recent migration
  version            print current schema version
  goto <version>     migrate up or down to a specific version
  force <version>    mark a version as applied (recovery only)`)
}
