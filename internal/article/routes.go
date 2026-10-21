package article

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	articles := router.Group("/articles")
	{
		articles.POST("", handler.Create)
		articles.GET("", handler.List)
		articles.GET("/:id", handler.Get)
		articles.PUT("/:id", handler.Update)
		articles.DELETE("/:id", handler.Delete)
		articles.PATCH("/:id/status", handler.UpdateStatus)
		articles.POST("/:id/views", handler.IncrementViews)
	}
}
