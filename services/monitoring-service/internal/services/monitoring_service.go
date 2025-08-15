package services

import (
	"time"
	"github.com/google/uuid"
	
	"monitoring-service/internal/models"
	"monitoring-service/internal/repositories"
)

type MonitoringService interface {
	// Metrics
	CreateMetric(tenantID string, request *models.CreateMetricRequest) (*models.Metric, error)
	GetMetrics(tenantID string, filters models.MetricFilters, page, limit int) (*models.MetricListResponse, error)
	QueryMetrics(tenantID string, query *models.MetricQueryRequest) ([]models.Metric, error)
	
	// Health Checks
	GetServiceHealth(serviceName string) (*models.ServiceHealth, error)
	GetAllServiceHealth() (*models.HealthCheckResponse, error)
	UpdateServiceHealth(health *models.ServiceHealth) error
	
	// Alerts
	CreateAlertRule(tenantID string, request *models.CreateAlertRuleRequest) (*models.AlertRule, error)
	GetAlertRules(tenantID string) ([]*models.AlertRule, error)
	UpdateAlertRule(tenantID, ruleID string, request *models.CreateAlertRuleRequest) (*models.AlertRule, error)
	DeleteAlertRule(tenantID, ruleID string) error
	
	GetAlerts(tenantID string, filters models.AlertFilters, page, limit int) (*models.AlertListResponse, error)
	UpdateAlert(tenantID, alertID string, request *models.UpdateAlertRequest) (*models.Alert, error)
	AcknowledgeAlert(tenantID, alertID, userID string) error
	ResolveAlert(tenantID, alertID string) error
	
	// Logs
	CreateLogEntry(entry *models.LogEntry) error
	GetLogs(tenantID string, filters models.LogFilters, page, limit int) (*models.LogListResponse, error)
	
	// Stats
	GetMonitoringStats(tenantID string) (*models.MonitoringStats, error)
}

type monitoringService struct {
	repo repositories.MonitoringRepository
}

func NewMonitoringService(repo repositories.MonitoringRepository) MonitoringService {
	return &monitoringService{
		repo: repo,
	}
}

func (s *monitoringService) CreateMetric(tenantID string, request *models.CreateMetricRequest) (*models.Metric, error) {
	now := time.Now()
	metric := &models.Metric{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Name:        request.Name,
		Type:        request.Type,
		Value:       request.Value,
		Labels:      request.Labels,
		Tags:        request.Tags,
		Unit:        request.Unit,
		Description: request.Description,
		Metadata:    request.Metadata,
		Timestamp:   now,
		CreatedAt:   now,
	}

	err := s.repo.CreateMetric(tenantID, metric)
	if err != nil {
		return nil, err
	}

	return metric, nil
}

