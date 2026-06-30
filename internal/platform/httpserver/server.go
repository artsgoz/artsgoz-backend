package httpserver

import (
	"github.com/gofiber/fiber/v3"

	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

// New returns a Fiber app pre-wired with the central apperr error handler.
func New() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: apperr.FiberErrorHandler,
	})
}
