package service

import (
	"context"
	"errors"
	"strconv"

	"paylater/services/merchant-service/internal/repository"
	"paylater/shared/auth"
)

var (
	ErrEmailExists       = errors.New("email already exists")
	ErrMerchantNotFound  = errors.New("merchant not found")
	ErrInvalidCommission = errors.New("commission must be between 3 and 20")
)

// MerchantService contains merchant domain business logic.
type MerchantService struct {
	repo repository.Repository
}

// NewMerchantService creates a MerchantService.
func NewMerchantService(repo repository.Repository) *MerchantService {
	return &MerchantService{repo: repo}
}

// CreateMerchant registers a merchant with a bcrypt-hashed password.
// Commission must be between 3 and 20 inclusive (monolith handler rule).
func (s *MerchantService) CreateMerchant(
	ctx context.Context,
	name, email, phone, password string,
	commission float64,
) (repository.MerchantView, error) {
	if err := validateCommission(commission); err != nil {
		return repository.MerchantView{}, err
	}

	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return repository.MerchantView{}, err
	}
	if exists {
		return repository.MerchantView{}, ErrEmailExists
	}

	hashed, err := auth.HashPassword(password)
	if err != nil {
		return repository.MerchantView{}, err
	}

	return s.repo.CreateMerchant(
		ctx,
		name,
		email,
		phone,
		hashed,
		formatCommission(commission),
	)
}

// GetMerchantByID returns a public merchant projection.
func (s *MerchantService) GetMerchantByID(
	ctx context.Context,
	merchantID int32,
) (repository.MerchantView, error) {
	merchant, err := s.repo.GetMerchantByID(ctx, merchantID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.MerchantView{}, ErrMerchantNotFound
		}
		return repository.MerchantView{}, err
	}
	return merchant, nil
}

// GetMerchantByEmail returns an internal merchant including password_hash.
func (s *MerchantService) GetMerchantByEmail(
	ctx context.Context,
	email string,
) (repository.MerchantInternal, error) {
	merchant, err := s.repo.GetMerchantByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.MerchantInternal{}, ErrMerchantNotFound
		}
		return repository.MerchantInternal{}, err
	}
	return merchant, nil
}

// ListMerchants returns all merchants without password hashes.
func (s *MerchantService) ListMerchants(ctx context.Context) ([]repository.MerchantView, error) {
	return s.repo.ListMerchants(ctx)
}

// UpdateMerchantCommission updates commission with the monolith 3–20% rule.
func (s *MerchantService) UpdateMerchantCommission(
	ctx context.Context,
	merchantID int32,
	commission float64,
) error {
	if err := validateCommission(commission); err != nil {
		return err
	}

	err := s.repo.UpdateMerchantCommission(
		ctx,
		merchantID,
		formatCommission(commission),
	)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrMerchantNotFound
		}
		return err
	}
	return nil
}

func validateCommission(commission float64) error {
	if commission < 3 || commission > 20 {
		return ErrInvalidCommission
	}
	return nil
}

func formatCommission(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}
