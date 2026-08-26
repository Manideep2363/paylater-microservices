package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const defaultCORSOrigin = "http://localhost:5173"

// CORS allows the Vite React frontend (dev) to call the gateway from the browser.
//
// Development default: http://localhost:5173
// Production: set CORS_ALLOWED_ORIGIN to your real frontend origin
// (e.g. https://app.paylater.example). Do not use "*".
func CORS() gin.HandlerFunc {
	allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = defaultCORSOrigin
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == allowedOrigin {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			c.Header("Vary", "Origin")
		}

		// Browsers send OPTIONS before cross-origin POST/PUT with JSON.
		// Answer here so preflight never hits the method allowlist / reverse proxy.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
