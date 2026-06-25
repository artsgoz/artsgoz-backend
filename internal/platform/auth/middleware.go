package auth

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

// FirebaseAuthMiddleware extracts the Firebase UID from the Bearer Token in the Authorization header
// or X-User-ID header, and saves it to c.Locals("uid") for sub-handlers.
func FirebaseAuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		var uid string
		if strings.HasPrefix(authHeader, "Bearer ") {
			uid = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Fallback to X-User-ID header (for local testing)
		if uid == "" {
			uid = c.Get("X-User-ID")
		}

		if uid == "" {
			return apperr.Unauthorized("missing authorization token or X-User-ID header")
		}

		c.Locals("uid", uid)
		return c.Next()
	}
}

// RequireAdmin guards routes by checking if the user in c.Locals("uid") has role = 'admin' in the database.
func RequireAdmin(pool *pgxpool.Pool) fiber.Handler {
	return func(c fiber.Ctx) error {
		uid, ok := c.Locals("uid").(string)
		if !ok || uid == "" {
			return apperr.Unauthorized("unauthorized: missing user identity context")
		}

		var role string
		const q = "SELECT COALESCE(role, 'student') FROM users WHERE id = $1"
		err := pool.QueryRow(c.Context(), q, uid).Scan(&role)
		if err != nil {
			return apperr.Unauthorized("unauthorized: administrative account verification failed")
		}

		if role != "admin" {
			return apperr.Forbidden("forbidden: administrative privileges required")
		}

		return c.Next()
	}
}
