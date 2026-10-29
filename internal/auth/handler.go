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

func (h *Handler) Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	identity, err := h.service.VerifyToken(c.Request.Context(), request.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid, revoked, or disabled token"})
		return
	}
	response, err := h.service.Login(c.Request.Context(), identity)
	if err != nil {
		internalError(c, "failed to load user profile", err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) Register(c *gin.Context) {
	identity, ok := requireIdentity(c)
	if !ok {
		return
	}

	profile, err := h.service.Register(c.Request.Context(), identity)
	if err != nil {
		internalError(c, "registration failed", err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"uid":                 profile.FirebaseUID,
		"role":                profile.Role,
		"message":             "registration successful",
		"onboarding_required": false,
	})
}

func (h *Handler) Me(c *gin.Context) {
	identity, ok := requireIdentity(c)
	if !ok {
		return
	}

	profile, err := h.service.GetProfile(c.Request.Context(), identity.UID)
	if err != nil {
		internalError(c, "failed to load user profile", err)
		return
	}
	if profile == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "onboarding required", "onboarding_required": true})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"uid":     profile.FirebaseUID,
		"role":    profile.Role,
		"message": "authenticated successfully",
	})
}

func (h *Handler) Logout(c *gin.Context) {
	identity, ok := requireIdentity(c)
	if !ok {
		return
	}

	if err := h.service.Logout(c.Request.Context(), identity.UID); err != nil {
		internalError(c, "logout failed", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func requireIdentity(c *gin.Context) (Identity, bool) {
	value, _ := c.Get("identity")
	identity, ok := value.(Identity)
	if !ok || identity.UID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return Identity{}, false
	}
	return identity, true
}

func internalError(c *gin.Context, message string, err error) {
	log.Printf("%s: %v", message, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}
