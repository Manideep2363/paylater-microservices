package routes_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"paylater/services/api-gateway/internal/proxy"
	"paylater/services/api-gateway/internal/routes"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type capture struct {
	method      string
	path        string
	rawQuery    string
	body        string
	auth        string
	contentType string
	accept      string
	requestID   string
	internalTok string
	hits        int32
}

func startBackend(t *testing.T, status int, respBody string, cap *capture) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&cap.hits, 1)
		body, _ := io.ReadAll(r.Body)
		cap.method = r.Method
		cap.path = r.URL.Path
		cap.rawQuery = r.URL.RawQuery
		cap.body = string(body)
		cap.auth = r.Header.Get("Authorization")
		cap.contentType = r.Header.Get("Content-Type")
		cap.accept = r.Header.Get("Accept")
		cap.requestID = r.Header.Get("X-Request-ID")
		cap.internalTok = r.Header.Get("X-Internal-Token")

		if status > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(respBody))
		}
	}))
}

func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	return u
}

func newGateway(t *testing.T, auth, user, merchant, ledger, report string, timeout time.Duration) *gin.Engine {
	t.Helper()
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	p := routes.Proxies{
		Auth:     proxy.New("auth", mustURL(t, auth), timeout),
		User:     proxy.New("user", mustURL(t, user), timeout),
		Merchant: proxy.New("merchant", mustURL(t, merchant), timeout),
		Ledger:   proxy.New("ledger", mustURL(t, ledger), timeout),
		Report:   proxy.New("report", mustURL(t, report), timeout),
	}
	r := gin.New()
	routes.Setup(r, p)
	return r
}

func unusedURL(t *testing.T) string {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected hit on unused backend: %s %s", r.Method, r.URL.Path)
		http.Error(w, "unexpected", 500)
	}))
	t.Cleanup(s.Close)
	return s.URL
}

func TestRouting_Allowlist(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		path    string
		service string // which capture should be hit
	}{
		{"register", http.MethodPost, "/register", "auth"},
		{"login", http.MethodPost, "/login", "auth"},
		{"purchases", http.MethodPost, "/purchases", "ledger"},
		{"payments", http.MethodPost, "/payments", "ledger"},
		{"merchant_profile", http.MethodGet, "/merchant/profile", "merchant"},
		{"admin_users", http.MethodGet, "/admin/users", "user"},
		{"admin_merchants", http.MethodGet, "/admin/merchants", "merchant"},
		{"report_outstanding", http.MethodGet, "/admin/reports/outstanding-balance", "report"},
		{"report_users_due", http.MethodGet, "/admin/reports/users-due", "report"},
		{"report_credit_limit", http.MethodGet, "/admin/reports/users-at-credit-limit", "report"},
		{"report_commissions", http.MethodGet, "/admin/reports/merchant-commissions", "report"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var authCap, userCap, merchantCap, ledgerCap, reportCap capture
			auth := startBackend(t, 200, `{"ok":"auth"}`, &authCap)
			user := startBackend(t, 200, `{"ok":"user"}`, &userCap)
			merchant := startBackend(t, 200, `{"ok":"merchant"}`, &merchantCap)
			ledger := startBackend(t, 200, `{"ok":"ledger"}`, &ledgerCap)
			report := startBackend(t, 200, `{"ok":"report"}`, &reportCap)
			defer auth.Close()
			defer user.Close()
			defer merchant.Close()
			defer ledger.Close()
			defer report.Close()

			gw := newGateway(t, auth.URL, user.URL, merchant.URL, ledger.URL, report.URL, 0)
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			gw.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}

			hits := map[string]int32{
				"auth":     atomic.LoadInt32(&authCap.hits),
				"user":     atomic.LoadInt32(&userCap.hits),
				"merchant": atomic.LoadInt32(&merchantCap.hits),
				"ledger":   atomic.LoadInt32(&ledgerCap.hits),
				"report":   atomic.LoadInt32(&reportCap.hits),
			}
			for name, n := range hits {
				want := int32(0)
				if name == tc.service {
					want = 1
				}
				if n != want {
					t.Errorf("%s hits=%d want=%d", name, n, want)
				}
			}
		})
	}
}

func TestSecurity_InternalBlocked(t *testing.T) {
	var hit int32
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hit, 1)
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"leaked":true}`))
	}))
	defer backend.Close()

	gw := newGateway(t, backend.URL, backend.URL, backend.URL, backend.URL, backend.URL, 0)

	paths := []string{
		"/internal",
		"/internal/",
		"/internal/users",
		"/internal/anything",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			before := atomic.LoadInt32(&hit)
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			gw.ServeHTTP(w, req)
			if w.Code != http.StatusNotFound {
				t.Fatalf("status=%d want=404", w.Code)
			}
			if atomic.LoadInt32(&hit) != before {
				t.Fatal("backend was contacted for /internal path")
			}
		})
	}
}

func TestSecurity_StripInternalToken_PreserveAuthorization(t *testing.T) {
	var cap capture
	backend := startBackend(t, 200, `{"ok":true}`, &cap)
	defer backend.Close()

	gw := newGateway(t, backend.URL, unusedURL(t), unusedURL(t), unusedURL(t), unusedURL(t), 0)
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"a@b.com"}`))
	req.Header.Set("Authorization", "Bearer secret-jwt")
	req.Header.Set("X-Internal-Token", "should-be-stripped")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
	if cap.auth != "Bearer secret-jwt" {
		t.Fatalf("Authorization not preserved: %q", cap.auth)
	}
	if cap.internalTok != "" {
		t.Fatalf("X-Internal-Token leaked to backend: %q", cap.internalTok)
	}
}

