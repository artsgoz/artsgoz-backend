package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/artsgoz/artsgoz-backend/config"
	"github.com/artsgoz/artsgoz-backend/internal/auth"
	"github.com/artsgoz/artsgoz-backend/internal/user"
)

func main() {
	// 1. Load local configuration (if present)
	if err := config.Load(); err != nil {
		log.Printf("Warning: failed to load local configuration: %v. Will rely on environment variables.", err)
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
	if err := db.AutoMigrate(&user.User{}); err != nil {
		log.Printf("Warning: AutoMigrate failed: %v", err)
	} else {
		log.Println("Database AutoMigrate executed successfully")
	}

	// 4. Initialize Gin Engine
	app := gin.Default()

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
	api := app.Group("/api/v1")
	userRepo := user.NewUserRepository(db)
	jwtManager := auth.NewJWTManager(config.JWTSecret(), 24*time.Hour)
	authService := auth.NewService(userRepo, jwtManager)

	auth.Register(api, authService)
	user.Register(api, db, auth.Middleware(authService))

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
