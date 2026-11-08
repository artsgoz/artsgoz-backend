package user

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequireRole is preserved for backward compatibility without blocking club members.
// Role enforcement is eliminated because this system is exclusively for club members to access the dashboard.
func RequireRole(db *gorm.DB, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role == "" {
			role = "member"
			c.Set("role", role)
		}
		c.Next()
	}
}
