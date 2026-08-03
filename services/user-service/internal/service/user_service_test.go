package service_test

import (
	"context"
	"errors"
	"testing"

	"paylater/services/user-service/internal/repository"
	"paylater/services/user-service/internal/service"
)

func TestCreateUserAndList(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := service.NewUserService(repo)
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, "Alice", "alice@example.com", "secret12")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.UserID == 0 || user.CreditLimit != "2000.00" || user.CurrentDue != "0.00" {
		t.Fatalf("unexpected user: %+v", user)
	}

	_, err = svc.CreateUser(ctx, "Alice2", "alice@example.com", "secret12")
	if !errors.Is(err, service.ErrEmailExists) {
		t.Fatalf("expected ErrEmailExists, got %v", err)
	}

	users, err := svc.ListUsers(ctx)
	if err != nil || len(users) != 1 {
		t.Fatalf("ListUsers: %v len=%d", err, len(users))
	}
}

func TestIncreaseDueCreditRules(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := service.NewUserService(repo)
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, "Bob", "bob@example.com", "secret12")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := svc.IncreaseDue(ctx, user.UserID, 0); !errors.Is(err, service.ErrAmountMustBePositive) {
		t.Fatalf("expected positive amount error, got %v", err)
	}

	if err := svc.IncreaseDue(ctx, user.UserID, 500); err != nil {
		t.Fatalf("IncreaseDue 500: %v", err)
	}

	got, _ := svc.GetUserByID(ctx, user.UserID)
	if got.CurrentDue != "500.00" {
		t.Fatalf("due want 500.00 got %s", got.CurrentDue)
	}

	// credit_limit 2000, due 500 -> available 1500
	if err := svc.IncreaseDue(ctx, user.UserID, 1500.01); !errors.Is(err, service.ErrInsufficientCredit) {
		t.Fatalf("expected insufficient credit, got %v", err)
	}

	if err := svc.IncreaseDue(ctx, user.UserID, 1500); err != nil {
		t.Fatalf("IncreaseDue to limit: %v", err)
	}
	got, _ = svc.GetUserByID(ctx, user.UserID)
	if got.CurrentDue != "2000.00" {
		t.Fatalf("due want 2000.00 got %s", got.CurrentDue)
	}
}

func TestDecreaseDueRules(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := service.NewUserService(repo)
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, "Carol", "carol@example.com", "secret12")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := svc.IncreaseDue(ctx, user.UserID, 300); err != nil {
		t.Fatalf("IncreaseDue: %v", err)
	}

	if err := svc.DecreaseDue(ctx, user.UserID, 0); !errors.Is(err, service.ErrAmountMustBePositive) {
		t.Fatalf("expected positive amount error, got %v", err)
	}

	if err := svc.DecreaseDue(ctx, user.UserID, 301); !errors.Is(err, service.ErrPaymentExceedsDue) {
		t.Fatalf("expected exceeds due, got %v", err)
	}

	if err := svc.DecreaseDue(ctx, user.UserID, 100); err != nil {
		t.Fatalf("DecreaseDue: %v", err)
	}

	got, _ := svc.GetUserByID(ctx, user.UserID)
	if got.CurrentDue != "200.00" {
		t.Fatalf("due want 200.00 got %s", got.CurrentDue)
	}
}

func TestGetUserNotFound(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := service.NewUserService(repo)

	_, err := svc.GetUserByID(context.Background(), 99)
	if !errors.Is(err, service.ErrUserNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
