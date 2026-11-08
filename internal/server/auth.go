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

// AuthRequired delegates authentication to the configured token verifier.
// The verifier must load user_id and user_role from trusted profile data.
func AuthRequired(verify gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if verify == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized", "message": "authentication is unavailable"}})
			return
		}
		verify(c)
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
