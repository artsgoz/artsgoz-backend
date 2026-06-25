package rest

import (
	"github.com/gofiber/fiber/v3"
	"github.com/artsgoz/artsgoz-backend/internal/platform/auth"
)

func RegisterRoutes(r fiber.Router, h *ClubHandler) {
	g := r.Group("/clubs")
	g.Get("/", h.List)
	g.Get("/:id", h.Get)

	// Admin routes
	adminGroup := g.Group("/", auth.FirebaseAuthMiddleware(), auth.RequireAdmin(h.Pool))
	adminGroup.Post("/", h.Create)
	adminGroup.Put("/:id", h.Update)
	adminGroup.Delete("/:id", h.Delete)
}
