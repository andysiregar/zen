package models

import (
	"time"
	"gorm.io/gorm"
)

// VMStatus represents the status of a virtual machine
type VMStatus string

const (
	VMStatusProvisioning VMStatus = "provisioning"
	VMStatusRunning      VMStatus = "running"
	VMStatusStopped      VMStatus = "stopped"
	VMStatusStarting     VMStatus = "starting"
	VMStatusStopping     VMStatus = "stopping"
	VMStatusError        VMStatus = "error"
)

// VMType represents the type of virtual machine
type VMType string

const (
	VMTypeDatabase  VMType = "database"
	VMTypeWebServer VMType = "web_server"
	VMTypeLoadBalancer VMType = "load_balancer"
)

// VirtualMachine represents a VM in the tenant infrastructure
type VirtualMachine struct {
	ID         string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TenantID   string    `json:"tenant_id" gorm:"type:uuid;not null;index"`
	VMIDProxy  string    `json:"vm_id_proxy" gorm:"not null;size:100"` // Proxmox VM ID
	Name       string    `json:"name" gorm:"not null;size:100"`
	Type       VMType    `json:"type" gorm:"type:varchar(20);not null"`
	Status     VMStatus  `json:"status" gorm:"type:varchar(20);default:'provisioning'"`
	
	// Hardware specifications
	CPUs    int `json:"cpus" gorm:"default:2"`
	Memory  int `json:"memory" gorm:"default:4096"` // MB
	Storage int `json:"storage" gorm:"default:50"`   // GB
	
	// Network configuration
	IPAddress    string `json:"ip_address" gorm:"size:45"`
	InternalIP   string `json:"internal_ip" gorm:"size:45"`
	PublicIP     string `json:"public_ip" gorm:"size:45"`
	
	// Proxmox details
	ProxmoxNode    string `json:"proxmox_node" gorm:"size:100"`
	ProxmoxCluster string `json:"proxmox_cluster" gorm:"size:100"`
	
	// Metadata
	Tags        JSONB `json:"tags" gorm:"type:jsonb;default:'{}'"`
	Config      JSONB `json:"config" gorm:"type:jsonb;default:'{}'"`
	
	// Timestamps
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	ProvisionedAt *time.Time     `json:"provisioned_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// TenantInfrastructure represents the complete infrastructure for a tenant
type TenantInfrastructure struct {
	ID               string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TenantID         string    `json:"tenant_id" gorm:"type:uuid;not null;unique;index"`
	Region           TenantRegion `json:"region" gorm:"type:varchar(20);not null"`
	
	// Infrastructure status
	Status           string `json:"status" gorm:"type:varchar(20);default:'provisioning'"`
	ProvisioningLogs JSONB  `json:"provisioning_logs" gorm:"type:jsonb;default:'[]'"`
	
	// Database VM details
	DatabaseVMID     string `json:"database_vm_id" gorm:"type:uuid"`
	DatabaseIP       string `json:"database_ip" gorm:"size:45"`
	DatabaseHost     string `json:"database_host" gorm:"size:255"`
	DatabasePort     int    `json:"database_port" gorm:"default:5432"`
	DatabaseName     string `json:"database_name" gorm:"not null;size:100"`
	DatabaseUser     string `json:"database_user" gorm:"not null;size:100"`
	DatabasePassword string `json:"-" gorm:"not null;type:text"` // Encrypted
	
	// Web cluster details
	LoadBalancerIP string `json:"load_balancer_ip" gorm:"size:45"`
	WebServerIPs   JSONB  `json:"web_server_ips" gorm:"type:jsonb;default:'[]'"`
	WebServerCount int    `json:"web_server_count" gorm:"default:3"`
	
	// Proxmox details
	ProxmoxCluster string `json:"proxmox_cluster" gorm:"size:100"`
	ProxmoxNode    string `json:"proxmox_node" gorm:"size:100"`
	
	// Timestamps
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	ProvisionedAt *time.Time     `json:"provisioned_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// InfrastructureLog represents a log entry for infrastructure operations
type InfrastructureLog struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TenantID     string    `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Operation    string    `json:"operation" gorm:"not null;size:50"` // provision, destroy, update, etc.
	Status       string    `json:"status" gorm:"not null;size:20"`    // success, failed, in_progress
	Message      string    `json:"message" gorm:"type:text"`
	Details      JSONB     `json:"details" gorm:"type:jsonb;default:'{}'"`
	
	CreatedAt time.Time `json:"created_at"`
}

// TableName overrides the table name for VirtualMachine
func (VirtualMachine) TableName() string {
	return "virtual_machines"
}

// TableName overrides the table name for TenantInfrastructure
func (TenantInfrastructure) TableName() string {
	return "tenant_infrastructure"
}

// TableName overrides the table name for InfrastructureLog
func (InfrastructureLog) TableName() string {
	return "infrastructure_logs"
}

// VMResponse represents the API response for VM information
type VMResponse struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	Type         VMType    `json:"type"`
	Status       VMStatus  `json:"status"`
	CPUs         int       `json:"cpus"`
	Memory       int       `json:"memory"`
	Storage      int       `json:"storage"`
	IPAddress    string    `json:"ip_address"`
	ProxmoxNode  string    `json:"proxmox_node"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ToResponse converts a VirtualMachine model to VMResponse
func (vm *VirtualMachine) ToResponse() VMResponse {
	return VMResponse{
		ID:          vm.ID,
		TenantID:    vm.TenantID,
		Name:        vm.Name,
		Type:        vm.Type,
		Status:      vm.Status,
		CPUs:        vm.CPUs,
		Memory:      vm.Memory,
		Storage:     vm.Storage,
		IPAddress:   vm.IPAddress,
		ProxmoxNode: vm.ProxmoxNode,
		CreatedAt:   vm.CreatedAt,
		UpdatedAt:   vm.UpdatedAt,
	}
}

// InfrastructureResponse represents the API response for tenant infrastructure
type InfrastructureResponse struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	Region           TenantRegion `json:"region"`
	Status           string     `json:"status"`
	DatabaseIP       string     `json:"database_ip"`
	LoadBalancerIP   string     `json:"load_balancer_ip"`
	WebServerIPs     JSONB      `json:"web_server_ips"`
	WebServerCount   int        `json:"web_server_count"`
	ProxmoxCluster   string     `json:"proxmox_cluster"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	ProvisionedAt    *time.Time `json:"provisioned_at"`
}

// ToResponse converts a TenantInfrastructure model to InfrastructureResponse
func (inf *TenantInfrastructure) ToResponse() InfrastructureResponse {
	return InfrastructureResponse{
		ID:             inf.ID,
		TenantID:       inf.TenantID,
		Region:         inf.Region,
		Status:         inf.Status,
		DatabaseIP:     inf.DatabaseIP,
		LoadBalancerIP: inf.LoadBalancerIP,
		WebServerIPs:   inf.WebServerIPs,
		WebServerCount: inf.WebServerCount,
		ProxmoxCluster: inf.ProxmoxCluster,
		CreatedAt:      inf.CreatedAt,
		UpdatedAt:      inf.UpdatedAt,
		ProvisionedAt:  inf.ProvisionedAt,
	}
}