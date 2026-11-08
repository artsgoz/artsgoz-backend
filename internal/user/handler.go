package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type registerRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type registerResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (h *Handler) Register(c *gin.Context) {
	var request registerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "email and password are required")
		return
	}

	item, err := h.service.Register(c.Request.Context(), RegisterInput{
		Email: request.Email, Password: request.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrInvalidPassword):
			writeError(c, http.StatusBadRequest, "validation", err.Error())
		case errors.Is(err, ErrEmailAlreadyExists):
			writeError(c, http.StatusConflict, "conflict", err.Error())
		default:
			_ = c.Error(err)
		}
		return
	}

	c.JSON(http.StatusCreated, registerResponse{ID: item.ID, Email: item.Email})
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
