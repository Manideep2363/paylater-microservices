package client

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"paylater/shared/httpclient"
)

var (
	ErrUnavailable = errors.New("service unavailable")
	ErrMalformed   = errors.New("malformed downstream response")
)

// UserReportsAPI is the user-service report client contract.
type UserReportsAPI interface {
	OutstandingBalance(ctx context.Context) (string, error)
	UsersDue(ctx context.Context) ([]UserDue, error)
	UsersAtCreditLimit(ctx context.Context) ([]UserAtLimit, error)
}

// LedgerReportsAPI is the ledger-service report client contract.
type LedgerReportsAPI interface {
	MerchantCommissions(ctx context.Context) ([]MerchantCommission, error)
}

// UserClient calls user-service internal report APIs.
type UserClient struct {
	http *httpclient.Client
}

// NewUserClient creates a UserClient.
func NewUserClient(baseURL, internalToken string) *UserClient {
	return &UserClient{
		http: httpclient.New(baseURL).WithHeader("X-Internal-Token", internalToken),
	}
}

type outstandingBalanceResponse struct {
	TotalOutstandingBalance string `json:"total_outstanding_balance"`
}

// UserDue matches the users-due report row.
type UserDue struct {
	UserID     int32  `json:"user_id"`
	Name       string `json:"name"`
	CurrentDue string `json:"current_due"`
}

// UserAtLimit matches the credit-limit report row.
type UserAtLimit struct {
	UserID      int32  `json:"user_id"`
	Name        string `json:"name"`
	CreditLimit string `json:"credit_limit"`
	CurrentDue  string `json:"current_due"`
}

// OutstandingBalance calls GET /internal/reports/outstanding-balance.
func (c *UserClient) OutstandingBalance(ctx context.Context) (string, error) {
	var out outstandingBalanceResponse
	if err := c.http.DoJSON(ctx, httpclient.Request{
		Method: http.MethodGet,
		Path:   "/internal/reports/outstanding-balance",
		Out:    &out,
	}); err != nil {
		return "", mapErr(err)
	}
	if out.TotalOutstandingBalance == "" {
		return "", ErrMalformed
	}
	return out.TotalOutstandingBalance, nil
}

// UsersDue calls GET /internal/reports/users-due.
func (c *UserClient) UsersDue(ctx context.Context) ([]UserDue, error) {
	var out []UserDue
	if err := c.http.DoJSON(ctx, httpclient.Request{
		Method: http.MethodGet,
		Path:   "/internal/reports/users-due",
		Out:    &out,
	}); err != nil {
		return nil, mapErr(err)
	}
	if out == nil {
		out = []UserDue{}
	}
	return out, nil
}

// UsersAtCreditLimit calls GET /internal/reports/users-at-credit-limit.
func (c *UserClient) UsersAtCreditLimit(ctx context.Context) ([]UserAtLimit, error) {
	var out []UserAtLimit
	if err := c.http.DoJSON(ctx, httpclient.Request{
		Method: http.MethodGet,
		Path:   "/internal/reports/users-at-credit-limit",
		Out:    &out,
	}); err != nil {
		return nil, mapErr(err)
	}
	if out == nil {
		out = []UserAtLimit{}
	}
	return out, nil
}

// LedgerClient calls ledger-service internal report APIs.
type LedgerClient struct {
	http *httpclient.Client
}

// NewLedgerClient creates a LedgerClient.
func NewLedgerClient(baseURL, internalToken string) *LedgerClient {
	return &LedgerClient{
		http: httpclient.New(baseURL).WithHeader("X-Internal-Token", internalToken),
	}
}

// MerchantCommission matches the commission summary row.
type MerchantCommission struct {
	MerchantID      int32  `json:"merchant_id"`
	TotalCommission string `json:"total_commission"`
}

// MerchantCommissions calls GET /internal/reports/merchant-commissions.
func (c *LedgerClient) MerchantCommissions(ctx context.Context) ([]MerchantCommission, error) {
	var out []MerchantCommission
	if err := c.http.DoJSON(ctx, httpclient.Request{
		Method: http.MethodGet,
		Path:   "/internal/reports/merchant-commissions",
		Out:    &out,
	}); err != nil {
		return nil, mapErr(err)
	}
	if out == nil {
		out = []MerchantCommission{}
	}
	return out, nil
}

func mapErr(err error) error {
	var httpErr *httpclient.HTTPError
	if errors.As(err, &httpErr) {
		return ErrUnavailable
	}
	if strings.Contains(err.Error(), "unmarshal") {
		return ErrMalformed
	}
	return ErrUnavailable
}