func (s *monitoringService) GetMetrics(tenantID string, filters models.MetricFilters, page, limit int) (*models.MetricListResponse, error) {
	offset := (page - 1) * limit
	metrics, total, err := s.repo.ListMetrics(tenantID, filters, limit, offset)
	if err != nil {
		return nil, err
	}

	return &models.MetricListResponse{
		Metrics:    metrics,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (s *monitoringService) QueryMetrics(tenantID string, query *models.MetricQueryRequest) ([]models.Metric, error) {
	// TODO: Implement metric query language parsing
	// For now, return empty slice
	return []models.Metric{}, nil
}

func (s *monitoringService) GetServiceHealth(serviceName string) (*models.ServiceHealth, error) {
	return s.repo.GetServiceHealth(serviceName)
}

func (s *monitoringService) GetAllServiceHealth() (*models.HealthCheckResponse, error) {
	services, err := s.repo.GetAllServiceHealth()
	if err != nil {
		return nil, err
	}

	// Calculate overall status and summary
	summary := models.HealthSummary{
		TotalServices: len(services),
	}

	overallStatus := models.HealthStatusHealthy
	for _, service := range services {
		switch service.Status {
		case models.HealthStatusHealthy:
			summary.HealthyServices++
		case models.HealthStatusUnhealthy:
			summary.UnhealthyServices++
			overallStatus = models.HealthStatusUnhealthy
		case models.HealthStatusDegraded:
			summary.DegradedServices++
			if overallStatus == models.HealthStatusHealthy {
				overallStatus = models.HealthStatusDegraded
			}
		}
	}

	return &models.HealthCheckResponse{
		Services: services,
		Overall:  overallStatus,
		Summary:  summary,
	}, nil
}

func (s *monitoringService) UpdateServiceHealth(health *models.ServiceHealth) error {
	health.UpdatedAt = time.Now()
	return s.repo.UpdateServiceHealth(health)
}

func (s *monitoringService) CreateAlertRule(tenantID string, request *models.CreateAlertRuleRequest) (*models.AlertRule, error) {
	now := time.Now()
	rule := &models.AlertRule{
		ID:                   uuid.New().String(),
		TenantID:             tenantID,
		Name:                 request.Name,
		Description:          request.Description,
		Severity:             request.Severity,
		Condition:            request.Condition,
		MetricQuery:          request.MetricQuery,
		Threshold:            request.Threshold,
		Duration:             request.Duration,
		Labels:               request.Labels,
		Annotations:          request.Annotations,
		NotificationChannels: request.NotificationChannels,
		IsActive:             true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	err := s.repo.CreateAlertRule(tenantID, rule)
	if err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *monitoringService) GetAlertRules(tenantID string) ([]*models.AlertRule, error) {
	return s.repo.GetAlertRules(tenantID)
}

func (s *monitoringService) UpdateAlertRule(tenantID, ruleID string, request *models.CreateAlertRuleRequest) (*models.AlertRule, error) {
	rule, err := s.repo.GetAlertRule(tenantID, ruleID)
	if err != nil {
		return nil, err
	}

	rule.Name = request.Name
	rule.Description = request.Description
	rule.Severity = request.Severity
	rule.Condition = request.Condition
	rule.MetricQuery = request.MetricQuery
	rule.Threshold = request.Threshold
	rule.Duration = request.Duration
	rule.Labels = request.Labels
	rule.Annotations = request.Annotations
	rule.NotificationChannels = request.NotificationChannels
	rule.UpdatedAt = time.Now()

	err = s.repo.UpdateAlertRule(tenantID, rule)
	if err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *monitoringService) DeleteAlertRule(tenantID, ruleID string) error {
	return s.repo.DeleteAlertRule(tenantID, ruleID)
}

func (s *monitoringService) GetAlerts(tenantID string, filters models.AlertFilters, page, limit int) (*models.AlertListResponse, error) {
	offset := (page - 1) * limit
	alerts, total, err := s.repo.ListAlerts(tenantID, filters, limit, offset)
	if err != nil {
		return nil, err
	}

	return &models.AlertListResponse{
		Alerts:     alerts,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (s *monitoringService) UpdateAlert(tenantID, alertID string, request *models.UpdateAlertRequest) (*models.Alert, error) {
	alert, err := s.repo.GetAlert(tenantID, alertID)
	if err != nil {
		return nil, err
	}

	if request.Status != "" {
		alert.Status = request.Status
		if request.Status == models.AlertStatusResolved {
			now := time.Now()
			alert.ResolvedAt = &now
		}
	}

	if request.AcknowledgedBy != "" {
		now := time.Now()
		alert.AcknowledgedAt = &now
		alert.AcknowledgedBy = request.AcknowledgedBy
	}

	alert.UpdatedAt = time.Now()

	err = s.repo.UpdateAlert(tenantID, alert)
	if err != nil {
		return nil, err
	}

	return alert, nil
}

func (s *monitoringService) AcknowledgeAlert(tenantID, alertID, userID string) error {
	alert, err := s.repo.GetAlert(tenantID, alertID)
	if err != nil {
		return err
	}

	now := time.Now()
	alert.AcknowledgedAt = &now
	alert.AcknowledgedBy = userID
	alert.UpdatedAt = time.Now()

	return s.repo.UpdateAlert(tenantID, alert)
}

func (s *monitoringService) ResolveAlert(tenantID, alertID string) error {
	alert, err := s.repo.GetAlert(tenantID, alertID)
	if err != nil {
		return err
	}

	now := time.Now()
	alert.Status = models.AlertStatusResolved
	alert.ResolvedAt = &now
	alert.UpdatedAt = time.Now()

	return s.repo.UpdateAlert(tenantID, alert)
}

func (s *monitoringService) CreateLogEntry(entry *models.LogEntry) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	entry.CreatedAt = time.Now()
	return s.repo.CreateLogEntry(entry)
}

func (s *monitoringService) GetLogs(tenantID string, filters models.LogFilters, page, limit int) (*models.LogListResponse, error) {
	offset := (page - 1) * limit
	logs, total, err := s.repo.ListLogs(tenantID, filters, limit, offset)
	if err != nil {
		return nil, err
	}

	return &models.LogListResponse{
		Logs:       logs,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (s *monitoringService) GetMonitoringStats(tenantID string) (*models.MonitoringStats, error) {
	return s.repo.GetMonitoringStats(tenantID)
}