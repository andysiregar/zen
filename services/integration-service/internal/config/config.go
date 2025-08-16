package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server         ServerConfig
	MasterDatabase DatabaseConfig
	Integration    IntegrationConfig
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

type IntegrationConfig struct {
	WebhookSecret          string
	MaxRetryAttempts       int
	RetryBackoffSeconds    int
	WebhookTimeoutSeconds  int
	MaxConcurrentWebhooks  int
	EnableAsyncProcessing  bool
	QueueSize             int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("INTEGRATION_SERVICE_PORT", "8009"),
		},
		MasterDatabase: DatabaseConfig{
			MasterHost:     getEnv("MASTER_DB_HOST", "localhost"),
			MasterPort:     getEnv("MASTER_DB_PORT", "5432"),
			MasterUser:     getEnv("MASTER_DB_USER", "saas_user"),
			MasterPassword: getEnv("MASTER_DB_PASSWORD", "saas_password"),
			MasterDatabase: getEnv("MASTER_DB_NAME", "master_db"),
			SSLMode:        getEnv("MASTER_DB_SSL_MODE", "disable"),
		},
		Integration: IntegrationConfig{
			WebhookSecret:          getEnv("WEBHOOK_SECRET", "your-webhook-secret-key"),
			MaxRetryAttempts:       getEnvAsInt("MAX_RETRY_ATTEMPTS", 3),
			RetryBackoffSeconds:    getEnvAsInt("RETRY_BACKOFF_SECONDS", 30),
			WebhookTimeoutSeconds:  getEnvAsInt("WEBHOOK_TIMEOUT_SECONDS", 30),
			MaxConcurrentWebhooks:  getEnvAsInt("MAX_CONCURRENT_WEBHOOKS", 10),
			EnableAsyncProcessing:  getEnvAsBool("ENABLE_ASYNC_PROCESSING", true),
			QueueSize:             getEnvAsInt("QUEUE_SIZE", 1000),
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