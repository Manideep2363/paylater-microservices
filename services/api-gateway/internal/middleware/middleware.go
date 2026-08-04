package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-ID"

// BlockInternal returns 404 for any /internal path (defense in depth).
func BlockInternal() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/internal" || strings.HasPrefix(path, "/internal/") {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Next()
	}
}

// StripInternalToken removes client-supplied X-Internal-Token before handlers/proxy.
func StripInternalToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Header.Del("X-Internal-Token")
		c.Next()
	}
}

// RequestID ensures every request has an X-Request-ID (generated if absent).
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader(requestIDHeader))
		if id == "" || !isReasonableRequestID(id) {
			id = newRequestID()
		}
		c.Request.Header.Set(requestIDHeader, id)
		c.Writer.Header().Set(requestIDHeader, id)
		c.Set("request_id", id)
		c.Set("downstream", "")
		c.Next()
	}
}

// AccessLog logs method, path, status, duration, request_id, and optional downstream.
// Never logs Authorization, tokens, or bodies.
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		downstream, _ := c.Get("downstream")
		requestID, _ := c.Get("request_id")
		log.Printf(
			"access method=%s path=%s status=%d duration_ms=%d request_id=%v downstream=%v",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start).Milliseconds(),
			requestID,
			downstream,
		)
	}
}

func isReasonableRequestID(id string) bool {
	if len(id) == 0 || len(id) > 128 {
		return false
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405.000000000")))
	}
	return hex.EncodeToString(b[:])
}
