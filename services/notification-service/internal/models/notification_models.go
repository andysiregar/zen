package models

import (
	"time"
)

type Notification struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	UserID      string    `json:"user_id"`
	Type        string    `json:"type"` // "email", "push", "in_app"
	Channel     string    `json:"channel"` // "email", "sms", "push", "in_app"
	Subject     string    `json:"subject"`
	Content     string    `json:"content"`
	Data        map[string]interface{} `json:"data"`
	Status      string    `json:"status"` // "pending", "sent", "failed", "read"
	Priority    string    `json:"priority"` // "low", "normal", "high", "urgent"
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type NotificationTemplate struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // "email", "push", "in_app"
	Subject     string    `json:"subject"`
	Content     string    `json:"content"`
	Variables   []string  `json:"variables"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type NotificationPreference struct {
	ID               string `json:"id"`
	TenantID         string `json:"tenant_id"`
	UserID           string `json:"user_id"`
	EmailEnabled     bool   `json:"email_enabled"`
	PushEnabled      bool   `json:"push_enabled"`
	InAppEnabled     bool   `json:"in_app_enabled"`
	SMSEnabled       bool   `json:"sms_enabled"`
	QuietHoursStart  string `json:"quiet_hours_start"` // HH:MM format
	QuietHoursEnd    string `json:"quiet_hours_end"`   // HH:MM format
	Timezone         string `json:"timezone"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type SendNotificationRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	Type        string                 `json:"type" binding:"required"`
	Channel     string                 `json:"channel" binding:"required"`
	Subject     string                 `json:"subject" binding:"required"`
	Content     string                 `json:"content" binding:"required"`
	Data        map[string]interface{} `json:"data"`
	Priority    string                 `json:"priority"`
	ScheduledAt *time.Time             `json:"scheduled_at"`
}

type SendBulkNotificationRequest struct {
	UserIDs     []string               `json:"user_ids" binding:"required"`
	Type        string                 `json:"type" binding:"required"`
	Channel     string                 `json:"channel" binding:"required"`
	Subject     string                 `json:"subject" binding:"required"`
	Content     string                 `json:"content" binding:"required"`
	Data        map[string]interface{} `json:"data"`
	Priority    string                 `json:"priority"`
	ScheduledAt *time.Time             `json:"scheduled_at"`
}

type NotificationResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	UserID      string    `json:"user_id"`
	Type        string    `json:"type"`
	Channel     string    `json:"channel"`
	Subject     string    `json:"subject"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}