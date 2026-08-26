package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"paylater/services/ledger-service/internal/db"
)

// ErrNotFound is returned when a ledger row does not exist.
var ErrNotFound = errors.New("not found")

// Transaction is a purchase record.
type Transaction struct {
	TransactionID        int32
	UserID               int32
	MerchantID           int32
	Amount               string
	CommissionPercentage string
	CommissionAmount     string
	CreatedAt            time.Time
}

// Payment is a repayment record.
type Payment struct {
	PaymentID int32
	UserID    int32
	Amount    string
	PaidAt    time.Time
}

// Store is the ledger persistence boundary.
type Store interface {
	CreateTransaction(
		ctx context.Context,
		userID, merchantID int32,
		amount, commissionPercentage, commissionAmount string,
	) (Transaction, error)
	ListTransactions(ctx context.Context) ([]Transaction, error)
	GetTransactionByID(ctx context.Context, id int32) (Transaction, error)
	ListUserTransactions(ctx context.Context, userID int32) ([]Transaction, error)
	ListMerchantTransactions(ctx context.Context, merchantID int32) ([]Transaction, error)

	CreatePayment(ctx context.Context, userID int32, amount string) (Payment, error)
	GetPaymentByID(ctx context.Context, id int32) (Payment, error)
	ListUserPayments(ctx context.Context, userID int32) ([]Payment, error)

	GetMerchantCommissionSummary(ctx context.Context) ([]MerchantCommissionRow, error)
}

// MerchantCommissionRow is an aggregated commission report row.
type MerchantCommissionRow struct {
	MerchantID      int32
	TotalCommission string
}

// SQLStore implements Store with SQLC.
type SQLStore struct {
	queries *db.Queries
}

// NewSQLStore creates a SQLC-backed store.
func NewSQLStore(dbConn *sql.DB) *SQLStore {
	return &SQLStore{queries: db.New(dbConn)}
}

func (s *SQLStore) CreateTransaction(
	ctx context.Context,
	userID, merchantID int32,
	amount, commissionPercentage, commissionAmount string,
) (Transaction, error) {
	result, err := s.queries.CreateTransaction(ctx, db.CreateTransactionParams{
		UserID:               userID,
		MerchantID:           merchantID,
		Amount:               amount,
		CommissionPercentage: commissionPercentage,
		CommissionAmount:     commissionAmount,
	})
	if err != nil {
		return Transaction{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Transaction{}, err
	}
	return s.GetTransactionByID(ctx, int32(id))
}

func (s *SQLStore) ListTransactions(ctx context.Context) ([]Transaction, error) {
	rows, err := s.queries.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Transaction, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapTransaction(row))
	}
	return out, nil
}

func (s *SQLStore) GetTransactionByID(ctx context.Context, id int32) (Transaction, error) {
	row, err := s.queries.GetTransactionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Transaction{}, ErrNotFound
		}
		return Transaction{}, err
	}
	return mapTransaction(row), nil
}

func (s *SQLStore) ListUserTransactions(ctx context.Context, userID int32) ([]Transaction, error) {
	rows, err := s.queries.ListUserTransactions(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Transaction, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapTransaction(row))
	}
	return out, nil
}

func (s *SQLStore) ListMerchantTransactions(ctx context.Context, merchantID int32) ([]Transaction, error) {
	rows, err := s.queries.ListMerchantTransactions(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	out := make([]Transaction, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapTransaction(row))
	}
	return out, nil
}

func (s *SQLStore) CreatePayment(ctx context.Context, userID int32, amount string) (Payment, error) {
	result, err := s.queries.CreatePayment(ctx, db.CreatePaymentParams{
		UserID: userID,
		Amount: amount,
	})
	if err != nil {
		return Payment{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Payment{}, err
	}
	return s.GetPaymentByID(ctx, int32(id))
}

func (s *SQLStore) GetPaymentByID(ctx context.Context, id int32) (Payment, error) {
	row, err := s.queries.GetPaymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Payment{}, ErrNotFound
		}
		return Payment{}, err
	}
	return mapPayment(row), nil
}

func (s *SQLStore) ListUserPayments(ctx context.Context, userID int32) ([]Payment, error) {
	rows, err := s.queries.ListUserPayments(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Payment, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapPayment(row))
	}
	return out, nil
}

func mapTransaction(row db.Transaction) Transaction {
	created := time.Time{}
	if row.CreatedAt.Valid {
		created = row.CreatedAt.Time
	}
	return Transaction{
		TransactionID:        row.TransactionID,
		UserID:               row.UserID,
		MerchantID:           row.MerchantID,
		Amount:               row.Amount,
		CommissionPercentage: row.CommissionPercentage,
		CommissionAmount:     row.CommissionAmount,
		CreatedAt:            created,
	}
}

func mapPayment(row db.Payment) Payment {
	paid := time.Time{}
	if row.PaidAt.Valid {
		paid = row.PaidAt.Time
	}
	return Payment{
		PaymentID: row.PaymentID,
		UserID:    row.UserID,
		Amount:    row.Amount,
		PaidAt:    paid,
	}
}

func (s *SQLStore) GetMerchantCommissionSummary(ctx context.Context) ([]MerchantCommissionRow, error) {
	rows, err := s.queries.GetMerchantCommissionSummary(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]MerchantCommissionRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, MerchantCommissionRow{
			MerchantID:      row.MerchantID,
			TotalCommission: row.TotalCommission,
		})
	}
	return out, nil
}
