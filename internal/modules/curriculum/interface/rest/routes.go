package rest

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(r fiber.Router, h *CurriculumHandler) {
	g := r.Group("/curricula")
	g.Get("/", h.List)
	g.Get("/:id", h.Get)
	g.Get("/:id/subjects", h.ListSubjects)
}
