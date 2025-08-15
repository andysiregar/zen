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

type TenantRegion string

const (
	TenantRegionAsia   TenantRegion = "asia"
	TenantRegionUS     TenantRegion = "us"
	TenantRegionEurope TenantRegion = "europe"
)

type Tenant struct {
	ID             string       `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrganizationID string       `json:"organization_id" gorm:"type:uuid;not null;index"`
	Name           string       `json:"name" gorm:"not null;size:255"`
	Slug           string       `json:"slug" gorm:"not null;size:100"` // URL-friendly workspace identifier
	Description    string       `json:"description" gorm:"type:text"`
	Status         TenantStatus `json:"status" gorm:"type:varchar(20);default:'active'"`
	Region         TenantRegion `json:"region" gorm:"type:varchar(20);not null;default:'us'"` // Geographic deployment region
	
	// Domain and subdomain (required by database schema)
	Domain        string `json:"domain" gorm:"not null;size:255"`
	Subdomain     string `json:"subdomain" gorm:"not null;size:63"`
	DatabaseName  string `json:"database_name" gorm:"not null;size:63"`
	
	// Infrastructure Details - Each tenant gets dedicated VMs
	DbHost              string `json:"db_host" gorm:"size:255;default:'localhost'"` // Dedicated DB VM IP
	DbPort              int    `json:"db_port" gorm:"default:5432"`
	DbName              string `json:"db_name" gorm:"not null;size:100"`            // Database name on dedicated VM
	DbUser              string `json:"db_user" gorm:"not null;size:100"`
	DbPasswordEncrypted string `json:"-" gorm:"not null;type:text"`                 // Encrypted connection password
	DbSslMode           string `json:"db_ssl_mode" gorm:"size:20;default:'disable'"`
	
	// Proxmox Infrastructure Details
	ProxmoxCluster      string `json:"proxmox_cluster" gorm:"column:proxmox_cluster;size:100"`      // Regional Proxmox cluster ID
	WebServerClusterIPs string `json:"web_cluster_ips" gorm:"column:web_cluster_ips;type:text"`     // JSON array of web server VM IPs
	DatabaseVMIP        string `json:"database_vm_ip" gorm:"column:database_vm_ip;size:45"`        // Dedicated database VM IP
	
	// Configuration
	Settings JSONB `json:"settings" gorm:"type:jsonb;default:'{}'"`   // Tenant-specific settings
	Features JSONB `json:"features" gorm:"type:jsonb;default:'[]'"`   // Enabled features
	
	// Status
	ProvisionedAt *time.Time `json:"provisioned_at"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type TenantCreateRequest struct {
	OrganizationID string       `json:"organization_id" binding:"required,uuid"`
	Name           string       `json:"name" binding:"required,min=2,max=255"`
	Slug           string       `json:"slug" binding:"required,min=2,max=100,alphanum"`
	Description    string       `json:"description,omitempty"`
	Region         TenantRegion `json:"region" binding:"required,oneof=asia us europe"`
}

type TenantUpdateRequest struct {
	Name        *string       `json:"name,omitempty" binding:"omitempty,min=2,max=255"`
	Description *string       `json:"description,omitempty"`
	Status      *TenantStatus `json:"status,omitempty"`
	Settings    *JSONB        `json:"settings,omitempty"`
	Features    *JSONB        `json:"features,omitempty"`
}

type TenantResponse struct {
	ID             string       `json:"id"`
	OrganizationID string       `json:"organization_id"`
	Name           string       `json:"name"`
	Slug           string       `json:"slug"`
	Description    string       `json:"description"`
	Status         TenantStatus `json:"status"`
	Region         TenantRegion `json:"region"`
	Settings       JSONB        `json:"settings"`
	Features       JSONB        `json:"features"`
	ProvisionedAt  *time.Time   `json:"provisioned_at"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
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
		Region:         t.Region,
		Settings:       t.Settings,
		Features:       t.Features,
		ProvisionedAt:  t.ProvisionedAt,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}