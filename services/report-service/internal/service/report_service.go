package service

import (
	"context"
	"errors"
	"log"

	"paylater/services/report-service/internal/client"
)

var (
	ErrServiceUnavailable = client.ErrUnavailable
	ErrInternal           = errors.New("internal error")
)

// ReportService orchestrates admin reports via owning services.
type ReportService struct {
	users  client.UserReportsAPI
	ledger client.LedgerReportsAPI
}

// NewReportService creates a ReportService.
func NewReportService(users client.UserReportsAPI, ledger client.LedgerReportsAPI) *ReportService {
	return &ReportService{users: users, ledger: ledger}
}

// OutstandingBalance forwards to user-service.
func (s *ReportService) OutstandingBalance(ctx context.Context) (string, error) {
	total, err := s.users.OutstandingBalance(ctx)
	if err != nil {
		return "", mapDownstream(err)
	}
	return total, nil
}

// UsersDue forwards to user-service.
func (s *ReportService) UsersDue(ctx context.Context) ([]client.UserDue, error) {
	rows, err := s.users.UsersDue(ctx)
	if err != nil {
		return nil, mapDownstream(err)
	}
	return rows, nil
}

// UsersAtCreditLimit forwards to user-service.
func (s *ReportService) UsersAtCreditLimit(ctx context.Context) ([]client.UserAtLimit, error) {
	rows, err := s.users.UsersAtCreditLimit(ctx)
	if err != nil {
		return nil, mapDownstream(err)
	}
	return rows, nil
}

// MerchantCommissions forwards to ledger-service.
func (s *ReportService) MerchantCommissions(ctx context.Context) ([]client.MerchantCommission, error) {
	rows, err := s.ledger.MerchantCommissions(ctx)
	if err != nil {
		return nil, mapDownstream(err)
	}
	return rows, nil
}

func mapDownstream(err error) error {
	if errors.Is(err, client.ErrUnavailable) {
		return ErrServiceUnavailable
	}
	if errors.Is(err, client.ErrMalformed) {
		log.Printf("ERROR malformed_downstream err=%v", err)
		return ErrInternal
	}
	log.Printf("ERROR report_downstream err=%v", err)
	return ErrInternal
}
