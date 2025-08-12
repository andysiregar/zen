package models

import (
	"time"
	"gorm.io/gorm"
)

type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusInactive  TenantStatus = "inactive"
)

type Tenant struct {
	ID          string       `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string       `json:"name" gorm:"not null;size:255"`
	Domain      string       `json:"domain" gorm:"uniqueIndex;not null;size:255"`
	Subdomain   string       `json:"subdomain" gorm:"uniqueIndex;not null;size:63"`
	Status      TenantStatus `json:"status" gorm:"type:varchar(20);default:'active'"`
	
	// Database configuration
	DatabaseName string `json:"database_name" gorm:"not null;size:63"`
	DatabaseHost string `json:"database_host" gorm:"size:255"`
	
	// Contact information
	ContactEmail string `json:"contact_email" gorm:"size:255"`
	ContactPhone string `json:"contact_phone" gorm:"size:50"`
	
	// Billing information
	PlanType        string    `json:"plan_type" gorm:"size:50;default:'basic'"`
	SubscriptionID  string    `json:"subscription_id" gorm:"size:255"`
	BillingEmail    string    `json:"billing_email" gorm:"size:255"`
	
	// Settings and limits
	MaxUsers      int  `json:"max_users" gorm:"default:10"`
	MaxStorage    int  `json:"max_storage" gorm:"default:1048576"` // 1GB in KB
	MaxBandwidth  int  `json:"max_bandwidth" gorm:"default:10240"` // 10GB in MB
	FeaturesJSON  string `json:"features_json" gorm:"type:text"`
	
	// Timestamps
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

type TenantCreateRequest struct {
	Name         string `json:"name" binding:"required,min=2,max=255"`
	Domain       string `json:"domain" binding:"required,min=3,max=255"`
	Subdomain    string `json:"subdomain" binding:"required,min=3,max=63,alphanum"`
	ContactEmail string `json:"contact_email" binding:"required,email"`
	ContactPhone string `json:"contact_phone,omitempty"`
	PlanType     string `json:"plan_type,omitempty"`
}

type TenantUpdateRequest struct {
	Name         *string       `json:"name,omitempty" binding:"omitempty,min=2,max=255"`
	Domain       *string       `json:"domain,omitempty" binding:"omitempty,min=3,max=255"`
	Status       *TenantStatus `json:"status,omitempty"`
	ContactEmail *string       `json:"contact_email,omitempty" binding:"omitempty,email"`
	ContactPhone *string       `json:"contact_phone,omitempty"`
	PlanType     *string       `json:"plan_type,omitempty"`
	MaxUsers     *int          `json:"max_users,omitempty" binding:"omitempty,min=1"`
	MaxStorage   *int          `json:"max_storage,omitempty" binding:"omitempty,min=1"`
}

type TenantResponse struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Domain       string       `json:"domain"`
	Subdomain    string       `json:"subdomain"`
	Status       TenantStatus `json:"status"`
	ContactEmail string       `json:"contact_email"`
	ContactPhone string       `json:"contact_phone"`
	PlanType     string       `json:"plan_type"`
	MaxUsers     int          `json:"max_users"`
	MaxStorage   int          `json:"max_storage"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// TableName overrides the table name used by Tenant to `tenants`
func (Tenant) TableName() string {
	return "tenants"
}

// ToResponse converts a Tenant model to TenantResponse
func (t *Tenant) ToResponse() TenantResponse {
	return TenantResponse{
		ID:           t.ID,
		Name:         t.Name,
		Domain:       t.Domain,
		Subdomain:    t.Subdomain,
		Status:       t.Status,
		ContactEmail: t.ContactEmail,
		ContactPhone: t.ContactPhone,
		PlanType:     t.PlanType,
		MaxUsers:     t.MaxUsers,
		MaxStorage:   t.MaxStorage,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}