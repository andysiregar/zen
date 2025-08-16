package repositories

import (
	"encoding/json"

	"github.com/zen/shared/pkg/database"
	"integration-service/internal/models"
)

type WebhookRepository interface {
	// Webhook CRUD operations
	CreateWebhook(tenantID string, webhook *models.Webhook) error
	GetWebhook(tenantID, webhookID string) (*models.Webhook, error)
	UpdateWebhook(tenantID string, webhook *models.Webhook) error
	DeleteWebhook(tenantID, webhookID string) error
	ListWebhooks(tenantID string, limit, offset int) ([]*models.Webhook, int64, error)
	GetActiveWebhooks(tenantID string) ([]*models.Webhook, error)
	
	// Webhook delivery operations
	CreateWebhookDelivery(tenantID string, delivery *models.WebhookDelivery) error
	GetWebhookDelivery(tenantID, deliveryID string) (*models.WebhookDelivery, error)
	UpdateWebhookDelivery(tenantID string, delivery *models.WebhookDelivery) error
	ListWebhookDeliveries(tenantID, webhookID string, limit, offset int) ([]*models.WebhookDelivery, int64, error)
	GetPendingRetries(tenantID string) ([]*models.WebhookDelivery, error)
	
	// Statistics
	GetWebhookStats(tenantID string) (*models.IntegrationStats, error)
}

type webhookRepository struct {
	tenantDBManager *database.TenantDatabaseManager
}

func NewWebhookRepository(tenantDBManager *database.TenantDatabaseManager) WebhookRepository {
	return &webhookRepository{
		tenantDBManager: tenantDBManager,
	}
}

func (r *webhookRepository) CreateWebhook(tenantID string, webhook *models.Webhook) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	eventsJSON, _ := json.Marshal(webhook.Events)
	headersJSON, _ := json.Marshal(webhook.Headers)
	metadataJSON, _ := json.Marshal(webhook.Metadata)

	result := db.Exec(`
		INSERT INTO webhooks (id, tenant_id, name, url, secret, events, active, headers, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, webhook.ID, webhook.TenantID, webhook.Name, webhook.URL, webhook.Secret,
		string(eventsJSON), webhook.Active, string(headersJSON), string(metadataJSON),
		webhook.CreatedAt, webhook.UpdatedAt)

	return result.Error
}

func (r *webhookRepository) GetWebhook(tenantID, webhookID string) (*models.Webhook, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var webhook models.Webhook
	var eventsJSON, headersJSON, metadataJSON string
	err = db.Raw(`
		SELECT id, tenant_id, name, url, secret, events, active, headers, metadata, created_at, updated_at
		FROM webhooks 
		WHERE id = ? AND tenant_id = ?
	`, webhookID, tenantID).Row().Scan(
		&webhook.ID, &webhook.TenantID, &webhook.Name, &webhook.URL, &webhook.Secret,
		&eventsJSON, &webhook.Active, &headersJSON, &metadataJSON,
		&webhook.CreatedAt, &webhook.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(eventsJSON), &webhook.Events)
	json.Unmarshal([]byte(headersJSON), &webhook.Headers)
	json.Unmarshal([]byte(metadataJSON), &webhook.Metadata)

	return &webhook, nil
}

func (r *webhookRepository) UpdateWebhook(tenantID string, webhook *models.Webhook) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	eventsJSON, _ := json.Marshal(webhook.Events)
	headersJSON, _ := json.Marshal(webhook.Headers)
	metadataJSON, _ := json.Marshal(webhook.Metadata)

	result := db.Exec(`
		UPDATE webhooks 
		SET name = ?, url = ?, secret = ?, events = ?, active = ?, headers = ?, metadata = ?, updated_at = ?
		WHERE id = ? AND tenant_id = ?
	`, webhook.Name, webhook.URL, webhook.Secret, string(eventsJSON), webhook.Active,
		string(headersJSON), string(metadataJSON), webhook.UpdatedAt, webhook.ID, tenantID)

	return result.Error
}

func (r *webhookRepository) DeleteWebhook(tenantID, webhookID string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		DELETE FROM webhooks 
		WHERE id = ? AND tenant_id = ?
	`, webhookID, tenantID)

	return result.Error
}

