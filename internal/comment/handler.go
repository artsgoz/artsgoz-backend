package comment

import (
	"errors"
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

// List handles GET /api/v1/comments
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("page_size", "10")))
	search := c.Query("search")
	if search == "" {
		search = c.Query("q")
	}
	articleID := c.Query("article_id")
	if articleID == "" {
		articleID = c.Param("article_id")
	}

	result, err := h.service.List(c.Request.Context(), page, limit, search, articleID)
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

type createCommentRequest struct {
	ArticleID    string `json:"article_id"`
	ArticleTitle string `json:"article_title"`
	AuthorID     string `json:"author_id"`
	AuthorName   string `json:"author_name" binding:"required"`
	AuthorAvatar string `json:"author_avatar"`
	Content      string `json:"content" binding:"required"`
}

// Create handles POST /api/v1/comments and POST /api/v1/articles/:article_id/comments
func (h *Handler) Create(c *gin.Context) {
	var req createCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	articleID := req.ArticleID
	if articleID == "" {
		articleID = c.Param("article_id")
	}

	item, err := h.service.Create(c.Request.Context(), CreateCommentInput{
		ArticleID:    articleID,
		ArticleTitle: req.ArticleTitle,
		AuthorID:     req.AuthorID,
		AuthorName:   req.AuthorName,
		AuthorAvatar: req.AuthorAvatar,
		Content:      req.Content,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidContent), errors.Is(err, ErrInvalidAuthor), errors.Is(err, ErrInvalidArticle):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, item)
}

// Delete handles DELETE /api/v1/comments/:id
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "comment deleted successfully"})
}

