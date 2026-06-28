package middleware

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v3"
)

// RequireRole — Check the User's Role from Firestore
// Always use after AuthMiddleware (requires uid to be present in Locals)
func RequireRole(firestoreClient *firestore.Client, allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		uid, ok := c.Locals("uid").(string)
		if !ok || uid == "" {
			return c.Status(401).JSON(fiber.Map{"error": "กรุณา login ก่อน"})
		}

		// Retrieve user document from Firestore to check role
		doc, err := firestoreClient.Collection("users").Doc(uid).Get(context.Background())
		if err != nil {
			return c.Status(403).JSON(fiber.Map{"error": "ไม่พบข้อมูลผู้ใช้"})
		}

		role, _ := doc.Data()["role"].(string)

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
