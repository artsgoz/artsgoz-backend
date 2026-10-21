package article

import (
	"errors"
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

type createArticleRequest struct {
	Title              string      `json:"title" binding:"required"`
	Slug               string      `json:"slug"`
	Subtitle           string      `json:"subtitle"`
	Excerpt            string      `json:"excerpt"`
	Content            string      `json:"content" binding:"required"`
	CoverImageURL      string      `json:"cover_image_url"`
	ImageURL           string      `json:"image_url"`
	CoverImage         string      `json:"cover_image"`
	Image              string      `json:"image"`
	AuthorID           string      `json:"author_id"`
	AuthorName         string      `json:"author_name"`
	Category           string      `json:"category"`
	Categories         []string    `json:"categories"`
	CategoryIDs        []string    `json:"category_ids"`
	Tags               []string    `json:"tags"`
	Status             string      `json:"status"`
	Visibility         string      `json:"visibility"`
	PublishedAt        *time.Time  `json:"published_at"`
	PublishDate        interface{} `json:"publish_date"`
	MetaDescription    string      `json:"meta_description"`
	SEOMetaDescription string      `json:"seo_meta_description"`
	References         []string    `json:"references"`
}

type updateArticleRequest struct {
	Title              *string     `json:"title"`
	Slug               *string     `json:"slug"`
	Subtitle           *string     `json:"subtitle"`
	Excerpt            *string     `json:"excerpt"`
	Content            *string     `json:"content"`
	CoverImageURL      *string     `json:"cover_image_url"`
	ImageURL           *string     `json:"image_url"`
	CoverImage         *string     `json:"cover_image"`
	Image              *string     `json:"image"`
	AuthorID           *string     `json:"author_id"`
	AuthorName         *string     `json:"author_name"`
	Category           *string     `json:"category"`
	Categories         []string    `json:"categories"`
	CategoryIDs        []string    `json:"category_ids"`
	Tags               []string    `json:"tags"`
	Status             *string     `json:"status"`
	Visibility         *string     `json:"visibility"`
	PublishedAt        *time.Time  `json:"published_at"`
	PublishDate        interface{} `json:"publish_date"`
	MetaDescription    *string     `json:"meta_description"`
	SEOMetaDescription *string     `json:"seo_meta_description"`
	References         []string    `json:"references"`
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// Create handles POST /api/articles
func (h *Handler) Create(c *gin.Context) {
	var req createArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	item, err := h.service.Create(c.Request.Context(), CreateArticleInput{
		Title:              req.Title,
		Slug:               req.Slug,
		Subtitle:           req.Subtitle,
		Excerpt:            req.Excerpt,
		Content:            req.Content,
		CoverImageURL:      req.CoverImageURL,
		ImageURL:           req.ImageURL,
		CoverImage:         req.CoverImage,
		Image:              req.Image,
		AuthorID:           req.AuthorID,
		AuthorName:         req.AuthorName,
		Category:           req.Category,
		Categories:         req.Categories,
		CategoryIDs:        req.CategoryIDs,
		Tags:               req.Tags,
		Status:             req.Status,
		Visibility:         req.Visibility,
		PublishedAt:        req.PublishedAt,
		PublishDate:        req.PublishDate,
		MetaDescription:    req.MetaDescription,
		SEOMetaDescription: req.SEOMetaDescription,
		References:         req.References,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTitle), errors.Is(err, ErrInvalidContent):
			writeError(c, http.StatusBadRequest, "validation", err.Error())
		case errors.Is(err, ErrSlugConflict):
			writeError(c, http.StatusConflict, "conflict", err.Error())
		default:
			_ = c.Error(err)
		}
		return
	}

	c.JSON(http.StatusCreated, item)
}

// Get handles GET /api/articles/:id
func (h *Handler) Get(c *gin.Context) {
	idOrSlug := c.Param("id")
	item, err := h.service.Get(c.Request.Context(), idOrSlug)
	if err != nil {
		if errors.Is(err, ErrArticleNotFound) {
			writeError(c, http.StatusNotFound, "not_found", "article not found")
			return
		}
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, item)
}

// Update handles PUT /api/articles/:id
func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	var req updateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	item, err := h.service.Update(c.Request.Context(), id, UpdateArticleInput{
		Title:              req.Title,
		Slug:               req.Slug,
		Subtitle:           req.Subtitle,
		Excerpt:            req.Excerpt,
		Content:            req.Content,
		CoverImageURL:      req.CoverImageURL,
		ImageURL:           req.ImageURL,
		CoverImage:         req.CoverImage,
		Image:              req.Image,
		AuthorID:           req.AuthorID,
		AuthorName:         req.AuthorName,
		Category:           req.Category,
		Categories:         req.Categories,
		CategoryIDs:        req.CategoryIDs,
		Tags:               req.Tags,
		Status:             req.Status,
		Visibility:         req.Visibility,
		PublishedAt:        req.PublishedAt,
		PublishDate:        req.PublishDate,
		MetaDescription:    req.MetaDescription,
		SEOMetaDescription: req.SEOMetaDescription,
		References:         req.References,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrArticleNotFound):
			writeError(c, http.StatusNotFound, "not_found", "article not found")
		case errors.Is(err, ErrInvalidTitle), errors.Is(err, ErrInvalidContent):
			writeError(c, http.StatusBadRequest, "validation", err.Error())
		case errors.Is(err, ErrSlugConflict):
			writeError(c, http.StatusConflict, "conflict", err.Error())
		default:
			_ = c.Error(err)
		}
		return
	}

	c.JSON(http.StatusOK, item)
}

// UpdateStatus handles PATCH /api/articles/:id/status
func (h *Handler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "status field is required")
		return
	}

	item, err := h.service.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		if errors.Is(err, ErrArticleNotFound) {
			writeError(c, http.StatusNotFound, "not_found", "article not found")
			return
		}
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, item)
}

// Delete handles DELETE /api/articles/:id
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	hard := c.Query("permanent") == "true" || c.Query("hard") == "true"

	if err := h.service.Delete(c.Request.Context(), id, hard); err != nil {
		if errors.Is(err, ErrArticleNotFound) {
			writeError(c, http.StatusNotFound, "not_found", "article not found")
			return
		}
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "article deleted successfully",
		"permanent": hard,
	})
}

// List handles GET /api/articles
func (h *Handler) List(c *gin.Context) {
	search := c.Query("search")
	if search == "" {
		search = c.Query("q")
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", c.DefaultQuery("limit", "10")))

	sortBy := c.DefaultQuery("sort", c.DefaultQuery("sort_by", "created_at"))
	order := c.DefaultQuery("order", "desc")

	includeDeleted := c.Query("include_deleted") == "true"

	params := ListParams{
		Search:         search,
		Status:         c.Query("status"),
		Category:       c.Query("category"),
		Author:         c.Query("author"),
		SortBy:         sortBy,
		Order:          order,
		Page:           page,
		PageSize:       pageSize,
		IncludeDeleted: includeDeleted,
	}

	result, err := h.service.List(c.Request.Context(), params)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result.Articles,
		"meta": gin.H{
			"page":        result.Page,
			"page_size":   result.PageSize,
			"total_items": result.TotalItems,
			"total_pages": result.TotalPages,
		},
	})
}

// IncrementViews handles POST /api/articles/:id/views
func (h *Handler) IncrementViews(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.IncrementViews(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
