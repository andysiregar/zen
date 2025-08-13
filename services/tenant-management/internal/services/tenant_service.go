package services

import (
	"fmt"
	"strings"
	"github.com/zen/shared/pkg/models"
	"tenant-management/internal/repositories"
)

type TenantService interface {
	CreateTenant(req models.TenantCreateRequest) (*models.TenantResponse, error)
	GetTenant(id string) (*models.TenantResponse, error)
	GetTenantBySlug(slug string) (*models.TenantResponse, error)
	UpdateTenant(id string, req models.TenantUpdateRequest) (*models.TenantResponse, error)
	DeleteTenant(id string) error
	ListTenants(limit, offset int) ([]*models.TenantResponse, int64, error)
}

type tenantService struct {
	repo repositories.TenantRepository
}

func NewTenantService(repo repositories.TenantRepository) TenantService {
	return &tenantService{repo: repo}
}

func (s *tenantService) CreateTenant(req models.TenantCreateRequest) (*models.TenantResponse, error) {
	// Generate database name from slug
	databaseName := fmt.Sprintf("tenant_%s", strings.ToLower(req.Slug))
	
	tenant := &models.Tenant{
		OrganizationID:      req.OrganizationID,
		Name:                req.Name,
		Slug:                req.Slug,
		Description:         req.Description,
		Status:              models.TenantStatusProvisioning,
		DbHost:              "localhost", // Default to localhost for now
		DbPort:              5432,
		DbName:              databaseName,
		DbUser:              "tenant_user", // Will be configured properly later
		DbPasswordEncrypted: "encrypted_password", // Will be encrypted properly later
		DbSslMode:           "require",
		Settings:            make(models.JSONB),
		Features:            make(models.JSONB),
	}
	
	err := s.repo.Create(tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}
	
	response := tenant.ToResponse()
	return &response, nil
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