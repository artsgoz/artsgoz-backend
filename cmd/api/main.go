package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
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

	databaseConfig := postgres.Config{
		Host: cfg.DBHost, Port: cfg.DBPort, User: cfg.DBUser,
		Password: cfg.DBPassword, Database: cfg.DBName, SSLMode: cfg.DBSSLMode,
	}
	database, err := postgres.New(startupCtx, databaseConfig)
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("close database", "error", err)
		}
	}()

	engine := server.New(logger, cfg.GinMode)
	api := engine.Group("/api/v1")
	userRepository := user.NewGormRepository(database.DB)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)
	user.RegisterRoutes(api, userHandler)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}
	listenErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "address", httpServer.Addr)
		listenErr <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-listenErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		logger.Info("shutting down http server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
		err := <-listenErr
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
