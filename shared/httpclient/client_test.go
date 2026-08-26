package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoJSON_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Token") != "secret" {
			t.Fatalf("missing internal token")
		}
		if r.URL.Path != "/internal/users/by-email" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if r.URL.Query().Get("email") != "a@example.com" {
			t.Fatalf("email query=%s", r.URL.Query().Get("email"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user_id": 1,
			"email":   "a@example.com",
		})
	}))
	defer server.Close()

	client := New(server.URL).WithHeader("X-Internal-Token", "secret")
	var out map[string]any
	err := client.DoJSON(context.Background(), Request{
		Method: http.MethodGet,
		Path:   "/internal/users/by-email",
		Query:  map[string][]string{"email": {"a@example.com"}},
		Out:    &out,
	})
	if err != nil {
		t.Fatalf("DoJSON: %v", err)
	}
	if out["email"] != "a@example.com" {
		t.Fatalf("out=%v", out)
	}
}

func TestDoJSON_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"user not found"}`))
	}))
	defer server.Close()

	client := New(server.URL)
	err := client.DoJSON(context.Background(), Request{
		Method: http.MethodGet,
		Path:   "/internal/users/by-email",
	})
	if !IsNotFound(err) {
		t.Fatalf("expected not found, got %v", err)
	}
	httpErr, ok := err.(*HTTPError)
	if !ok || httpErr.Message != "user not found" {
		t.Fatalf("httpErr=%v", err)
	}
}
