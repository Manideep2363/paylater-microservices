package repository

import (
	"context"
	"sync"
	"time"
)

// MemoryStore is an in-memory Store for unit tests.
type MemoryStore struct {
	mu           sync.Mutex
	nextTxID     int32
	nextPayID    int32
	transactions map[int32]Transaction
	payments     map[int32]Payment
	FailNextTx   bool
	FailNextPay  bool
}

// NewMemoryStore creates an empty memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		nextTxID:     1,
		nextPayID:    1,
		transactions: make(map[int32]Transaction),
		payments:     make(map[int32]Payment),
	}
}

func (m *MemoryStore) CreateTransaction(
	ctx context.Context,
	userID, merchantID int32,
	amount, commissionPercentage, commissionAmount string,
) (Transaction, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.FailNextTx {
		m.FailNextTx = false
		return Transaction{}, errorsNew("insert failed")
	}
	tx := Transaction{
		TransactionID:        m.nextTxID,
		UserID:               userID,
		MerchantID:           merchantID,
		Amount:               amount,
		CommissionPercentage: commissionPercentage,
		CommissionAmount:     commissionAmount,
		CreatedAt:            time.Now().UTC(),
	}
	m.nextTxID++
	m.transactions[tx.TransactionID] = tx
	return tx, nil
}

func (m *MemoryStore) ListTransactions(ctx context.Context) ([]Transaction, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Transaction, 0, len(m.transactions))
	for id := int32(1); id < m.nextTxID; id++ {
		if tx, ok := m.transactions[id]; ok {
			out = append(out, tx)
		}
	}
	return out, nil
}

func (m *MemoryStore) GetTransactionByID(ctx context.Context, id int32) (Transaction, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	tx, ok := m.transactions[id]
	if !ok {
		return Transaction{}, ErrNotFound
	}
	return tx, nil
}

func (m *MemoryStore) ListUserTransactions(ctx context.Context, userID int32) ([]Transaction, error) {
	all, err := m.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Transaction, 0)
	for _, tx := range all {
		if tx.UserID == userID {
			out = append(out, tx)
		}
	}
	return out, nil
}

func (m *MemoryStore) ListMerchantTransactions(ctx context.Context, merchantID int32) ([]Transaction, error) {
	all, err := m.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Transaction, 0)
	for _, tx := range all {
		if tx.MerchantID == merchantID {
			out = append(out, tx)
		}
	}
	return out, nil
}

func (m *MemoryStore) CreatePayment(ctx context.Context, userID int32, amount string) (Payment, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.FailNextPay {
		m.FailNextPay = false
		return Payment{}, errorsNew("insert failed")
	}
	p := Payment{
		PaymentID: m.nextPayID,
		UserID:    userID,
		Amount:    amount,
		PaidAt:    time.Now().UTC(),
	}
	m.nextPayID++
	m.payments[p.PaymentID] = p
	return p, nil
}

func (m *MemoryStore) GetPaymentByID(ctx context.Context, id int32) (Payment, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.payments[id]
	if !ok {
		return Payment{}, ErrNotFound
	}
	return p, nil
}

func (m *MemoryStore) ListUserPayments(ctx context.Context, userID int32) ([]Payment, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Payment, 0)
	for id := int32(1); id < m.nextPayID; id++ {
		if p, ok := m.payments[id]; ok && p.UserID == userID {
			out = append(out, p)
		}
	}
	return out, nil
}

func errorsNew(msg string) error {
	return &storeError{msg: msg}
}

type storeError struct{ msg string }

func (e *storeError) Error() string { return e.msg }
