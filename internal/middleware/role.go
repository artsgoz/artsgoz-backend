package middleware

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v3"
)

// RequireRole — ตรวจสอบ Role ของ User จาก Firestore
// ใช้ต่อจาก AuthMiddleware เสมอ (ต้องมี uid ใน Locals แล้ว)
func RequireRole(firestoreClient *firestore.Client, allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		uid, ok := c.Locals("uid").(string)
		if !ok || uid == "" {
			return c.Status(401).JSON(fiber.Map{"error": "กรุณา login ก่อน"})
		}

		// ดึง user document จาก Firestore เพื่อเช็ค role
		doc, err := firestoreClient.Collection("users").Doc(uid).Get(context.Background())
		if err != nil {
			return c.Status(403).JSON(fiber.Map{"error": "ไม่พบข้อมูลผู้ใช้"})
		}

		role, _ := doc.Data()["role"].(string)

		// เช็คว่า role ของ user ตรงกับ allowedRoles ที่กำหนดไว้หรือไม่
		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Locals("role", role)
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{"error": "คุณไม่มีสิทธิ์เข้าถึง"})
	}
}
