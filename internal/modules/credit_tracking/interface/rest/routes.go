package rest

import (
	"github.com/gofiber/fiber/v3"
	"github.com/artsgoz/artsgoz-backend/internal/platform/auth"
)

func RegisterRoutes(r fiber.Router, h *CreditTrackingHandler) {
	meGroup := r.Group("/me", auth.FirebaseAuthMiddleware())
	
	meGroup.Get("/subjects", h.GetSubjects)
	meGroup.Put("/subjects/:subjectId", h.UpdateSubjectCompletion)
	meGroup.Post("/subjects", h.AddCustomSubject)
	meGroup.Delete("/subjects/:subjectId", h.DeleteCustomSubject)
	
	meGroup.Get("/profile", h.GetProfile)
	meGroup.Put("/profile", h.UpdateProfile)
}
