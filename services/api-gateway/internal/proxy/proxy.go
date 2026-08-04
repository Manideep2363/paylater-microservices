package proxy

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// ServiceProxy wraps a ReverseProxy for one downstream service.
type ServiceProxy struct {
	Name  string
	Proxy *httputil.ReverseProxy
}

// New creates a ReverseProxy targeting baseURL with a shared timeout transport.
func New(name string, baseURL *url.URL, timeout time.Duration) *ServiceProxy {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: timeout,
	}

	proxy := httputil.NewSingleHostReverseProxy(baseURL)
	proxy.Transport = &timeoutRoundTripper{
		base:    transport,
		timeout: timeout,
	}

	originalDirector := proxy.Director
	targetHost := baseURL.Host
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetHost
		req.Header.Del("X-Internal-Token")
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf(
			"proxy_error service=%s method=%s path=%s request_id=%s",
			name,
			r.Method,
			r.URL.Path,
			r.Header.Get("X-Request-ID"),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "service unavailable",
		})
	}

	return &ServiceProxy{Name: name, Proxy: proxy}
}

type timeoutRoundTripper struct {
	base    http.RoundTripper
	timeout time.Duration
}

func (t *timeoutRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if t.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}
	return t.base.RoundTrip(req)
}
