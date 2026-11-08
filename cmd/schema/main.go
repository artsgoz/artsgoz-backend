// Schema synchronization is an explicit administrative command.
// Run it once against the target database before deploying model changes.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/activity"
	"github.com/artsgoz/artsgoz-backend/internal/article"
	"github.com/artsgoz/artsgoz-backend/internal/comment"
	"github.com/artsgoz/artsgoz-backend/internal/community"
	"github.com/artsgoz/artsgoz-backend/internal/dashboard"
	"github.com/artsgoz/artsgoz-backend/internal/platform/config"
	"github.com/artsgoz/artsgoz-backend/internal/platform/postgres"
	"github.com/artsgoz/artsgoz-backend/internal/user"
)

func main() {
	if err := run(); err != nil {
		log.Printf("schema synchronization failed: %v", err)
		os.Exit(1)
	}
	log.Print("schema synchronization complete")
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	startupCtx, cancel := context.WithTimeout(ctx, cfg.StartupTimeout)
	defer cancel()
	database, err := postgres.New(startupCtx, postgres.Config{
		Host: cfg.DBHost, Port: cfg.DBPort, User: cfg.DBUser,
		Password: cfg.DBPassword, Database: cfg.DBName, SSLMode: cfg.DBSSLMode,
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()
	migrationCtx, cancelMigration := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelMigration()
	if err := database.DB.WithContext(migrationCtx).AutoMigrate(
		&user.User{},
		&article.Article{}, &article.Category{}, &article.ActivityLog{}, &article.Tag{},
		&community.FAQ{}, &community.Contact{}, &community.Review{},
		&comment.Comment{}, &activity.ActivityLog{}, &dashboard.TrafficRecord{},
	); err != nil {
		return fmt.Errorf("synchronize schema: %w", err)
	}
	return nil
}
