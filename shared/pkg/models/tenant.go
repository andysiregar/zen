package models

import (
	"time"
	"gorm.io/gorm"
)

type TenantStatus string

const (
	TenantStatusActive       TenantStatus = "active"
	TenantStatusSuspended    TenantStatus = "suspended"
	TenantStatusInactive     TenantStatus = "inactive"
	TenantStatusProvisioning TenantStatus = "provisioning"
)

type Tenant struct {
	ID             string       `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrganizationID string       `json:"organization_id" gorm:"type:uuid;not null;index"`
	Name           string       `json:"name" gorm:"not null;size:255"`
	Slug           string       `json:"slug" gorm:"not null;size:100"` // URL-friendly workspace identifier
	Description    string       `json:"description" gorm:"type:text"`
	Status         TenantStatus `json:"status" gorm:"type:varchar(20);default:'active'"`
	
	// Database Connection Details (CRITICAL)
	DbHost              string `json:"db_host" gorm:"not null;size:255"`
	DbPort              int    `json:"db_port" gorm:"default:5432"`
	DbName              string `json:"db_name" gorm:"not null;size:100"`
	DbUser              string `json:"db_user" gorm:"not null;size:100"`
	DbPasswordEncrypted string `json:"-" gorm:"not null;type:text"` // Encrypted connection password
	DbSslMode           string `json:"db_ssl_mode" gorm:"size:20;default:'require'"`
	
	// Configuration
	Settings JSONB `json:"settings" gorm:"type:jsonb;default:'{}'"`   // Tenant-specific settings
	Features JSONB `json:"features" gorm:"type:jsonb;default:'{}'"`   // Enabled features
	
	// Status
	ProvisionedAt *time.Time `json:"provisioned_at"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type TenantCreateRequest struct {
	OrganizationID string `json:"organization_id" binding:"required,uuid"`
	Name           string `json:"name" binding:"required,min=2,max=255"`
	Slug           string `json:"slug" binding:"required,min=2,max=100,alphanum"`
	Description    string `json:"description,omitempty"`
}

type TenantUpdateRequest struct {
	Name        *string       `json:"name,omitempty" binding:"omitempty,min=2,max=255"`
	Description *string       `json:"description,omitempty"`
	Status      *TenantStatus `json:"status,omitempty"`
	Settings    *JSONB        `json:"settings,omitempty"`
	Features    *JSONB        `json:"features,omitempty"`
}

type TenantResponse struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	Name           string     `json:"name"`
	Slug           string     `json:"slug"`
	Description    string     `json:"description"`
	Status         TenantStatus `json:"status"`
	Settings       JSONB      `json:"settings"`
	Features       JSONB      `json:"features"`
	ProvisionedAt  *time.Time `json:"provisioned_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TableName overrides the table name used by Tenant to `tenants`
func (Tenant) TableName() string {
	return "tenants"
}

// ToResponse converts a Tenant model to TenantResponse
func (t *Tenant) ToResponse() TenantResponse {
	return TenantResponse{
		ID:             t.ID,
		OrganizationID: t.OrganizationID,
		Name:           t.Name,
		Slug:           t.Slug,
		Description:    t.Description,
		Status:         t.Status,
		Settings:       t.Settings,
		Features:       t.Features,
		ProvisionedAt:  t.ProvisionedAt,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}