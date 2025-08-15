package models

import (
	"errors"
	"time"
)

var (
	ErrIntegrationNotFound = errors.New("integration not found")
	ErrWebhookNotFound     = errors.New("webhook not found")
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID          string             `json:"id"`
	TenantID    string             `json:"tenant_id"`
	Name        string             `json:"name"`
	URL         string             `json:"url"`
	Secret      string             `json:"secret,omitempty"`
	Events      []string           `json:"events"`
	Active      bool               `json:"active"`
	Headers     map[string]string  `json:"headers"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	WebhookID     string    `json:"webhook_id"`
	Event         string    `json:"event"`
	Payload       string    `json:"payload"`
	URL           string    `json:"url"`
	Method        string    `json:"method"`
	Headers       map[string]string `json:"headers"`
	StatusCode    int       `json:"status_code"`
	Response      string    `json:"response"`
	AttemptCount  int       `json:"attempt_count"`
	MaxAttempts   int       `json:"max_attempts"`
	NextRetryAt   *time.Time `json:"next_retry_at,omitempty"`
	DeliveredAt   *time.Time `json:"delivered_at,omitempty"`
	FailedAt      *time.Time `json:"failed_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Integration represents a general integration (alias for ThirdPartyIntegration for compatibility)
type Integration = ThirdPartyIntegration

// ThirdPartyIntegration represents a third-party API integration
type ThirdPartyIntegration struct {
	ID              string             `json:"id"`
	TenantID        string             `json:"tenant_id"`
	Name            string             `json:"name"`
	Provider        string             `json:"provider"` // slack, teams, discord, email, etc.
	Configuration   map[string]interface{} `json:"configuration"`
	Credentials     map[string]string  `json:"credentials,omitempty"`
	Active          bool               `json:"active"`
	SyncEnabled     bool               `json:"sync_enabled"`
	LastSyncAt      *time.Time         `json:"last_sync_at,omitempty"`
	NextSyncAt      *time.Time         `json:"next_sync_at,omitempty"`
	SyncInterval    int                `json:"sync_interval"` // minutes
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// DataSync represents a data synchronization record
type DataSync struct {
	ID              string             `json:"id"`
	TenantID        string             `json:"tenant_id"`
	IntegrationID   string             `json:"integration_id"`
	SyncType        string             `json:"sync_type"` // import, export, bidirectional
	EntityType      string             `json:"entity_type"` // tickets, users, projects, etc.
	Status          string             `json:"status"` // pending, running, completed, failed
	Progress        int                `json:"progress"` // 0-100
	RecordsTotal    int                `json:"records_total"`
	RecordsProcessed int               `json:"records_processed"`
	RecordsFailed   int                `json:"records_failed"`
	StartedAt       *time.Time         `json:"started_at,omitempty"`
	CompletedAt     *time.Time         `json:"completed_at,omitempty"`
	ErrorMessage    string             `json:"error_message,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// API Request/Response Models

type CreateWebhookRequest struct {
	Name     string             `json:"name" binding:"required"`
	URL      string             `json:"url" binding:"required,url"`
	Secret   string             `json:"secret,omitempty"`
	Events   []string           `json:"events" binding:"required,min=1"`
	Active   bool               `json:"active"`
	Headers  map[string]string  `json:"headers"`
	Metadata map[string]interface{} `json:"metadata"`
}

type UpdateWebhookRequest struct {
	Name     string             `json:"name"`
	URL      string             `json:"url" binding:"omitempty,url"`
	Secret   string             `json:"secret"`
	Events   []string           `json:"events"`
	Active   *bool              `json:"active"`
	Headers  map[string]string  `json:"headers"`
	Metadata map[string]interface{} `json:"metadata"`
}

type WebhookTestRequest struct {
	Event   string                 `json:"event" binding:"required"`
	Payload map[string]interface{} `json:"payload"`
}

type CreateIntegrationRequest struct {
	Name            string             `json:"name" binding:"required"`
	Provider        string             `json:"provider" binding:"required"`
	Configuration   map[string]interface{} `json:"configuration"`
	Credentials     map[string]string  `json:"credentials"`
	Active          bool               `json:"active"`
	SyncEnabled     bool               `json:"sync_enabled"`
	SyncInterval    int                `json:"sync_interval"`
	Metadata        map[string]interface{} `json:"metadata"`
}

type UpdateIntegrationRequest struct {
	Name            string             `json:"name"`
	Configuration   map[string]interface{} `json:"configuration"`
	Credentials     map[string]string  `json:"credentials"`
	Active          *bool              `json:"active"`
	SyncEnabled     *bool              `json:"sync_enabled"`
	SyncInterval    int                `json:"sync_interval"`
	Metadata        map[string]interface{} `json:"metadata"`
}

type TriggerSyncRequest struct {
	SyncType   string `json:"sync_type" binding:"required"`
	EntityType string `json:"entity_type" binding:"required"`
}

type WebhookListResponse struct {
	Webhooks   []*Webhook `json:"webhooks"`
	TotalCount int64      `json:"total_count"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
}

type WebhookDeliveryListResponse struct {
	Deliveries []*WebhookDelivery `json:"deliveries"`
	TotalCount int64              `json:"total_count"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
}

type IntegrationListResponse struct {
	Integrations []*ThirdPartyIntegration `json:"integrations"`
	TotalCount   int64                    `json:"total_count"`
	Page         int                      `json:"page"`
	Limit        int                      `json:"limit"`
}

type DataSyncListResponse struct {
	Syncs      []*DataSync `json:"syncs"`
	TotalCount int64       `json:"total_count"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
}

type IntegrationStats struct {
	TotalWebhooks       int64 `json:"total_webhooks"`
	ActiveWebhooks      int64 `json:"active_webhooks"`
	TotalDeliveries     int64 `json:"total_deliveries"`
	SuccessfulDeliveries int64 `json:"successful_deliveries"`
	FailedDeliveries    int64 `json:"failed_deliveries"`
	PendingRetries      int64 `json:"pending_retries"`
	TotalIntegrations   int64 `json:"total_integrations"`
	ActiveIntegrations  int64 `json:"active_integrations"`
	TotalSyncs          int64 `json:"total_syncs"`
	RunningSyncs        int64 `json:"running_syncs"`
}

// Event payload for webhooks
type WebhookEvent struct {
	Event     string                 `json:"event"`
	TenantID  string                 `json:"tenant_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	ID        string                 `json:"id"`
}