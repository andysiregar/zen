package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"integration-service/internal/config"
	"integration-service/internal/models"
	"integration-service/internal/repositories"
)

type WebhookService interface {
	// Webhook management
	CreateWebhook(tenantID string, request *models.CreateWebhookRequest) (*models.Webhook, error)
	GetWebhook(tenantID, webhookID string) (*models.Webhook, error)
	UpdateWebhook(tenantID, webhookID string, request *models.UpdateWebhookRequest) (*models.Webhook, error)
	DeleteWebhook(tenantID, webhookID string) error
	ListWebhooks(tenantID string, page, limit int) (*models.WebhookListResponse, error)
	
	// Webhook delivery
	DeliverWebhook(tenantID string, event *models.WebhookEvent) error
	TestWebhook(tenantID, webhookID string, request *models.WebhookTestRequest) error
	RetryFailedWebhooks(tenantID string) error
	
	// Webhook deliveries
	ListWebhookDeliveries(tenantID, webhookID string, page, limit int) (*models.WebhookDeliveryListResponse, error)
	GetWebhookStats(tenantID string) (*models.IntegrationStats, error)
}

type webhookService struct {
	webhookRepo repositories.WebhookRepository
	config      *config.Config
	httpClient  *http.Client
}

func NewWebhookService(webhookRepo repositories.WebhookRepository, cfg *config.Config) WebhookService {
	return &webhookService{
		webhookRepo: webhookRepo,
		config:      cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Integration.WebhookTimeoutSeconds) * time.Second,
		},
	}
}

