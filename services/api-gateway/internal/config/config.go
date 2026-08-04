package config

import (
	"fmt"
	"net/url"
	"os"
	"time"
)

// Config holds gateway settings (no JWT, DB, or internal token).
type Config struct {
	ServerPort string

	AuthServiceURL     *url.URL
	UserServiceURL     *url.URL
	MerchantServiceURL *url.URL
	LedgerServiceURL   *url.URL
	ReportServiceURL   *url.URL

	ProxyTimeout      time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

// Load reads and validates gateway configuration.
func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		ProxyTimeout:      15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	var err error
	if cfg.AuthServiceURL, err = parseURL("AUTH_SERVICE_URL", "http://localhost:8081"); err != nil {
		return nil, err
	}
	if cfg.UserServiceURL, err = parseURL("USER_SERVICE_URL", "http://localhost:8082"); err != nil {
		return nil, err
	}
	if cfg.MerchantServiceURL, err = parseURL("MERCHANT_SERVICE_URL", "http://localhost:8083"); err != nil {
		return nil, err
	}
	if cfg.LedgerServiceURL, err = parseURL("LEDGER_SERVICE_URL", "http://localhost:8084"); err != nil {
		return nil, err
	}
	if cfg.ReportServiceURL, err = parseURL("REPORT_SERVICE_URL", "http://localhost:8085"); err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseURL(envKey, fallback string) (*url.URL, error) {
	raw := getEnv(envKey, fallback)
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid URL %q: %w", envKey, raw, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("%s: URL must include scheme and host: %q", envKey, raw)
	}
	return u, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
