package activity

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// List handles GET /api/v1/activities
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("page_size", "10")))
	search := c.Query("search")
	if search == "" {
		search = c.Query("q")
	}

	result, err := h.service.List(c.Request.Context(), page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     result.Items,
		"has_more": result.HasMore,
		"meta": gin.H{
			"page":        result.Page,
			"limit":       result.Limit,
			"total_items": result.TotalItems,
		},
	})
}

type createActivityRequest struct {
	Type         Type   `json:"type" binding:"required"`
	Action       string `json:"action" binding:"required"`
	Detail       string `json:"detail"`
	ArticleID    string `json:"article_id"`
	ArticleTitle string `json:"article_title"`
	ActorID      string `json:"actor_id"`
	ActorName    string `json:"actor_name"`
}

// Create handles POST /api/v1/activities
func (h *Handler) Create(c *gin.Context) {
	var req createActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.Log(c.Request.Context(), LogInput{
		Type:         req.Type,
		Action:       req.Action,
		Detail:       req.Detail,
		ArticleID:    req.ArticleID,
		ArticleTitle: req.ArticleTitle,
		ActorID:      req.ActorID,
		ActorName:    req.ActorName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

