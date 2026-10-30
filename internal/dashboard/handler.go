package dashboard

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetOverview handles GET /api/v1/dashboard/overview
func (h *Handler) GetOverview(c *gin.Context) {
	data, err := h.service.GetOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetTraffic handles GET /api/v1/dashboard/traffic
func (h *Handler) GetTraffic(c *gin.Context) {
	filter := c.DefaultQuery("filter", c.DefaultQuery("period", "7d"))
	data, err := h.service.GetTraffic(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetCategories handles GET /api/v1/dashboard/categories
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

