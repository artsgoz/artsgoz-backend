package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Role represents user roles for authorization.
type Role string

const (
	RoleEditor     Role = "editor"
	RoleSuperAdmin Role = "superadmin"
)

// contextKeyRole is the gin context key for the authenticated user's role.
const contextKeyRole = "user_role"

// contextKeyUserID is the gin context key for the authenticated user's ID.
const contextKeyUserID = "user_id"

// AuthRequired is a middleware that validates Bearer Token authentication.
// It extracts the token from the Authorization header and sets user context.
// TODO: Integrate with a proper JWT library (e.g. golang-jwt/jwt/v5) for token
// verification once the auth module is fully implemented.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "unauthorized",
					"message": "authorization header is required",
				},
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "unauthorized",
					"message": "invalid authorization header format, expected 'Bearer <token>'",
				},
			})
			return
		}

		token := strings.TrimSpace(parts[1])

		// TODO: Replace with real JWT verification.
		// For now, set token as user_id and default role.
		// In production, decode the JWT claims to extract user_id and role.
		c.Set(contextKeyUserID, token)
		c.Set(contextKeyRole, string(RoleEditor))

		c.Next()
	}
}

// RequireRoles is a middleware that checks whether the authenticated user has one of the allowed roles.
func RequireRoles(allowed ...Role) gin.HandlerFunc {
	allowedSet := make(map[Role]bool, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = true
	}

	return func(c *gin.Context) {
		roleStr, exists := c.Get(contextKeyRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "forbidden",
					"message": "insufficient permissions",
				},
			})
			return
		}

		userRole := Role(strings.ToLower(roleStr.(string)))
		if !allowedSet[userRole] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "forbidden",
					"message": "insufficient permissions, required roles: " + joinRoles(allowed),
				},
			})
			return
		}

		c.Next()
	}
}

func joinRoles(roles []Role) string {
	s := make([]string, len(roles))
	for i, r := range roles {
		s[i] = string(r)
	}
	return strings.Join(s, ", ")
}

