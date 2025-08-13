package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server         ServerConfig
	MasterDatabase DatabaseConfig
	Redis          RedisConfig
	Logger         LoggerConfig
	EncryptionKey  string
}

type ServerConfig struct {
	Port         string
	Environment  string
	ReadTimeout  int
	WriteTimeout int
}

type DatabaseConfig struct {
	MasterHost     string
	MasterUser     string
	MasterPassword string
	MasterDatabase string
	MasterPort     string
	SSLMode        string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	Database int
}

type LoggerConfig struct {
	Level string
	Format string // json or console
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("TICKET_SERVICE_PORT", "8004"),
			Environment:  getEnv("ENVIRONMENT", "development"),
			ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 30),
			WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 30),
		},
		MasterDatabase: DatabaseConfig{
			MasterHost:     getEnv("MASTER_DB_HOST", "localhost"),
			MasterUser:     getEnv("MASTER_DB_USER", "saas_user"),
			MasterPassword: getEnv("MASTER_DB_PASSWORD", "saas_password"),
			MasterDatabase: getEnv("MASTER_DB_NAME", "master_db"),
			MasterPort:     getEnv("MASTER_DB_PORT", "5432"),
			SSLMode:        getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			Database: getEnvAsInt("REDIS_DB", 0),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		EncryptionKey: getEnv("ENCRYPTION_KEY", "your-32-byte-encryption-key-here"),
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