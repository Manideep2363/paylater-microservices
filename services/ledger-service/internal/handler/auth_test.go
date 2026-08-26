package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"paylater/services/ledger-service/internal/client"
	"paylater/services/ledger-service/internal/handler"
	"paylater/services/ledger-service/internal/repository"
	"paylater/services/ledger-service/internal/routes"
	"paylater/services/ledger-service/internal/service"
	"paylater/shared/auth"
)

type stubUsers struct {
	due, limit float64
}

func (s *stubUsers) IncreaseDue(ctx context.Context, userID int32, amount float64) error {
	_ = ctx
	_ = userID
	if s.due+amount > s.limit {
		return client.ErrInsufficientCredit
	}
	s.due += amount
	return nil
}

func (s *stubUsers) DecreaseDue(ctx context.Context, userID int32, amount float64) error {
	_ = ctx
	_ = userID
	if amount > s.due {
		return client.ErrPaymentExceedsDue
	}
	s.due -= amount
	return nil
}

type stubMerchants struct{}

func (s *stubMerchants) GetMerchantByID(ctx context.Context, merchantID int32) (client.Merchant, error) {
	_ = ctx
	if merchantID != 1 {
		return client.Merchant{}, client.ErrNotFound
	}
	return client.Merchant{MerchantID: 1, CommissionPercentage: "5.00"}, nil
}

func setupRouter(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	store := repository.NewMemoryStore()
	svc := service.NewLedgerService(store, &stubUsers{due: 0, limit: 2000}, &stubMerchants{})
	h := handler.NewLedgerHandler(svc)
	r := gin.New()
	secret := "test-secret"
	routes.Setup(r, h, secret)
	return r, secret
}

func bearer(t *testing.T, secret string, id int32, role string) string {
	t.Helper()
	token, err := auth.GenerateToken(id, "test@example.com", role, secret)
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + token
}

func TestPurchaseUsesJWTUserID(t *testing.T) {
	r, secret := setupRouter(t)
	body := map[string]any{"merchant_id": 1, "amount": 50, "user_id": 999}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/purchases", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer(t, secret, 7, "user"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUserCannotAccessAdmin(t *testing.T) {
	r, secret := setupRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/admin/purchases", nil)
	req.Header.Set("Authorization", bearer(t, secret, 1, "user"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestMerchantTransactionsUsesJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := repository.NewMemoryStore()
	_, _ = store.CreateTransaction(context.Background(), 1, 5, "10.00", "5.00", "0.50")
	_, _ = store.CreateTransaction(context.Background(), 1, 9, "20.00", "5.00", "1.00")
	svc := service.NewLedgerService(store, &stubUsers{limit: 2000}, &stubMerchants{})
	h := handler.NewLedgerHandler(svc)
	r := gin.New()
	secret := "test-secret"
	routes.Setup(r, h, secret)

	req := httptest.NewRequest(http.MethodGet, "/merchant/transactions", nil)
	req.Header.Set("Authorization", bearer(t, secret, 5, "merchant"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if len(out) != 1 {
		t.Fatalf("expected only merchant 5 txs, got %v", out)
	}
	if int(out[0]["merchant_id"].(float64)) != 5 {
		t.Fatalf("got %v", out[0])
	}
}

func TestMerchantCannotAccessOtherViaPath(t *testing.T) {
	// There is no path param for merchant id; ensure admin path is forbidden.
	r, secret := setupRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/admin/purchases", nil)
	req.Header.Set("Authorization", bearer(t, secret, 5, "merchant"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rec.Code)
	}
}
