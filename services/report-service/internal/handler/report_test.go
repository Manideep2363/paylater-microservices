package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"paylater/services/report-service/internal/client"
	"paylater/services/report-service/internal/handler"
	"paylater/services/report-service/internal/routes"
	"paylater/services/report-service/internal/service"
	"paylater/shared/auth"
)

type fakeUsers struct {
	balance string
	dues    []client.UserDue
	limits  []client.UserAtLimit
	err     error
}

func (f *fakeUsers) OutstandingBalance(ctx context.Context) (string, error) {
	_ = ctx
	if f.err != nil {
		return "", f.err
	}
	return f.balance, nil
}

func (f *fakeUsers) UsersDue(ctx context.Context) ([]client.UserDue, error) {
	_ = ctx
	if f.err != nil {
		return nil, f.err
	}
	if f.dues == nil {
		return []client.UserDue{}, nil
	}
	return f.dues, nil
}

func (f *fakeUsers) UsersAtCreditLimit(ctx context.Context) ([]client.UserAtLimit, error) {
	_ = ctx
	if f.err != nil {
		return nil, f.err
	}
	if f.limits == nil {
		return []client.UserAtLimit{}, nil
	}
	return f.limits, nil
}

type fakeLedger struct {
	rows []client.MerchantCommission
	err  error
}

func (f *fakeLedger) MerchantCommissions(ctx context.Context) ([]client.MerchantCommission, error) {
	_ = ctx
	if f.err != nil {
		return nil, f.err
	}
	if f.rows == nil {
		return []client.MerchantCommission{}, nil
	}
	return f.rows, nil
}

func setup(t *testing.T, users client.UserReportsAPI, ledger client.LedgerReportsAPI) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	svc := service.NewReportService(users, ledger)
	h := handler.NewReportHandler(svc)
	r := gin.New()
	secret := "report-secret"
	routes.Setup(r, h, secret)
	return r, secret
}

func adminAuth(t *testing.T, secret string) string {
	t.Helper()
	tok, err := auth.GenerateToken(0, "admin@paylater.com", "admin", secret)
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + tok
}

func userAuth(t *testing.T, secret string) string {
	t.Helper()
	tok, err := auth.GenerateToken(1, "u@example.com", "user", secret)
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + tok
}

func TestOutstandingBalanceSuccess(t *testing.T) {
	r, secret := setup(t, &fakeUsers{balance: "123.45"}, &fakeLedger{})
	req := httptest.NewRequest(http.MethodGet, "/admin/reports/outstanding-balance", nil)
	req.Header.Set("Authorization", adminAuth(t, secret))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["total_outstanding_balance"] != "123.45" {
		t.Fatalf("body=%v", body)
	}
}

func TestUsersDueAndCreditLimitAndCommissions(t *testing.T) {
	users := &fakeUsers{
		dues:   []client.UserDue{{UserID: 1, Name: "A", CurrentDue: "10.00"}},
		limits: []client.UserAtLimit{{UserID: 1, Name: "A", CreditLimit: "10.00", CurrentDue: "10.00"}},
	}
	ledger := &fakeLedger{rows: []client.MerchantCommission{{MerchantID: 9, TotalCommission: "5.00"}}}
	r, secret := setup(t, users, ledger)

	for _, path := range []string{
		"/admin/reports/users-due",
		"/admin/reports/users-at-credit-limit",
		"/admin/reports/merchant-commissions",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", adminAuth(t, secret))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", path, rec.Code, rec.Body.String())
		}
	}
}

func TestEmptyResponses(t *testing.T) {
	r, secret := setup(t, &fakeUsers{balance: "0.00"}, &fakeLedger{})
	req := httptest.NewRequest(http.MethodGet, "/admin/reports/users-due", nil)
	req.Header.Set("Authorization", adminAuth(t, secret))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "[]" {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDownstreamUnavailable(t *testing.T) {
	r, secret := setup(t, &fakeUsers{err: client.ErrUnavailable}, &fakeLedger{})
	req := httptest.NewRequest(http.MethodGet, "/admin/reports/outstanding-balance", nil)
	req.Header.Set("Authorization", adminAuth(t, secret))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d", rec.Code)
	}

	r2, secret2 := setup(t, &fakeUsers{balance: "1.00"}, &fakeLedger{err: client.ErrUnavailable})
	req2 := httptest.NewRequest(http.MethodGet, "/admin/reports/merchant-commissions", nil)
	req2.Header.Set("Authorization", adminAuth(t, secret2))
	rec2 := httptest.NewRecorder()
	r2.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d", rec2.Code)
	}
}

func TestMalformedDownstream(t *testing.T) {
	r, secret := setup(t, &fakeUsers{err: client.ErrMalformed}, &fakeLedger{})
	req := httptest.NewRequest(http.MethodGet, "/admin/reports/outstanding-balance", nil)
	req.Header.Set("Authorization", adminAuth(t, secret))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestAuthZ(t *testing.T) {
	r, secret := setup(t, &fakeUsers{balance: "0.00"}, &fakeLedger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/reports/outstanding-balance", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing jwt status=%d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/admin/reports/outstanding-balance", nil)
	req2.Header.Set("Authorization", userAuth(t, secret))
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Fatalf("non-admin status=%d", rec2.Code)
	}
}

func TestClientSendsInternalToken(t *testing.T) {
	var got string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Internal-Token")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"total_outstanding_balance": "0.00",
		})
	}))
	defer server.Close()

	c := client.NewUserClient(server.URL, "tok-123")
	_, err := c.OutstandingBalance(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != "tok-123" {
		t.Fatalf("token=%q", got)
	}
}
