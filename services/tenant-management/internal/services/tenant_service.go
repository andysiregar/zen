package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"github.com/zen/shared/pkg/models"
	"github.com/zen/shared/pkg/database"
	"github.com/zen/shared/pkg/proxmox"
	"tenant-management/internal/repositories"
)

type TenantService interface {
	CreateTenant(req models.TenantCreateRequest) (*models.TenantResponse, error)
	GetTenant(id string) (*models.TenantResponse, error)
	GetTenantBySlug(slug string) (*models.TenantResponse, error)
	UpdateTenant(id string, req models.TenantUpdateRequest) (*models.TenantResponse, error)
	DeleteTenant(id string) error
	ListTenants(limit, offset int) ([]*models.TenantResponse, int64, error)
	
	// Infrastructure management
	GetTenantInfrastructure(tenantID string) (*models.TenantResponse, error)
	DestroyTenantInfrastructure(tenantID string) error
}

type tenantService struct {
	repo           repositories.TenantRepository
	dbManager      *database.DatabaseManager
	proxmoxClients map[models.TenantRegion]*proxmox.DummyProxmoxClient
}

func NewTenantService(repo repositories.TenantRepository, dbManager *database.DatabaseManager) TenantService {
	// Initialize Proxmox clients for each region
	proxmoxClients := map[models.TenantRegion]*proxmox.DummyProxmoxClient{
		models.TenantRegionAsia:   proxmox.NewDummyProxmoxClient("asia"),
		models.TenantRegionUS:     proxmox.NewDummyProxmoxClient("us"),
		models.TenantRegionEurope: proxmox.NewDummyProxmoxClient("europe"),
	}

	return &tenantService{
		repo:           repo,
		dbManager:      dbManager,
		proxmoxClients: proxmoxClients,
	}
}

func (s *tenantService) CreateTenant(req models.TenantCreateRequest) (*models.TenantResponse, error) {
	// Validate slug uniqueness
	existingTenant, err := s.repo.GetBySlug(req.Slug)
	if err == nil && existingTenant != nil {
		return nil, fmt.Errorf("tenant with slug '%s' already exists", req.Slug)
	}

	// Generate database name from slug
	databaseName := fmt.Sprintf("tenant_%s", strings.ToLower(req.Slug))
	
	// Generate domain and subdomain from slug
	domain := fmt.Sprintf("%s.localhost", req.Slug)
	subdomain := req.Slug
	
	// Get regional Proxmox cluster
	proxmoxCluster := s.getRegionalProxmoxCluster(req.Region)

	tenant := &models.Tenant{
		OrganizationID:      req.OrganizationID,
		Name:                req.Name,
		Slug:                req.Slug,
		Description:         req.Description,
		Status:              models.TenantStatusProvisioning,
		Region:              req.Region,
		
		// Infrastructure will be provisioned via Proxmox API
		ProxmoxCluster:      proxmoxCluster,
		DbHost:              "", // Will be set after VM provisioning
		DbPort:              5432,
		DbName:              databaseName,
		DbUser:              "tenant_user",
		DbPasswordEncrypted: "encrypted_password",
		DbSslMode:           "disable",
		DatabaseVMIP:        "", // Will be set after VM provisioning
		WebServerClusterIPs: "", // Will be set after web cluster provisioning
		
		// Required schema fields
		Domain:        domain,
		Subdomain:     subdomain, 
		DatabaseName:  databaseName,
		
		// Settings
		Settings:            make(models.JSONB),
		Features:            make(models.JSONB),
	}
	
	// Create tenant record first
	err = s.repo.Create(tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Provision tenant database (async in production)
	go s.provisionTenantDatabase(tenant)
	
	response := tenant.ToResponse()
	return &response, nil
}

// provisionTenantDatabase creates dedicated VMs and database infrastructure using dummy Proxmox client
func (s *tenantService) provisionTenantDatabase(tenant *models.Tenant) {
	// Get the appropriate Proxmox client for the tenant's region
	proxmoxClient, exists := s.proxmoxClients[tenant.Region]
	if !exists {
		// Fallback to US region client
		proxmoxClient = s.proxmoxClients[models.TenantRegionUS]
	}
	
	// Provision the infrastructure using the dummy Proxmox client
	result, err := proxmoxClient.ProvisionTenantInfrastructure(tenant.Slug, tenant.Region)
	if err != nil {
		// In production, you'd log this error and set tenant status to error
		fmt.Printf("Failed to provision infrastructure for tenant %s: %v\n", tenant.Slug, err)
		tenant.Status = models.TenantStatusInactive
		s.repo.Update(tenant)
		return
	}
	
	// Update tenant with provisioned infrastructure details
	tenant.DatabaseVMIP = result.DatabaseVM.IP
	tenant.DbHost = result.DatabaseVM.IP
	
	// Convert web cluster IPs to JSON
	webIPs := make([]string, len(result.WebCluster.VMs))
	for i, vm := range result.WebCluster.VMs {
		webIPs[i] = vm.IP
	}
	webIPsJSON, _ := json.Marshal(webIPs)
	tenant.WebServerClusterIPs = string(webIPsJSON)
	
	// Simulate database creation on the database VM
	err = proxmoxClient.CreateDatabase(result.DatabaseVM.IP, tenant.Slug)
	if err != nil {
		fmt.Printf("Failed to create database for tenant %s: %v\n", tenant.Slug, err)
	}
	
	// Simulate web server setup
	err = proxmoxClient.SetupWebServers(result.WebCluster, tenant.Slug)
	if err != nil {
		fmt.Printf("Failed to setup web servers for tenant %s: %v\n", tenant.Slug, err)
	}
	
	// Update tenant status to active
	tenant.Status = models.TenantStatusActive
	now := time.Now()
	tenant.ProvisionedAt = &now
	
	// Update tenant in database
	err = s.repo.Update(tenant)
	if err != nil {
		fmt.Printf("Failed to update tenant %s after provisioning: %v\n", tenant.Slug, err)
	}
	
	fmt.Printf("✅ Successfully provisioned infrastructure for tenant %s in %s region:\n", tenant.Slug, tenant.Region)
	fmt.Printf("   Database VM: %s\n", result.DatabaseVM.IP)
	fmt.Printf("   Web Cluster: %v\n", webIPs)
	fmt.Printf("   Load Balancer: %s\n", result.WebCluster.LoadBalance)
}

func (s *tenantService) GetTenant(id string) (*models.TenantResponse, error) {
	tenant, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	
	response := tenant.ToResponse()
	return &response, nil
}

func (s *tenantService) GetTenantBySlug(slug string) (*models.TenantResponse, error) {
	tenant, err := s.repo.GetBySlug(slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by slug: %w", err)
	}
	
	response := tenant.ToResponse()
	return &response, nil
}

func (s *tenantService) UpdateTenant(id string, req models.TenantUpdateRequest) (*models.TenantResponse, error) {
	tenant, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	
	// Update fields if provided
	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if req.Description != nil {
		tenant.Description = *req.Description
	}
	if req.Status != nil {
		tenant.Status = *req.Status
	}
	if req.Settings != nil {
		tenant.Settings = *req.Settings
	}
	if req.Features != nil {
		tenant.Features = *req.Features
	}
	
	err = s.repo.Update(tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}
	
	response := tenant.ToResponse()
	return &response, nil
}

func (s *tenantService) DeleteTenant(id string) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}
	
	err = s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	
	return nil
}

