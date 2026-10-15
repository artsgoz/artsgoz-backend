package middleware

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RequireRole — Check the User's Role from PostgreSQL
// Always use after AuthMiddleware (requires uid to be present in Locals)
func RequireRole(pool *pgxpool.Pool, allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		uid, ok := c.Locals("uid").(string)
		if !ok || uid == "" {
			return c.Status(401).JSON(fiber.Map{"error": "กรุณา login ก่อน"})
		}

		// Retrieve user role from PostgreSQL
		var role string
		err := pool.QueryRow(context.Background(),
			`SELECT role FROM users WHERE firebase_uid = $1`, uid).Scan(&role)
		if err != nil {
			return c.Status(403).JSON(fiber.Map{"error": "ไม่พบข้อมูลผู้ใช้"})
		}

		// Check if the user's role matches the specified allowedRoles
		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Locals("role", role)
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{"error": "คุณไม่มีสิทธิ์เข้าถึง"})
	}
}
