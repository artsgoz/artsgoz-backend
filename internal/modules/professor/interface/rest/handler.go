package rest

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/professor/application"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

// ProfessorHandler handles HTTP requests for the professor resource.
type ProfessorHandler struct {
	list   *application.ListProfessors
	get    *application.GetProfessor
	create *application.CreateProfessor
	update *application.UpdateProfessor
	delete *application.DeleteProfessor
	Pool   *pgxpool.Pool
}

func NewProfessorHandler(
	list *application.ListProfessors,
	get *application.GetProfessor,
	create *application.CreateProfessor,
	update *application.UpdateProfessor,
	delete *application.DeleteProfessor,
	pool *pgxpool.Pool,
) *ProfessorHandler {
	return &ProfessorHandler{
		list:   list,
		get:    get,
		create: create,
		update: update,
		delete: delete,
		Pool:   pool,
	}
}

// List handles GET /api/v1/professors
func (h *ProfessorHandler) List(c fiber.Ctx) error {
	var in application.ListProfessorsInput
	if err := c.Bind().Query(&in); err != nil {
		return apperr.Validation("invalid query params", err.Error())
	}
	out, err := h.list.Execute(c.Context(), in)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

// Get handles GET /api/v1/professors/:id
func (h *ProfessorHandler) Get(c fiber.Ctx) error {
	id := c.Params("id")
	out, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

// Create handles POST /api/v1/professors (Admin only)
func (h *ProfessorHandler) Create(c fiber.Ctx) error {
	var in application.CreateProfessorInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	out, err := h.create.Execute(c.Context(), in)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": out})
}

// Update handles PUT /api/v1/professors/:id (Admin only)
func (h *ProfessorHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var in application.UpdateProfessorInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	out, err := h.update.Execute(c.Context(), id, in)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

// Delete handles DELETE /api/v1/professors/:id (Admin only)
func (h *ProfessorHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.delete.Execute(c.Context(), id); err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Professor deleted successfully",
	})
}
