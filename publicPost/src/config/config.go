// src/config/config.go

package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

func LoadConfig() *Config {
	accessExpiry, _ := strconv.Atoi(getEnv("ACCESS_TOKEN_EXPIRY", "15"))
	refreshExpiry, _ := strconv.Atoi(getEnv("REFRESH_TOKEN_EXPIRY", "10080"))

	return &Config{
		JWTSecret:          getEnv("JWT_SECRET", "your_default_secret"),
		AccessTokenExpiry:  time.Duration(accessExpiry) * time.Minute,
		RefreshTokenExpiry: time.Duration(refreshExpiry) * time.Minute,
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
