package config

import (
	"os"

	sharedconfig "paylater/shared/config"
)

// Config holds report-service settings (no database).
type Config struct {
	*sharedconfig.Config

	UserServiceURL   string
	LedgerServiceURL string
	InternalAPIToken string
}

// Load reads report-service configuration from the environment.
func Load() *Config {
	base := sharedconfig.LoadConfig()
	cfg := &Config{
		Config:           base,
		UserServiceURL:   getEnv("USER_SERVICE_URL", "http://localhost:8082"),
		LedgerServiceURL: getEnv("LEDGER_SERVICE_URL", "http://localhost:8084"),
		InternalAPIToken: getEnv("INTERNAL_API_TOKEN", ""),
	}
	if os.Getenv("SERVER_PORT") == "" {
		cfg.ServerPort = "8085"
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
