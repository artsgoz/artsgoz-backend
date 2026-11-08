package auth

import "github.com/gin-gonic/gin"

func Register(router *gin.RouterGroup, service *Service) {
	handler := NewHandler(service)
	requireAuth := Middleware(service)

	routes := router.Group("/auth")
	routes.POST("/register", handler.FirebaseLogin)
	router.POST("/users/register", handler.FirebaseLogin)
	routes.POST("/login", handler.FirebaseLogin)
	routes.POST("/firebase", handler.FirebaseLogin)
	routes.POST("/firebase/login", handler.FirebaseLogin)
	routes.GET("/me", requireAuth, handler.Me)
	routes.POST("/logout", requireAuth, handler.Logout)

	// Keep root-level aliases for existing clients.
	router.POST("/login", handler.FirebaseLogin)
	router.POST("/register", handler.FirebaseLogin)
	router.GET("/me", requireAuth, handler.Me)
	router.POST("/logout", requireAuth, handler.Logout)
}
