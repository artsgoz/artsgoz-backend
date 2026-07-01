package main

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"

	"github.com/artsgoz/artsgoz-backend/config"
	"github.com/artsgoz/artsgoz-backend/internal/modules/contact"
	contactrest "github.com/artsgoz/artsgoz-backend/internal/modules/contact/interface/rest"
	"github.com/artsgoz/artsgoz-backend/internal/platform/httpserver"
)

func main() {
	ctx := context.Background()

	if err := config.LoadFromSecretManager(ctx); err != nil {
		log.Fatalf("load secrets: %v", err)
	}

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = "chula-artsgoz-website"
	}

	// 1. Initialize Firestore Client
	fsClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore client init: %v", err)
	}
	defer fsClient.Close()

	// 2. Initialize Modules
	contactMod := contact.New(fsClient)

	// 3. Register HTTP Routes
	app := httpserver.New()
	api := app.Group("/api/v1")
	contactrest.RegisterRoutes(api, contactMod.Handler)

	// 4. Listen on PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
