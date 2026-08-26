package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"paylater/shared/response"
)

// RequireInternalToken gates /internal/* routes (fail closed).
func RequireInternalToken() gin.HandlerFunc {
	expected := os.Getenv("INTERNAL_API_TOKEN")

	return func(c *gin.Context) {
		if expected == "" {
			response.Error(c, http.StatusUnauthorized, "internal api token is not configured")
			c.Abort()
			return
		}

		provided := c.GetHeader("X-Internal-Token")
		if provided == "" {
			response.Error(c, http.StatusUnauthorized, "internal token is required")
			c.Abort()
			return
		}

		if provided != expected {
			response.Error(c, http.StatusUnauthorized, "invalid internal token")
			c.Abort()
			return
		}

		c.Next()
	}
}
