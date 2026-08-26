package repository

import "errors"

// ErrNotFound is returned when a user or merchant lookup fails.
var ErrNotFound = errors.New("not found")

// ErrEmailExists is returned when creating a record with a duplicate email.
var ErrEmailExists = errors.New("email already exists")

// ErrUnavailable is returned when a downstream service cannot be reached
// or returns an unexpected failure (mapped to HTTP 503 by handlers).
var ErrUnavailable = errors.New("service unavailable")
