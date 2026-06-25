package rest

import (
	"github.com/gofiber/fiber/v3"

	"github.com/artsgoz/artsgoz-backend/internal/modules/curriculum/application"
)

type CurriculumHandler struct {
	list     *application.ListCurricula
	get      *application.GetCurriculum
	subjects *application.ListSubjects
}

func NewCurriculumHandler(
	list *application.ListCurricula,
	get *application.GetCurriculum,
	subjects *application.ListSubjects,
) *CurriculumHandler {
	return &CurriculumHandler{list: list, get: get, subjects: subjects}
}

func (h *CurriculumHandler) List(c fiber.Ctx) error {
	out, err := h.list.Execute(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

func (h *CurriculumHandler) Get(c fiber.Ctx) error {
	id := c.Params("id")
	out, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

func (h *CurriculumHandler) ListSubjects(c fiber.Ctx) error {
	id := c.Params("id")
	out, err := h.subjects.Execute(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}
