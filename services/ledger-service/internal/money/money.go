package money

import (
	"fmt"
	"math"
	"strconv"
)

// NormalizeAmount rounds a monetary value to two decimal places.
// The same normalized value must be used for due mutations and compensation.
func NormalizeAmount(amount float64) (float64, string, error) {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, "", fmt.Errorf("invalid amount")
	}
	normalized := math.Round(amount*100) / 100
	return normalized, Format(normalized), nil
}

// Format renders a monetary value with exactly two decimal places.
func Format(amount float64) string {
	return strconv.FormatFloat(amount, 'f', 2, 64)
}

// Parse converts a DECIMAL string to float64.
func Parse(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

// CommissionAmount calculates amount * percentage / 100 and normalizes to 2dp.
func CommissionAmount(amount, percentage float64) (float64, string) {
	raw := amount * percentage / 100
	normalized := math.Round(raw*100) / 100
	return normalized, Format(normalized)
}
