package rest

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/club/application"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type ClubHandler struct {
	list   *application.ListClubs
	get    *application.GetClub
	create *application.CreateClub
	update *application.UpdateClub
	delete *application.DeleteClub
	Pool   *pgxpool.Pool
}

func NewClubHandler(
	list *application.ListClubs,
	get *application.GetClub,
	create *application.CreateClub,
	update *application.UpdateClub,
	delete *application.DeleteClub,
	pool *pgxpool.Pool,
) *ClubHandler {
	return &ClubHandler{
		list:   list,
		get:    get,
		create: create,
		update: update,
		delete: delete,
		Pool:   pool,
	}
}

func (h *ClubHandler) List(c fiber.Ctx) error {
	var in application.ListClubsInput
	if err := c.Bind().Query(&in); err != nil {
		return apperr.Validation("invalid query params", err.Error())
	}
	out, err := h.list.Execute(c.Context(), in)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

func (h *ClubHandler) Get(c fiber.Ctx) error {
	id := c.Params("id")
	out, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

func (h *ClubHandler) Create(c fiber.Ctx) error {
	var in application.CreateClubInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	out, err := h.create.Execute(c.Context(), in)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": out})
}

func (h *ClubHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var in application.UpdateClubInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	out, err := h.update.Execute(c.Context(), id, in)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

func (h *ClubHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.delete.Execute(c.Context(), id); err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Club deleted successfully",
	})
}
