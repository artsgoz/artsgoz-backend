package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	loginService    LoginService
	registerService RegisterService
}

func NewAuthHandler(l LoginService, r RegisterService) *AuthHandler {
	return &AuthHandler{
		loginService:    l,
		registerService: r,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	uid, role, err := h.loginService.Authenticate(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"uid": uid, "role": role})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.registerService.Register(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "สมัครสมาชิกสำเร็จ"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	val, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "กรุณา login ก่อน"})
		return
	}

	uid, ok := val.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "กรุณา login ก่อน"})
		return
	}

	user, err := h.loginService.GetUserProfile(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลผู้ใช้ได้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"uid":     user.FirebaseUID,
		"role":    user.Role,
		"message": "authenticated successfully",
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	if val, exists := c.Get("uid"); exists {
		if uid, ok := val.(string); ok && uid != "" {
			_ = h.loginService.Logout(c.Request.Context(), uid)
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "ออกจากระบบสำเร็จ"})
}

type AdminHandler struct {
	adminService AdminService
}

func NewAdminHandler(u AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: u,
	}
}

func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	users, err := h.adminService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลผู้ใช้ได้: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุ uid"})
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.adminService.UpdateUserRole(uid, req.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "อัปเดตบทบาทผู้ใช้สำเร็จ"})
}

func (h *AdminHandler) GetProfile(c *gin.Context) {
	val, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "กรุณา login ก่อน"})
		return
	}

	uid, ok := val.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "กรุณา login ก่อน"})
		return
	}

	user, err := h.adminService.GetProfile(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลผู้ใช้ได้: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"firebase_uid": user.FirebaseUID,
		"uid":          user.FirebaseUID,
		"email":        user.Email,
		"role":         user.Role,
		"created_at":   user.CreatedAt,
		"updated_at":   user.UpdatedAt,
	})
}

func (h *AdminHandler) UpdateProfile(c *gin.Context) {
	val, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "กรุณา login ก่อน"})
		return
	}

	uid, ok := val.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "กรุณา login ก่อน"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.adminService.UpdateProfile(uid, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "แก้ไขข้อมูลส่วนตัวสำเร็จ",
		"firebase_uid": user.FirebaseUID,
		"uid":          user.FirebaseUID,
		"email":        user.Email,
		"role":         user.Role,
		"updated_at":   user.UpdatedAt,
	})
}
