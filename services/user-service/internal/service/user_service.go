package service

import (
	"context"
	"errors"
	"strconv"

	"paylater/services/user-service/internal/repository"
	"paylater/shared/auth"
)

var (
	ErrEmailExists          = errors.New("email already exists")
	ErrAmountMustBePositive = errors.New("amount must be greater than zero")
	ErrInsufficientCredit   = errors.New("insufficient credit limit")
	ErrPaymentExceedsDue    = errors.New("payment exceeds outstanding due")
	ErrUserNotFound         = errors.New("user not found")
)

// UserService contains user domain business logic.
type UserService struct {
	repo repository.Repository
}

// NewUserService creates a UserService.
func NewUserService(repo repository.Repository) *UserService {
	return &UserService{repo: repo}
}

// CreateUser registers a user with a bcrypt-hashed password.
func (s *UserService) CreateUser(
	ctx context.Context,
	name, email, password string,
) (repository.UserView, error) {
	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return repository.UserView{}, err
	}
	if exists {
		return repository.UserView{}, ErrEmailExists
	}

	hashed, err := auth.HashPassword(password)
	if err != nil {
		return repository.UserView{}, err
	}

	return s.repo.CreateUser(ctx, name, email, hashed)
}

// GetUserByID returns a public user projection.
func (s *UserService) GetUserByID(ctx context.Context, userID int32) (repository.UserView, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.UserView{}, ErrUserNotFound
		}
		return repository.UserView{}, err
	}
	return user, nil
}

// GetUserByEmail returns an internal user record including password hash.
func (s *UserService) GetUserByEmail(
	ctx context.Context,
	email string,
) (repository.UserInternal, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.UserInternal{}, ErrUserNotFound
		}
		return repository.UserInternal{}, err
	}
	return user, nil
}

// ListUsers returns all users without password hashes.
func (s *UserService) ListUsers(ctx context.Context) ([]repository.UserView, error) {
	return s.repo.ListUsers(ctx)
}

// GetOutstandingBalance returns SUM(current_due) as a DECIMAL string.
func (s *UserService) GetOutstandingBalance(ctx context.Context) (string, error) {
	total, err := s.repo.GetOutstandingBalance(ctx)
	if err != nil {
		return "", err
	}
	if total == "" {
		return "0.00", nil
	}
	return total, nil
}

// GetUserOutstandingDues returns all users ordered by current_due DESC.
func (s *UserService) GetUserOutstandingDues(ctx context.Context) ([]repository.UserDueRow, error) {
	return s.repo.GetUserOutstandingDues(ctx)
}

// GetUsersAtCreditLimit returns users with current_due >= credit_limit.
func (s *UserService) GetUsersAtCreditLimit(ctx context.Context) ([]repository.UserAtLimitRow, error) {
	return s.repo.GetUsersAtCreditLimit(ctx)
}

// IncreaseDue atomically raises current_due after a credit-limit check.
// Mirrors monolith purchase credit rules.
func (s *UserService) IncreaseDue(ctx context.Context, userID int32, amount float64) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	return s.repo.WithinTx(ctx, func(txRepo repository.TxRepository) error {
		user, err := txRepo.GetUserByIDForUpdate(ctx, userID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrUserNotFound
			}
			return err
		}

		creditLimit, err := parseDecimal(user.CreditLimit)
		if err != nil {
			return err
		}
		currentDue, err := parseDecimal(user.CurrentDue)
		if err != nil {
			return err
		}

		availableCredit := creditLimit - currentDue
		if amount > availableCredit {
			return ErrInsufficientCredit
		}

		newDue := currentDue + amount
		return txRepo.IncreaseUserDue(ctx, user.UserID, formatDecimal(newDue))
	})
}

// DecreaseDue atomically lowers current_due without going negative.
// Mirrors monolith repayment rules.
func (s *UserService) DecreaseDue(ctx context.Context, userID int32, amount float64) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	return s.repo.WithinTx(ctx, func(txRepo repository.TxRepository) error {
		user, err := txRepo.GetUserByIDForUpdate(ctx, userID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrUserNotFound
			}
			return err
		}

		currentDue, err := parseDecimal(user.CurrentDue)
		if err != nil {
			return err
		}

		if amount > currentDue {
			return ErrPaymentExceedsDue
		}

		newDue := currentDue - amount
		return txRepo.DecreaseUserDue(ctx, user.UserID, formatDecimal(newDue))
	})
}

func parseDecimal(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

func formatDecimal(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}
