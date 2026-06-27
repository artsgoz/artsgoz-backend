package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user"

	firebase "firebase.google.com/go/v4"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	// 1. เชื่อมต่อ Firebase
	opt := option.WithCredentialsFile("firebase-service-account.json")
	firebaseApp, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: "account-test-c25f0",
	}, opt)
	if err != nil {
		log.Fatal("Failed to init Firebase:", err)
	}

	// 2. Firebase Auth
	firebaseAuth, err := firebaseApp.Auth(ctx)
	if err != nil {
		log.Fatal("Failed to init Firebase Auth:", err)
	}

	// 3. Firestore
	firestoreClient, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatal("Failed to init Firestore:", err)
	}
	defer firestoreClient.Close()

	// 4. Fiber App
	app := fiber.New()

	// 4.1 CORS — อนุญาต Frontend (Next.js) เข้าถึง API
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3001"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
	}))

	// 4.2 Rate Limiter — ป้องกัน brute force
	app.Use(limiter.New(limiter.Config{
		Max:        20,
		Expiration: 60 * time.Second,
	}))

	// 5. Register User module routes
	user.Register(app, firestoreClient, firebaseAuth)

	// 6. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(":" + port))
}
