package repository

import "context"

// User is the auth-relevant user record owned by user-service.
type User struct {
	UserID   int32
	Name     string
	Email    string
	Password string // bcrypt hash from user-service
}

// UserRepository abstracts user credential access via user-service REST.
type UserRepository interface {
	// CreateUser creates a user with a plain password; user-service hashes it.
	CreateUser(ctx context.Context, name, email, password string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}
