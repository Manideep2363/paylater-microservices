package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds shared environment settings used by every service.
// Database fields are loaded for later phases; Phase 0 does not connect to MySQL.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	ServerPort string
	JWTSecret  string

	AdminEmail    string
	AdminPassword string
}

// LoadConfig reads configuration from environment variables.
// A missing .env file is ignored so services can start with OS env or defaults.
func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "paylater"),

		ServerPort: getEnv("SERVER_PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", "dev-secret-change-me"),

		AdminEmail:    getEnv("ADMIN_EMAIL", ""),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
