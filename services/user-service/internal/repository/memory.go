package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// MemoryRepository is an in-memory Repository for unit tests.
type MemoryRepository struct {
	mu      sync.Mutex
	nextID  int32
	byID    map[int32]memoryUser
	byEmail map[string]int32
}

type memoryUser struct {
	UserID      int32
	Name        string
	Email       string
	Password    string
	CreditLimit string
	CurrentDue  string
}

// NewMemoryRepository creates an empty in-memory store.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		nextID:  1,
		byID:    make(map[int32]memoryUser),
		byEmail: make(map[string]int32),
	}
}

func (m *MemoryRepository) CreateUser(
	ctx context.Context,
	name, email, passwordHash string,
) (UserView, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	key := strings.ToLower(email)
	if _, ok := m.byEmail[key]; ok {
		return UserView{}, fmt.Errorf("email already exists")
	}

	u := memoryUser{
		UserID:      m.nextID,
		Name:        name,
		Email:       email,
		Password:    passwordHash,
		CreditLimit: "2000.00",
		CurrentDue:  "0.00",
	}
	m.nextID++
	m.byID[u.UserID] = u
	m.byEmail[key] = u.UserID

	return UserView{
		UserID:      u.UserID,
		Name:        u.Name,
		Email:       u.Email,
		CreditLimit: u.CreditLimit,
		CurrentDue:  u.CurrentDue,
	}, nil
}

func (m *MemoryRepository) GetUserByID(ctx context.Context, userID int32) (UserView, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	u, ok := m.byID[userID]
	if !ok {
		return UserView{}, ErrNotFound
	}
	return UserView{
		UserID:      u.UserID,
		Name:        u.Name,
		Email:       u.Email,
		CreditLimit: u.CreditLimit,
		CurrentDue:  u.CurrentDue,
	}, nil
}

func (m *MemoryRepository) GetUserByEmail(ctx context.Context, email string) (UserInternal, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	id, ok := m.byEmail[strings.ToLower(email)]
	if !ok {
		return UserInternal{}, ErrNotFound
	}
	u := m.byID[id]
	return UserInternal{
		UserID:      u.UserID,
		Name:        u.Name,
		Email:       u.Email,
		Password:    u.Password,
		CreditLimit: u.CreditLimit,
		CurrentDue:  u.CurrentDue,
	}, nil
}

func (m *MemoryRepository) ListUsers(ctx context.Context) ([]UserView, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]UserView, 0, len(m.byID))
	for id := int32(1); id < m.nextID; id++ {
		u, ok := m.byID[id]
		if !ok {
			continue
		}
		out = append(out, UserView{
			UserID:      u.UserID,
			Name:        u.Name,
			Email:       u.Email,
			CreditLimit: u.CreditLimit,
			CurrentDue:  u.CurrentDue,
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

func (m *MemoryRepository) WithinTx(
	ctx context.Context,
	fn func(txRepo TxRepository) error,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fn(&memoryTxRepository{parent: m})
}

type memoryTxRepository struct {
	parent *MemoryRepository
}

func (t *memoryTxRepository) GetUserByIDForUpdate(
	ctx context.Context,
	userID int32,
) (LockedUser, error) {
	_ = ctx
	u, ok := t.parent.byID[userID]
	if !ok {
		return LockedUser{}, ErrNotFound
	}
	return LockedUser{
		UserID:      u.UserID,
		Name:        u.Name,
		Email:       u.Email,
		CreditLimit: u.CreditLimit,
		CurrentDue:  u.CurrentDue,
	}, nil
}

func (t *memoryTxRepository) IncreaseUserDue(
	ctx context.Context,
	userID int32,
	newDue string,
) error {
	_ = ctx
	u, ok := t.parent.byID[userID]
	if !ok {
		return ErrNotFound
	}
	u.CurrentDue = newDue
	t.parent.byID[userID] = u
	return nil
}

func (t *memoryTxRepository) DecreaseUserDue(
	ctx context.Context,
	userID int32,
	newDue string,
) error {
	_ = ctx
	u, ok := t.parent.byID[userID]
	if !ok {
		return ErrNotFound
	}
	u.CurrentDue = newDue
	t.parent.byID[userID] = u
	return nil
}

// SetDueForTest sets current_due directly (test helper).
func (m *MemoryRepository) SetDueForTest(userID int32, due float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[userID]
	if !ok {
		return ErrNotFound
	}
	u.CurrentDue = strconv.FormatFloat(due, 'f', 2, 64)
	m.byID[userID] = u
	return nil
}
