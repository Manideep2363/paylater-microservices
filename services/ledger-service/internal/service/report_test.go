package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"paylater/services/ledger-service/internal/handler"
	"paylater/services/ledger-service/internal/repository"
	"paylater/services/ledger-service/internal/routes"
	"paylater/services/ledger-service/internal/service"
)

func TestMerchantCommissionSummary(t *testing.T) {
	store := repository.NewMemoryStore()
	ctx := context.Background()
	_, _ = store.CreateTransaction(ctx, 1, 10, "100.00", "5.00", "5.00")
	_, _ = store.CreateTransaction(ctx, 1, 10, "50.00", "5.00", "2.50")
	_, _ = store.CreateTransaction(ctx, 2, 20, "200.00", "10.00", "20.00")

	svc := service.NewLedgerService(store, nil, nil)
	rows, err := svc.MerchantCommissionSummary(ctx)
	if err != nil || len(rows) != 2 {
		t.Fatalf("rows=%+v err=%v", rows, err)
	}
	if rows[0].MerchantID != 20 || rows[0].TotalCommission != "20.00" {
		t.Fatalf("expected merchant 20 first: %+v", rows)
	}
	if rows[1].MerchantID != 10 || rows[1].TotalCommission != "7.50" {
		t.Fatalf("expected merchant 10 total 7.50: %+v", rows)
	}
}

func TestMerchantCommissionEmpty(t *testing.T) {
	svc := service.NewLedgerService(repository.NewMemoryStore(), nil, nil)
	rows, err := svc.MerchantCommissionSummary(context.Background())
	if err != nil || len(rows) != 0 {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
}

func TestInternalCommissionToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("INTERNAL_API_TOKEN", "ledger-secret")

	store := repository.NewMemoryStore()
	svc := service.NewLedgerService(store, nil, nil)
	h := handler.NewLedgerHandler(svc)
	r := gin.New()
	routes.Setup(r, h, "jwt")

	req := httptest.NewRequest(http.MethodGet, "/internal/reports/merchant-commissions", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/internal/reports/merchant-commissions", nil)
	req2.Header.Set("X-Internal-Token", "wrong")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/internal/reports/merchant-commissions", nil)
	req3.Header.Set("X-Internal-Token", "ledger-secret")
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec3.Code, rec3.Body.String())
	}
}
