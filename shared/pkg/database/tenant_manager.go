package database

import (
	"fmt"
	"sync"
	"gorm.io/gorm"
	"github.com/zen/shared/pkg/models"
	"github.com/zen/shared/pkg/tenant_models"
)

// TenantDatabaseManager manages connections to multiple tenant databases
type TenantDatabaseManager struct {
	masterDB      *gorm.DB
	tenantDBs     map[string]*gorm.DB
	tenantConfigs map[string]models.Tenant
	mutex         sync.RWMutex
	encryptionKey string
}

// NewTenantDatabaseManager creates a new multi-tenant database manager
func NewTenantDatabaseManager(masterDB *gorm.DB, encryptionKey string) *TenantDatabaseManager {
	return &TenantDatabaseManager{
		masterDB:      masterDB,
		tenantDBs:     make(map[string]*gorm.DB),
		tenantConfigs: make(map[string]models.Tenant),
		encryptionKey: encryptionKey,
	}
}

// GetTenantDB returns a database connection for a specific tenant
func (tdm *TenantDatabaseManager) GetTenantDB(tenantID string) (*gorm.DB, error) {
	tdm.mutex.RLock()
	if db, exists := tdm.tenantDBs[tenantID]; exists {
		tdm.mutex.RUnlock()
		return db, nil
	}
	tdm.mutex.RUnlock()

	// Database not cached, fetch tenant config and create connection
	return tdm.connectToTenantDB(tenantID)
}

// connectToTenantDB establishes a new connection to a tenant's database
func (tdm *TenantDatabaseManager) connectToTenantDB(tenantID string) (*gorm.DB, error) {
	tdm.mutex.Lock()
	defer tdm.mutex.Unlock()

	// Double-check pattern - connection might have been created while waiting for lock
	if db, exists := tdm.tenantDBs[tenantID]; exists {
		return db, nil
	}

	// Fetch tenant configuration from master database
	var tenant models.Tenant
	err := tdm.masterDB.Where("id = ? AND status = ?", tenantID, models.TenantStatusActive).First(&tenant).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant %s: %w", tenantID, err)
	}

	// Create database configuration for tenant
	config := Config{
		Host:     tenant.DbHost,
		Port:     tenant.DbPort,
		User:     tenant.DbUser,
		Password: tenant.DbPasswordEncrypted, // In production, decrypt this
		DBName:   tenant.DbName,
		SSLMode:  tenant.DbSslMode,
	}

	// Create database connection
	db, err := NewConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tenant %s database: %w", tenantID, err)
	}

	// Auto-migrate tenant-specific tables
	err = tdm.migrateTenantTables(db)
	if err != nil {
		db.DB()
		return nil, fmt.Errorf("failed to migrate tenant %s tables: %w", tenantID, err)
	}

	// Cache the connection and tenant config
	tdm.tenantDBs[tenantID] = db
	tdm.tenantConfigs[tenantID] = tenant

	return db, nil
}

// migrateTenantTables runs auto-migration for tenant-specific tables
func (tdm *TenantDatabaseManager) migrateTenantTables(db *gorm.DB) error {
	// Import tenant models for migration
	err := db.AutoMigrate(
		&tenant_models.Ticket{},
		&tenant_models.Comment{},
		&tenant_models.Attachment{},
		&tenant_models.Project{},
		&tenant_models.ProjectMember{},
		// Add more tenant-specific models as needed
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate tables: %w", err)
	}

	return nil
}

// GetTenantConfig returns the cached tenant configuration
func (tdm *TenantDatabaseManager) GetTenantConfig(tenantID string) (models.Tenant, bool) {
	tdm.mutex.RLock()
	defer tdm.mutex.RUnlock()
	
	config, exists := tdm.tenantConfigs[tenantID]
	return config, exists
}

// RefreshTenantConnection forces a refresh of a tenant's database connection
func (tdm *TenantDatabaseManager) RefreshTenantConnection(tenantID string) error {
	tdm.mutex.Lock()
	defer tdm.mutex.Unlock()

	// Close existing connection if it exists
	if db, exists := tdm.tenantDBs[tenantID]; exists {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
		delete(tdm.tenantDBs, tenantID)
		delete(tdm.tenantConfigs, tenantID)
	}

	// Reconnect will happen on next GetTenantDB call
	return nil
}

// CloseTenantConnection closes and removes a tenant database connection
func (tdm *TenantDatabaseManager) CloseTenantConnection(tenantID string) error {
	tdm.mutex.Lock()
	defer tdm.mutex.Unlock()

	if db, exists := tdm.tenantDBs[tenantID]; exists {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
		delete(tdm.tenantDBs, tenantID)
		delete(tdm.tenantConfigs, tenantID)
	}

	return nil
}

// CloseAllConnections closes all tenant database connections
func (tdm *TenantDatabaseManager) CloseAllConnections() error {
	tdm.mutex.Lock()
	defer tdm.mutex.Unlock()

	for tenantID, db := range tdm.tenantDBs {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
		delete(tdm.tenantDBs, tenantID)
		delete(tdm.tenantConfigs, tenantID)
	}

	return nil
}

// GetActiveTenantsCount returns the number of active tenant connections
func (tdm *TenantDatabaseManager) GetActiveTenantsCount() int {
	tdm.mutex.RLock()
	defer tdm.mutex.RUnlock()
	
	return len(tdm.tenantDBs)
}

// ListActiveTenants returns a list of tenant IDs with active connections
func (tdm *TenantDatabaseManager) ListActiveTenants() []string {
	tdm.mutex.RLock()
	defer tdm.mutex.RUnlock()
	
	tenants := make([]string, 0, len(tdm.tenantDBs))
	for tenantID := range tdm.tenantDBs {
		tenants = append(tenants, tenantID)
	}
	
	return tenants
}