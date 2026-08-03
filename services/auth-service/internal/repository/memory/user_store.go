package memory

import (
	"context"
	"strings"
	"sync"

	"paylater/services/auth-service/internal/repository"
)

// UserStore is an in-memory UserRepository for Phase 1.
// Replace with a REST client to user-service in a later phase.
type UserStore struct {
	mu     sync.RWMutex
	nextID int32
	byEmail map[string]repository.User
}

// NewUserStore creates an empty in-memory user store.
func NewUserStore() *UserStore {
	return &UserStore{
		nextID:  1,
		byEmail: make(map[string]repository.User),
	}
}

func (s *UserStore) CreateUser(
	ctx context.Context,
	name, email, passwordHash string,
) (repository.User, error) {
	_ = ctx

	s.mu.Lock()
	defer s.mu.Unlock()

	key := strings.ToLower(email)
	if _, exists := s.byEmail[key]; exists {
		return repository.User{}, repository.ErrEmailExists
	}

	user := repository.User{
		UserID:   s.nextID,
		Name:     name,
		Email:    email,
		Password: passwordHash,
	}
	s.nextID++
	s.byEmail[key] = user
	return user, nil
}

func (s *UserStore) GetUserByEmail(
	ctx context.Context,
	email string,
) (repository.User, error) {
	_ = ctx

	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.byEmail[strings.ToLower(email)]
	if !ok {
		return repository.User{}, repository.ErrNotFound
	}
	return user, nil
}

func (s *UserStore) EmailExists(
	ctx context.Context,
	email string,
) (bool, error) {
	_ = ctx

	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.byEmail[strings.ToLower(email)]
	return ok, nil
}
