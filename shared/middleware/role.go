package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"paylater/shared/response"
)

// RequireRole allows the request only if the JWT role is one of the given roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentRole, exists := c.Get("role")
		if !exists {
			response.Error(c, http.StatusForbidden, "role not found")
			c.Abort()
			return
		}

		allowed := false
		for _, role := range roles {
			if currentRole.(string) == role {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Error(c, http.StatusForbidden, "access denied")
			c.Abort()
			return
		}

		c.Next()
	}
}
