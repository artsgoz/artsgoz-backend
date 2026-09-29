package rest

import (
	"strings"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/application"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	loginUsecase    application.LoginUsecase
	registerUsecase application.RegisterUsecase
}

func NewAuthHandler(l application.LoginUsecase, r application.RegisterUsecase) *AuthHandler {
	return &AuthHandler{
		loginUsecase:    l,
		registerUsecase: r,
	}
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req application.LoginRequest

	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	if strings.TrimSpace(req.IDToken) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "กรุณาระบุ id_token"})
	}

	uid, role, err := h.loginUsecase.Authenticate(req)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"uid": uid, "role": role})
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req application.RegisterRequest

	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	if err := h.registerUsecase.Register(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"message": "สมัครสมาชิกสำเร็จ"})
}

func (h *AuthHandler) Me(c fiber.Ctx) error {
	uid, ok := c.Locals("uid").(string)
	if !ok || uid == "" {
		return c.Status(401).JSON(fiber.Map{"error": "กรุณา login ก่อน"})
	}

	user, err := h.loginUsecase.GetUserProfile(uid)
	if err != nil {
		if err.Error() == "ไม่พบผู้ใช้" || err.Error() == "ไม่พบผู้ใช้ในระบบ" {
			return c.Status(404).JSON(fiber.Map{"error": "ไม่พบข้อมูลผู้ใช้"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "ไม่สามารถดึงข้อมูลผู้ใช้ได้"})
	}

	return c.JSON(fiber.Map{
		"uid":     user.FirebaseUID,
		"role":    user.Role,
		"message": "authenticated successfully",
	})
}

type AdminHandler struct {
	adminUsecase application.AdminUsecase
}

func NewAdminHandler(u application.AdminUsecase) *AdminHandler {
	return &AdminHandler{
		adminUsecase: u,
	}
}

func (h *AdminHandler) GetAllUsers(c fiber.Ctx) error {
	users, err := h.adminUsecase.GetAllUsers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ไม่สามารถดึงข้อมูลผู้ใช้ได้: " + err.Error()})
	}
	return c.JSON(users)
}

func (h *AdminHandler) UpdateUserRole(c fiber.Ctx) error {
	uid := strings.TrimSpace(c.Params("uid"))
	if uid == "" {
		return c.Status(400).JSON(fiber.Map{"error": "กรุณาระบุ uid"})
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	req.Role = strings.TrimSpace(req.Role)
	if req.Role == "" {
		return c.Status(400).JSON(fiber.Map{"error": "กรุณาระบุ role"})
	}

	if err := h.adminUsecase.UpdateUserRole(uid, req.Role); err != nil {
		if err.Error() == "บทบาทผู้ใช้งานไม่ถูกต้อง" || err.Error() == "กรุณาระบุ uid" {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "อัปเดตบทบาทผู้ใช้งานสำเร็จ"})
}
