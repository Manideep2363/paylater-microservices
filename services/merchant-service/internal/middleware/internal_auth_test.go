package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireInternalToken_FailClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("rejects when token env is empty", func(t *testing.T) {
		t.Setenv("INTERNAL_API_TOKEN", "")
		rec := runInternalRequest(t, "")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("rejects when header missing", func(t *testing.T) {
		t.Setenv("INTERNAL_API_TOKEN", "secret-token")
		rec := runInternalRequest(t, "")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("rejects when header incorrect", func(t *testing.T) {
		t.Setenv("INTERNAL_API_TOKEN", "secret-token")
		rec := runInternalRequest(t, "wrong-token")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("allows matching token", func(t *testing.T) {
		t.Setenv("INTERNAL_API_TOKEN", "secret-token")
		rec := runInternalRequest(t, "secret-token")
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
	})
}

func runInternalRequest(t *testing.T, tokenHeader string) *httptest.ResponseRecorder {
	t.Helper()

	r := gin.New()
	r.Use(RequireInternalToken())
	r.GET("/internal/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/internal/ping", nil)
	if tokenHeader != "" {
		req.Header.Set("X-Internal-Token", tokenHeader)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}
