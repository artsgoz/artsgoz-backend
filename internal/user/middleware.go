package user

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware verifies the Firebase ID token and rejects revoked tokens and disabled users.
func AuthMiddleware(firebaseAuth FirebaseAuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := firebaseAuth.VerifyIDTokenAndCheckRevoked(c.Request.Context(), idToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid, revoked, or disabled token"})
			return
		}

		c.Set("uid", token.UID)
		c.Next()
	}
}

// RequireRole checks the user's role in PostgreSQL.
// Always use after AuthMiddleware (requires uid to be present in context)
func RequireRole(db *gorm.DB, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("uid")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		uid, ok := val.(string)
		if !ok || uid == "" {
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
