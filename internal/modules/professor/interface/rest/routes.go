package rest

import (
	"github.com/gofiber/fiber/v3"
	"github.com/artsgoz/artsgoz-backend/internal/platform/auth"
)

// RegisterRoutes mounts professor routes onto the given router group.
func RegisterRoutes(r fiber.Router, h *ProfessorHandler) {
	g := r.Group("/professors")
	g.Get("/", h.List)
	g.Get("/:id", h.Get)

	// Admin routes
	adminGroup := g.Group("/", auth.FirebaseAuthMiddleware(), auth.RequireAdmin(h.Pool))
	adminGroup.Post("/", h.Create)
	adminGroup.Put("/:id", h.Update)
	adminGroup.Delete("/:id", h.Delete)
}
