package repository

import "context"

// User is the auth-relevant user record.
// In later phases this will be owned by user-service.
type User struct {
	UserID   int32
	Name     string
	Email    string
	Password string // bcrypt hash
}

// UserRepository abstracts user credential storage.
// Temporary in-memory impl today; future REST client to user-service.
type UserRepository interface {
	CreateUser(ctx context.Context, name, email, passwordHash string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}
