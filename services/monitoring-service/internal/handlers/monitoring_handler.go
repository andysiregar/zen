package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"monitoring-service/internal/models"
	"monitoring-service/internal/services"
	"github.com/zen/shared/pkg/utils"
)

type MonitoringHandler struct {
	monitoringService services.MonitoringService
	logger            *zap.Logger
}

func NewMonitoringHandler(monitoringService services.MonitoringService, logger *zap.Logger) *MonitoringHandler {
	return &MonitoringHandler{
		monitoringService: monitoringService,
		logger:            logger,
	}
}

// Health Check Endpoints

func (h *MonitoringHandler) GetHealthCheck(c *gin.Context) {
	health, err := h.monitoringService.GetAllServiceHealth()
	if err != nil {
		h.logger.Error("Failed to get health check", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get health check")
		return
	}

	utils.SuccessResponse(c, health)
}

func (h *MonitoringHandler) GetServiceHealth(c *gin.Context) {
	serviceName := c.Param("service")
	if serviceName == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Service name required")
		return
	}

	health, err := h.monitoringService.GetServiceHealth(serviceName)
	if err != nil {
		h.logger.Error("Failed to get service health", zap.Error(err))
		utils.ErrorResponse(c, http.StatusNotFound, "Service not found")
		return
	}

	utils.SuccessResponse(c, health)
}

// Metrics Endpoints

func (h *MonitoringHandler) GetMetrics(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	// Parse pagination
	page := 1
	limit := 50
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// Parse filters
	filters := models.MetricFilters{
		Name: c.Query("name"),
		Type: c.Query("type"),
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}

	metrics, err := h.monitoringService.GetMetrics(tenantID, filters, page, limit)
	if err != nil {
		h.logger.Error("Failed to get metrics", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get metrics")
		return
	}

	utils.SuccessResponse(c, metrics)
}

func (h *MonitoringHandler) CreateMetric(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	var req models.CreateMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	metric, err := h.monitoringService.CreateMetric(tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to create metric", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create metric")
		return
	}

	utils.SuccessResponse(c, metric)
}

func (h *MonitoringHandler) QueryMetrics(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	var req models.MetricQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	metrics, err := h.monitoringService.QueryMetrics(tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to query metrics", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to query metrics")
		return
	}

	utils.SuccessResponse(c, metrics)
}

// Alert Endpoints

func (h *MonitoringHandler) GetAlerts(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	// Parse pagination
	page := 1
	limit := 20
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse filters
	filters := models.AlertFilters{
		Status:      models.AlertStatus(c.Query("status")),
		Severity:    models.AlertSeverity(c.Query("severity")),
		ServiceName: c.Query("service_name"),
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}

	alerts, err := h.monitoringService.GetAlerts(tenantID, filters, page, limit)
	if err != nil {
		h.logger.Error("Failed to get alerts", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get alerts")
		return
	}

	utils.SuccessResponse(c, alerts)
}

func (h *MonitoringHandler) UpdateAlert(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	alertID := c.Param("id")
	if alertID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Alert ID required")
		return
	}

	var req models.UpdateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	alert, err := h.monitoringService.UpdateAlert(tenantID, alertID, &req)
	if err != nil {
		h.logger.Error("Failed to update alert", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update alert")
		return
	}

	utils.SuccessResponse(c, alert)
}

func (h *MonitoringHandler) AcknowledgeAlert(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	alertID := c.Param("id")
	if alertID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Alert ID required")
		return
	}

	userID := c.GetString("user_id") // From JWT middleware
	if userID == "" {
		userID = "system"
	}

	err := h.monitoringService.AcknowledgeAlert(tenantID, alertID, userID)
	if err != nil {
		h.logger.Error("Failed to acknowledge alert", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to acknowledge alert")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Alert acknowledged successfully"})
}

func (h *MonitoringHandler) ResolveAlert(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	alertID := c.Param("id")
	if alertID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Alert ID required")
		return
	}

	err := h.monitoringService.ResolveAlert(tenantID, alertID)
	if err != nil {
		h.logger.Error("Failed to resolve alert", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to resolve alert")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Alert resolved successfully"})
}

// Alert Rules Endpoints

func (h *MonitoringHandler) GetAlertRules(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	rules, err := h.monitoringService.GetAlertRules(tenantID)
	if err != nil {
		h.logger.Error("Failed to get alert rules", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get alert rules")
		return
	}

	utils.SuccessResponse(c, rules)
}

func (h *MonitoringHandler) CreateAlertRule(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	var req models.CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	rule, err := h.monitoringService.CreateAlertRule(tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to create alert rule", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create alert rule")
		return
	}

	utils.SuccessResponse(c, rule)
}

func (h *MonitoringHandler) UpdateAlertRule(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	ruleID := c.Param("id")
	if ruleID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Rule ID required")
		return
	}

	var req models.CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	rule, err := h.monitoringService.UpdateAlertRule(tenantID, ruleID, &req)
	if err != nil {
		h.logger.Error("Failed to update alert rule", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update alert rule")
		return
	}

	utils.SuccessResponse(c, rule)
}

func (h *MonitoringHandler) DeleteAlertRule(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	ruleID := c.Param("id")
	if ruleID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Rule ID required")
		return
	}

	err := h.monitoringService.DeleteAlertRule(tenantID, ruleID)
	if err != nil {
		h.logger.Error("Failed to delete alert rule", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete alert rule")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Alert rule deleted successfully"})
}

// Logs Endpoints

func (h *MonitoringHandler) GetLogs(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	// Parse pagination
	page := 1
	limit := 100
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// Parse filters
	filters := models.LogFilters{
		Level:       c.Query("level"),
		ServiceName: c.Query("service_name"),
		TraceID:     c.Query("trace_id"),
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}

	logs, err := h.monitoringService.GetLogs(tenantID, filters, page, limit)
	if err != nil {
		h.logger.Error("Failed to get logs", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get logs")
		return
	}

	utils.SuccessResponse(c, logs)
}

func (h *MonitoringHandler) CreateLogEntry(c *gin.Context) {
	var entry models.LogEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	err := h.monitoringService.CreateLogEntry(&entry)
	if err != nil {
		h.logger.Error("Failed to create log entry", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create log entry")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Log entry created successfully"})
}

// Stats Endpoints

func (h *MonitoringHandler) GetMonitoringStats(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	stats, err := h.monitoringService.GetMonitoringStats(tenantID)
	if err != nil {
		h.logger.Error("Failed to get monitoring stats", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get monitoring stats")
		return
	}

	utils.SuccessResponse(c, stats)
}