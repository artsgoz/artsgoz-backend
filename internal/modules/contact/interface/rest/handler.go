package rest

import (
	"github.com/gofiber/fiber/v3"

	"github.com/artsgoz/artsgoz-backend/internal/modules/contact/application"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type ContactSubmissionHandler struct {
	submit *application.SubmitContact
}

func NewContactSubmissionHandler(submit *application.SubmitContact) *ContactSubmissionHandler {
	return &ContactSubmissionHandler{submit: submit}
}

func (h *ContactSubmissionHandler) Submit(c fiber.Ctx) error {
	var in application.SubmitContactInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	out, err := h.submit.Execute(c.Context(), in)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    out,
	})
}
