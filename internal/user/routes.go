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
	// Admin backoffice group for club members (role restriction removed: backoffice is for all club members)
	adminGroup := router.Group("/admin", requireAuth)
	adminGroup.GET("/dashboard", func(c *gin.Context) {
		role, _ := c.Get("role")
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to admin dashboard",
			"role":    role,
		})
	})
	adminGroup.GET("/users", adminHandler.GetAllUsers)
	adminGroup.PUT("/users/:id/role", adminHandler.UpdateUserRole)
	adminGroup.GET("/me", adminHandler.GetProfile)
	adminGroup.PUT("/me", adminHandler.UpdateProfile)
	adminGroup.PUT("/me/password", adminHandler.ChangePassword)

	// Direct user profile aliases
	router.PUT("/me", requireAuth, adminHandler.UpdateProfile)
	router.PUT("/me/password", requireAuth, adminHandler.ChangePassword)
}
