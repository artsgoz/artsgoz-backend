package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Register(router *gin.RouterGroup, db *gorm.DB, requireAuth gin.HandlerFunc) {
	repository := NewUserRepository(db)
	handler := NewAdminHandler(NewAdminService(repository))
	RegisterRoutes(router, handler, requireAuth, db)
}

func RegisterRoutes(router *gin.RouterGroup, adminHandler *AdminHandler, requireAuth gin.HandlerFunc, db *gorm.DB) {
	// Admin access requires a stored admin role.
	adminGroup := router.Group("/admin", requireAuth, RequireRole(db, "admin"))
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
