package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequireRole checks the user's role.
// It first checks if role is already stored in Gin context (e.g. from JWT),
// falling back to querying PostgreSQL by user ID.
func RequireRole(db *gorm.DB, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role == "" {
			userID := c.GetString("user_id")
			if userID == "" {
				userID = c.GetString("uid")
			}
			if userID == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
				return
			}

			var u User
			err := db.WithContext(c.Request.Context()).Select("role").Where("id = ?", userID).Take(&u).Error
			if err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user not found"})
				return
			}
			role = u.Role
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Set("role", role)
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
	}
}
