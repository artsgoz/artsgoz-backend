package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/artsgoz/artsgoz-backend/config"
	"github.com/artsgoz/artsgoz-backend/internal/article"
	"github.com/artsgoz/artsgoz-backend/internal/auth"
	"github.com/artsgoz/artsgoz-backend/internal/media"
	"github.com/artsgoz/artsgoz-backend/internal/user"
)

func main() {
	// 1. Load local configuration (if present)
	if err := config.Load(); err != nil {
		log.Printf("Warning: failed to load local configuration: %v. Will rely on environment variables.", err)
	}

	// Firebase is required for email/password and Google sign-in.
	firebaseClient, err := auth.InitFirebase(config.FirebaseCredentialsPath(), config.FirebaseProjectID())
	if err != nil {
		log.Fatalf("Failed to initialize Firebase authentication: %v", err)
	}

	// 2. Connect to PostgreSQL
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

	// 3. Auto Migrate Database Tables
	if err := db.AutoMigrate(&user.User{}, &article.Article{}, &article.Category{}, &article.ActivityLog{}); err != nil {
		log.Printf("Warning: AutoMigrate failed: %v", err)
	} else {
		log.Println("Database AutoMigrate executed successfully")
	}

	// 4. Initialize Gin Engine
	app := gin.Default()

	// Ensure uploads directory exists and mount static route
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	_ = os.MkdirAll(filepath.Join(uploadDir, "avatars"), 0755)
	app.Static("/uploads", uploadDir)

	// 4.1 Allow frontend access through CORS
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

	// 5. Wire authentication and user modules under the API prefix
	api := app.Group("/api")
	apiV1 := app.Group("/api/v1")
	userRepo := user.NewUserRepository(db)

	authService := auth.NewService(userRepo, firebaseClient)

	auth.Register(api, authService)
	user.Register(api, db, auth.Middleware(authService))
	auth.Register(apiV1, authService)
	user.Register(apiV1, db, auth.Middleware(authService))

	articleHandler := article.NewHandler(article.NewService(article.NewGormRepository(db)))
	article.RegisterRoutes(api, articleHandler)
	article.RegisterRoutes(apiV1, articleHandler)

	uploadBaseURL := os.Getenv("UPLOAD_BASE_URL")
	if uploadBaseURL == "" {
		uploadBaseURL = "/uploads"
	}
	var storageProvider media.StorageProvider = media.NewLocalStorageProvider(uploadDir, uploadBaseURL)
	if bucket := os.Getenv("STORAGE_BUCKET"); bucket != "" {
		storageProvider = media.NewCloudStorageProvider(bucket, "", storageProvider)
	}
	mediaHandler := media.NewHandler(media.NewService(storageProvider))
	media.RegisterRoutes(api, mediaHandler)
	media.RegisterRoutes(apiV1, mediaHandler)

	// 6. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("Server is running on port %s", port)
	if err := app.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
