package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	Port        string
	Environment string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnvOrDefault("DATABASE_URL", "postgres://error_logs_user:error_logs_password@localhost:5432/error_logs?sslmode=disable"),
		RedisURL:    getEnvOrDefault("REDIS_URL", "redis://localhost:6379"),
		Port:        getEnvOrDefault("PORT", "8080"),
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