func (s *tenantService) ListTenants(limit, offset int) ([]*models.TenantResponse, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	
	tenants, err := s.repo.List(limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}
	
	count, err := s.repo.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}
	
	var responses []*models.TenantResponse
	for _, tenant := range tenants {
		response := tenant.ToResponse()
		responses = append(responses, &response)
	}
	
	return responses, count, nil
}

// GetTenantInfrastructure returns detailed infrastructure information for a tenant
func (s *tenantService) GetTenantInfrastructure(tenantID string) (*models.TenantResponse, error) {
	tenant, err := s.repo.GetByID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	
	response := tenant.ToResponse()
	return &response, nil
}

// DestroyTenantInfrastructure destroys all VMs and infrastructure for a tenant
func (s *tenantService) DestroyTenantInfrastructure(tenantID string) error {
	tenant, err := s.repo.GetByID(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}
	
	// Get the appropriate Proxmox client for the tenant's region
	proxmoxClient, exists := s.proxmoxClients[tenant.Region]
	if !exists {
		proxmoxClient = s.proxmoxClients[models.TenantRegionUS]
	}
	
	// Destroy the infrastructure
	err = proxmoxClient.DestroyTenantInfrastructure(tenant.Slug)
	if err != nil {
		return fmt.Errorf("failed to destroy infrastructure for tenant %s: %w", tenant.Slug, err)
	}
	
	// Update tenant status
	tenant.Status = models.TenantStatusInactive
	tenant.DatabaseVMIP = ""
	tenant.DbHost = ""
	tenant.WebServerClusterIPs = "[]"
	
	err = s.repo.Update(tenant)
	if err != nil {
		return fmt.Errorf("failed to update tenant after destroying infrastructure: %w", err)
	}
	
	fmt.Printf("✅ Successfully destroyed infrastructure for tenant %s\n", tenant.Slug)
	return nil
}

// getRegionalProxmoxCluster returns the Proxmox cluster for the given region
func (s *tenantService) getRegionalProxmoxCluster(region models.TenantRegion) string {
	switch region {
	case models.TenantRegionAsia:
		return "asia-proxmox-cluster"    // Asia Proxmox cluster
	case models.TenantRegionUS:
		return "us-proxmox-cluster"      // US Proxmox cluster  
	case models.TenantRegionEurope:
		return "eu-proxmox-cluster"      // EU Proxmox cluster
	default:
		return "us-proxmox-cluster"      // Default to US
	}
}

// GetTenantRegionalEndpoint returns the regional API endpoint for a tenant
func (s *tenantService) GetTenantRegionalEndpoint(tenant *models.Tenant) string {
	switch tenant.Region {
	case models.TenantRegionAsia:
		return "https://asia.chilldesk.io"
	case models.TenantRegionUS:
		return "https://us.chilldesk.io"
	case models.TenantRegionEurope:
		return "https://eu.chilldesk.io"
	default:
		return "https://us.chilldesk.io"
	}
}