package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "INTERNAL_SERVER_ERROR", "message": "failed to retrieve user data: " + err.Error()},
			"message": "failed to retrieve user data: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
		"users":   users,
	})
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		id = c.Param("uid")
	}
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "user ID is required"},
			"message": "user ID is required",
		})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "invalid request body: " + err.Error()},
			"message": "invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.adminService.UpdateUserRole(id, req.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": err.Error()},
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user role updated successfully",
	})
}

func (h *AdminHandler) GetProfile(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   gin.H{"code": "UNAUTHORIZED", "message": "authentication required"},
			"message": "authentication required",
		})
		return
	}

	user, err := h.adminService.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "INTERNAL_SERVER_ERROR", "message": "failed to retrieve user data: " + err.Error()},
			"message": "failed to retrieve user data: " + err.Error(),
		})
		return
	}

	userData := gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"avatar_url": user.AvatarURL,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"avatar_url": user.AvatarURL,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
		"user":       userData,
		"data":       userData,
	})
}

func (h *AdminHandler) UpdateProfile(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   gin.H{"code": "UNAUTHORIZED", "message": "authentication required"},
			"message": "authentication required",
		})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "invalid request body: " + err.Error()},
			"message": "invalid request body: " + err.Error(),
		})
		return
	}

	user, err := h.adminService.UpdateProfile(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": err.Error()},
			"message": err.Error(),
		})
		return
	}

	userData := gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"avatar_url": user.AvatarURL,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "profile updated successfully",
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"avatar_url": user.AvatarURL,
		"updated_at": user.UpdatedAt,
		"user":       userData,
		"data":       userData,
	})
}

func (h *AdminHandler) ChangePassword(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   gin.H{"code": "UNAUTHORIZED", "message": "authentication required"},
			"message": "authentication required",
		})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "invalid request body: " + err.Error()},
			"message": "invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.adminService.ChangePassword(userID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": err.Error()},
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "password updated successfully",
	})
}

func getUserIDFromContext(c *gin.Context) string {
	if val := c.GetString("user_id"); val != "" {
		return val
	}
	return c.GetString("uid")
}
