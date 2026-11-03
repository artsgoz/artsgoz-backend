package auth

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var request RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	user, token, err := h.service.Register(c.Request.Context(), request)
	if err != nil {
		if err.Error() == "username is already taken" || err.Error() == "email is already registered" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "registration successful",
		"token":   token,
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}

func (h *Handler) Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	response, err := h.service.Login(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": response.Token,
		"user": gin.H{
			"id":         response.User.ID,
			"username":   response.User.Username,
			"email":      response.User.Email,
			"role":       response.User.Role,
			"created_at": response.User.CreatedAt,
			"updated_at": response.User.UpdatedAt,
		},
	})
}

func (h *Handler) Me(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	profile, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		internalError(c, "failed to load user profile", err)
		return
	}
	if profile == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         profile.ID,
			"username":   profile.Username,
			"email":      profile.Email,
			"role":       profile.Role,
			"created_at": profile.CreatedAt,
			"updated_at": profile.UpdatedAt,
		},
	})
}

func (h *Handler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func getUserID(c *gin.Context) string {
	if val := c.GetString("user_id"); val != "" {
		return val
	}
	return c.GetString("uid")
}

func internalError(c *gin.Context, msg string, err error) {
	log.Printf("ERROR: %s: %v", msg, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
}
