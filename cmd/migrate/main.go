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
// DATABASE_URL is read from the same Secret Manager blob as the main app, so
// running migrations uses the exact same auth path as running the server.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/artsgoz/artsgoz-backend/config"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]

	ctx := context.Background()
	if err := config.LoadFromSecretManager(ctx); err != nil {
		log.Fatalf("load secrets: %v", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	// golang-migrate's pgx driver expects a "pgx5://" scheme.
	driverDSN := "pgx5" + dsn[len("postgres"):]

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
