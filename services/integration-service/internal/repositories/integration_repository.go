package repositories

import (
	"encoding/json"

	"github.com/zen/shared/pkg/database"
	"integration-service/internal/models"
)

type IntegrationRepository interface {
	// Third-party integration operations
	CreateIntegration(tenantID string, integration *models.ThirdPartyIntegration) error
	GetIntegration(tenantID, integrationID string) (*models.ThirdPartyIntegration, error)
	UpdateIntegration(tenantID string, integration *models.ThirdPartyIntegration) error
	DeleteIntegration(tenantID, integrationID string) error
	ListIntegrations(tenantID string, limit, offset int) ([]*models.ThirdPartyIntegration, int64, error)
	GetActiveIntegrations(tenantID string) ([]*models.ThirdPartyIntegration, error)
	
	// Data sync operations
	CreateDataSync(tenantID string, dataSync *models.DataSync) error
	GetDataSync(tenantID, syncID string) (*models.DataSync, error)
	UpdateDataSync(tenantID string, dataSync *models.DataSync) error
	ListDataSyncs(tenantID, integrationID string, limit, offset int) ([]*models.DataSync, int64, error)
	GetPendingSyncs(tenantID string) ([]*models.DataSync, error)
	GetRunningSyncs(tenantID string) ([]*models.DataSync, error)
	
	// Statistics
	GetIntegrationStats(tenantID string) (*models.IntegrationStats, error)
}

type integrationRepository struct {
	tenantDBManager *database.TenantDatabaseManager
}

func NewIntegrationRepository(tenantDBManager *database.TenantDatabaseManager) IntegrationRepository {
	return &integrationRepository{
		tenantDBManager: tenantDBManager,
	}
}

func (r *integrationRepository) CreateIntegration(tenantID string, integration *models.ThirdPartyIntegration) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	configJSON, _ := json.Marshal(integration.Configuration)
	credentialsJSON, _ := json.Marshal(integration.Credentials)
	metadataJSON, _ := json.Marshal(integration.Metadata)

	result := db.Exec(`
		INSERT INTO third_party_integrations (id, tenant_id, name, provider, configuration, credentials, active, sync_enabled, last_sync_at, next_sync_at, sync_interval, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, integration.ID, integration.TenantID, integration.Name, integration.Provider,
		string(configJSON), string(credentialsJSON), integration.Active, integration.SyncEnabled,
		integration.LastSyncAt, integration.NextSyncAt, integration.SyncInterval,
		string(metadataJSON), integration.CreatedAt, integration.UpdatedAt)

	return result.Error
}

func (r *integrationRepository) GetIntegration(tenantID, integrationID string) (*models.ThirdPartyIntegration, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var integration models.ThirdPartyIntegration
	var configJSON, credentialsJSON, metadataJSON string
	err = db.Raw(`
		SELECT id, tenant_id, name, provider, configuration, credentials, active, sync_enabled, last_sync_at, next_sync_at, sync_interval, metadata, created_at, updated_at
		FROM third_party_integrations 
		WHERE id = ? AND tenant_id = ?
	`, integrationID, tenantID).Row().Scan(
		&integration.ID, &integration.TenantID, &integration.Name, &integration.Provider,
		&configJSON, &credentialsJSON, &integration.Active, &integration.SyncEnabled,
		&integration.LastSyncAt, &integration.NextSyncAt, &integration.SyncInterval,
		&metadataJSON, &integration.CreatedAt, &integration.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(configJSON), &integration.Configuration)
	json.Unmarshal([]byte(credentialsJSON), &integration.Credentials)
	json.Unmarshal([]byte(metadataJSON), &integration.Metadata)

	return &integration, nil
}

func (r *integrationRepository) UpdateIntegration(tenantID string, integration *models.ThirdPartyIntegration) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	configJSON, _ := json.Marshal(integration.Configuration)
	credentialsJSON, _ := json.Marshal(integration.Credentials)
	metadataJSON, _ := json.Marshal(integration.Metadata)

	result := db.Exec(`
		UPDATE third_party_integrations 
		SET name = ?, configuration = ?, credentials = ?, active = ?, sync_enabled = ?, last_sync_at = ?, next_sync_at = ?, sync_interval = ?, metadata = ?, updated_at = ?
		WHERE id = ? AND tenant_id = ?
	`, integration.Name, string(configJSON), string(credentialsJSON), integration.Active,
		integration.SyncEnabled, integration.LastSyncAt, integration.NextSyncAt,
		integration.SyncInterval, string(metadataJSON), integration.UpdatedAt,
		integration.ID, tenantID)

	return result.Error
}

func (r *integrationRepository) DeleteIntegration(tenantID, integrationID string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		DELETE FROM third_party_integrations 
		WHERE id = ? AND tenant_id = ?
	`, integrationID, tenantID)

	return result.Error
}

