package user

import (
	"net/http"

	"github.com/artsgoz/artsgoz-backend/internal/server"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Register(router *gin.RouterGroup, db *gorm.DB, requireAuth gin.HandlerFunc) {
	repository := NewUserRepository(db)
	handler := NewAdminHandler(NewAdminService(repository))
	RegisterAdminRoutes(router, handler, requireAuth, db)
}

func RegisterAdminRoutes(router *gin.RouterGroup, adminHandler *AdminHandler, requireAuth gin.HandlerFunc, db *gorm.DB) {
	// Admin backoffice group for club members (role restriction removed: backoffice is for all club members)
	adminGroup := router.Group("/admin", requireAuth)
	adminGroup.GET("/dashboard", func(c *gin.Context) {
		role, _ := c.Get("role")
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to admin dashboard",
			"role":    role,
		})
	})
	adminGroup.GET("/users", server.RequireRoles(server.RoleEditor, server.RoleSuperAdmin, server.Role("admin")), adminHandler.GetAllUsers)
	adminGroup.PUT("/users/:id/role", server.RequireRoles(server.RoleSuperAdmin, server.Role("admin")), adminHandler.UpdateUserRole)
	adminGroup.GET("/me", adminHandler.GetProfile)
	adminGroup.PUT("/me", adminHandler.UpdateProfile)
	adminGroup.PUT("/me/password", adminHandler.ChangePassword)
	adminGroup.POST("/me/avatar", adminHandler.UploadAvatar)
	adminGroup.POST("/avatar", adminHandler.UploadAvatar)

	// User routes and aliases for personal profile
	userGroup := router.Group("/user", requireAuth)
	{
		userGroup.GET("/me", adminHandler.GetProfile)
		userGroup.PUT("/me", adminHandler.UpdateProfile)
		userGroup.PUT("/password", adminHandler.ChangePassword)
		userGroup.POST("/avatar", adminHandler.UploadAvatar)
	}

	// Direct user profile aliases
	router.PUT("/me", requireAuth, adminHandler.UpdateProfile)
	router.PUT("/me/password", requireAuth, adminHandler.ChangePassword)
	router.POST("/me/avatar", requireAuth, adminHandler.UploadAvatar)
}
