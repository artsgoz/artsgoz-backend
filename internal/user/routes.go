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
	// 1. Personal profile settings for any authenticated user (e.g. saraniyakorn, student, editor, etc.)
	userGroup := router.Group("/user", requireAuth)
	{
		userGroup.GET("/me", adminHandler.GetProfile)
		userGroup.PUT("/me", adminHandler.UpdateProfile)
		userGroup.PUT("/password", adminHandler.ChangePassword)
		userGroup.POST("/avatar", adminHandler.UploadAvatar)
	}

	// 2. Admin access requires a stored admin role.
	adminGroup := router.Group("/admin", requireAuth, RequireRole(db, "admin"))
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
}
