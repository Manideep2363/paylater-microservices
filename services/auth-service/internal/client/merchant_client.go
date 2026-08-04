package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"paylater/services/auth-service/internal/repository"
	"paylater/shared/httpclient"
)

// MerchantClient talks to merchant-service internal APIs.
type MerchantClient struct {
	http *httpclient.Client
}

// NewMerchantClient creates a MerchantClient for the given base URL and internal token.
func NewMerchantClient(baseURL, internalToken string) *MerchantClient {
	return &MerchantClient{
		http: httpclient.New(baseURL).WithHeader("X-Internal-Token", internalToken),
	}
}

type createMerchantBody struct {
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Phone      string  `json:"phone"`
	Password   string  `json:"password"`
	Commission float64 `json:"commission"`
}

type merchantResponse struct {
	MerchantID           int32  `json:"merchant_id"`
	Name                 string `json:"name"`
	Email                string `json:"email"`
	Phone                string `json:"phone"`
	PasswordHash         string `json:"password_hash"`
	CommissionPercentage string `json:"commission_percentage"`
}

// CreateMerchant implements repository.MerchantRepository via POST /internal/merchants.
func (c *MerchantClient) CreateMerchant(
	ctx context.Context,
	name, email, phone, password string,
	commission float64,
) (repository.Merchant, error) {
	var out merchantResponse
	err := c.http.DoJSON(ctx, httpclient.Request{
		Method: http.MethodPost,
		Path:   "/internal/merchants",
		Body: createMerchantBody{
			Name:       name,
			Email:      email,
			Phone:      phone,
			Password:   password,
			Commission: commission,
		},
		Out: &out,
	})
	if err != nil {
		return repository.Merchant{}, mapMerchantError(err)
	}
	return toMerchant(out), nil
}

// GetMerchantByEmail implements repository.MerchantRepository via GET /internal/merchants/by-email.
func (c *MerchantClient) GetMerchantByEmail(
	ctx context.Context,
	email string,
) (repository.Merchant, error) {
	var out merchantResponse
	err := c.http.DoJSON(ctx, httpclient.Request{
		Method: http.MethodGet,
		Path:   "/internal/merchants/by-email",
		Query:  url.Values{"email": {email}},
		Out:    &out,
	})
	if err != nil {
		return repository.Merchant{}, mapMerchantError(err)
	}
	return toMerchant(out), nil
}

func toMerchant(out merchantResponse) repository.Merchant {
	return repository.Merchant{
		MerchantID:           out.MerchantID,
		Name:                 out.Name,
		Email:                out.Email,
		Phone:                out.Phone,
		PasswordHash:         out.PasswordHash,
		CommissionPercentage: out.CommissionPercentage,
	}
}

func mapMerchantError(err error) error {
	var httpErr *httpclient.HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case http.StatusNotFound:
			return repository.ErrNotFound
		case http.StatusBadRequest:
			if httpErr.Message == "email already exists" {
				return repository.ErrEmailExists
			}
			return errors.New(httpErr.Message)
		default:
			return repository.ErrUnavailable
		}
	}
	return repository.ErrUnavailable
}
