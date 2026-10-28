package user

import (
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthMiddleware — Verifies Firebase ID Token from Authorization header
func AuthMiddleware(firebaseAuth *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "กรุณา login ก่อน"})
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := firebaseAuth.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token ไม่ถูกต้อง"})
			return
		}

		c.Set("uid", token.UID)
		c.Next()
	}
}

// RequireRole — Check the User's Role from PostgreSQL
// Always use after AuthMiddleware (requires uid to be present in context)
func RequireRole(pool *pgxpool.Pool, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("uid")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "กรุณา login ก่อน"})
			return
		}

		uid, ok := val.(string)
		if !ok || uid == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "กรุณา login ก่อน"})
			return
		}

		// Retrieve user role from PostgreSQL
		var role string
		err := pool.QueryRow(c.Request.Context(),
			`SELECT role FROM users WHERE firebase_uid = $1`, uid).Scan(&role)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "ไม่พบข้อมูลผู้ใช้"})
			return
		}

		// Check if the user's role matches the specified allowedRoles
		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Set("role", role)
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "คุณไม่มีสิทธิ์เข้าถึง"})
	}
}
