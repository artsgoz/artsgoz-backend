package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"path/filepath"

	"github.com/artsgoz/artsgoz-backend/internal/activity"
	"github.com/artsgoz/artsgoz-backend/internal/article"
	"github.com/artsgoz/artsgoz-backend/internal/auth"
	"github.com/artsgoz/artsgoz-backend/internal/comment"
	"github.com/artsgoz/artsgoz-backend/internal/community"
	"github.com/artsgoz/artsgoz-backend/internal/dashboard"
	"github.com/artsgoz/artsgoz-backend/internal/media"
	"github.com/artsgoz/artsgoz-backend/internal/platform/config"
	"github.com/artsgoz/artsgoz-backend/internal/platform/postgres"
	"github.com/artsgoz/artsgoz-backend/internal/server"
	"github.com/artsgoz/artsgoz-backend/internal/user"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	firebaseClient, err := auth.InitFirebase(cfg.FirebaseCredentials, cfg.FirebaseProjectID)
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
		&article.Article{}, &article.Category{}, &article.ActivityLog{}, &article.Tag{},
		&community.FAQ{}, &community.Contact{}, &community.Review{},
		&comment.Comment{},
		&activity.ActivityLog{},
		&dashboard.TrafficRecord{},
	); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}

	engine := server.New(cfg.GinMode, cfg.CORSAllowedOrigins...)
	if err := os.MkdirAll(filepath.Join(cfg.UploadDir, "avatars"), 0755); err != nil {
		return err
	}
	engine.Static("/uploads", cfg.UploadDir)

	RegisterRoutes(engine, database.DB, firebaseClient, cfg)

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

// RegisterRoutes is shared by both entrypoints and can be checked without starting a database.
func RegisterRoutes(engine *gin.Engine, db *gorm.DB, firebaseClient auth.FirebaseClient, cfg config.Config) {
	authService := auth.NewService(user.NewGormRepository(db), firebaseClient)
	requireAuth := auth.Middleware(authService)
	activityService := activity.NewService(activity.NewGormRepository(db))
	activityHandler := activity.NewHandler(activityService)
	commentRepository := comment.NewGormRepository(db)
	commentHandler := comment.NewHandler(comment.NewService(commentRepository, activityService))
	dashboardHandler := dashboard.NewHandler(dashboard.NewService(dashboard.NewGormRepository(db), commentRepository, activityService))
	articleHandler := article.NewHandler(article.NewService(article.NewGormRepository(db)))
	var storage media.StorageProvider = media.NewLocalStorageProvider(cfg.UploadDir, cfg.UploadBaseURL)
	if cfg.StorageBucket != "" {
		storage = media.NewCloudStorageProvider(cfg.StorageBucket, "", storage)
	}
	mediaHandler := media.NewHandler(media.NewService(storage))
	for _, prefix := range []string{"/api", "/api/v1"} {
		api := engine.Group(prefix)
		auth.Register(api, authService)
		user.Register(api, db, requireAuth)
		article.RegisterRoutes(api, articleHandler, requireAuth, server.RequireRoles(server.Role("writer"), server.RoleEditor, server.RoleSuperAdmin, server.Role("admin")))
		media.RegisterRoutes(api.Group("", requireAuth), mediaHandler)
		activity.RegisterRoutes(api.Group("", requireAuth), activityHandler)
		comment.RegisterRoutes(api, commentHandler)
		dashboard.RegisterRoutes(api, dashboardHandler, false)
		admin := api.Group("/admin", requireAuth, server.RequireRoles(server.RoleEditor, server.RoleSuperAdmin, server.Role("admin")))
		dashboard.RegisterAdminRoutes(admin, dashboardHandler)
		community.Register(api, db, requireAuth)
	}
}
