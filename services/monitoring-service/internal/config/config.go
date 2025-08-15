package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server         ServerConfig
	MasterDatabase DatabaseConfig
	Redis          RedisConfig
	Monitoring     MonitoringConfig
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

type MonitoringConfig struct {
	MetricsPort             string
	HealthCheckInterval     int
	MetricsRetentionDays    int
	AlertingEnabled         bool
	AlertWebhookURL         string
	ServiceDiscoveryEnabled bool
	ServicesToMonitor       []string
	LogLevel                string
	MaxLogSizeMB           int
	MaxLogFiles            int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("MONITORING_SERVICE_PORT", "8011"),
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
		Monitoring: MonitoringConfig{
			MetricsPort:             getEnv("METRICS_PORT", "9011"),
			HealthCheckInterval:     getEnvAsInt("HEALTH_CHECK_INTERVAL", 30),
			MetricsRetentionDays:    getEnvAsInt("METRICS_RETENTION_DAYS", 30),
			AlertingEnabled:         getEnvAsBool("ALERTING_ENABLED", true),
			AlertWebhookURL:         getEnv("ALERT_WEBHOOK_URL", ""),
			ServiceDiscoveryEnabled: getEnvAsBool("SERVICE_DISCOVERY_ENABLED", true),
			ServicesToMonitor:       getEnvAsStringSlice("SERVICES_TO_MONITOR", []string{"api-gateway", "auth-service", "tenant-management", "ticket-service", "chat-service"}),
			LogLevel:                getEnv("LOG_LEVEL", "info"),
			MaxLogSizeMB:           getEnvAsInt("MAX_LOG_SIZE_MB", 100),
			MaxLogFiles:            getEnvAsInt("MAX_LOG_FILES", 10),
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

func getEnvAsStringSlice(key string, defaultValue []string) []string {
	// Simple implementation - in production you'd want proper parsing
	return defaultValue
}