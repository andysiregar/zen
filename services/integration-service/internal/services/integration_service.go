package services

import (
	"time"
	"github.com/google/uuid"
	
	"integration-service/internal/models"
	"integration-service/internal/repositories"
)

type IntegrationService interface {
	CreateIntegration(tenantID, name, provider string, config map[string]interface{}, credentials map[string]string, isActive bool, syncEnabled bool, syncInterval int) (*models.Integration, error)
	GetIntegration(integrationID, tenantID string) (*models.Integration, error)
	GetIntegrationsByTenant(tenantID string) ([]*models.Integration, error)
	UpdateIntegration(integrationID, tenantID, name string, config map[string]interface{}, credentials map[string]string, isActive bool, syncEnabled bool, syncInterval int) (*models.Integration, error)
	DeleteIntegration(integrationID, tenantID string) error
	TestIntegration(integrationID, tenantID string) error
}

type integrationService struct {
	repo repositories.IntegrationRepository
}

func NewIntegrationService(repo repositories.IntegrationRepository) IntegrationService {
	return &integrationService{
		repo: repo,
	}
}

func (s *integrationService) CreateIntegration(tenantID, name, provider string, config map[string]interface{}, credentials map[string]string, isActive bool, syncEnabled bool, syncInterval int) (*models.Integration, error) {
	integration := &models.Integration{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		Name:          name,
		Provider:      provider,
		Configuration: config,
		Credentials:   credentials,
		Active:        isActive,
		SyncEnabled:   syncEnabled,
		SyncInterval:  syncInterval,
		Metadata:      make(map[string]interface{}),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	err := s.repo.CreateIntegration(tenantID, integration)
	if err != nil {
		return nil, err
	}
	
	return integration, nil
}

func (s *integrationService) GetIntegration(integrationID, tenantID string) (*models.Integration, error) {
	integration, err := s.repo.GetIntegration(tenantID, integrationID)
	if err != nil {
		return nil, err
	}
	
	return integration, nil
}

func (s *integrationService) GetIntegrationsByTenant(tenantID string) ([]*models.Integration, error) {
	integrations, _, err := s.repo.ListIntegrations(tenantID, 100, 0) // Default limit
	return integrations, err
}

func (s *integrationService) UpdateIntegration(integrationID, tenantID, name string, config map[string]interface{}, credentials map[string]string, isActive bool, syncEnabled bool, syncInterval int) (*models.Integration, error) {
	integration, err := s.repo.GetIntegration(tenantID, integrationID)
	if err != nil {
		return nil, err
	}
	
	if name != "" {
		integration.Name = name
	}
	if config != nil {
		integration.Configuration = config
	}
	if credentials != nil {
		integration.Credentials = credentials
	}
	integration.Active = isActive
	integration.SyncEnabled = syncEnabled
	if syncInterval > 0 {
		integration.SyncInterval = syncInterval
	}
	integration.UpdatedAt = time.Now()
	
	err = s.repo.UpdateIntegration(tenantID, integration)
	if err != nil {
		return nil, err
	}
	
	return integration, nil
}

func (s *integrationService) DeleteIntegration(integrationID, tenantID string) error {
	return s.repo.DeleteIntegration(tenantID, integrationID)
}

func (s *integrationService) TestIntegration(integrationID, tenantID string) error {
	integration, err := s.GetIntegration(integrationID, tenantID)
	if err != nil {
		return err
	}
	
	// TODO: Implement integration testing logic based on integration type
	_ = integration
	
	return nil
}