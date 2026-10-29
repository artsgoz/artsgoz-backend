package auth

import "github.com/gin-gonic/gin"

func Register(router *gin.RouterGroup, service *Service) {
	handler := NewHandler(service)
	requireAuth := Middleware(service)

	routes := router.Group("/auth")
	routes.POST("/login", handler.Login)
	routes.POST("/register", requireAuth, handler.Register)
	routes.POST("/logout", requireAuth, handler.Logout)

	// Keep root-level aliases for existing clients.
	router.POST("/login", handler.Login)
	router.POST("/register", requireAuth, handler.Register)
	router.GET("/me", requireAuth, handler.Me)
}