func (r *webhookRepository) ListWebhooks(tenantID string, limit, offset int) ([]*models.Webhook, int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	var count int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM webhooks 
		WHERE tenant_id = ?
	`, tenantID).Scan(&count)

	// Get webhooks
	rows, err := db.Raw(`
		SELECT id, tenant_id, name, url, secret, events, active, headers, metadata, created_at, updated_at
		FROM webhooks 
		WHERE tenant_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, tenantID, limit, offset).Rows()

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var webhooks []*models.Webhook
	for rows.Next() {
		var webhook models.Webhook
		var eventsJSON, headersJSON, metadataJSON string
		rows.Scan(
			&webhook.ID, &webhook.TenantID, &webhook.Name, &webhook.URL, &webhook.Secret,
			&eventsJSON, &webhook.Active, &headersJSON, &metadataJSON,
			&webhook.CreatedAt, &webhook.UpdatedAt,
		)

		json.Unmarshal([]byte(eventsJSON), &webhook.Events)
		json.Unmarshal([]byte(headersJSON), &webhook.Headers)
		json.Unmarshal([]byte(metadataJSON), &webhook.Metadata)

		webhooks = append(webhooks, &webhook)
	}

	return webhooks, count, nil
}

func (r *webhookRepository) GetActiveWebhooks(tenantID string) ([]*models.Webhook, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	rows, err := db.Raw(`
		SELECT id, tenant_id, name, url, secret, events, active, headers, metadata, created_at, updated_at
		FROM webhooks 
		WHERE tenant_id = ? AND active = true
		ORDER BY created_at DESC
	`, tenantID).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []*models.Webhook
	for rows.Next() {
		var webhook models.Webhook
		var eventsJSON, headersJSON, metadataJSON string
		rows.Scan(
			&webhook.ID, &webhook.TenantID, &webhook.Name, &webhook.URL, &webhook.Secret,
			&eventsJSON, &webhook.Active, &headersJSON, &metadataJSON,
			&webhook.CreatedAt, &webhook.UpdatedAt,
		)

		json.Unmarshal([]byte(eventsJSON), &webhook.Events)
		json.Unmarshal([]byte(headersJSON), &webhook.Headers)
		json.Unmarshal([]byte(metadataJSON), &webhook.Metadata)

		webhooks = append(webhooks, &webhook)
	}

	return webhooks, nil
}

func (r *webhookRepository) CreateWebhookDelivery(tenantID string, delivery *models.WebhookDelivery) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	headersJSON, _ := json.Marshal(delivery.Headers)

	result := db.Exec(`
		INSERT INTO webhook_deliveries (id, tenant_id, webhook_id, event, payload, url, method, headers, status_code, response, attempt_count, max_attempts, next_retry_at, delivered_at, failed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, delivery.ID, delivery.TenantID, delivery.WebhookID, delivery.Event, delivery.Payload,
		delivery.URL, delivery.Method, string(headersJSON), delivery.StatusCode, delivery.Response,
		delivery.AttemptCount, delivery.MaxAttempts, delivery.NextRetryAt, delivery.DeliveredAt,
		delivery.FailedAt, delivery.CreatedAt, delivery.UpdatedAt)

	return result.Error
}

func (r *webhookRepository) GetWebhookDelivery(tenantID, deliveryID string) (*models.WebhookDelivery, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var delivery models.WebhookDelivery
	var headersJSON string
	err = db.Raw(`
		SELECT id, tenant_id, webhook_id, event, payload, url, method, headers, status_code, response, attempt_count, max_attempts, next_retry_at, delivered_at, failed_at, created_at, updated_at
		FROM webhook_deliveries 
		WHERE id = ? AND tenant_id = ?
	`, deliveryID, tenantID).Row().Scan(
		&delivery.ID, &delivery.TenantID, &delivery.WebhookID, &delivery.Event, &delivery.Payload,
		&delivery.URL, &delivery.Method, &headersJSON, &delivery.StatusCode, &delivery.Response,
		&delivery.AttemptCount, &delivery.MaxAttempts, &delivery.NextRetryAt, &delivery.DeliveredAt,
		&delivery.FailedAt, &delivery.CreatedAt, &delivery.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(headersJSON), &delivery.Headers)

	return &delivery, nil
}

func (r *webhookRepository) UpdateWebhookDelivery(tenantID string, delivery *models.WebhookDelivery) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		UPDATE webhook_deliveries 
		SET status_code = ?, response = ?, attempt_count = ?, next_retry_at = ?, delivered_at = ?, failed_at = ?, updated_at = ?
		WHERE id = ? AND tenant_id = ?
	`, delivery.StatusCode, delivery.Response, delivery.AttemptCount, delivery.NextRetryAt,
		delivery.DeliveredAt, delivery.FailedAt, delivery.UpdatedAt, delivery.ID, tenantID)

	return result.Error
}

