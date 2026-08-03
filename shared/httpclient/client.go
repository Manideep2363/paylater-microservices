package httpclient

import (
	"net/http"
	"time"
)

// Client is a thin HTTP wrapper for future service-to-service REST calls.
// Phase 0 only provides the type; no inter-service calls are implemented yet.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// New creates a Client with a sensible default timeout.
func New(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
