package apperr

import (
	"github.com/gofiber/fiber/v3"
)

func statusFor(kind Kind) int {
	switch kind {
	case KindValidation:
		return fiber.StatusBadRequest
	case KindConflict:
		return fiber.StatusConflict
	case KindNotFound:
		return fiber.StatusNotFound
	case KindUnauthorized:
		return fiber.StatusUnauthorized
	case KindForbidden:
		return fiber.StatusForbidden
	default:
		return fiber.StatusInternalServerError
	}
}

// FiberErrorHandler wires *apperr.Error into Fiber's centralised error pipeline.
// Plain errors fall back to 500.
func FiberErrorHandler(c fiber.Ctx, err error) error {
	if appErr, ok := As(err); ok {
		body := fiber.Map{
			"kind":    string(appErr.Kind),
			"message": appErr.Message,
		}
		if appErr.Details != nil {
			body["details"] = appErr.Details
		}
		return c.Status(statusFor(appErr.Kind)).JSON(body)
	}

	if fe, ok := err.(*fiber.Error); ok {
		return c.Status(fe.Code).JSON(fiber.Map{
			"kind":    "internal",
			"message": fe.Message,
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"kind":    "internal",
		"message": "internal server error",
	})
}
