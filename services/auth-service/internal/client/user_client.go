package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"paylater/services/auth-service/internal/repository"
	"paylater/shared/httpclient"
)

// UserClient talks to user-service internal APIs.
type UserClient struct {
	http *httpclient.Client
}

// NewUserClient creates a UserClient for the given base URL and internal token.
func NewUserClient(baseURL, internalToken string) *UserClient {
	return &UserClient{
		http: httpclient.New(baseURL).WithHeader("X-Internal-Token", internalToken),
	}
}

type createUserBody struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	UserID   int32  `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateUser implements repository.UserRepository via POST /internal/users.
func (c *UserClient) CreateUser(
	ctx context.Context,
	name, email, password string,
) (repository.User, error) {
	var out userResponse
	err := c.http.DoJSON(ctx, httpclient.Request{
		Method: http.MethodPost,
		Path:   "/internal/users",
		Body: createUserBody{
			Name:     name,
			Email:    email,
			Password: password,
		},
		Out: &out,
	})
	if err != nil {
		return repository.User{}, mapUserError(err)
	}
	return repository.User{
		UserID:   out.UserID,
		Name:     out.Name,
		Email:    out.Email,
		Password: out.Password,
	}, nil
}

// GetUserByEmail implements repository.UserRepository via GET /internal/users/by-email.
func (c *UserClient) GetUserByEmail(ctx context.Context, email string) (repository.User, error) {
	var out userResponse
	err := c.http.DoJSON(ctx, httpclient.Request{
		Method: http.MethodGet,
		Path:   "/internal/users/by-email",
		Query:  url.Values{"email": {email}},
		Out:    &out,
	})
	if err != nil {
		return repository.User{}, mapUserError(err)
	}
	return repository.User{
		UserID:   out.UserID,
		Name:     out.Name,
		Email:    out.Email,
		Password: out.Password,
	}, nil
}

// EmailExists checks whether a user email is already registered.
func (c *UserClient) EmailExists(ctx context.Context, email string) (bool, error) {
	_, err := c.GetUserByEmail(ctx, email)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, repository.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func mapUserError(err error) error {
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
	// Network / timeout / dial errors.
	return repository.ErrUnavailable
}
