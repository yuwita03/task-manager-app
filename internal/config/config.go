package config

import (
	"os"
	"time"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	TokenExpiry time.Duration
	Port        string
}

func Load() Config {
	tokenExpiry := 24 * time.Hour
	if v := os.Getenv("TOKEN_EXPIRY_HOURS"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			tokenExpiry = d
		}
	}

	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		TokenExpiry: tokenExpiry,
		Port:        getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}