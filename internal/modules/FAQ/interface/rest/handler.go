package rest

import (
	"github.com/gofiber/fiber/v3"

	"github.com/artsgoz/artsgoz-backend/internal/modules/FAQ/application"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type FAQHandler struct {
	list   *application.ListFAQs
	create *application.CreateFAQ
}

func NewFAQHandler(list *application.ListFAQs, create *application.CreateFAQ) *FAQHandler {
	return &FAQHandler{list: list, create: create}
}

func (h *FAQHandler) List(c fiber.Ctx) error {
	var in application.ListFAQsInput
	if err := c.Bind().Query(&in); err != nil {
		return apperr.Validation("invalid query params", err.Error())
	}

	out, err := h.list.Execute(c.Context(), in)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    out,
	})
}

func (h *FAQHandler) Create(c fiber.Ctx) error {
	var in application.CreateFAQInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	out, err := h.create.Execute(c.Context(), in)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    out,
	})
}
