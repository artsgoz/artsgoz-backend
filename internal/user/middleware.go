package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequireRole checks the user's role in PostgreSQL.
// Use after authentication middleware has supplied the Firebase UID.
func RequireRole(db *gorm.DB, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("uid")
		if uid == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		// Retrieve user role from PostgreSQL
		var user User
		err := db.WithContext(c.Request.Context()).Select("role").Where("firebase_uid = ?", uid).Take(&user).Error
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user not found"})
			return
		}

		// Check if the user's role matches the specified allowedRoles
		for _, allowed := range allowedRoles {
			if user.Role == allowed {
				c.Set("role", user.Role)
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
	}
}
