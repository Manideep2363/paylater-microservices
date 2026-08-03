package service_test

import (
	"context"
	"errors"
	"testing"

	"paylater/services/merchant-service/internal/repository"
	"paylater/services/merchant-service/internal/service"
)

func TestCreateMerchantSuccess(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := service.NewMerchantService(repo)
	ctx := context.Background()

	m, err := svc.CreateMerchant(ctx, "Shop", "shop@example.com", "9999999999", "secret12", 5)
	if err != nil {
		t.Fatalf("CreateMerchant: %v", err)
	}
	if m.MerchantID == 0 || m.CommissionPercentage != "5.00" {
		t.Fatalf("unexpected merchant: %+v", m)
	}
	if m.Email != "shop@example.com" {
		t.Fatalf("email=%s", m.Email)
	}
}

func TestCreateMerchantDuplicateEmail(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := service.NewMerchantService(repo)
	ctx := context.Background()

	_, err := svc.CreateMerchant(ctx, "Shop", "shop@example.com", "111", "secret12", 5)
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err = svc.CreateMerchant(ctx, "Other", "shop@example.com", "222", "secret12", 6)
	if !errors.Is(err, service.ErrEmailExists) {
		t.Fatalf("expected ErrEmailExists, got %v", err)
	}
}

func TestCommissionValidationOnCreate(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := service.NewMerchantService(repo)
	ctx := context.Background()

	cases := []float64{2.99, 20.01, 0, 100, -1}
	for _, commission := range cases {
		_, err := svc.CreateMerchant(ctx, "Shop", "a@example.com", "111", "secret12", commission)
		if !errors.Is(err, service.ErrInvalidCommission) {
			t.Fatalf("commission=%v expected ErrInvalidCommission got %v", commission, err)
		}
	}

	_, err := svc.CreateMerchant(ctx, "Shop", "ok@example.com", "111", "secret12", 3)
	if err != nil {
		t.Fatalf("commission=3 should be valid: %v", err)
	}
	_, err = svc.CreateMerchant(ctx, "Shop2", "ok2@example.com", "222", "secret12", 20)
	if err != nil {
		t.Fatalf("commission=20 should be valid: %v", err)
	}
}

func TestUpdateMerchantCommission(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := service.NewMerchantService(repo)
	ctx := context.Background()

	m, err := svc.CreateMerchant(ctx, "Shop", "shop@example.com", "999", "secret12", 5)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := svc.UpdateMerchantCommission(ctx, m.MerchantID, 2); !errors.Is(err, service.ErrInvalidCommission) {
		t.Fatalf("expected invalid commission, got %v", err)
	}

	if err := svc.UpdateMerchantCommission(ctx, m.MerchantID, 12); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := svc.GetMerchantByID(ctx, m.MerchantID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.CommissionPercentage != "12.00" {
		t.Fatalf("commission=%s", got.CommissionPercentage)
	}

	if err := svc.UpdateMerchantCommission(ctx, 999, 10); !errors.Is(err, service.ErrMerchantNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestGetMerchantByEmailIncludesHash(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := service.NewMerchantService(repo)
	ctx := context.Background()

	_, err := svc.CreateMerchant(ctx, "Shop", "shop@example.com", "999", "secret12", 5)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	internal, err := svc.GetMerchantByEmail(ctx, "shop@example.com")
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if internal.PasswordHash == "" || internal.PasswordHash == "secret12" {
		t.Fatalf("expected bcrypt hash, got %q", internal.PasswordHash)
	}
}
