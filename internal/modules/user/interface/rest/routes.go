package rest

import (
	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v3"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user/middleware"
)

func RegisterRoutes(
	router fiber.Router,
	authHandler *AuthHandler,
	adminHandler *AdminHandler,
	firebaseAuth *auth.Client,
	firestoreClient *firestore.Client,
) {
	// 1. Public routes
	router.Post("/register", authHandler.Register)
	router.Post("/login", authHandler.Login)

	// 2. Protected routes (requires login)
	protected := router.Group("", middleware.AuthMiddleware(firebaseAuth))
	protected.Get("/me", authHandler.Me)

	// 3. Admin-only routes (requires role = "admin")
	adminGroup := router.Group("/admin", middleware.AuthMiddleware(firebaseAuth), middleware.RequireRole(firestoreClient, "admin"))
	adminGroup.Get("/dashboard", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to admin dashboard",
			"role":    c.Locals("role"),
		})
	})
	adminGroup.Get("/users", adminHandler.GetAllUsers)
}
