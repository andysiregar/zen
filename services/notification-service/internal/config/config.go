package config

import (
	"os"
)

type Config struct {
	Server         ServerConfig
	MasterDatabase DatabaseConfig
	EncryptionKey  string
	SMTPConfig     SMTPConfig
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

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("NOTIFICATION_SERVICE_PORT", "8007"),
		},
		MasterDatabase: DatabaseConfig{
			MasterHost:     getEnv("MASTER_DB_HOST", "localhost"),
			MasterPort:     getEnv("MASTER_DB_PORT", "5432"),
			MasterUser:     getEnv("MASTER_DB_USER", "saas_user"),
			MasterPassword: getEnv("MASTER_DB_PASSWORD", "saas_password"),
			MasterDatabase: getEnv("MASTER_DB_NAME", "master_db"),
			SSLMode:        getEnv("MASTER_DB_SSL_MODE", "disable"),
		},
		EncryptionKey: getEnv("ENCRYPTION_KEY", "your-32-char-encryption-key-here"),
		SMTPConfig: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "localhost"),
			Port:     getEnv("SMTP_PORT", "587"),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@yourdomain.com"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}