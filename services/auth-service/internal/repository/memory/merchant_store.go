package memory

import (
	"context"
	"strings"
	"sync"

	"paylater/services/auth-service/internal/repository"
)

// MerchantStore is an in-memory MerchantRepository for Phase 1.
// Replace with a REST client to merchant-service in a later phase.
type MerchantStore struct {
	mu      sync.RWMutex
	nextID  int32
	byEmail map[string]repository.Merchant
}

// NewMerchantStore creates an empty in-memory merchant store.
func NewMerchantStore() *MerchantStore {
	return &MerchantStore{
		nextID:  1,
		byEmail: make(map[string]repository.Merchant),
	}
}

func (s *MerchantStore) CreateMerchant(
	ctx context.Context,
	name, email, phone, passwordHash, commissionPercentage string,
) (repository.Merchant, error) {
	_ = ctx

	s.mu.Lock()
	defer s.mu.Unlock()

	key := strings.ToLower(email)
	if _, exists := s.byEmail[key]; exists {
		return repository.Merchant{}, repository.ErrEmailExists
	}

	merchant := repository.Merchant{
		MerchantID:           s.nextID,
		Name:                 name,
		Email:                email,
		Phone:                phone,
		PasswordHash:         passwordHash,
		CommissionPercentage: commissionPercentage,
	}
	s.nextID++
	s.byEmail[key] = merchant
	return merchant, nil
}

func (s *MerchantStore) GetMerchantByEmail(
	ctx context.Context,
	email string,
) (repository.Merchant, error) {
	_ = ctx

	s.mu.RLock()
	defer s.mu.RUnlock()

	merchant, ok := s.byEmail[strings.ToLower(email)]
	if !ok {
		return repository.Merchant{}, repository.ErrNotFound
	}
	return merchant, nil
}
