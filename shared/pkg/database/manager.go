package database

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"

	"gorm.io/gorm"
)

// DatabaseManager manages connections to master and tenant databases
type DatabaseManager struct {
	masterDB        *gorm.DB
	tenantConns     map[string]*gorm.DB
	tenantConnsMutex sync.RWMutex
	encryptionKey   []byte
}

// TenantConnectionInfo holds the connection details for a tenant database
type TenantConnectionInfo struct {
	TenantID            string
	Host                string
	Port                int
	User                string
	EncryptedPassword   string
	DBName              string
	SSLMode             string
}

// NewDatabaseManager creates a new database manager with master database connection
func NewDatabaseManager(masterConfig Config, encryptionKey string) (*DatabaseManager, error) {
	masterDB, err := NewConnection(masterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master database: %w", err)
	}

	// Convert encryption key to 32 bytes for AES-256
	key := make([]byte, 32)
	copy(key, []byte(encryptionKey))

	return &DatabaseManager{
		masterDB:      masterDB,
		tenantConns:   make(map[string]*gorm.DB),
		encryptionKey: key,
	}, nil
}

// GetMasterDB returns the master database connection
func (dm *DatabaseManager) GetMasterDB() *gorm.DB {
	return dm.masterDB
}

// GetTenantDB returns a tenant database connection, creating it if it doesn't exist
func (dm *DatabaseManager) GetTenantDB(connInfo TenantConnectionInfo) (*gorm.DB, error) {
	dm.tenantConnsMutex.RLock()
	if conn, exists := dm.tenantConns[connInfo.TenantID]; exists {
		dm.tenantConnsMutex.RUnlock()
		return conn, nil
	}
	dm.tenantConnsMutex.RUnlock()

	// Need to create connection
	dm.tenantConnsMutex.Lock()
	defer dm.tenantConnsMutex.Unlock()

	// Double check in case another goroutine created it
	if conn, exists := dm.tenantConns[connInfo.TenantID]; exists {
		return conn, nil
	}

	// Decrypt password
	password, err := dm.decryptPassword(connInfo.EncryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt tenant database password: %w", err)
	}

	// Create tenant database connection
	tenantConfig := Config{
		Host:     connInfo.Host,
		Port:     connInfo.Port,
		User:     connInfo.User,
		Password: password,
		DBName:   connInfo.DBName,
		SSLMode:  connInfo.SSLMode,
	}

	tenantDB, err := NewConnection(tenantConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tenant database %s: %w", connInfo.TenantID, err)
	}

	// Store the connection
	dm.tenantConns[connInfo.TenantID] = tenantDB

	return tenantDB, nil
}

// CloseTenantConnection closes and removes a tenant database connection
func (dm *DatabaseManager) CloseTenantConnection(tenantID string) error {
	dm.tenantConnsMutex.Lock()
	defer dm.tenantConnsMutex.Unlock()

	if conn, exists := dm.tenantConns[tenantID]; exists {
		sqlDB, err := conn.DB()
		if err == nil {
			sqlDB.Close()
		}
		delete(dm.tenantConns, tenantID)
	}

	return nil
}

// CloseAllConnections closes all database connections
func (dm *DatabaseManager) CloseAllConnections() error {
	dm.tenantConnsMutex.Lock()
	defer dm.tenantConnsMutex.Unlock()

	// Close all tenant connections
	for tenantID, conn := range dm.tenantConns {
		sqlDB, err := conn.DB()
		if err == nil {
			sqlDB.Close()
		}
		delete(dm.tenantConns, tenantID)
	}

	// Close master connection
	if dm.masterDB != nil {
		sqlDB, err := dm.masterDB.DB()
		if err == nil {
			return sqlDB.Close()
		}
	}

	return nil
}

// EncryptPassword encrypts a password for storage
func (dm *DatabaseManager) EncryptPassword(password string) (string, error) {
	block, err := aes.NewCipher(dm.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptPassword decrypts an encrypted password
func (dm *DatabaseManager) decryptPassword(encryptedPassword string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(dm.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// HealthCheckAll performs health checks on all database connections
func (dm *DatabaseManager) HealthCheckAll() map[string]error {
	results := make(map[string]error)

	// Check master database
	results["master"] = HealthCheck(dm.masterDB)

	// Check tenant databases
	dm.tenantConnsMutex.RLock()
	for tenantID, conn := range dm.tenantConns {
		results[fmt.Sprintf("tenant_%s", tenantID)] = HealthCheck(conn)
	}
	dm.tenantConnsMutex.RUnlock()

	return results
}

// GetActiveTenantConnections returns the number of active tenant connections
func (dm *DatabaseManager) GetActiveTenantConnections() int {
	dm.tenantConnsMutex.RLock()
	defer dm.tenantConnsMutex.RUnlock()
	return len(dm.tenantConns)
}

// CleanupIdleConnections removes tenant connections that have been idle too long
func (dm *DatabaseManager) CleanupIdleConnections(maxIdleTime time.Duration) error {
	dm.tenantConnsMutex.Lock()
	defer dm.tenantConnsMutex.Unlock()

	for tenantID, conn := range dm.tenantConns {
		sqlDB, err := conn.DB()
		if err != nil {
			continue
		}

		// Check if connection is still alive
		if err := sqlDB.Ping(); err != nil {
			// Connection is dead, remove it
			sqlDB.Close()
			delete(dm.tenantConns, tenantID)
			continue
		}

		// Check idle time (this is a simplified check)
		stats := sqlDB.Stats()
		if stats.Idle > 0 {
			// If there are idle connections and we want to clean up
			// This is a basic implementation - you might want to add more sophisticated tracking
			sqlDB.SetMaxIdleConns(1) // Reduce idle connections
		}
	}

	return nil
}