package user

import (
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Register(router *gin.RouterGroup, db *gorm.DB, firebaseAuth *auth.Client) {
	// Initialize Repository (PostgreSQL)
	userRepo := NewUserRepository(db)

	// Initialize Services
	loginService := NewLoginService(userRepo, firebaseAuth)
	registerService := NewRegisterService(userRepo, firebaseAuth)
	adminService := NewAdminService(userRepo)

	// Initialize Handlers
	authHandler := NewAuthHandler(loginService, registerService)
	adminHandler := NewAdminHandler(adminService)

	// Register Routes
	RegisterRoutes(router, authHandler, adminHandler, firebaseAuth, db)
}

func RegisterRoutes(
	router *gin.RouterGroup,
	authHandler *AuthHandler,
	adminHandler *AdminHandler,
	firebaseAuth FirebaseAuthClient,
	db *gorm.DB,
) {
	// 1. Auth routes
	authGroup := router.Group("/auth")
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/logout", AuthMiddleware(firebaseAuth), authHandler.Logout)

	// Keep existing root-level auth endpoints for backward compatibility
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	// 2. Protected routes (requires login)
	protected := router.Group("", AuthMiddleware(firebaseAuth))
	protected.GET("/me", authHandler.Me)

	// 3. Admin-only routes (requires role = "admin")
	adminGroup := router.Group("/admin", AuthMiddleware(firebaseAuth), RequireRole(db, "admin"))
	adminGroup.GET("/dashboard", func(c *gin.Context) {
		role, _ := c.Get("role")
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to admin dashboard",
			"role":    role,
		})
	})
	adminGroup.GET("/users", adminHandler.GetAllUsers)
	adminGroup.PUT("/users/:uid/role", adminHandler.UpdateUserRole)
	adminGroup.GET("/me", adminHandler.GetProfile)
	adminGroup.PUT("/me", adminHandler.UpdateProfile)
}
