package config

import (
	"os"
)

// Config holds all application configuration.
type Config struct {
	Env         string
	Port        string
	DatabaseURL string
}

// Load reads configuration from environment variables with sane defaults.
func Load() *Config {
	return &Config{
		Env:         getEnv("APP_ENV", "development"),
		Port:        getEnv("APP_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/userdb?sslmode=disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
