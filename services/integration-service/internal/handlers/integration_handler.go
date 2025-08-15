package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"integration-service/internal/services"
	"github.com/zen/shared/pkg/utils"
)

type IntegrationHandler struct {
	integrationService services.IntegrationService
	webhookService     services.WebhookService
	logger             *zap.Logger
}

func NewIntegrationHandler(integrationService services.IntegrationService, webhookService services.WebhookService, logger *zap.Logger) *IntegrationHandler {
	return &IntegrationHandler{
		integrationService: integrationService,
		webhookService:     webhookService,
		logger:             logger,
	}
}

func (h *IntegrationHandler) GetIntegrations(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	integrations, err := h.integrationService.GetIntegrationsByTenant(tenantID)
	if err != nil {
		h.logger.Error("Failed to get integrations", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get integrations")
		return
	}

	utils.SuccessResponse(c, integrations)
}

func (h *IntegrationHandler) CreateIntegration(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	var req struct {
		Name         string                 `json:"name" binding:"required"`
		Provider     string                 `json:"provider" binding:"required"`
		Config       map[string]interface{} `json:"config"`
		Credentials  map[string]string      `json:"credentials"`
		IsActive     bool                   `json:"is_active"`
		SyncEnabled  bool                   `json:"sync_enabled"`
		SyncInterval int                    `json:"sync_interval"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	integration, err := h.integrationService.CreateIntegration(tenantID, req.Name, req.Provider, req.Config, req.Credentials, req.IsActive, req.SyncEnabled, req.SyncInterval)
	if err != nil {
		h.logger.Error("Failed to create integration", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create integration")
		return
	}

	utils.SuccessResponse(c, integration)
}

func (h *IntegrationHandler) GetIntegration(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	integrationID := c.Param("id")
	if integrationID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid integration ID")
		return
	}

	integration, err := h.integrationService.GetIntegration(integrationID, tenantID)
	if err != nil {
		h.logger.Error("Failed to get integration", zap.Error(err))
		utils.ErrorResponse(c, http.StatusNotFound, "Integration not found")
		return
	}

	utils.SuccessResponse(c, integration)
}

func (h *IntegrationHandler) UpdateIntegration(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	integrationID := c.Param("id")
	if integrationID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid integration ID")
		return
	}

	var req struct {
		Name         string                 `json:"name"`
		Config       map[string]interface{} `json:"config"`
		Credentials  map[string]string      `json:"credentials"`
		IsActive     bool                   `json:"is_active"`
		SyncEnabled  bool                   `json:"sync_enabled"`
		SyncInterval int                    `json:"sync_interval"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	integration, err := h.integrationService.UpdateIntegration(integrationID, tenantID, req.Name, req.Config, req.Credentials, req.IsActive, req.SyncEnabled, req.SyncInterval)
	if err != nil {
		h.logger.Error("Failed to update integration", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update integration")
		return
	}

	utils.SuccessResponse(c, integration)
}

func (h *IntegrationHandler) DeleteIntegration(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	integrationID := c.Param("id")
	if integrationID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid integration ID")
		return
	}

	err := h.integrationService.DeleteIntegration(integrationID, tenantID)
	if err != nil {
		h.logger.Error("Failed to delete integration", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete integration")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Integration deleted successfully"})
}