package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"paylater/services/ledger-service/internal/client"
	"paylater/services/ledger-service/internal/repository"
	"paylater/services/ledger-service/internal/service"
)

type fakeUsers struct {
	mu            sync.Mutex
	due           float64
	limit         float64
	increaseCalls []float64
	decreaseCalls []float64
	failIncrease  error
	failDecrease  error
	failNextDec   error
	failNextInc   error
}

func newFakeUsers(due, limit float64) *fakeUsers {
	return &fakeUsers{due: due, limit: limit}
}

func (f *fakeUsers) IncreaseDue(ctx context.Context, userID int32, amount float64) error {
	_ = ctx
	_ = userID
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failNextInc != nil {
		err := f.failNextInc
		f.failNextInc = nil
		return err
	}
	if f.failIncrease != nil {
		return f.failIncrease
	}
	if amount <= 0 {
		return client.ErrAmountMustBePositive
	}
	if f.due+amount > f.limit {
		return client.ErrInsufficientCredit
	}
	f.due += amount
	f.increaseCalls = append(f.increaseCalls, amount)
	return nil
}

func (f *fakeUsers) DecreaseDue(ctx context.Context, userID int32, amount float64) error {
	_ = ctx
	_ = userID
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failNextDec != nil {
		err := f.failNextDec
		f.failNextDec = nil
		return err
	}
	if f.failDecrease != nil {
		return f.failDecrease
	}
	if amount <= 0 {
		return client.ErrAmountMustBePositive
	}
	if amount > f.due {
		return client.ErrPaymentExceedsDue
	}
	f.due -= amount
	f.decreaseCalls = append(f.decreaseCalls, amount)
	return nil
}

type fakeMerchants struct {
	byID        map[int32]client.Merchant
	unavailable bool
	notFoundAll bool
}

func (f *fakeMerchants) GetMerchantByID(ctx context.Context, merchantID int32) (client.Merchant, error) {
	_ = ctx
	if f.unavailable {
		return client.Merchant{}, client.ErrUnavailable
	}
	if f.notFoundAll {
		return client.Merchant{}, client.ErrNotFound
	}
	m, ok := f.byID[merchantID]
	if !ok {
		return client.Merchant{}, client.ErrNotFound
	}
	return m, nil
}

func TestPurchaseAmountNonPositive(t *testing.T) {
	svc := service.NewLedgerService(repository.NewMemoryStore(), newFakeUsers(0, 2000), &fakeMerchants{})
	_, err := svc.Purchase(context.Background(), 1, 1, 0)
	if !errors.Is(err, service.ErrAmountMustBePositive) {
		t.Fatalf("got %v", err)
	}
}

func TestPurchaseInvalidMerchant(t *testing.T) {
	users := newFakeUsers(0, 2000)
	merchants := &fakeMerchants{byID: map[int32]client.Merchant{}}
	svc := service.NewLedgerService(repository.NewMemoryStore(), users, merchants)
	_, err := svc.Purchase(context.Background(), 1, 99, 100)
	if !errors.Is(err, service.ErrMerchantNotFound) {
		t.Fatalf("got %v", err)
	}
	if len(users.increaseCalls) != 0 {
		t.Fatal("should not increase due for invalid merchant")
	}
}

func TestPurchaseMerchantUnavailable(t *testing.T) {
	svc := service.NewLedgerService(
		repository.NewMemoryStore(),
		newFakeUsers(0, 2000),
		&fakeMerchants{unavailable: true},
	)
	_, err := svc.Purchase(context.Background(), 1, 1, 100)
	if !errors.Is(err, service.ErrServiceUnavailable) {
		t.Fatalf("got %v", err)
	}
}

func TestPurchaseUserUnavailable(t *testing.T) {
	users := newFakeUsers(0, 2000)
	users.failIncrease = client.ErrUnavailable
	merchants := &fakeMerchants{byID: map[int32]client.Merchant{
		1: {MerchantID: 1, CommissionPercentage: "5.00"},
	}}
	svc := service.NewLedgerService(repository.NewMemoryStore(), users, merchants)
	_, err := svc.Purchase(context.Background(), 1, 1, 100)
	if !errors.Is(err, service.ErrServiceUnavailable) {
		t.Fatalf("got %v", err)
	}
}

func TestPurchaseInsufficientCredit(t *testing.T) {
	users := newFakeUsers(1900, 2000)
	merchants := &fakeMerchants{byID: map[int32]client.Merchant{
		1: {MerchantID: 1, CommissionPercentage: "5.00"},
	}}
	svc := service.NewLedgerService(repository.NewMemoryStore(), users, merchants)
	_, err := svc.Purchase(context.Background(), 1, 1, 200)
	if !errors.Is(err, service.ErrInsufficientCredit) {
		t.Fatalf("got %v", err)
	}
}

