package dashboard

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetOverview handles GET /api/v1/dashboard/overview (legacy)
func (h *Handler) GetOverview(c *gin.Context) {
	data, err := h.service.GetOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetTraffic handles GET /api/v1/dashboard/traffic (legacy)
func (h *Handler) GetTraffic(c *gin.Context) {
	filter := c.DefaultQuery("filter", c.DefaultQuery("period", "7d"))
	data, err := h.service.GetTraffic(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetCategories handles GET /api/v1/dashboard/categories (legacy)
func (h *Handler) GetCategories(c *gin.Context) {
	data, err := h.service.GetCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// Export handles GET /api/v1/dashboard/export
func (h *Handler) Export(c *gin.Context) {
	csvBytes, err := h.service.ExportCSV(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("saraniyakorn-dashboard-export-%s.csv", time.Now().Format("20060102-150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvBytes)
}

type recordTrafficRequest struct {
	ArticleID string `json:"article_id"`
	Source    string `json:"source"`
}

// RecordTraffic handles POST /api/v1/dashboard/traffic/record
func (h *Handler) RecordTraffic(c *gin.Context) {
	var req recordTrafficRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If empty body, record default direct traffic
		req.Source = "direct"
	}

	if err := h.service.RecordTraffic(c.Request.Context(), req.ArticleID, req.Source); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ---------------------------------------------------------------------------
// Admin Dashboard Analytics endpoints (spec-compliant)
// ---------------------------------------------------------------------------

// GetStats handles GET /api/v1/admin/dashboard/stats
// Returns 4 key metric cards: total views, published articles, pending comments, total categories.
func (h *Handler) GetStats(c *gin.Context) {
	data, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetMonthlyTrends handles GET /api/v1/admin/dashboard/monthly-trends
// Query parameter: year (integer, default: current year)
func (h *Handler) GetMonthlyTrends(c *gin.Context) {
	yearStr := c.Query("year")
	year := time.Now().Year()
	if yearStr != "" {
		if parsed, err := strconv.Atoi(yearStr); err == nil && parsed >= 2000 && parsed <= 2100 {
			year = parsed
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year parameter"})
			return
		}
	}

	data, err := h.service.GetMonthlyTrends(c.Request.Context(), year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetTopArticlesHandler handles GET /api/v1/admin/dashboard/top-articles
// Query parameter: limit (integer, default: 5, max: 20)
func (h *Handler) GetTopArticlesHandler(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	data, err := h.service.GetTopArticlesList(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetTopCategoriesHandler handles GET /api/v1/admin/dashboard/top-categories
// Returns category distribution for Donut Chart.
func (h *Handler) GetTopCategoriesHandler(c *gin.Context) {
	data, err := h.service.GetTopCategoriesDistribution(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetRecentActivities handles GET /api/v1/admin/dashboard/recent-activities
// Returns the 10 most recent audit trail log entries.
func (h *Handler) GetRecentActivities(c *gin.Context) {
	data, err := h.service.GetRecentActivities(c.Request.Context(), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
