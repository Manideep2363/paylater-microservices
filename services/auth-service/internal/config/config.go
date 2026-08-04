package config

import (
	"os"

	sharedconfig "paylater/shared/config"
)

// Config extends shared config with auth-service dependency URLs.
type Config struct {
	*sharedconfig.Config

	UserServiceURL     string
	MerchantServiceURL string
	InternalAPIToken   string
}

// Load reads shared env vars plus auth-service service discovery settings.
func Load() *Config {
	base := sharedconfig.LoadConfig()

	cfg := &Config{
		Config:             base,
		UserServiceURL:     getEnv("USER_SERVICE_URL", "http://localhost:8082"),
		MerchantServiceURL: getEnv("MERCHANT_SERVICE_URL", "http://localhost:8083"),
		InternalAPIToken:   getEnv("INTERNAL_API_TOKEN", ""),
	}

	if os.Getenv("SERVER_PORT") == "" {
		cfg.ServerPort = "8081"
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
