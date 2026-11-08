package comment

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	comments := router.Group("/comments")
	{
		comments.GET("", handler.List)
		comments.POST("", handler.Create)
		comments.DELETE("/:id", handler.Delete)
	}

	articles := router.Group("/articles/:id/comments")
	{
		articles.GET("", handler.List)
		articles.POST("", handler.Create)
	}
}
