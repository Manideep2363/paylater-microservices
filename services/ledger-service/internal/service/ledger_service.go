package service

import (
	"context"
	"errors"
	"log"

	"paylater/services/ledger-service/internal/client"
	"paylater/services/ledger-service/internal/money"
	"paylater/services/ledger-service/internal/repository"
)

var (
	ErrAmountMustBePositive = errors.New("amount must be greater than zero")
	ErrMerchantNotFound     = errors.New("merchant not found")
	ErrServiceUnavailable   = errors.New("service unavailable")
	ErrInsufficientCredit   = client.ErrInsufficientCredit
	ErrPaymentExceedsDue    = client.ErrPaymentExceedsDue
	ErrUserNotFound         = client.ErrUserNotFound
	ErrNotFound             = repository.ErrNotFound
	ErrInternal             = errors.New("internal error")
)

// LedgerService orchestrates purchases and repayments with REST compensation.
type LedgerService struct {
	store     repository.Store
	users     client.UserDueAPI
	merchants client.MerchantAPI
}

// NewLedgerService creates a LedgerService.
func NewLedgerService(
	store repository.Store,
	users client.UserDueAPI,
	merchants client.MerchantAPI,
) *LedgerService {
	return &LedgerService{
		store:     store,
		users:     users,
		merchants: merchants,
	}
}

// Purchase creates a transaction after increasing the user's due via user-service.
func (s *LedgerService) Purchase(
	ctx context.Context,
	userID, merchantID int32,
	amount float64,
) (repository.Transaction, error) {
	normalized, amountStr, err := money.NormalizeAmount(amount)
	if err != nil || normalized <= 0 {
		return repository.Transaction{}, ErrAmountMustBePositive
	}

	merchant, err := s.merchants.GetMerchantByID(ctx, merchantID)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return repository.Transaction{}, ErrMerchantNotFound
		}
		if errors.Is(err, client.ErrUnavailable) {
			return repository.Transaction{}, ErrServiceUnavailable
		}
		return repository.Transaction{}, ErrServiceUnavailable
	}

	commissionPct, err := money.Parse(merchant.CommissionPercentage)
	if err != nil {
		return repository.Transaction{}, ErrInternal
	}
	_, commissionAmountStr := money.CommissionAmount(normalized, commissionPct)
	commissionPctStr := money.Format(commissionPct)

	if err := s.users.IncreaseDue(ctx, userID, normalized); err != nil {
		return repository.Transaction{}, mapUserErr(err)
	}

	tx, err := s.store.CreateTransaction(
		ctx,
		userID,
		merchantID,
		amountStr,
		commissionPctStr,
		commissionAmountStr,
	)
	if err != nil {
		if compErr := s.users.DecreaseDue(ctx, userID, normalized); compErr != nil {
			log.Printf(
				"CRITICAL compensation_failed operation=purchase user_id=%d merchant_id=%d amount=%s original_error=%v compensation_error=%v",
				userID, merchantID, amountStr, err, compErr,
			)
			return repository.Transaction{}, ErrInternal
		}
		log.Printf(
			"ERROR ledger_insert_failed_compensated operation=purchase user_id=%d merchant_id=%d amount=%s original_error=%v",
			userID, merchantID, amountStr, err,
		)
		return repository.Transaction{}, ErrInternal
	}

	return tx, nil
}

// Repay records a payment after decreasing the user's due via user-service.
func (s *LedgerService) Repay(
	ctx context.Context,
	userID int32,
	amount float64,
) (repository.Payment, error) {
	normalized, amountStr, err := money.NormalizeAmount(amount)
	if err != nil || normalized <= 0 {
		return repository.Payment{}, ErrAmountMustBePositive
	}

	if err := s.users.DecreaseDue(ctx, userID, normalized); err != nil {
		return repository.Payment{}, mapUserErr(err)
	}

	payment, err := s.store.CreatePayment(ctx, userID, amountStr)
	if err != nil {
		if compErr := s.users.IncreaseDue(ctx, userID, normalized); compErr != nil {
			log.Printf(
				"CRITICAL compensation_failed operation=repayment user_id=%d amount=%s original_error=%v compensation_error=%v",
				userID, amountStr, err, compErr,
			)
			return repository.Payment{}, ErrInternal
		}
		log.Printf(
			"ERROR ledger_insert_failed_compensated operation=repayment user_id=%d amount=%s original_error=%v",
			userID, amountStr, err,
		)
		return repository.Payment{}, ErrInternal
	}

	return payment, nil
}

func mapUserErr(err error) error {
	switch {
	case errors.Is(err, client.ErrInsufficientCredit):
		return ErrInsufficientCredit
	case errors.Is(err, client.ErrPaymentExceedsDue):
		return ErrPaymentExceedsDue
	case errors.Is(err, client.ErrAmountMustBePositive):
		return ErrAmountMustBePositive
	case errors.Is(err, client.ErrUserNotFound):
		return ErrUserNotFound
	case errors.Is(err, client.ErrUnavailable):
		return ErrServiceUnavailable
	default:
		return ErrServiceUnavailable
	}
}

// ListTransactions returns all purchases (admin).
func (s *LedgerService) ListTransactions(ctx context.Context) ([]repository.Transaction, error) {
	return s.store.ListTransactions(ctx)
}

// GetTransactionByID returns one purchase.
func (s *LedgerService) GetTransactionByID(ctx context.Context, id int32) (repository.Transaction, error) {
	tx, err := s.store.GetTransactionByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.Transaction{}, ErrNotFound
		}
		return repository.Transaction{}, err
	}
	return tx, nil
}

// ListUserTransactions returns purchases for a user.
func (s *LedgerService) ListUserTransactions(ctx context.Context, userID int32) ([]repository.Transaction, error) {
	return s.store.ListUserTransactions(ctx, userID)
}

// ListMerchantTransactions returns purchases for a merchant.
func (s *LedgerService) ListMerchantTransactions(ctx context.Context, merchantID int32) ([]repository.Transaction, error) {
	return s.store.ListMerchantTransactions(ctx, merchantID)
}

// GetPaymentByID returns one payment.
func (s *LedgerService) GetPaymentByID(ctx context.Context, id int32) (repository.Payment, error) {
	p, err := s.store.GetPaymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.Payment{}, ErrNotFound
		}
		return repository.Payment{}, err
	}
	return p, nil
}

// ListUserPayments returns payments for a user.
func (s *LedgerService) ListUserPayments(ctx context.Context, userID int32) ([]repository.Payment, error) {
	return s.store.ListUserPayments(ctx, userID)
}
