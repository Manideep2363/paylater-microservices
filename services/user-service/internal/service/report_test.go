package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"paylater/services/user-service/internal/handler"
	"paylater/services/user-service/internal/repository"
	"paylater/services/user-service/internal/routes"
	"paylater/services/user-service/internal/service"
)

func TestReportOutstandingBalanceAndDues(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := service.NewUserService(repo)
	ctx := context.Background()

	total, err := svc.GetOutstandingBalance(ctx)
	if err != nil || total != "0.00" {
		t.Fatalf("empty balance got %q err=%v", total, err)
	}

	u1, _ := svc.CreateUser(ctx, "Alice", "a@example.com", "secret12")
	u2, _ := svc.CreateUser(ctx, "Bob", "b@example.com", "secret12")
	_ = repo.SetDueForTest(u1.UserID, 100)
	_ = repo.SetDueForTest(u2.UserID, 50)

	total, err = svc.GetOutstandingBalance(ctx)
	if err != nil || total != "150.00" {
		t.Fatalf("balance=%q err=%v", total, err)
	}

	dues, err := svc.GetUserOutstandingDues(ctx)
	if err != nil || len(dues) != 2 {
		t.Fatalf("dues=%v err=%v", dues, err)
	}
	if dues[0].CurrentDue != "100.00" || dues[1].CurrentDue != "50.00" {
		t.Fatalf("order wrong: %+v", dues)
	}

	// zero-due user must still appear
	u3, _ := svc.CreateUser(ctx, "Zero", "z@example.com", "secret12")
	dues, _ = svc.GetUserOutstandingDues(ctx)
	foundZero := false
	for _, d := range dues {
		if d.UserID == u3.UserID && d.CurrentDue == "0.00" {
			foundZero = true
		}
	}
	if !foundZero {
		t.Fatal("expected zero-due user in users-due report")
	}
}

func TestUsersAtCreditLimit(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := service.NewUserService(repo)
	ctx := context.Background()

	u, _ := svc.CreateUser(ctx, "Full", "f@example.com", "secret12")
	_ = repo.SetDueForTest(u.UserID, 2000)
	at, err := svc.GetUsersAtCreditLimit(ctx)
	if err != nil || len(at) != 1 || at[0].UserID != u.UserID {
		t.Fatalf("at=%+v err=%v", at, err)
	}
}

func TestInternalReportTokenRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("INTERNAL_API_TOKEN", "secret-token")

	repo := repository.NewMemoryRepository()
	svc := service.NewUserService(repo)
	h := handler.NewUserHandler(svc)
	r := gin.New()
	routes.Setup(r, h, "jwt-secret")

	req := httptest.NewRequest(http.MethodGet, "/internal/reports/outstanding-balance", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status=%d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/internal/reports/outstanding-balance", nil)
	req2.Header.Set("X-Internal-Token", "wrong")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token status=%d", rec2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/internal/reports/outstanding-balance", nil)
	req3.Header.Set("X-Internal-Token", "secret-token")
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("valid token status=%d body=%s", rec3.Code, rec3.Body.String())
	}
}
