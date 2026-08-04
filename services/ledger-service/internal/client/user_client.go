package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"paylater/shared/httpclient"
)

var (
	// ErrUnavailable means the downstream service could not be reached or returned 5xx.
	ErrUnavailable = errors.New("service unavailable")
	// ErrNotFound means the resource was not found downstream.
	ErrNotFound = errors.New("not found")
	// ErrInsufficientCredit is returned when user-service rejects a due increase.
	ErrInsufficientCredit = errors.New("insufficient credit limit")
	// ErrPaymentExceedsDue is returned when user-service rejects a due decrease.
	ErrPaymentExceedsDue = errors.New("payment exceeds outstanding due")
	// ErrAmountMustBePositive mirrors user-service validation.
	ErrAmountMustBePositive = errors.New("amount must be greater than zero")
	// ErrUserNotFound mirrors user-service not found.
	ErrUserNotFound = errors.New("user not found")
)

// UserClient calls user-service internal due APIs.
type UserClient struct {
	http *httpclient.Client
}

// NewUserClient creates a UserClient.
func NewUserClient(baseURL, internalToken string) *UserClient {
	return &UserClient{
		http: httpclient.New(baseURL).WithHeader("X-Internal-Token", internalToken),
	}
}

type dueBody struct {
	Amount float64 `json:"amount"`
}

// UserDueAPI is the interface purchase/repay services depend on (mockable).
type UserDueAPI interface {
	IncreaseDue(ctx context.Context, userID int32, amount float64) error
	DecreaseDue(ctx context.Context, userID int32, amount float64) error
}

// IncreaseDue calls POST /internal/users/:id/due/increase.
func (c *UserClient) IncreaseDue(ctx context.Context, userID int32, amount float64) error {
	err := c.http.DoJSON(ctx, httpclient.Request{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/internal/users/%d/due/increase", userID),
		Body:   dueBody{Amount: amount},
	})
	return mapDueError(err)
}

// DecreaseDue calls POST /internal/users/:id/due/decrease.
func (c *UserClient) DecreaseDue(ctx context.Context, userID int32, amount float64) error {
	err := c.http.DoJSON(ctx, httpclient.Request{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/internal/users/%d/due/decrease", userID),
		Body:   dueBody{Amount: amount},
	})
	return mapDueError(err)
}

func mapDueError(err error) error {
	if err == nil {
		return nil
	}
	var httpErr *httpclient.HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case http.StatusNotFound:
			return ErrUserNotFound
		case http.StatusBadRequest:
			switch httpErr.Message {
			case "insufficient credit limit":
				return ErrInsufficientCredit
			case "payment exceeds outstanding due":
				return ErrPaymentExceedsDue
			case "amount must be greater than zero":
				return ErrAmountMustBePositive
			default:
				return errors.New(httpErr.Message)
			}
		default:
			return ErrUnavailable
		}
	}
	return ErrUnavailable
}
