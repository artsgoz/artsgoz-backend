package rest

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/article/application"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type ArticleHandler struct {
	list   *application.ListArticles
	get    *application.GetArticle
	create *application.CreateArticle
	update *application.UpdateArticle
	delete *application.DeleteArticle
	Pool   *pgxpool.Pool
}

func NewArticleHandler(
	list *application.ListArticles,
	get *application.GetArticle,
	create *application.CreateArticle,
	update *application.UpdateArticle,
	delete *application.DeleteArticle,
	pool *pgxpool.Pool,
) *ArticleHandler {
	return &ArticleHandler{
		list:   list,
		get:    get,
		create: create,
		update: update,
		delete: delete,
		Pool:   pool,
	}
}

func (h *ArticleHandler) List(c fiber.Ctx) error {
	var in application.ListArticlesInput
	if err := c.Bind().Query(&in); err != nil {
		return apperr.Validation("invalid query params", err.Error())
	}
	out, err := h.list.Execute(c.Context(), in)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

func (h *ArticleHandler) Get(c fiber.Ctx) error {
	id := c.Params("id")
	out, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

func (h *ArticleHandler) Create(c fiber.Ctx) error {
	var in application.CreateArticleInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	out, err := h.create.Execute(c.Context(), in)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": out})
}

func (h *ArticleHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var in application.UpdateArticleInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	out, err := h.update.Execute(c.Context(), id, in)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

func (h *ArticleHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.delete.Execute(c.Context(), id); err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Article deleted successfully",
	})
}
