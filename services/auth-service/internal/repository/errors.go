package repository

import "errors"

// ErrNotFound is returned when a user or merchant lookup fails.
var ErrNotFound = errors.New("not found")

// ErrEmailExists is returned when creating a record with a duplicate email.
var ErrEmailExists = errors.New("email already exists")