func TestPurchaseSuccessAndCommission(t *testing.T) {
	store := repository.NewMemoryStore()
	users := newFakeUsers(0, 2000)
	merchants := &fakeMerchants{byID: map[int32]client.Merchant{
		2: {MerchantID: 2, CommissionPercentage: "5.00"},
	}}
	svc := service.NewLedgerService(store, users, merchants)

	tx, err := svc.Purchase(context.Background(), 10, 2, 100)
	if err != nil {
		t.Fatalf("Purchase: %v", err)
	}
	if tx.Amount != "100.00" || tx.CommissionPercentage != "5.00" || tx.CommissionAmount != "5.00" {
		t.Fatalf("tx=%+v", tx)
	}
	if users.due != 100 {
		t.Fatalf("due=%v", users.due)
	}
}

func TestPurchaseInsertFailCompensates(t *testing.T) {
	store := repository.NewMemoryStore()
	store.FailNextTx = true
	users := newFakeUsers(0, 2000)
	merchants := &fakeMerchants{byID: map[int32]client.Merchant{
		1: {MerchantID: 1, CommissionPercentage: "10.00"},
	}}
	svc := service.NewLedgerService(store, users, merchants)

	_, err := svc.Purchase(context.Background(), 1, 1, 50)
	if !errors.Is(err, service.ErrInternal) {
		t.Fatalf("got %v", err)
	}
	if users.due != 0 {
		t.Fatalf("expected compensated due=0 got %v", users.due)
	}
	if len(users.increaseCalls) != 1 || len(users.decreaseCalls) != 1 {
		t.Fatalf("increase=%v decrease=%v", users.increaseCalls, users.decreaseCalls)
	}
	if users.increaseCalls[0] != users.decreaseCalls[0] {
		t.Fatalf("compensation amount mismatch")
	}
}

func TestPurchaseInsertFailCompensationFails(t *testing.T) {
	store := repository.NewMemoryStore()
	store.FailNextTx = true
	users := newFakeUsers(0, 2000)
	users.failNextDec = client.ErrUnavailable
	merchants := &fakeMerchants{byID: map[int32]client.Merchant{
		1: {MerchantID: 1, CommissionPercentage: "5.00"},
	}}
	svc := service.NewLedgerService(store, users, merchants)

	_, err := svc.Purchase(context.Background(), 1, 1, 40)
	if !errors.Is(err, service.ErrInternal) {
		t.Fatalf("got %v", err)
	}
	if users.due != 40 {
		t.Fatalf("due should remain increased when compensate fails, got %v", users.due)
	}
}

func TestRepayAmountNonPositive(t *testing.T) {
	svc := service.NewLedgerService(repository.NewMemoryStore(), newFakeUsers(100, 2000), &fakeMerchants{})
	_, err := svc.Repay(context.Background(), 1, -1)
	if !errors.Is(err, service.ErrAmountMustBePositive) {
		t.Fatalf("got %v", err)
	}
}

func TestRepayExceedsDue(t *testing.T) {
	svc := service.NewLedgerService(repository.NewMemoryStore(), newFakeUsers(50, 2000), &fakeMerchants{})
	_, err := svc.Repay(context.Background(), 1, 60)
	if !errors.Is(err, service.ErrPaymentExceedsDue) {
		t.Fatalf("got %v", err)
	}
}

func TestRepayUserUnavailable(t *testing.T) {
	users := newFakeUsers(100, 2000)
	users.failDecrease = client.ErrUnavailable
	svc := service.NewLedgerService(repository.NewMemoryStore(), users, &fakeMerchants{})
	_, err := svc.Repay(context.Background(), 1, 10)
	if !errors.Is(err, service.ErrServiceUnavailable) {
		t.Fatalf("got %v", err)
	}
}

func TestRepaySuccess(t *testing.T) {
	store := repository.NewMemoryStore()
	users := newFakeUsers(100, 2000)
	svc := service.NewLedgerService(store, users, &fakeMerchants{})
	p, err := svc.Repay(context.Background(), 5, 25)
	if err != nil || p.Amount != "25.00" || users.due != 75 {
		t.Fatalf("p=%+v due=%v err=%v", p, users.due, err)
	}
}

func TestRepayInsertFailCompensates(t *testing.T) {
	store := repository.NewMemoryStore()
	store.FailNextPay = true
	users := newFakeUsers(100, 2000)
	svc := service.NewLedgerService(store, users, &fakeMerchants{})
	_, err := svc.Repay(context.Background(), 1, 30)
	if !errors.Is(err, service.ErrInternal) {
		t.Fatalf("got %v", err)
	}
	if users.due != 100 {
		t.Fatalf("expected compensated due=100 got %v", users.due)
	}
}

func TestRepayInsertFailCompensationFails(t *testing.T) {
	store := repository.NewMemoryStore()
	store.FailNextPay = true
	users := newFakeUsers(100, 2000)
	users.failNextInc = client.ErrUnavailable
	svc := service.NewLedgerService(store, users, &fakeMerchants{})
	_, err := svc.Repay(context.Background(), 1, 20)
	if !errors.Is(err, service.ErrInternal) {
		t.Fatalf("got %v", err)
	}
	if users.due != 80 {
		t.Fatalf("due should remain decreased when compensate fails, got %v", users.due)
	}
}
