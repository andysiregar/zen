package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server         ServerConfig
	MasterDatabase DatabaseConfig
	Redis          RedisConfig
	Jobs           JobsConfig
	EncryptionKey  string
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	MasterHost     string
	MasterPort     string
	MasterUser     string
	MasterPassword string
	MasterDatabase string
	SSLMode        string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JobsConfig struct {
	MaxWorkers          int
	MaxRetries          int
	RetryDelaySeconds   int
	JobTimeoutSeconds   int
	QueueCheckInterval  int
	EnableMetrics       bool
	MetricsPort         string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("BACKGROUND_JOBS_SERVICE_PORT", "8010"),
		},
		MasterDatabase: DatabaseConfig{
			MasterHost:     getEnv("MASTER_DB_HOST", "localhost"),
			MasterPort:     getEnv("MASTER_DB_PORT", "5432"),
			MasterUser:     getEnv("MASTER_DB_USER", "saas_user"),
			MasterPassword: getEnv("MASTER_DB_PASSWORD", "saas_password"),
			MasterDatabase: getEnv("MASTER_DB_NAME", "master_db"),
			SSLMode:        getEnv("MASTER_DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Jobs: JobsConfig{
			MaxWorkers:          getEnvAsInt("MAX_WORKERS", 10),
			MaxRetries:          getEnvAsInt("MAX_RETRIES", 3),
			RetryDelaySeconds:   getEnvAsInt("RETRY_DELAY_SECONDS", 60),
			JobTimeoutSeconds:   getEnvAsInt("JOB_TIMEOUT_SECONDS", 300),
			QueueCheckInterval:  getEnvAsInt("QUEUE_CHECK_INTERVAL", 5),
			EnableMetrics:       getEnvAsBool("ENABLE_METRICS", true),
			MetricsPort:         getEnv("METRICS_PORT", "9010"),
		},
		EncryptionKey: getEnv("ENCRYPTION_KEY", "your-32-char-encryption-key-here"),
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}