package config

import (
	"os"
	"strconv"

	"github.com/zen/shared/pkg/database"
	"github.com/zen/shared/pkg/redis"
)

type Config struct {
	Environment string
	Port        string
	Database    *database.Config
	Redis       *redis.Config
	JWT         *JWTConfig
}

type JWTConfig struct {
	SecretKey        string
	AccessTokenTTL   int // hours
	RefreshTokenTTL  int // hours
}

func Load() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PLATFORM_ADMIN_PORT", "8014"),
		Database: &database.Config{
			MasterHost:     getEnv("MASTER_DB_HOST", "localhost"),
			MasterPort:     getEnv("MASTER_DB_PORT", "5432"),
			MasterUser:     getEnv("MASTER_DB_USER", "saas_user"),
			MasterPassword: getEnv("MASTER_DB_PASSWORD", "saas_password"),
			MasterDBName:   getEnv("MASTER_DB_NAME", "master_db"),
			SSLMode:        getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: &redis.Config{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: &JWTConfig{
			SecretKey:       getEnv("PLATFORM_JWT_SECRET", "your-platform-admin-secret-key"),
			AccessTokenTTL:  getEnvAsInt("PLATFORM_JWT_ACCESS_TTL", 24),  // 24 hours
			RefreshTokenTTL: getEnvAsInt("PLATFORM_JWT_REFRESH_TTL", 168), // 7 days
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}