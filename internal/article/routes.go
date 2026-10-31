package article

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	// Standard / legacy articles
	articles := router.Group("/articles")
	{
		articles.POST("", handler.Create)
		articles.GET("", handler.List)
		articles.GET("/:id", handler.Get)
		articles.PUT("/:id", handler.Update)
		articles.DELETE("/:id", handler.Delete)
		articles.PATCH("/:id/status", handler.UpdateStatus)
		articles.POST("/:id/views", handler.IncrementViews)
		articles.GET("/:id/logs", handler.ListActivityLogs)
	}

	// Public categories
	categories := router.Group("/categories")
	{
		categories.GET("", handler.ListCategories)
		categories.GET("/:id", handler.GetCategory)
	}

	// Admin Article & Category Management endpoints
	admin := router.Group("/admin")
	{
		adminArticles := admin.Group("/articles")
		{
			adminArticles.GET("", handler.List)
			adminArticles.POST("", handler.Create)
			adminArticles.GET("/:id", handler.Get)
			adminArticles.PUT("/:id", handler.Update)
			adminArticles.DELETE("/:id", handler.Delete)
			adminArticles.PATCH("/:id/status", handler.UpdateStatus)
			adminArticles.GET("/:id/logs", handler.ListActivityLogs)
		}

		adminCategories := admin.Group("/categories")
		{
			adminCategories.GET("", handler.ListCategories)
			adminCategories.POST("", handler.CreateCategory)
			adminCategories.GET("/:id", handler.GetCategory)
			adminCategories.PUT("/:id", handler.UpdateCategory)
			adminCategories.DELETE("/:id", handler.DeleteCategory)
		}
	}
}
