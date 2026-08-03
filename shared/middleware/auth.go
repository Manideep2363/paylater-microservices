package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"paylater/shared/auth"
	"paylater/shared/response"
)

// AuthMiddleware validates a Bearer JWT and stores claims on the Gin context.
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "authorization header is required")
			c.Abort()
			return
		}

		const bearer = "Bearer "
		if !strings.HasPrefix(authHeader, bearer) {
			response.Error(c, http.StatusUnauthorized, "invalid authorization header")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearer)
		claims, err := auth.ValidateToken(tokenString, secret)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}