func TestProxy_PreservesMethodQueryBody(t *testing.T) {
	var cap capture
	backend := startBackend(t, 200, `{"ok":true}`, &cap)
	defer backend.Close()

	gw := newGateway(t, unusedURL(t), unusedURL(t), unusedURL(t), backend.URL, unusedURL(t), 0)
	body := `{"amount":10.5,"merchant_id":1}`
	req := httptest.NewRequest(http.MethodPost, "/purchases?trace=1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
	if cap.method != http.MethodPost {
		t.Fatalf("method=%s", cap.method)
	}
	if cap.path != "/purchases" {
		t.Fatalf("path=%s", cap.path)
	}
	if cap.rawQuery != "trace=1" {
		t.Fatalf("query=%s", cap.rawQuery)
	}
	if cap.body != body {
		t.Fatalf("body=%q", cap.body)
	}
	if cap.contentType != "application/json" {
		t.Fatalf("content-type=%s", cap.contentType)
	}
	if cap.accept != "application/json" {
		t.Fatalf("accept=%s", cap.accept)
	}
}

func TestProxy_DownstreamStatusPassthrough(t *testing.T) {
	statuses := []int{400, 401, 403, 500}
	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			respBody := `{"error":"downstream"}`
			var cap capture
			backend := startBackend(t, status, respBody, &cap)
			defer backend.Close()

			gw := newGateway(t, backend.URL, unusedURL(t), unusedURL(t), unusedURL(t), unusedURL(t), 0)
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			gw.ServeHTTP(w, req)

			if w.Code != status {
				t.Fatalf("status=%d want=%d", w.Code, status)
			}
			if strings.TrimSpace(w.Body.String()) != respBody {
				t.Fatalf("body=%q", w.Body.String())
			}
		})
	}
}

func TestProxy_Unavailable503(t *testing.T) {
	// Closed listener URL — connection refused.
	badURL := "http://127.0.0.1:1"
	gw := newGateway(t, badURL, unusedURL(t), unusedURL(t), unusedURL(t), unusedURL(t), 500*time.Millisecond)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d want=503 body=%s", w.Code, w.Body.String())
	}
	var payload map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json: %v", err)
	}
	if payload["error"] != "service unavailable" {
		t.Fatalf("payload=%v", payload)
	}
	body := w.Body.String()
	if strings.Contains(body, "127.0.0.1") || strings.Contains(strings.ToLower(body), "connection") {
		t.Fatalf("leaked network details: %s", body)
	}
}

func TestProxy_Timeout503(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer backend.Close()

	gw := newGateway(t, backend.URL, unusedURL(t), unusedURL(t), unusedURL(t), unusedURL(t), 50*time.Millisecond)
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d want=503 body=%s", w.Code, w.Body.String())
	}
	var payload map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json: %v", err)
	}
	if payload["error"] != "service unavailable" {
		t.Fatalf("payload=%v", payload)
	}
}

func TestRequestID_GeneratedAndForwarded(t *testing.T) {
	var cap capture
	backend := startBackend(t, 200, `{"ok":true}`, &cap)
	defer backend.Close()

	gw := newGateway(t, backend.URL, unusedURL(t), unusedURL(t), unusedURL(t), unusedURL(t), 0)
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	got := w.Header().Get("X-Request-ID")
	if got == "" {
		t.Fatal("missing X-Request-ID on response")
	}
	if cap.requestID != got {
		t.Fatalf("downstream got %q client got %q", cap.requestID, got)
	}
}

func TestRequestID_Preserved(t *testing.T) {
	var cap capture
	backend := startBackend(t, 200, `{"ok":true}`, &cap)
	defer backend.Close()

	gw := newGateway(t, backend.URL, unusedURL(t), unusedURL(t), unusedURL(t), unusedURL(t), 0)
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.Header.Set("X-Request-ID", "client-req-123")
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != "client-req-123" {
		t.Fatalf("response id=%q", w.Header().Get("X-Request-ID"))
	}
	if cap.requestID != "client-req-123" {
		t.Fatalf("downstream id=%q", cap.requestID)
	}
}

func TestHealth_WorksWhenDownstreamsDown(t *testing.T) {
	bad := "http://127.0.0.1:1"
	gw := newGateway(t, bad, bad, bad, bad, bad, 100*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
	var payload map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json: %v", err)
	}
	if payload["status"] != "ok" || payload["service"] != "api-gateway" {
		t.Fatalf("payload=%v", payload)
	}
}

func TestUnregisteredRoute_404(t *testing.T) {
	var hit int32
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hit, 1)
	}))
	defer backend.Close()

	gw := newGateway(t, backend.URL, backend.URL, backend.URL, backend.URL, backend.URL, 0)
	req := httptest.NewRequest(http.MethodGet, "/not-a-real-route", nil)
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status=%d", w.Code)
	}
	if atomic.LoadInt32(&hit) != 0 {
		t.Fatal("backend contacted for unregistered route")
	}
}
