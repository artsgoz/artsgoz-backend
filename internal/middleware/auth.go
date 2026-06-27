package middleware

import (
	"context"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v3"
)

// AuthMiddleware — ตรวจสอบ Firebase ID Token จาก Authorization header
func AuthMiddleware(firebaseAuth *auth.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "กรุณา login ก่อน"})
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{"error": "invalid authorization format"})
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := firebaseAuth.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "token ไม่ถูกต้อง"})
		}

		c.Locals("uid", token.UID)
		return c.Next()
	}
}
