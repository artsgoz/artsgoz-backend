package rest

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(r fiber.Router, h *UserHandler) {
	g := r.Group("/users")
	g.Post("/register", h.Register)
}
