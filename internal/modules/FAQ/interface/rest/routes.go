package rest

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(r fiber.Router, h *FAQHandler) {
	g := r.Group("/faqs")
	g.Get("/", h.List)
	g.Post("/", h.Create)
}
