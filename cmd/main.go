package main

import (
	"context"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/option"

	"github.com/artsgoz/artsgoz-backend/config"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user"
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
	pool, err := pgxpool.New(ctx, config.DatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping PostgreSQL: %v", err)
	}
	log.Println("Connected to PostgreSQL successfully")

	// 5. Initialize Gin Engine
	app := gin.Default()

	// 5.1 CORS — Allow Frontend (Next.js) to access the API
	app.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:3001")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 6. Register User Module Routes under api prefix
	api := app.Group("/api/v1")
	user.Register(api, pool, firebaseAuth)

	// 7. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("Server starting on port %s...", port)
	if err := app.Run(":" + port); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
