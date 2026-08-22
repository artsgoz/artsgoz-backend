package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	users := router.Group("/users")
	users.POST("/register", handler.Register)
}
