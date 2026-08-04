package repository

import (
	"context"
	"database/sql"
	"errors"

	"paylater/services/user-service/internal/db"
)

// ErrNotFound is returned when a user row does not exist.
var ErrNotFound = errors.New("user not found")

// UserView is a public user projection without password.
type UserView struct {
	UserID      int32
	Name        string
	Email       string
	CreditLimit string
	CurrentDue  string
}

// UserInternal includes the password hash for internal auth lookups.
type UserInternal struct {
	UserID      int32
	Name        string
	Email       string
	Password    string
	CreditLimit string
	CurrentDue  string
}

// LockedUser is the row shape loaded under SELECT ... FOR UPDATE.
type LockedUser struct {
	UserID      int32
	Name        string
	Email       string
	CreditLimit string
	CurrentDue  string
}

// TxRepository exposes queries that must run inside a DB transaction.
type TxRepository interface {
	GetUserByIDForUpdate(ctx context.Context, userID int32) (LockedUser, error)
	IncreaseUserDue(ctx context.Context, userID int32, newDue string) error
	DecreaseUserDue(ctx context.Context, userID int32, newDue string) error
}

// UserDueRow is a users-due report row (no email/password).
type UserDueRow struct {
	UserID     int32
	Name       string
	CurrentDue string
}

// UserAtLimitRow is a users-at-credit-limit report row.
type UserAtLimitRow struct {
	UserID      int32
	Name        string
	CreditLimit string
	CurrentDue  string
}

// Repository is the user-service data access boundary.
type Repository interface {
	CreateUser(ctx context.Context, name, email, passwordHash string) (UserView, error)
	GetUserByID(ctx context.Context, userID int32) (UserView, error)
	GetUserByEmail(ctx context.Context, email string) (UserInternal, error)
	ListUsers(ctx context.Context) ([]UserView, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	WithinTx(ctx context.Context, fn func(txRepo TxRepository) error) error

	GetOutstandingBalance(ctx context.Context) (string, error)
	GetUserOutstandingDues(ctx context.Context) ([]UserDueRow, error)
	GetUsersAtCreditLimit(ctx context.Context) ([]UserAtLimitRow, error)
}

// SQLRepository implements Repository using SQLC + database/sql.
type SQLRepository struct {
	db      *sql.DB
	queries *db.Queries
}

// NewSQLRepository creates a SQLC-backed repository.
func NewSQLRepository(dbConn *sql.DB) *SQLRepository {
	return &SQLRepository{
		db:      dbConn,
		queries: db.New(dbConn),
	}
}

func (r *SQLRepository) CreateUser(
	ctx context.Context,
	name, email, passwordHash string,
) (UserView, error) {
	result, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Name:     name,
		Email:    email,
		Password: passwordHash,
	})
	if err != nil {
		return UserView{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return UserView{}, err
	}

	return r.GetUserByID(ctx, int32(id))
}

func (r *SQLRepository) GetUserByID(ctx context.Context, userID int32) (UserView, error) {
	row, err := r.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserView{}, ErrNotFound
		}
		return UserView{}, err
	}
	return UserView{
		UserID:      row.UserID,
		Name:        row.Name,
		Email:       row.Email,
		CreditLimit: row.CreditLimit,
		CurrentDue:  row.CurrentDue,
	}, nil
}

func (r *SQLRepository) GetUserByEmail(ctx context.Context, email string) (UserInternal, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserInternal{}, ErrNotFound
		}
		return UserInternal{}, err
	}
	return UserInternal{
		UserID:      row.UserID,
		Name:        row.Name,
		Email:       row.Email,
		Password:    row.Password,
		CreditLimit: row.CreditLimit,
		CurrentDue:  row.CurrentDue,
	}, nil
}

func (r *SQLRepository) ListUsers(ctx context.Context) ([]UserView, error) {
	rows, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	users := make([]UserView, 0, len(rows))
	for _, row := range rows {
		users = append(users, UserView{
			UserID:      row.UserID,
			Name:        row.Name,
			Email:       row.Email,
			CreditLimit: row.CreditLimit,
			CurrentDue:  row.CurrentDue,
		})
	}
	return users, nil
}

func (r *SQLRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.queries.CheckEmailExists(ctx, email)
}

func (r *SQLRepository) WithinTx(
	ctx context.Context,
	fn func(txRepo TxRepository) error,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txRepo := &sqlTxRepository{queries: r.queries.WithTx(tx)}
	if err := fn(txRepo); err != nil {
		return err
	}
	return tx.Commit()
}

type sqlTxRepository struct {
	queries *db.Queries
}

func (t *sqlTxRepository) GetUserByIDForUpdate(
	ctx context.Context,
	userID int32,
) (LockedUser, error) {
	row, err := t.queries.GetUserByIDForUpdate(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LockedUser{}, ErrNotFound
		}
		return LockedUser{}, err
	}
	return LockedUser{
		UserID:      row.UserID,
		Name:        row.Name,
		Email:       row.Email,
		CreditLimit: row.CreditLimit,
		CurrentDue:  row.CurrentDue,
	}, nil
}

func (t *sqlTxRepository) IncreaseUserDue(
	ctx context.Context,
	userID int32,
	newDue string,
) error {
	_, err := t.queries.IncreaseUserDue(ctx, db.IncreaseUserDueParams{
		CurrentDue: newDue,
		UserID:     userID,
	})
	return err
}

func (t *sqlTxRepository) DecreaseUserDue(
	ctx context.Context,
	userID int32,
	newDue string,
) error {
	_, err := t.queries.DecreaseUserDue(ctx, db.DecreaseUserDueParams{
		CurrentDue: newDue,
		UserID:     userID,
	})
	return err
}

func (r *SQLRepository) GetOutstandingBalance(ctx context.Context) (string, error) {
	total, err := r.queries.GetOutstandingBalance(ctx)
	if err != nil {
		return "", err
	}
	if total == "" {
		return "0.00", nil
	}
	return total, nil
}

func (r *SQLRepository) GetUserOutstandingDues(ctx context.Context) ([]UserDueRow, error) {
	rows, err := r.queries.GetUserOutstandingDues(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]UserDueRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, UserDueRow{
			UserID:     row.UserID,
			Name:       row.Name,
			CurrentDue: row.CurrentDue,
		})
	}
	return out, nil
}

func (r *SQLRepository) GetUsersAtCreditLimit(ctx context.Context) ([]UserAtLimitRow, error) {
	rows, err := r.queries.GetUsersAtCreditLimit(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]UserAtLimitRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, UserAtLimitRow{
			UserID:      row.UserID,
			Name:        row.Name,
			CreditLimit: row.CreditLimit,
			CurrentDue:  row.CurrentDue,
		})
	}
	return out, nil
}
