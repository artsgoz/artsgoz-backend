package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/activity"
	"github.com/artsgoz/artsgoz-backend/internal/comment"
	"github.com/artsgoz/artsgoz-backend/internal/dashboard"
	"github.com/artsgoz/artsgoz-backend/internal/platform/config"
	"github.com/artsgoz/artsgoz-backend/internal/platform/postgres"
	"github.com/artsgoz/artsgoz-backend/internal/server"
	"github.com/artsgoz/artsgoz-backend/internal/user"
)

func main() {
	if err := run(); err != nil {
		log.Printf("application stopped: %v", err)
		os.Exit(1)
	}
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
			log.Printf("close database: %v", err)
		}
	}()

	// Auto-migrate tables
	if err := database.DB.AutoMigrate(
		&user.User{},
		&comment.Comment{},
		&activity.ActivityLog{},
		&dashboard.TrafficRecord{},
	); err != nil {
		log.Printf("warning: auto-migrate: %v", err)
	}

	engine := server.New(cfg.GinMode)

	// API groups: register on both /api and /api/v1
	api := engine.Group("/api")
	apiV1 := engine.Group("/api/v1")

	// User module
	userRepository := user.NewGormRepository(database.DB)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)
	user.RegisterRoutes(api, userHandler)
	user.RegisterRoutes(apiV1, userHandler)

	// Activity module
	activityRepository := activity.NewGormRepository(database.DB)
	activityService := activity.NewService(activityRepository)
	activityHandler := activity.NewHandler(activityService)
	activity.RegisterRoutes(api, activityHandler)
	activity.RegisterRoutes(apiV1, activityHandler)

	// Comment module
	commentRepository := comment.NewGormRepository(database.DB)
	commentService := comment.NewService(commentRepository, activityService)
	commentHandler := comment.NewHandler(commentService)
	comment.RegisterRoutes(api, commentHandler)
	comment.RegisterRoutes(apiV1, commentHandler)

	// Dashboard module
	dashboardRepository := dashboard.NewGormRepository(database.DB)
	dashboardService := dashboard.NewService(dashboardRepository, commentRepository, activityService)
	dashboardHandler := dashboard.NewHandler(dashboardService)

	// Legacy dashboard routes (backward compatibility)
	dashboard.RegisterRoutes(api, dashboardHandler)
	dashboard.RegisterRoutes(apiV1, dashboardHandler)

	// Admin dashboard routes with authentication (spec-compliant)
	// [GET] /api/v1/admin/dashboard/stats
	// [GET] /api/v1/admin/dashboard/monthly-trends?year=2026
	// [GET] /api/v1/admin/dashboard/top-articles?limit=5
	// [GET] /api/v1/admin/dashboard/top-categories
	// [GET] /api/v1/admin/dashboard/recent-activities
	adminV1 := apiV1.Group("/admin",
		server.AuthRequired(),
		server.RequireRoles(server.RoleEditor, server.RoleSuperAdmin),
	)
	dashboard.RegisterAdminRoutes(adminV1, dashboardHandler)

	// Also register on /api/admin for consistency
	adminAPI := api.Group("/admin",
		server.AuthRequired(),
		server.RequireRoles(server.RoleEditor, server.RoleSuperAdmin),
	)
	dashboard.RegisterAdminRoutes(adminAPI, dashboardHandler)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}
	listenErr := make(chan error, 1)
	go func() {
		log.Printf("http server listening on %s", httpServer.Addr)
		listenErr <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-listenErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		log.Print("shutting down http server")
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
