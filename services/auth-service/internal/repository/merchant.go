package repository

import "context"

// Merchant is the auth-relevant merchant record owned by merchant-service.
type Merchant struct {
	MerchantID           int32
	Name                 string
	Email                string
	Phone                string
	PasswordHash         string
	CommissionPercentage string
}

// MerchantRepository abstracts merchant credential access via merchant-service REST.
type MerchantRepository interface {
	// CreateMerchant creates a merchant with a plain password; merchant-service hashes it.
	CreateMerchant(
		ctx context.Context,
		name, email, phone, password string,
		commission float64,
	) (Merchant, error)
	GetMerchantByEmail(ctx context.Context, email string) (Merchant, error)
}
