package dashboard

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	dash := router.Group("/dashboard")
	{
		dash.GET("/overview", handler.GetOverview)
		dash.GET("/traffic", handler.GetTraffic)
		dash.GET("/categories", handler.GetCategories)
		dash.GET("/export", handler.Export)
		dash.POST("/traffic/record", handler.RecordTraffic)
	}
}

