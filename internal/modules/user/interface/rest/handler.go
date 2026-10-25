package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/application"
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

func (h *AuthHandler) Login(c *gin.Context) {
	var req application.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	uid, role, err := h.loginUsecase.Authenticate(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"uid": uid, "role": role})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req application.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.registerUsecase.Register(req); err != nil {
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

	user, err := h.loginUsecase.GetUserProfile(uid)
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

type AdminHandler struct {
	adminUsecase application.AdminUsecase
}

func NewAdminHandler(u application.AdminUsecase) *AdminHandler {
	return &AdminHandler{
		adminUsecase: u,
	}
}

func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	users, err := h.adminUsecase.GetAllUsers()
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

	if err := h.adminUsecase.UpdateUserRole(uid, req.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "อัปเดตบทบาทผู้ใช้สำเร็จ"})
}
