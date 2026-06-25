package rest

import (
	"github.com/gofiber/fiber/v3"
	"github.com/artsgoz/artsgoz-backend/internal/platform/auth"
)

func RegisterRoutes(r fiber.Router, h *YellowCardHandler) {
	// /api/v1/me/... group requiring authentication
	meGroup := r.Group("/me", auth.FirebaseAuthMiddleware())
	
	// /api/v1/me/yellow-card/... group
	ycGroup := meGroup.Group("/yellow-card")
	ycGroup.Get("", h.Get)
	ycGroup.Put("/profile", h.UpdateProfile)
	ycGroup.Put("/subjects", h.UpdateSubjects)
	ycGroup.Post("/pdpa", h.RecordPDPA)
}
