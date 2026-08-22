package controller

import (
	"errors"

	"github.com/gofiber/fiber/v3"

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

func (h *Handler) RegisterUser(c fiber.Ctx) error {
	var request registerUserRequest
	if err := c.Bind().Body(&request); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_request", "invalid request body")
	}

	output, err := h.registerUser.Execute(c.Context(), usecase.RegisterUserInput{
		Email: request.Email, Password: request.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidEmail), errors.Is(err, usecase.ErrInvalidPassword):
			return writeError(c, fiber.StatusBadRequest, "validation", err.Error())
		case errors.Is(err, usecase.ErrEmailAlreadyExists):
			return writeError(c, fiber.StatusConflict, "conflict", err.Error())
		default:
			return err
		}
	}

	return c.Status(fiber.StatusCreated).JSON(registerUserResponse{ID: output.ID, Email: output.Email})
}

func writeError(c fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": fiber.Map{"code": code, "message": message},
	})
}
