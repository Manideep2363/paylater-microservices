package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"paylater/services/auth-service/internal/client"
	"paylater/services/auth-service/internal/repository"
)

func TestUserClient_CreateAndGetByEmail(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/internal/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Token") != "tok" {
			t.Fatalf("token")
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["password"] == "" {
			t.Fatal("expected plain password")
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user_id":  7,
			"name":     body["name"],
			"email":    body["email"],
			"password": "hashed",
		})
	})
	mux.HandleFunc("/internal/users/by-email", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user_id":  7,
			"name":     "Alice",
			"email":    r.URL.Query().Get("email"),
			"password": "hashed",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := client.NewUserClient(server.URL, "tok")
	created, err := c.CreateUser(context.Background(), "Alice", "a@example.com", "secret12")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.UserID != 7 {
		t.Fatalf("id=%d", created.UserID)
	}

	got, err := c.GetUserByEmail(context.Background(), "a@example.com")
	if err != nil || got.Password != "hashed" {
		t.Fatalf("GetUserByEmail: %+v err=%v", got, err)
	}
}

func TestUserClient_Unavailable(t *testing.T) {
	c := client.NewUserClient("http://127.0.0.1:1", "tok")
	_, err := c.GetUserByEmail(context.Background(), "a@example.com")
	if !errors.Is(err, repository.ErrUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
}

func TestMerchantClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/merchants" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["commission"].(float64) != 5 {
			t.Fatalf("commission=%v", body["commission"])
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"merchant_id":           3,
			"name":                  body["name"],
			"email":                 body["email"],
			"phone":                 body["phone"],
			"password_hash":         "mhash",
			"commission_percentage": "5.00",
		})
	}))
	defer server.Close()

	c := client.NewMerchantClient(server.URL, "tok")
	m, err := c.CreateMerchant(context.Background(), "Shop", "s@example.com", "999", "secret12", 5)
	if err != nil || m.MerchantID != 3 || m.PasswordHash != "mhash" {
		t.Fatalf("CreateMerchant: %+v err=%v", m, err)
	}
}