func (s *webhookService) CreateWebhook(tenantID string, request *models.CreateWebhookRequest) (*models.Webhook, error) {
	now := time.Now()
	webhook := &models.Webhook{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Name:      request.Name,
		URL:       request.URL,
		Secret:    request.Secret,
		Events:    request.Events,
		Active:    request.Active,
		Headers:   request.Headers,
		Metadata:  request.Metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.webhookRepo.CreateWebhook(tenantID, webhook); err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	return webhook, nil
}

func (s *webhookService) GetWebhook(tenantID, webhookID string) (*models.Webhook, error) {
	return s.webhookRepo.GetWebhook(tenantID, webhookID)
}

func (s *webhookService) UpdateWebhook(tenantID, webhookID string, request *models.UpdateWebhookRequest) (*models.Webhook, error) {
	webhook, err := s.webhookRepo.GetWebhook(tenantID, webhookID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if request.Name != "" {
		webhook.Name = request.Name
	}
	if request.URL != "" {
		webhook.URL = request.URL
	}
	if request.Secret != "" {
		webhook.Secret = request.Secret
	}
	if len(request.Events) > 0 {
		webhook.Events = request.Events
	}
	if request.Active != nil {
		webhook.Active = *request.Active
	}
	if request.Headers != nil {
		webhook.Headers = request.Headers
	}
	if request.Metadata != nil {
		webhook.Metadata = request.Metadata
	}
	webhook.UpdatedAt = time.Now()

	if err := s.webhookRepo.UpdateWebhook(tenantID, webhook); err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}

	return webhook, nil
}

func (s *webhookService) DeleteWebhook(tenantID, webhookID string) error {
	return s.webhookRepo.DeleteWebhook(tenantID, webhookID)
}

func (s *webhookService) ListWebhooks(tenantID string, page, limit int) (*models.WebhookListResponse, error) {
	offset := (page - 1) * limit
	webhooks, total, err := s.webhookRepo.ListWebhooks(tenantID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &models.WebhookListResponse{
		Webhooks:   webhooks,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (s *webhookService) DeliverWebhook(tenantID string, event *models.WebhookEvent) error {
	// Get active webhooks for this tenant
	webhooks, err := s.webhookRepo.GetActiveWebhooks(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get active webhooks: %w", err)
	}

	// Filter webhooks that are subscribed to this event
	for _, webhook := range webhooks {
		if s.webhookSubscribedToEvent(webhook, event.Event) {
			if err := s.deliverToWebhook(webhook, event); err != nil {
				// Log error but continue with other webhooks
				continue
			}
		}
	}

	return nil
}

func (s *webhookService) TestWebhook(tenantID, webhookID string, request *models.WebhookTestRequest) error {
	webhook, err := s.webhookRepo.GetWebhook(tenantID, webhookID)
	if err != nil {
		return err
	}

	// Create test event
	event := &models.WebhookEvent{
		ID:        uuid.New().String(),
		Event:     request.Event,
		TenantID:  tenantID,
		Data:      request.Payload,
		Timestamp: time.Now(),
	}

	return s.deliverToWebhook(webhook, event)
}

func (s *webhookService) RetryFailedWebhooks(tenantID string) error {
	pendingRetries, err := s.webhookRepo.GetPendingRetries(tenantID)
	if err != nil {
		return err
	}

	for _, delivery := range pendingRetries {
		webhook, err := s.webhookRepo.GetWebhook(tenantID, delivery.WebhookID)
		if err != nil {
			continue
		}

		// Recreate event from delivery
		var eventData map[string]interface{}
		json.Unmarshal([]byte(delivery.Payload), &eventData)
		
		event := &models.WebhookEvent{
			ID:        uuid.New().String(),
			Event:     delivery.Event,
			TenantID:  tenantID,
			Data:      eventData,
			Timestamp: time.Now(),
		}

		s.deliverToWebhook(webhook, event)
	}

	return nil
}

func (s *webhookService) ListWebhookDeliveries(tenantID, webhookID string, page, limit int) (*models.WebhookDeliveryListResponse, error) {
	offset := (page - 1) * limit
	deliveries, total, err := s.webhookRepo.ListWebhookDeliveries(tenantID, webhookID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &models.WebhookDeliveryListResponse{
		Deliveries: deliveries,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (s *webhookService) GetWebhookStats(tenantID string) (*models.IntegrationStats, error) {
	return s.webhookRepo.GetWebhookStats(tenantID)
}

// Helper methods

func (s *webhookService) webhookSubscribedToEvent(webhook *models.Webhook, event string) bool {
	for _, subscribedEvent := range webhook.Events {
		if subscribedEvent == event || subscribedEvent == "*" {
			return true
		}
	}
	return false
}

func (s *webhookService) deliverToWebhook(webhook *models.Webhook, event *models.WebhookEvent) error {
	// Create delivery record
	delivery := &models.WebhookDelivery{
		ID:           uuid.New().String(),
		TenantID:     webhook.TenantID,
		WebhookID:    webhook.ID,
		Event:        event.Event,
		URL:          webhook.URL,
		Method:       "POST",
		Headers:      webhook.Headers,
		AttemptCount: 1,
		MaxAttempts:  s.config.Integration.MaxRetryAttempts,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Prepare payload
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}
	delivery.Payload = string(payloadBytes)

	// Create HTTP request
	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		delivery.StatusCode = 0
		delivery.Response = fmt.Sprintf("Failed to create request: %v", err)
		delivery.FailedAt = &delivery.CreatedAt
		s.webhookRepo.CreateWebhookDelivery(webhook.TenantID, delivery)
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ZenPlatform-Webhooks/1.0")
	
	// Add custom headers
	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}

	// Add signature if secret is provided
	if webhook.Secret != "" {
		signature := s.generateSignature(payloadBytes, webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Make request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		delivery.StatusCode = 0
		delivery.Response = fmt.Sprintf("Request failed: %v", err)
		delivery.NextRetryAt = s.calculateNextRetry(delivery.AttemptCount)
		s.webhookRepo.CreateWebhookDelivery(webhook.TenantID, delivery)
		return err
	}
	defer resp.Body.Close()

	// Read response
	responseBody := make([]byte, 1024) // Limit response size
	resp.Body.Read(responseBody)

	delivery.StatusCode = resp.StatusCode
	delivery.Response = string(responseBody)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success
		now := time.Now()
		delivery.DeliveredAt = &now
	} else {
		// Failed
		if delivery.AttemptCount < delivery.MaxAttempts {
			delivery.NextRetryAt = s.calculateNextRetry(delivery.AttemptCount)
		} else {
			now := time.Now()
			delivery.FailedAt = &now
		}
	}

	return s.webhookRepo.CreateWebhookDelivery(webhook.TenantID, delivery)
}

func (s *webhookService) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

func (s *webhookService) calculateNextRetry(attemptCount int) *time.Time {
	backoffSeconds := s.config.Integration.RetryBackoffSeconds * attemptCount
	nextRetry := time.Now().Add(time.Duration(backoffSeconds) * time.Second)
	return &nextRetry
}