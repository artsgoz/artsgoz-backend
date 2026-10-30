package activity

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	activities := router.Group("/activities")
	{
		activities.GET("", handler.List)
		activities.POST("", handler.Create)
	}
}

