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

// RegisterAdminRoutes registers the spec-compliant admin dashboard analytics endpoints.
// These routes should be mounted under an authenticated /admin group with Editor/SuperAdmin access.
//
//	[GET] /admin/dashboard/stats           — Key Metrics (4 cards)
//	[GET] /admin/dashboard/monthly-trends  — 12-Month Visitor Trends
//	[GET] /admin/dashboard/top-articles    — Top Articles by Views
//	[GET] /admin/dashboard/top-categories  — Top Categories (Donut Chart)
//	[GET] /admin/dashboard/recent-activities — Audit Trail Logs
func RegisterAdminRoutes(router *gin.RouterGroup, handler *Handler) {
	dash := router.Group("/dashboard")
	{
		dash.GET("/stats", handler.GetStats)
		dash.GET("/monthly-trends", handler.GetMonthlyTrends)
		dash.GET("/top-articles", handler.GetTopArticlesHandler)
		dash.GET("/top-categories", handler.GetTopCategoriesHandler)
		dash.GET("/recent-activities", handler.GetRecentActivities)
	}
}
