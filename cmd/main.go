package main

import (
	"context"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/artsgoz/artsgoz-backend/config"
	"github.com/artsgoz/artsgoz-backend/internal/article"
	"github.com/artsgoz/artsgoz-backend/internal/media"
	"github.com/artsgoz/artsgoz-backend/internal/user"
)

func main() {
	ctx := context.Background()

	// 1. Load local configuration (if present)
	if err := config.Load(); err != nil {
		log.Printf("Warning: failed to load local configuration: %v. Will rely on environment variables.", err)
	}

	// 2. Connect to Firebase (Auth only — user data is now stored in PostgreSQL)
	var opt option.ClientOption
	if _, err := os.Stat("firebase-service-account.json"); err == nil {
		opt = option.WithCredentialsFile("firebase-service-account.json")
	}

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = "account-test-c25f0"
	}

	firebaseConfig := &firebase.Config{
		ProjectID: projectID,
	}

	var firebaseApp *firebase.App
	var err error
	if opt != nil {
		firebaseApp, err = firebase.NewApp(ctx, firebaseConfig, opt)
	} else {
		firebaseApp, err = firebase.NewApp(ctx, firebaseConfig)
	}
	if err != nil {
		log.Fatalf("Failed to init Firebase: %v", err)
	}

	// 3. Firebase Auth Client
	firebaseAuth, err := firebaseApp.Auth(ctx)
	if err != nil {
		log.Fatalf("Failed to init Firebase Auth: %v", err)
	}

	// 4. Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(config.DatabaseURL()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get PostgreSQL connection pool: %v", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	defer sqlDB.Close()
	log.Println("Connected to PostgreSQL successfully")

	// 5. Auto Migrate Database Tables
	if err := db.AutoMigrate(
		&user.User{},
		&article.Article{},
		&article.Category{},
		&article.ActivityLog{},
	); err != nil {
		log.Printf("Warning: AutoMigrate: %v", err)
	} else {
		log.Println("Database AutoMigrate executed successfully")
	}

	// 6. Initialize Gin Engine
	app := gin.Default()

	// Upload directory & static file serving
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	_ = os.MkdirAll(uploadDir, 0755)
	app.Static("/uploads", uploadDir)

	// 6.1 CORS — Allow Frontend to access the API
	app.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 7. Register Module Routes under api and api/v1 prefixes
	api := app.Group("/api")
	apiV1 := app.Group("/api/v1")

	// User module
	user.Register(apiV1, db, firebaseAuth)

	// Article module
	articleRepo := article.NewGormRepository(db)
	articleSvc := article.NewService(articleRepo)
	articleHandler := article.NewHandler(articleSvc)
	article.RegisterRoutes(api, articleHandler)
	article.RegisterRoutes(apiV1, articleHandler)

	// Media module
	uploadBaseURL := os.Getenv("UPLOAD_BASE_URL")
	if uploadBaseURL == "" {
		uploadBaseURL = "/uploads"
	}
	var storageProvider media.StorageProvider = media.NewLocalStorageProvider(uploadDir, uploadBaseURL)
	if bucket := os.Getenv("STORAGE_BUCKET"); bucket != "" {
		storageProvider = media.NewCloudStorageProvider(bucket, "", storageProvider)
	}
	mediaSvc := media.NewService(storageProvider)
	mediaHandler := media.NewHandler(mediaSvc)
	media.RegisterRoutes(api, mediaHandler)
	media.RegisterRoutes(apiV1, mediaHandler)

	// 8. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("Server starting on port %s...", port)
	if err := app.Run(":" + port); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
