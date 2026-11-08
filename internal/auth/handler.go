package auth

import (
	"errors"
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

func (h *Handler) FirebaseLogin(c *gin.Context) {
	var request FirebaseLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "invalid request body: " + err.Error()},
			"message": "invalid request body: " + err.Error(),
		})
		return
	}

	response, err := h.service.FirebaseLogin(c.Request.Context(), request)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	userData := gin.H{
		"id":           response.User.ID,
		"firebase_uid": response.User.FirebaseUID,
		"username":     response.User.Username,
		"email":        response.User.Email,
		"role":         response.User.Role,
		"avatar_url":   response.User.AvatarURL,
		"created_at":   response.User.CreatedAt,
		"updated_at":   response.User.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "firebase login successful",
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
		"id":           profile.ID,
		"firebase_uid": profile.FirebaseUID,
		"username":     profile.Username,
		"email":        profile.Email,
		"role":         profile.Role,
		"avatar_url":   profile.AvatarURL,
		"created_at":   profile.CreatedAt,
		"updated_at":   profile.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    userData,
		"data":    userData,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	if err := h.service.Logout(c.Request.Context(), c.GetString("firebase_uid")); err != nil {
		internalError(c, "failed to sign out Firebase sessions", err)
		return
	}
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

func writeAuthError(c *gin.Context, err error) {
	status, code := http.StatusUnauthorized, "UNAUTHORIZED"
	switch {
	case errors.Is(err, ErrFirebaseUnavailable):
		status, code = http.StatusServiceUnavailable, "AUTH_UNAVAILABLE"
	case errors.Is(err, ErrAccountConflict):
		status, code = http.StatusConflict, "ACCOUNT_CONFLICT"
	case errors.Is(err, ErrInvalidUsername):
		status, code = http.StatusBadRequest, "BAD_REQUEST"
	case !errors.Is(err, ErrInvalidToken):
		internalError(c, "authentication failed", err)
		return
	}
	c.JSON(status, gin.H{"success": false, "error": gin.H{"code": code, "message": err.Error()}, "message": err.Error()})
}
