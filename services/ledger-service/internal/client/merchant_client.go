package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"paylater/shared/httpclient"
)

// Merchant is a subset of merchant fields needed for purchases.
type Merchant struct {
	MerchantID           int32
	CommissionPercentage string
}

// MerchantAPI is the interface purchase services depend on (mockable).
type MerchantAPI interface {
	GetMerchantByID(ctx context.Context, merchantID int32) (Merchant, error)
}

// MerchantClient calls merchant-service internal APIs.
type MerchantClient struct {
	http *httpclient.Client
}

// NewMerchantClient creates a MerchantClient.
func NewMerchantClient(baseURL, internalToken string) *MerchantClient {
	return &MerchantClient{
		http: httpclient.New(baseURL).WithHeader("X-Internal-Token", internalToken),
	}
}

type merchantResponse struct {
	MerchantID           int32  `json:"merchant_id"`
	CommissionPercentage string `json:"commission_percentage"`
}

// GetMerchantByID calls GET /internal/merchants/:id.
func (c *MerchantClient) GetMerchantByID(ctx context.Context, merchantID int32) (Merchant, error) {
	var out merchantResponse
	err := c.http.DoJSON(ctx, httpclient.Request{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/internal/merchants/%d", merchantID),
		Out:    &out,
	})
	if err != nil {
		return Merchant{}, mapMerchantError(err)
	}
	return Merchant{
		MerchantID:           out.MerchantID,
		CommissionPercentage: out.CommissionPercentage,
	}, nil
}

func mapMerchantError(err error) error {
	var httpErr *httpclient.HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case http.StatusNotFound:
			return ErrNotFound
		default:
			return ErrUnavailable
		}
	}
	return ErrUnavailable
}
