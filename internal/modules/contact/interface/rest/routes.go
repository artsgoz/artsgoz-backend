package rest

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(r fiber.Router, h *ContactSubmissionHandler) {
	g := r.Group("/contacts")
	g.Post("/", h.Submit)
}
