package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"integration-service/internal/models"
	"integration-service/internal/services"
	"github.com/zen/shared/pkg/utils"
)

type WebhookHandler struct {
	webhookService services.WebhookService
	logger         *zap.Logger
}

func NewWebhookHandler(webhookService services.WebhookService, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
		logger:         logger,
	}
}

func (h *WebhookHandler) GetWebhooks(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	// Parse pagination parameters
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

	webhooks, err := h.webhookService.ListWebhooks(tenantID, page, limit)
	if err != nil {
		h.logger.Error("Failed to get webhooks", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get webhooks")
		return
	}

	utils.SuccessResponse(c, webhooks)
}

func (h *WebhookHandler) CreateWebhook(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	var req models.CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	webhook, err := h.webhookService.CreateWebhook(tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to create webhook", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create webhook")
		return
	}

	utils.SuccessResponse(c, webhook)
}

func (h *WebhookHandler) GetWebhook(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	webhookID := c.Param("id")
	if webhookID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid webhook ID")
		return
	}

	webhook, err := h.webhookService.GetWebhook(tenantID, webhookID)
	if err != nil {
		h.logger.Error("Failed to get webhook", zap.Error(err))
		utils.ErrorResponse(c, http.StatusNotFound, "Webhook not found")
		return
	}

	utils.SuccessResponse(c, webhook)
}

func (h *WebhookHandler) UpdateWebhook(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	webhookID := c.Param("id")
	if webhookID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid webhook ID")
		return
	}

	var req models.UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	webhook, err := h.webhookService.UpdateWebhook(tenantID, webhookID, &req)
	if err != nil {
		h.logger.Error("Failed to update webhook", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update webhook")
		return
	}

	utils.SuccessResponse(c, webhook)
}

func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	webhookID := c.Param("id")
	if webhookID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid webhook ID")
		return
	}

	err := h.webhookService.DeleteWebhook(tenantID, webhookID)
	if err != nil {
		h.logger.Error("Failed to delete webhook", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete webhook")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Webhook deleted successfully"})
}

func (h *WebhookHandler) TestWebhook(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	webhookID := c.Param("id")
	if webhookID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid webhook ID")
		return
	}

	var req models.WebhookTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	err := h.webhookService.TestWebhook(tenantID, webhookID, &req)
	if err != nil {
		h.logger.Error("Failed to test webhook", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to test webhook")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Webhook test sent successfully"})
}

func (h *WebhookHandler) ReceiveWebhook(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Token required")
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// TODO: Implement webhook processing logic based on token
	// This is a placeholder for receiving incoming webhooks from external systems
	h.logger.Info("Received webhook", 
		zap.String("token", token),
		zap.Any("payload", payload),
	)

	utils.SuccessResponse(c, gin.H{"message": "Webhook received successfully"})
}