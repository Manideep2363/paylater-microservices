package repository

import (
	"context"
	"database/sql"
	"errors"

	"paylater/services/merchant-service/internal/db"
)

// ErrNotFound is returned when a merchant row does not exist.
var ErrNotFound = errors.New("merchant not found")

// MerchantView is a public/admin merchant projection without password_hash.
type MerchantView struct {
	MerchantID           int32
	Name                 string
	Email                string
	Phone                string
	CommissionPercentage string
}

// MerchantInternal includes password_hash for auth lookups.
type MerchantInternal struct {
	MerchantID           int32
	Name                 string
	Email                string
	Phone                string
	PasswordHash         string
	CommissionPercentage string
}

// Repository is the merchant-service data access boundary.
type Repository interface {
	CreateMerchant(
		ctx context.Context,
		name, email, phone, passwordHash, commissionPercentage string,
	) (MerchantView, error)
	GetMerchantByID(ctx context.Context, merchantID int32) (MerchantView, error)
	GetMerchantByEmail(ctx context.Context, email string) (MerchantInternal, error)
	ListMerchants(ctx context.Context) ([]MerchantView, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	UpdateMerchantCommission(ctx context.Context, merchantID int32, commissionPercentage string) error
}

// SQLRepository implements Repository using SQLC + database/sql.
type SQLRepository struct {
	queries *db.Queries
}

// NewSQLRepository creates a SQLC-backed repository.
func NewSQLRepository(dbConn *sql.DB) *SQLRepository {
	return &SQLRepository{queries: db.New(dbConn)}
}

func (r *SQLRepository) CreateMerchant(
	ctx context.Context,
	name, email, phone, passwordHash, commissionPercentage string,
) (MerchantView, error) {
	result, err := r.queries.CreateMerchant(ctx, db.CreateMerchantParams{
		Name:                 name,
		Email:                email,
		Phone:                phone,
		PasswordHash:         passwordHash,
		CommissionPercentage: commissionPercentage,
	})
	if err != nil {
		return MerchantView{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return MerchantView{}, err
	}

	return r.GetMerchantByID(ctx, int32(id))
}

func (r *SQLRepository) GetMerchantByID(
	ctx context.Context,
	merchantID int32,
) (MerchantView, error) {
	row, err := r.queries.GetMerchantByID(ctx, merchantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return MerchantView{}, ErrNotFound
		}
		return MerchantView{}, err
	}
	return MerchantView{
		MerchantID:           row.MerchantID,
		Name:                 row.Name,
		Email:                row.Email,
		Phone:                row.Phone,
		CommissionPercentage: row.CommissionPercentage,
	}, nil
}

func (r *SQLRepository) GetMerchantByEmail(
	ctx context.Context,
	email string,
) (MerchantInternal, error) {
	row, err := r.queries.GetMerchantByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return MerchantInternal{}, ErrNotFound
		}
		return MerchantInternal{}, err
	}
	return MerchantInternal{
		MerchantID:           row.MerchantID,
		Name:                 row.Name,
		Email:                row.Email,
		Phone:                row.Phone,
		PasswordHash:         row.PasswordHash,
		CommissionPercentage: row.CommissionPercentage,
	}, nil
}

func (r *SQLRepository) ListMerchants(ctx context.Context) ([]MerchantView, error) {
	rows, err := r.queries.ListMerchants(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]MerchantView, 0, len(rows))
	for _, row := range rows {
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

func (r *SQLRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.queries.CheckMerchantEmailExists(ctx, email)
}

func (r *SQLRepository) UpdateMerchantCommission(
	ctx context.Context,
	merchantID int32,
	commissionPercentage string,
) error {
	result, err := r.queries.UpdateMerchantCommission(ctx, db.UpdateMerchantCommissionParams{
		CommissionPercentage: commissionPercentage,
		MerchantID:           merchantID,
	})
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
