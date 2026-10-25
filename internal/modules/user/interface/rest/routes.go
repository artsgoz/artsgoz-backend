package rest

import (
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/middleware"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	authHandler *AuthHandler,
	adminHandler *AdminHandler,
	firebaseAuth *auth.Client,
	pool *pgxpool.Pool,
) {
	// 1. Auth routes
	authGroup := router.Group("/auth")
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/logout", authHandler.Logout)

	// Keep existing root-level auth endpoints for backward compatibility
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	// 2. Protected routes (requires login)
	protected := router.Group("", middleware.AuthMiddleware(firebaseAuth))
	protected.GET("/me", authHandler.Me)

	// 3. Admin-only routes (requires role = "admin")
	adminGroup := router.Group("/admin", middleware.AuthMiddleware(firebaseAuth), middleware.RequireRole(pool, "admin"))
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
