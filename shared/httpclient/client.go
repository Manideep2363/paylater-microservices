package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultTimeout = 10 * time.Second

// Client is a reusable JSON HTTP client for service-to-service REST calls.
type Client struct {
	BaseURL        string
	HTTPClient     *http.Client
	DefaultHeaders map[string]string
}

// New creates a Client with a default timeout.
func New(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
		},
		DefaultHeaders: make(map[string]string),
	}
}

// WithTimeout overrides the HTTP client timeout.
func (c *Client) WithTimeout(d time.Duration) *Client {
	c.HTTPClient.Timeout = d
	return c
}

// WithHeader sets a default header sent on every request.
func (c *Client) WithHeader(key, value string) *Client {
	if c.DefaultHeaders == nil {
		c.DefaultHeaders = make(map[string]string)
	}
	c.DefaultHeaders[key] = value
	return c
}

// Request describes a JSON HTTP call.
type Request struct {
	Method  string
	Path    string
	Query   url.Values
	Headers map[string]string
	Body    any
	Out     any // optional; unmarshaled from response body on 2xx
}

// HTTPError is returned for non-2xx responses.
type HTTPError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *HTTPError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("http status %d", e.StatusCode)
}

// DoJSON performs an HTTP request with JSON encode/decode and shared error handling.
func (c *Client) DoJSON(ctx context.Context, req Request) error {
	fullURL, err := c.buildURL(req.Path, req.Query)
	if err != nil {
		return err
	}

	var bodyReader io.Reader
	if req.Body != nil {
		payload, err := json.Marshal(req.Body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("Accept", "application/json")

	for k, v := range c.DefaultHeaders {
		httpReq.Header.Set(k, v)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    extractErrorMessage(respBody),
			Body:       string(respBody),
		}
	}

	if req.Out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, req.Out); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) buildURL(path string, query url.Values) (string, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}
	return u.String(), nil
}

func extractErrorMessage(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && payload.Error != "" {
		return payload.Error
	}
	return strings.TrimSpace(string(body))
}

// IsNotFound reports whether err is an HTTP 404.
func IsNotFound(err error) bool {
	var httpErr *HTTPError
	if ok := asHTTPError(err, &httpErr); ok {
		return httpErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsConflictOrBadRequest reports 400/409 style client errors from downstream.
func IsBadRequest(err error) bool {
	var httpErr *HTTPError
	if ok := asHTTPError(err, &httpErr); ok {
		return httpErr.StatusCode == http.StatusBadRequest
	}
	return false
}

func asHTTPError(err error, target **HTTPError) bool {
	if err == nil {
		return false
	}
	httpErr, ok := err.(*HTTPError)
	if !ok {
		return false
	}
	*target = httpErr
	return true
}
