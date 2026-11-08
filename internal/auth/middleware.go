package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Middleware verifies JWT or Firebase tokens before allowing access to authenticated routes.
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
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   gin.H{"code": "UNAUTHORIZED", "message": "invalid authorization format"},
				"message": "invalid authorization format",
			})
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))

		// 1. Try local JWT token first
		claims, err := service.ValidateToken(tokenString)
		if err == nil {
			c.Set("user_id", claims.UserID)
			c.Set("uid", claims.UserID) // compatibility
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			c.Next()
			return
		}

		// 2. If local JWT validation fails and Firebase is configured, try Firebase ID token
		if service.HasFirebase() {
			profile, err := service.VerifyAndGetUserFromFirebase(c.Request.Context(), tokenString)
			if err == nil && profile != nil {
				c.Set("user_id", profile.ID)
				c.Set("uid", profile.ID)
				c.Set("username", profile.Username)
				c.Set("role", profile.Role)
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   gin.H{"code": "UNAUTHORIZED", "message": "invalid or expired token"},
			"message": "invalid or expired token",
		})
	}
}
