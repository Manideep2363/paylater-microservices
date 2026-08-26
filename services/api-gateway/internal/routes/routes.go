package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"paylater/services/api-gateway/internal/middleware"
	"paylater/services/api-gateway/internal/proxy"
	"paylater/shared/response"
)

// Proxies holds one ReverseProxy per downstream service.
type Proxies struct {
	Auth     *proxy.ServiceProxy
	User     *proxy.ServiceProxy
	Merchant *proxy.ServiceProxy
	Ledger   *proxy.ServiceProxy
	Report   *proxy.ServiceProxy
}

// Setup registers the explicit public allowlist and gateway health.
func Setup(router *gin.Engine, p Proxies) {
	router.Use(
		middleware.CORS(),
		middleware.RequestID(),
		middleware.StripInternalToken(),
		middleware.BlockInternal(),
		middleware.AccessLog(),
	)

	router.GET("/health", func(c *gin.Context) {
		response.JSON(c, http.StatusOK, gin.H{
			"status":  "ok",
			"service": "api-gateway",
		})
	})

	// Auth
	forward(router, http.MethodPost, "/register", p.Auth)
	forward(router, http.MethodPost, "/login", p.Auth)
	forward(router, http.MethodPost, "/merchant/register", p.Auth)
	forward(router, http.MethodPost, "/merchant/login", p.Auth)
	forward(router, http.MethodPost, "/admin/login", p.Auth)

	// User
	forward(router, http.MethodGet, "/users/:id", p.User)
	forward(router, http.MethodGet, "/admin/users", p.User)
	forward(router, http.MethodPost, "/admin/users", p.User)

	// Merchant
	forward(router, http.MethodGet, "/merchants", p.Merchant)
	forward(router, http.MethodGet, "/merchant/profile", p.Merchant)
	forward(router, http.MethodPost, "/admin/merchants", p.Merchant)
	forward(router, http.MethodGet, "/admin/merchants", p.Merchant)
	forward(router, http.MethodGet, "/admin/merchants/:id", p.Merchant)
	forward(router, http.MethodPut, "/admin/merchants/:id/commission", p.Merchant)

	// Ledger
	forward(router, http.MethodPost, "/purchases", p.Ledger)
	forward(router, http.MethodPost, "/payments", p.Ledger)
	forward(router, http.MethodGet, "/payments", p.Ledger)
	forward(router, http.MethodGet, "/merchant/transactions", p.Ledger)
	forward(router, http.MethodGet, "/admin/purchases", p.Ledger)
	forward(router, http.MethodGet, "/admin/purchases/:id", p.Ledger)
	forward(router, http.MethodGet, "/admin/users/:id/purchases", p.Ledger)
	forward(router, http.MethodGet, "/admin/payments/:id", p.Ledger)
	forward(router, http.MethodGet, "/admin/users/:id/payments", p.Ledger)

	// Report
	forward(router, http.MethodGet, "/admin/reports/outstanding-balance", p.Report)
	forward(router, http.MethodGet, "/admin/reports/users-due", p.Report)
	forward(router, http.MethodGet, "/admin/reports/users-at-credit-limit", p.Report)
	forward(router, http.MethodGet, "/admin/reports/merchant-commissions", p.Report)
}

func forward(router *gin.Engine, method, path string, svc *proxy.ServiceProxy) {
	router.Handle(method, path, func(c *gin.Context) {
		c.Set("downstream", svc.Name)
		// Wrap writer so ReverseProxy CloseNotify works with httptest.ResponseRecorder.
		svc.Proxy.ServeHTTP(&proxyWriter{ResponseWriter: c.Writer}, c.Request)
	})
}

// proxyWriter satisfies http.CloseNotifier without panicking when the
// underlying writer (e.g. httptest.ResponseRecorder) does not.
type proxyWriter struct {
	gin.ResponseWriter
}

func (w *proxyWriter) CloseNotify() <-chan bool {
	return make(chan bool) // never closed; request lifecycle ends with ServeHTTP
}
