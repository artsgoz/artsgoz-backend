package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/artsgoz/artsgoz-backend/internal/user/usecase"
)

type Handler struct {
	registerUser *usecase.RegisterUser
}

func NewHandler(registerUser *usecase.RegisterUser) *Handler {
	return &Handler{registerUser: registerUser}
}

type registerUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (h *Handler) RegisterUser(c *gin.Context) {
	var request registerUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	output, err := h.registerUser.Execute(c.Request.Context(), usecase.RegisterUserInput{
		Email: request.Email, Password: request.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidEmail), errors.Is(err, usecase.ErrInvalidPassword):
			writeError(c, http.StatusBadRequest, "validation", err.Error())
		case errors.Is(err, usecase.ErrEmailAlreadyExists):
			writeError(c, http.StatusConflict, "conflict", err.Error())
		default:
			_ = c.Error(err)
		}
		return
	}

	c.JSON(http.StatusCreated, registerUserResponse{ID: output.ID, Email: output.Email})
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
