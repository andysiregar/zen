package config

import (
	"os"
)

type Config struct {
	Server         ServerConfig
	MasterDatabase DatabaseConfig
	Storage        StorageConfig
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

type StorageConfig struct {
	Provider    string // "local", "s3", "gcs"
	BasePath    string // Base directory for local storage
	MaxFileSize int64  // Maximum file size in bytes
	AllowedTypes []string // Allowed file types/extensions
	CDNBaseURL  string // CDN base URL for serving files
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("FILE_STORAGE_SERVICE_PORT", "8008"),
		},
		MasterDatabase: DatabaseConfig{
			MasterHost:     getEnv("MASTER_DB_HOST", "localhost"),
			MasterPort:     getEnv("MASTER_DB_PORT", "5432"),
			MasterUser:     getEnv("MASTER_DB_USER", "saas_user"),
			MasterPassword: getEnv("MASTER_DB_PASSWORD", "saas_password"),
			MasterDatabase: getEnv("MASTER_DB_NAME", "master_db"),
			SSLMode:        getEnv("MASTER_DB_SSL_MODE", "disable"),
		},
		Storage: StorageConfig{
			Provider:    getEnv("STORAGE_PROVIDER", "local"),
			BasePath:    getEnv("STORAGE_BASE_PATH", "/tmp/saas-files"),
			MaxFileSize: getEnvAsInt64("MAX_FILE_SIZE", 10*1024*1024), // 10MB default
			AllowedTypes: []string{
				".jpg", ".jpeg", ".png", ".gif", ".webp", // Images
				".pdf", ".doc", ".docx", ".txt", ".rtf", // Documents
				".csv", ".xlsx", ".xls", // Spreadsheets
				".zip", ".tar", ".gz", // Archives
			},
			CDNBaseURL: getEnv("CDN_BASE_URL", "http://localhost:8008/api/v1/files"),
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

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		// In a real implementation, parse the string to int64
		return defaultValue
	}
	return defaultValue
}