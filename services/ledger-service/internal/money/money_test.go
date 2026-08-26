package money_test

import (
	"testing"

	"paylater/services/ledger-service/internal/money"
)

func TestCommissionAmountRounding(t *testing.T) {
	cases := []struct {
		amount, pct float64
		want        string
	}{
		{100, 5, "5.00"},
		{33.33, 10, "3.33"},
		{10.00, 3.5, "0.35"},
		{99.99, 7.5, "7.50"},
		{1.00, 3, "0.03"},
	}
	for _, tc := range cases {
		_, got := money.CommissionAmount(tc.amount, tc.pct)
		if got != tc.want {
			t.Fatalf("amount=%v pct=%v got=%s want=%s", tc.amount, tc.pct, got, tc.want)
		}
	}
}

func TestNormalizeAmount(t *testing.T) {
	n, s, err := money.NormalizeAmount(10.456)
	if err != nil || n != 10.46 || s != "10.46" {
		t.Fatalf("got n=%v s=%s err=%v", n, s, err)
	}
}
