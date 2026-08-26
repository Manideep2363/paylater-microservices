package config

import (
	"os"

	sharedconfig "paylater/shared/config"
)

// Config holds ledger-service settings including downstream service URLs.
type Config struct {
	*sharedconfig.Config

	UserServiceURL     string
	MerchantServiceURL string
	InternalAPIToken   string
}

// Load reads environment configuration for ledger-service.
func Load() *Config {
	base := sharedconfig.LoadConfig()

	cfg := &Config{
		Config:             base,
		UserServiceURL:     getEnv("USER_SERVICE_URL", "http://localhost:8082"),
		MerchantServiceURL: getEnv("MERCHANT_SERVICE_URL", "http://localhost:8083"),
		InternalAPIToken:   getEnv("INTERNAL_API_TOKEN", ""),
	}

	if os.Getenv("DB_NAME") == "" {
		cfg.DBName = "paylater_ledger"
	}
	if os.Getenv("SERVER_PORT") == "" {
		cfg.ServerPort = "8084"
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
