package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/platform/config"
	"github.com/artsgoz/artsgoz-backend/internal/platform/logging"
	"github.com/artsgoz/artsgoz-backend/internal/platform/postgres"
	"github.com/artsgoz/artsgoz-backend/internal/server"
	"github.com/artsgoz/artsgoz-backend/internal/user"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(ctx)
	if err != nil {
		return err
	}

	logger := logging.New(cfg.LogLevel)
	slog.SetDefault(logger)

	startupCtx, cancel := context.WithTimeout(ctx, cfg.StartupTimeout)
	defer cancel()

	pool, err := postgres.New(startupCtx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	app := server.New(logger)
	api := app.Group("/api/v1")
	user.New(pool).RegisterRoutes(api)

	listenErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "address", cfg.HTTPAddress)
		listenErr <- app.Listen(cfg.HTTPAddress)
	}()

	select {
	case err := <-listenErr:
		return err
	case <-ctx.Done():
		logger.Info("shutting down http server")
		if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
			return err
		}
		err := <-listenErr
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	}
}