func (r *webhookRepository) ListWebhookDeliveries(tenantID, webhookID string, limit, offset int) ([]*models.WebhookDelivery, int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	var count int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM webhook_deliveries 
		WHERE tenant_id = ? AND webhook_id = ?
	`, tenantID, webhookID).Scan(&count)

	// Get deliveries
	rows, err := db.Raw(`
		SELECT id, tenant_id, webhook_id, event, payload, url, method, headers, status_code, response, attempt_count, max_attempts, next_retry_at, delivered_at, failed_at, created_at, updated_at
		FROM webhook_deliveries 
		WHERE tenant_id = ? AND webhook_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, tenantID, webhookID, limit, offset).Rows()

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var deliveries []*models.WebhookDelivery
	for rows.Next() {
		var delivery models.WebhookDelivery
		var headersJSON string
		rows.Scan(
			&delivery.ID, &delivery.TenantID, &delivery.WebhookID, &delivery.Event, &delivery.Payload,
			&delivery.URL, &delivery.Method, &headersJSON, &delivery.StatusCode, &delivery.Response,
			&delivery.AttemptCount, &delivery.MaxAttempts, &delivery.NextRetryAt, &delivery.DeliveredAt,
			&delivery.FailedAt, &delivery.CreatedAt, &delivery.UpdatedAt,
		)

		json.Unmarshal([]byte(headersJSON), &delivery.Headers)
		deliveries = append(deliveries, &delivery)
	}

	return deliveries, count, nil
}

func (r *webhookRepository) GetPendingRetries(tenantID string) ([]*models.WebhookDelivery, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	rows, err := db.Raw(`
		SELECT id, tenant_id, webhook_id, event, payload, url, method, headers, status_code, response, attempt_count, max_attempts, next_retry_at, delivered_at, failed_at, created_at, updated_at
		FROM webhook_deliveries 
		WHERE tenant_id = ? AND next_retry_at IS NOT NULL AND next_retry_at <= NOW() AND delivered_at IS NULL AND failed_at IS NULL
		ORDER BY next_retry_at ASC
	`, tenantID).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []*models.WebhookDelivery
	for rows.Next() {
		var delivery models.WebhookDelivery
		var headersJSON string
		rows.Scan(
			&delivery.ID, &delivery.TenantID, &delivery.WebhookID, &delivery.Event, &delivery.Payload,
			&delivery.URL, &delivery.Method, &headersJSON, &delivery.StatusCode, &delivery.Response,
			&delivery.AttemptCount, &delivery.MaxAttempts, &delivery.NextRetryAt, &delivery.DeliveredAt,
			&delivery.FailedAt, &delivery.CreatedAt, &delivery.UpdatedAt,
		)

		json.Unmarshal([]byte(headersJSON), &delivery.Headers)
		deliveries = append(deliveries, &delivery)
	}

	return deliveries, nil
}

func (r *webhookRepository) GetWebhookStats(tenantID string) (*models.IntegrationStats, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	stats := &models.IntegrationStats{}

	// Total and active webhooks
	db.Raw(`
		SELECT COUNT(*) as total_webhooks, COUNT(CASE WHEN active = true THEN 1 END) as active_webhooks
		FROM webhooks 
		WHERE tenant_id = ?
	`, tenantID).Scan(stats)

	// Delivery statistics
	db.Raw(`
		SELECT 
			COUNT(*) as total_deliveries,
			COUNT(CASE WHEN delivered_at IS NOT NULL THEN 1 END) as successful_deliveries,
			COUNT(CASE WHEN failed_at IS NOT NULL THEN 1 END) as failed_deliveries,
			COUNT(CASE WHEN next_retry_at IS NOT NULL AND next_retry_at > NOW() THEN 1 END) as pending_retries
		FROM webhook_deliveries 
		WHERE tenant_id = ?
	`, tenantID).Scan(stats)

	return stats, nil
}