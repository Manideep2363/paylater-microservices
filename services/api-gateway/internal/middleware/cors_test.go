package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"paylater/services/api-gateway/internal/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCORS_AllowsViteOriginAndPreflight(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGIN", "http://localhost:5173")

	r := gin.New()
	r.Use(middleware.CORS())
	r.POST("/admin/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"token": "x"})
	})

	// Preflight
	pre := httptest.NewRequest(http.MethodOptions, "/admin/login", nil)
	pre.Header.Set("Origin", "http://localhost:5173")
	pre.Header.Set("Access-Control-Request-Method", "POST")
	pre.Header.Set("Access-Control-Request-Headers", "content-type,authorization")
	pw := httptest.NewRecorder()
	r.ServeHTTP(pw, pre)

	if pw.Code != http.StatusNoContent {
		t.Fatalf("preflight status=%d want=204", pw.Code)
	}
	if got := pw.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("Allow-Origin=%q", got)
	}
	if got := pw.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("missing Allow-Methods")
	}
	if got := pw.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatal("missing Allow-Headers")
	}

	// Actual POST
	req := httptest.NewRequest(http.MethodPost, "/admin/login", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("post status=%d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("Allow-Origin=%q", got)
	}
}

func TestCORS_RejectsOtherOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGIN", "http://localhost:5173")

	r := gin.New()
	r.Use(middleware.CORS())
	r.POST("/login", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.Header.Set("Origin", "http://evil.example")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected Allow-Origin=%q", got)
	}
}
