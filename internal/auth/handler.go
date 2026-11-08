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
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "invalid request body: " + err.Error()},
			"message": "invalid request body: " + err.Error(),
		})
		return
	}

	user, token, err := h.service.Register(c.Request.Context(), request)
	if err != nil {
		if err.Error() == "username is already taken" || err.Error() == "email is already registered" {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   gin.H{"code": "CONFLICT", "message": err.Error()},
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": err.Error()},
			"message": err.Error(),
		})
		return
	}

	userData := gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"avatar_url": user.AvatarURL,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "registration successful",
		"token":   token,
		"user":    userData,
		"data": gin.H{
			"token": token,
			"user":  userData,
		},
	})
}

func (h *Handler) Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "invalid request body: " + err.Error()},
			"message": "invalid request body: " + err.Error(),
		})
		return
	}

	response, err := h.service.Login(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   gin.H{"code": "UNAUTHORIZED", "message": err.Error()},
			"message": err.Error(),
		})
		return
	}

	userData := gin.H{
		"id":         response.User.ID,
		"username":   response.User.Username,
		"email":      response.User.Email,
		"role":       response.User.Role,
		"avatar_url": response.User.AvatarURL,
		"created_at": response.User.CreatedAt,
		"updated_at": response.User.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   response.Token,
		"user":    userData,
		"data": gin.H{
			"token": response.Token,
			"user":  userData,
		},
	})
}

func (h *Handler) Me(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   gin.H{"code": "UNAUTHORIZED", "message": "authentication required"},
			"message": "authentication required",
		})
		return
	}

	profile, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		internalError(c, "failed to load user profile", err)
		return
	}
	if profile == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"code": "NOT_FOUND", "message": "user not found"},
			"message": "user not found",
		})
		return
	}

	userData := gin.H{
		"id":         profile.ID,
		"username":   profile.Username,
		"email":      profile.Email,
		"role":       profile.Role,
		"avatar_url": profile.AvatarURL,
		"created_at": profile.CreatedAt,
		"updated_at": profile.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    userData,
		"data":    userData,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "logout successful",
		"data":    nil,
	})
}

func getUserID(c *gin.Context) string {
	if val := c.GetString("user_id"); val != "" {
		return val
	}
	return c.GetString("uid")
}

func internalError(c *gin.Context, msg string, err error) {
	log.Printf("ERROR: %s: %v", msg, err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error":   gin.H{"code": "INTERNAL_SERVER_ERROR", "message": msg},
		"message": msg,
	})
}
