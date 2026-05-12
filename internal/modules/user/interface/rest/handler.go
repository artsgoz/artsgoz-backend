package rest

import (
	"github.com/gofiber/fiber/v3"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/application"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type UserHandler struct {
	register *application.RegisterUser
}

func NewUserHandler(register *application.RegisterUser) *UserHandler {
	return &UserHandler{register: register}
}

func (h *UserHandler) Register(c fiber.Ctx) error {
	var in application.RegisterUserInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}
	out, err := h.register.Execute(c.Context(), in)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(out)
}
