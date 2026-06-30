package main

import (
	"context"
	"log"
	"os"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"google.golang.org/api/option"

	"github.com/artsgoz/artsgoz-backend/config"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user"
	"github.com/artsgoz/artsgoz-backend/internal/platform/httpserver"
)

func main() {
	ctx := context.Background()

	// 1. Load configuration secrets from GCP Secret Manager (if configured)
	if err := config.LoadFromSecretManager(ctx); err != nil {
		log.Printf("Warning: failed to load secrets from secret manager: %v. Will rely on local environment.", err)
	}

	// 2. Connect to Firebase
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

	// 4. Firestore Client
	firestoreClient, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("Failed to init Firestore: %v", err)
	}
	defer firestoreClient.Close()

	// 5. Initialize Fiber App
	app := httpserver.New()

	// 5.1 CORS — Allow Frontend (Next.js) to access the API
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3001"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
	}))

	// 5.2 Rate Limiter — Protect against brute force
	app.Use(limiter.New(limiter.Config{
		Max:        20,
		Expiration: 60 * time.Second,
	}))

	// 6. Register User Module Routes under api prefix
	api := app.Group("/api/v1")
	user.Register(api, firestoreClient, firebaseAuth)

	// 7. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("Server starting on port %s...", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
