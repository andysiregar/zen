package services

import (
	"fmt"
	"strings"
	"tenant-management/internal/models"
	"tenant-management/internal/repositories"
)

type TenantService interface {
	CreateTenant(req models.TenantCreateRequest) (*models.TenantResponse, error)
	GetTenant(id string) (*models.TenantResponse, error)
	GetTenantByDomain(domain string) (*models.TenantResponse, error)
	GetTenantBySubdomain(subdomain string) (*models.TenantResponse, error)
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
	// Generate database name from subdomain
	databaseName := fmt.Sprintf("tenant_%s", strings.ToLower(req.Subdomain))
	
	tenant := &models.Tenant{
		Name:         req.Name,
		Domain:       req.Domain,
		Subdomain:    req.Subdomain,
		DatabaseName: databaseName,
		DatabaseHost: "localhost", // Default to localhost for now
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
		Status:       models.TenantStatusActive,
	}
	
	if req.PlanType != "" {
		tenant.PlanType = req.PlanType
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

func (s *tenantService) GetTenantByDomain(domain string) (*models.TenantResponse, error) {
	tenant, err := s.repo.GetByDomain(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by domain: %w", err)
	}
	
	response := tenant.ToResponse()
	return &response, nil
}

func (s *tenantService) GetTenantBySubdomain(subdomain string) (*models.TenantResponse, error) {
	tenant, err := s.repo.GetBySubdomain(subdomain)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by subdomain: %w", err)
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
	if req.Domain != nil {
		tenant.Domain = *req.Domain
	}
	if req.Status != nil {
		tenant.Status = *req.Status
	}
	if req.ContactEmail != nil {
		tenant.ContactEmail = *req.ContactEmail
	}
	if req.ContactPhone != nil {
		tenant.ContactPhone = *req.ContactPhone
	}
	if req.PlanType != nil {
		tenant.PlanType = *req.PlanType
	}
	if req.MaxUsers != nil {
		tenant.MaxUsers = *req.MaxUsers
	}
	if req.MaxStorage != nil {
		tenant.MaxStorage = *req.MaxStorage
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