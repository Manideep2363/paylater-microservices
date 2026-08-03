package repository

import "context"

// Merchant is the auth-relevant merchant record.
// In later phases this will be owned by merchant-service.
type Merchant struct {
	MerchantID           int32
	Name                 string
	Email                string
	Phone                string
	PasswordHash         string
	CommissionPercentage string
}

// MerchantRepository abstracts merchant credential storage.
// Temporary in-memory impl today; future REST client to merchant-service.
type MerchantRepository interface {
	CreateMerchant(
		ctx context.Context,
		name, email, phone, passwordHash, commissionPercentage string,
	) (Merchant, error)
	GetMerchantByEmail(ctx context.Context, email string) (Merchant, error)
}
