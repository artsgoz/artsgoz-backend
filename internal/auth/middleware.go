package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Middleware verifies Firebase ID tokens and loads authorization roles from PostgreSQL.
func Middleware(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   gin.H{"code": "UNAUTHORIZED", "message": "authentication required"},
				"message": "authentication required",
			})
			return
		}
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   gin.H{"code": "UNAUTHORIZED", "message": "invalid authorization format"},
				"message": "invalid authorization format",
			})
			return
		}

		tokenString := parts[1]

		profile, err := service.VerifyAndGetUserFromFirebase(c.Request.Context(), tokenString)
		if err == nil && profile != nil {
			c.Set("user_id", profile.ID)
			c.Set("uid", profile.ID)
			c.Set("firebase_uid", profile.FirebaseUID)
			c.Set("username", profile.Username)
			c.Set("role", profile.Role)
			c.Set("user_role", profile.Role)
			c.Next()
			return
		}

		writeAuthError(c, err)
		c.Abort()

	}
}
