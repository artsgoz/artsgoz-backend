package rest

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/document/application"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type DocumentHandler struct {
	list   *application.ListDocuments
	get    *application.GetDocument
	create *application.CreateDocument
	update *application.UpdateDocument
	delete *application.DeleteDocument
	Pool   *pgxpool.Pool
}

func NewDocumentHandler(
	list *application.ListDocuments,
	get *application.GetDocument,
	create *application.CreateDocument,
	update *application.UpdateDocument,
	delete *application.DeleteDocument,
	pool *pgxpool.Pool,
) *DocumentHandler {
	return &DocumentHandler{
		list:   list,
		get:    get,
		create: create,
		update: update,
		delete: delete,
		Pool:   pool,
	}
}

func (h *DocumentHandler) List(c fiber.Ctx) error {
	var in application.ListDocumentsInput
	if err := c.Bind().Query(&in); err != nil {
		return apperr.Validation("invalid query params", err.Error())
	}
	out, err := h.list.Execute(c.Context(), in)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

func (h *DocumentHandler) Get(c fiber.Ctx) error {
	id := c.Params("id")
	out, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

func (h *DocumentHandler) Create(c fiber.Ctx) error {
	var in application.CreateDocumentInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	out, err := h.create.Execute(c.Context(), in)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": out})
}

func (h *DocumentHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var in application.UpdateDocumentInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	out, err := h.update.Execute(c.Context(), id, in)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

func (h *DocumentHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.delete.Execute(c.Context(), id); err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Document deleted successfully",
	})
}
