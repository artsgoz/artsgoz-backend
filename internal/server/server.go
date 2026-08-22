package server

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

// New creates the HTTP server. Feature routes are registered by their modules.
func New(logger *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler(logger),
	})
	registerMiddleware(app, logger)

	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	return app
}

func errorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		code := "internal"
		message := "internal server error"

		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			status = fiberErr.Code
			message = fiberErr.Message
			if status < fiber.StatusInternalServerError {
				code = "request_error"
			}
		}

		if status >= fiber.StatusInternalServerError {
			logger.Error("http request failed", "error", err, "method", c.Method(), "path", c.Path())
		}

		return c.Status(status).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    code,
				"message": message,
			},
		})
	}
}