func (r *integrationRepository) ListIntegrations(tenantID string, limit, offset int) ([]*models.ThirdPartyIntegration, int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	var count int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM third_party_integrations 
		WHERE tenant_id = ?
	`, tenantID).Scan(&count)

	// Get integrations
	rows, err := db.Raw(`
		SELECT id, tenant_id, name, provider, configuration, credentials, active, sync_enabled, last_sync_at, next_sync_at, sync_interval, metadata, created_at, updated_at
		FROM third_party_integrations 
		WHERE tenant_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, tenantID, limit, offset).Rows()

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var integrations []*models.ThirdPartyIntegration
	for rows.Next() {
		var integration models.ThirdPartyIntegration
		var configJSON, credentialsJSON, metadataJSON string
		rows.Scan(
			&integration.ID, &integration.TenantID, &integration.Name, &integration.Provider,
			&configJSON, &credentialsJSON, &integration.Active, &integration.SyncEnabled,
			&integration.LastSyncAt, &integration.NextSyncAt, &integration.SyncInterval,
			&metadataJSON, &integration.CreatedAt, &integration.UpdatedAt,
		)

		json.Unmarshal([]byte(configJSON), &integration.Configuration)
		json.Unmarshal([]byte(credentialsJSON), &integration.Credentials)
		json.Unmarshal([]byte(metadataJSON), &integration.Metadata)

		integrations = append(integrations, &integration)
	}

	return integrations, count, nil
}

func (r *integrationRepository) GetActiveIntegrations(tenantID string) ([]*models.ThirdPartyIntegration, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	rows, err := db.Raw(`
		SELECT id, tenant_id, name, provider, configuration, credentials, active, sync_enabled, last_sync_at, next_sync_at, sync_interval, metadata, created_at, updated_at
		FROM third_party_integrations 
		WHERE tenant_id = ? AND active = true
		ORDER BY created_at DESC
	`, tenantID).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []*models.ThirdPartyIntegration
	for rows.Next() {
		var integration models.ThirdPartyIntegration
		var configJSON, credentialsJSON, metadataJSON string
		rows.Scan(
			&integration.ID, &integration.TenantID, &integration.Name, &integration.Provider,
			&configJSON, &credentialsJSON, &integration.Active, &integration.SyncEnabled,
			&integration.LastSyncAt, &integration.NextSyncAt, &integration.SyncInterval,
			&metadataJSON, &integration.CreatedAt, &integration.UpdatedAt,
		)

		json.Unmarshal([]byte(configJSON), &integration.Configuration)
		json.Unmarshal([]byte(credentialsJSON), &integration.Credentials)
		json.Unmarshal([]byte(metadataJSON), &integration.Metadata)

		integrations = append(integrations, &integration)
	}

	return integrations, nil
}

