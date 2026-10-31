package article

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func getUserContext(c *gin.Context) UserContext {
	id := c.GetString("user_id")
	if id == "" {
		id = c.GetHeader("X-User-ID")
	}
	role := c.GetString("user_role")
	if role == "" {
		role = c.GetHeader("X-User-Role")
	}
	name := c.GetString("user_name")
	if name == "" {
		name = c.GetHeader("X-User-Name")
	}
	return UserContext{
		ID:   strings.TrimSpace(id),
		Name: strings.TrimSpace(name),
		Role: strings.TrimSpace(role),
	}
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

type createCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Order       int    `json:"order"`
}

type updateCategoryRequest struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
	Order       *int    `json:"order"`
}

// Create handles POST /api/articles and POST /api/v1/admin/articles
func (h *Handler) Create(c *gin.Context) {
	var req createArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	userCtx := getUserContext(c)

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
		User:               userCtx,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTitle), errors.Is(err, ErrInvalidContent):
			writeError(c, http.StatusBadRequest, "validation", err.Error())
		case errors.Is(err, ErrSlugConflict):
			writeError(c, http.StatusConflict, "conflict", err.Error())
		case errors.Is(err, ErrForbidden):
			writeError(c, http.StatusForbidden, "forbidden", err.Error())
		default:
			_ = c.Error(err)
		}
		return
	}

	c.Header("Location", fmt.Sprintf("/api/v1/admin/articles/%s", item.ID))
	c.JSON(http.StatusCreated, item)
}

// Get handles GET /api/articles/:id and GET /api/v1/admin/articles/:id
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

// Update handles PUT /api/articles/:id and PUT /api/v1/admin/articles/:id
func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	var req updateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	userCtx := getUserContext(c)

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
		User:               userCtx,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrArticleNotFound):
			writeError(c, http.StatusNotFound, "not_found", "article not found")
		case errors.Is(err, ErrInvalidTitle), errors.Is(err, ErrInvalidContent):
			writeError(c, http.StatusBadRequest, "validation", err.Error())
		case errors.Is(err, ErrSlugConflict):
			writeError(c, http.StatusConflict, "conflict", err.Error())
		case errors.Is(err, ErrForbidden):
			writeError(c, http.StatusForbidden, "forbidden", err.Error())
		default:
			_ = c.Error(err)
		}
		return
	}

	c.JSON(http.StatusOK, item)
}

// UpdateStatus handles PATCH /api/articles/:id/status and PATCH /api/v1/admin/articles/:id/status
func (h *Handler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "status field is required")
		return
	}

	userCtx := getUserContext(c)

	item, err := h.service.UpdateStatus(c.Request.Context(), id, req.Status, userCtx)
	if err != nil {
		switch {
		case errors.Is(err, ErrArticleNotFound):
			writeError(c, http.StatusNotFound, "not_found", "article not found")
		case errors.Is(err, ErrForbidden):
			writeError(c, http.StatusForbidden, "forbidden", err.Error())
		default:
			_ = c.Error(err)
		}
		return
	}

	c.JSON(http.StatusOK, item)
}

// Delete handles DELETE /api/articles/:id and DELETE /api/v1/admin/articles/:id
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	hard := c.Query("permanent") == "true" || c.Query("hard") == "true"
	userCtx := getUserContext(c)

	if err := h.service.Delete(c.Request.Context(), id, hard, userCtx); err != nil {
		switch {
		case errors.Is(err, ErrArticleNotFound):
			writeError(c, http.StatusNotFound, "not_found", "article not found")
		case errors.Is(err, ErrForbidden):
			writeError(c, http.StatusForbidden, "forbidden", err.Error())
		default:
			_ = c.Error(err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "article deleted successfully",
		"permanent": hard,
	})
}

// List handles GET /api/articles and GET /api/v1/admin/articles
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
		CategoryID:     c.Query("category_id"),
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
			"limit":       result.PageSize,
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

// ListActivityLogs handles GET /api/v1/admin/articles/:id/logs
func (h *Handler) ListActivityLogs(c *gin.Context) {
	id := c.Param("id")
	logs, err := h.service.ListActivityLogs(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": logs,
	})
}

// Category handlers

// CreateCategory handles POST /api/v1/admin/categories
func (h *Handler) CreateCategory(c *gin.Context) {
	var req createCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	item, err := h.service.CreateCategory(c.Request.Context(), CreateCategoryInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Color:       req.Color,
		Order:       req.Order,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCategoryName):
			writeError(c, http.StatusBadRequest, "validation", err.Error())
		case errors.Is(err, ErrCategoryConflict):
			writeError(c, http.StatusConflict, "conflict", err.Error())
		default:
			_ = c.Error(err)
		}
		return
	}

	c.Header("Location", fmt.Sprintf("/api/v1/admin/categories/%s", item.ID))
	c.JSON(http.StatusCreated, item)
}

// GetCategory handles GET /api/v1/admin/categories/:id and GET /api/categories/:id
func (h *Handler) GetCategory(c *gin.Context) {
	idOrSlug := c.Param("id")
	item, err := h.service.GetCategory(c.Request.Context(), idOrSlug)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			writeError(c, http.StatusNotFound, "not_found", "category not found")
			return
		}
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, item)
}

// UpdateCategory handles PUT /api/v1/admin/categories/:id
func (h *Handler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var req updateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	item, err := h.service.UpdateCategory(c.Request.Context(), id, UpdateCategoryInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Color:       req.Color,
		Order:       req.Order,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrCategoryNotFound):
			writeError(c, http.StatusNotFound, "not_found", "category not found")
		case errors.Is(err, ErrInvalidCategoryName):
			writeError(c, http.StatusBadRequest, "validation", err.Error())
		case errors.Is(err, ErrCategoryConflict):
			writeError(c, http.StatusConflict, "conflict", err.Error())
		default:
			_ = c.Error(err)
		}
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteCategory handles DELETE /api/v1/admin/categories/:id
func (h *Handler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteCategory(c.Request.Context(), id); err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			writeError(c, http.StatusNotFound, "not_found", "category not found")
			return
		}
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "category deleted successfully",
	})
}

// ListCategories handles GET /api/v1/admin/categories and GET /api/categories
func (h *Handler) ListCategories(c *gin.Context) {
	items, err := h.service.ListCategories(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": items,
	})
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
