package repository

import (
	"context"
	"strings"
	"sync"
)

// MemoryRepository is an in-memory Repository for unit tests.
type MemoryRepository struct {
	mu      sync.Mutex
	nextID  int32
	byID    map[int32]memoryMerchant
	byEmail map[string]int32
}

type memoryMerchant struct {
	MerchantID           int32
	Name                 string
	Email                string
	Phone                string
	PasswordHash         string
	CommissionPercentage string
}

// NewMemoryRepository creates an empty in-memory store.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		nextID:  1,
		byID:    make(map[int32]memoryMerchant),
		byEmail: make(map[string]int32),
	}
}

func (m *MemoryRepository) CreateMerchant(
	ctx context.Context,
	name, email, phone, passwordHash, commissionPercentage string,
) (MerchantView, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	key := strings.ToLower(email)

	row := memoryMerchant{
		MerchantID:           m.nextID,
		Name:                 name,
		Email:                email,
		Phone:                phone,
		PasswordHash:         passwordHash,
		CommissionPercentage: commissionPercentage,
	}
	m.nextID++
	m.byID[row.MerchantID] = row
	m.byEmail[key] = row.MerchantID

	return MerchantView{
		MerchantID:           row.MerchantID,
		Name:                 row.Name,
		Email:                row.Email,
		Phone:                row.Phone,
		CommissionPercentage: row.CommissionPercentage,
	}, nil
}

func (m *MemoryRepository) GetMerchantByID(
	ctx context.Context,
	merchantID int32,
) (MerchantView, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	row, ok := m.byID[merchantID]
	if !ok {
		return MerchantView{}, ErrNotFound
	}
	return MerchantView{
		MerchantID:           row.MerchantID,
		Name:                 row.Name,
		Email:                row.Email,
		Phone:                row.Phone,
		CommissionPercentage: row.CommissionPercentage,
	}, nil
}

func (m *MemoryRepository) GetMerchantByEmail(
	ctx context.Context,
	email string,
) (MerchantInternal, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	id, ok := m.byEmail[strings.ToLower(email)]
	if !ok {
		return MerchantInternal{}, ErrNotFound
	}
	row := m.byID[id]
	return MerchantInternal{
		MerchantID:           row.MerchantID,
		Name:                 row.Name,
		Email:                row.Email,
		Phone:                row.Phone,
		PasswordHash:         row.PasswordHash,
		CommissionPercentage: row.CommissionPercentage,
	}, nil
}

func (m *MemoryRepository) ListMerchants(ctx context.Context) ([]MerchantView, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]MerchantView, 0, len(m.byID))
	for id := int32(1); id < m.nextID; id++ {
		row, ok := m.byID[id]
		if !ok {
			continue
		}
		out = append(out, MerchantView{
			MerchantID:           row.MerchantID,
			Name:                 row.Name,
			Email:                row.Email,
			Phone:                row.Phone,
			CommissionPercentage: row.CommissionPercentage,
		})
	}
	return out, nil
}

func (m *MemoryRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.byEmail[strings.ToLower(email)]
	return ok, nil
}

func (m *MemoryRepository) UpdateMerchantCommission(
	ctx context.Context,
	merchantID int32,
	commissionPercentage string,
) error {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	row, ok := m.byID[merchantID]
	if !ok {
		return ErrNotFound
	}
	row.CommissionPercentage = commissionPercentage
	m.byID[merchantID] = row
	return nil
}