func (r *integrationRepository) CreateDataSync(tenantID string, dataSync *models.DataSync) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	metadataJSON, _ := json.Marshal(dataSync.Metadata)

	result := db.Exec(`
		INSERT INTO data_syncs (id, tenant_id, integration_id, sync_type, entity_type, status, progress, records_total, records_processed, records_failed, started_at, completed_at, error_message, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, dataSync.ID, dataSync.TenantID, dataSync.IntegrationID, dataSync.SyncType, dataSync.EntityType,
		dataSync.Status, dataSync.Progress, dataSync.RecordsTotal, dataSync.RecordsProcessed,
		dataSync.RecordsFailed, dataSync.StartedAt, dataSync.CompletedAt, dataSync.ErrorMessage,
		string(metadataJSON), dataSync.CreatedAt, dataSync.UpdatedAt)

	return result.Error
}

func (r *integrationRepository) GetDataSync(tenantID, syncID string) (*models.DataSync, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var dataSync models.DataSync
	var metadataJSON string
	err = db.Raw(`
		SELECT id, tenant_id, integration_id, sync_type, entity_type, status, progress, records_total, records_processed, records_failed, started_at, completed_at, error_message, metadata, created_at, updated_at
		FROM data_syncs 
		WHERE id = ? AND tenant_id = ?
	`, syncID, tenantID).Row().Scan(
		&dataSync.ID, &dataSync.TenantID, &dataSync.IntegrationID, &dataSync.SyncType,
		&dataSync.EntityType, &dataSync.Status, &dataSync.Progress, &dataSync.RecordsTotal,
		&dataSync.RecordsProcessed, &dataSync.RecordsFailed, &dataSync.StartedAt,
		&dataSync.CompletedAt, &dataSync.ErrorMessage, &metadataJSON,
		&dataSync.CreatedAt, &dataSync.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(metadataJSON), &dataSync.Metadata)

	return &dataSync, nil
}

func (r *integrationRepository) UpdateDataSync(tenantID string, dataSync *models.DataSync) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	metadataJSON, _ := json.Marshal(dataSync.Metadata)

	result := db.Exec(`
		UPDATE data_syncs 
		SET status = ?, progress = ?, records_total = ?, records_processed = ?, records_failed = ?, started_at = ?, completed_at = ?, error_message = ?, metadata = ?, updated_at = ?
		WHERE id = ? AND tenant_id = ?
	`, dataSync.Status, dataSync.Progress, dataSync.RecordsTotal, dataSync.RecordsProcessed,
		dataSync.RecordsFailed, dataSync.StartedAt, dataSync.CompletedAt, dataSync.ErrorMessage,
		string(metadataJSON), dataSync.UpdatedAt, dataSync.ID, tenantID)

	return result.Error
}

func (r *integrationRepository) ListDataSyncs(tenantID, integrationID string, limit, offset int) ([]*models.DataSync, int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	var count int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM data_syncs 
		WHERE tenant_id = ? AND integration_id = ?
	`, tenantID, integrationID).Scan(&count)

	// Get syncs
	rows, err := db.Raw(`
		SELECT id, tenant_id, integration_id, sync_type, entity_type, status, progress, records_total, records_processed, records_failed, started_at, completed_at, error_message, metadata, created_at, updated_at
		FROM data_syncs 
		WHERE tenant_id = ? AND integration_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, tenantID, integrationID, limit, offset).Rows()

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var syncs []*models.DataSync
	for rows.Next() {
		var dataSync models.DataSync
		var metadataJSON string
		rows.Scan(
			&dataSync.ID, &dataSync.TenantID, &dataSync.IntegrationID, &dataSync.SyncType,
			&dataSync.EntityType, &dataSync.Status, &dataSync.Progress, &dataSync.RecordsTotal,
			&dataSync.RecordsProcessed, &dataSync.RecordsFailed, &dataSync.StartedAt,
			&dataSync.CompletedAt, &dataSync.ErrorMessage, &metadataJSON,
			&dataSync.CreatedAt, &dataSync.UpdatedAt,
		)

		json.Unmarshal([]byte(metadataJSON), &dataSync.Metadata)
		syncs = append(syncs, &dataSync)
	}

	return syncs, count, nil
}

func (r *integrationRepository) GetPendingSyncs(tenantID string) ([]*models.DataSync, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	rows, err := db.Raw(`
		SELECT id, tenant_id, integration_id, sync_type, entity_type, status, progress, records_total, records_processed, records_failed, started_at, completed_at, error_message, metadata, created_at, updated_at
		FROM data_syncs 
		WHERE tenant_id = ? AND status = 'pending'
		ORDER BY created_at ASC
	`, tenantID).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var syncs []*models.DataSync
	for rows.Next() {
		var dataSync models.DataSync
		var metadataJSON string
		rows.Scan(
			&dataSync.ID, &dataSync.TenantID, &dataSync.IntegrationID, &dataSync.SyncType,
			&dataSync.EntityType, &dataSync.Status, &dataSync.Progress, &dataSync.RecordsTotal,
			&dataSync.RecordsProcessed, &dataSync.RecordsFailed, &dataSync.StartedAt,
			&dataSync.CompletedAt, &dataSync.ErrorMessage, &metadataJSON,
			&dataSync.CreatedAt, &dataSync.UpdatedAt,
		)

		json.Unmarshal([]byte(metadataJSON), &dataSync.Metadata)
		syncs = append(syncs, &dataSync)
	}

	return syncs, nil
}

func (r *integrationRepository) GetRunningSyncs(tenantID string) ([]*models.DataSync, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	rows, err := db.Raw(`
		SELECT id, tenant_id, integration_id, sync_type, entity_type, status, progress, records_total, records_processed, records_failed, started_at, completed_at, error_message, metadata, created_at, updated_at
		FROM data_syncs 
		WHERE tenant_id = ? AND status = 'running'
		ORDER BY started_at ASC
	`, tenantID).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var syncs []*models.DataSync
	for rows.Next() {
		var dataSync models.DataSync
		var metadataJSON string
		rows.Scan(
			&dataSync.ID, &dataSync.TenantID, &dataSync.IntegrationID, &dataSync.SyncType,
			&dataSync.EntityType, &dataSync.Status, &dataSync.Progress, &dataSync.RecordsTotal,
			&dataSync.RecordsProcessed, &dataSync.RecordsFailed, &dataSync.StartedAt,
			&dataSync.CompletedAt, &dataSync.ErrorMessage, &metadataJSON,
			&dataSync.CreatedAt, &dataSync.UpdatedAt,
		)

		json.Unmarshal([]byte(metadataJSON), &dataSync.Metadata)
		syncs = append(syncs, &dataSync)
	}

	return syncs, nil
}

func (r *integrationRepository) GetIntegrationStats(tenantID string) (*models.IntegrationStats, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	stats := &models.IntegrationStats{}

	// Integration statistics
	db.Raw(`
		SELECT 
			COUNT(*) as total_integrations,
			COUNT(CASE WHEN active = true THEN 1 END) as active_integrations
		FROM third_party_integrations 
		WHERE tenant_id = ?
	`, tenantID).Scan(stats)

	// Sync statistics
	db.Raw(`
		SELECT 
			COUNT(*) as total_syncs,
			COUNT(CASE WHEN status = 'running' THEN 1 END) as running_syncs
		FROM data_syncs 
		WHERE tenant_id = ?
	`, tenantID).Scan(stats)

	return stats, nil
}