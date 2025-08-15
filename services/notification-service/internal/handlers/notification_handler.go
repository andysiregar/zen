package handlers

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zen/shared/pkg/middleware"
	"github.com/zen/shared/pkg/utils"
	"notification-service/internal/models"
	"notification-service/internal/services"
)

type NotificationHandler struct {
	notificationService services.NotificationService
	logger              *zap.Logger
}

func NewNotificationHandler(notificationService services.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		logger:              logger,
	}
}

// Helper function to get user and tenant context
func (h *NotificationHandler) getUserAndTenantContext(c *gin.Context) (string, string, error) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		return "", "", fmt.Errorf("user ID not found in context")
	}

	tenantContext, err := middleware.GetTenantContext(c)
	if err != nil {
		return "", "", fmt.Errorf("tenant context not found: %w", err)
	}

	return userID, tenantContext.TenantID, nil
}

// SendNotification handles POST /notifications
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	var req models.SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	notification, err := h.notificationService.SendNotification(userID, tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to send notification", zap.Error(err))
		utils.InternalServerErrorResponse(c, "Failed to send notification")
		return
	}

	utils.CreatedResponse(c, notification, "Notification sent successfully")
}

// SendBulkNotification handles POST /notifications/bulk
func (h *NotificationHandler) SendBulkNotification(c *gin.Context) {
	_, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	var req models.SendBulkNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	notifications, err := h.notificationService.SendBulkNotification(tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to send bulk notifications", zap.Error(err))
		utils.InternalServerErrorResponse(c, "Failed to send bulk notifications")
		return
	}

	utils.CreatedResponse(c, notifications, fmt.Sprintf("Sent %d notifications successfully", len(notifications)))
}

// GetUserNotifications handles GET /notifications
func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	// Parse pagination parameters
	limit := 50 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
		if limit > 100 {
			limit = 100
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	notifications, err := h.notificationService.GetUserNotifications(userID, tenantID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get user notifications", zap.Error(err))
		utils.InternalServerErrorResponse(c, "Failed to retrieve notifications")
		return
	}

	utils.SuccessResponse(c, notifications, "Notifications retrieved successfully")
}

// MarkAsRead handles PUT /notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		utils.BadRequestResponse(c, "Notification ID is required")
		return
	}

	err = h.notificationService.MarkAsRead(userID, tenantID, notificationID)
	if err != nil {
		h.logger.Error("Failed to mark notification as read", zap.Error(err))
		utils.InternalServerErrorResponse(c, "Failed to mark notification as read")
		return
	}

	utils.SuccessResponse(c, nil, "Notification marked as read")
}

// GetUnreadCount handles GET /notifications/unread/count
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	count, err := h.notificationService.GetUnreadCount(userID, tenantID)
	if err != nil {
		h.logger.Error("Failed to get unread count", zap.Error(err))
		utils.InternalServerErrorResponse(c, "Failed to get unread count")
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{"count": count}, "Unread count retrieved successfully")
}

// CreateTemplate handles POST /templates
func (h *NotificationHandler) CreateTemplate(c *gin.Context) {
	_, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	var template models.NotificationTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	err = h.notificationService.CreateTemplate(tenantID, &template)
	if err != nil {
		h.logger.Error("Failed to create template", zap.Error(err))
		utils.InternalServerErrorResponse(c, "Failed to create template")
		return
	}

	utils.CreatedResponse(c, template, "Template created successfully")
}

// GetTemplate handles GET /templates/:id
func (h *NotificationHandler) GetTemplate(c *gin.Context) {
	_, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	templateID := c.Param("id")
	if templateID == "" {
		utils.BadRequestResponse(c, "Template ID is required")
		return
	}

	template, err := h.notificationService.GetTemplate(tenantID, templateID)
	if err != nil {
		utils.NotFoundResponse(c, "Template not found")
		return
	}

	utils.SuccessResponse(c, template, "Template retrieved successfully")
}

// UpdateTemplate handles PUT /templates/:id
func (h *NotificationHandler) UpdateTemplate(c *gin.Context) {
	_, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	templateID := c.Param("id")
	if templateID == "" {
		utils.BadRequestResponse(c, "Template ID is required")
		return
	}

	var template models.NotificationTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	template.ID = templateID
	err = h.notificationService.UpdateTemplate(tenantID, &template)
	if err != nil {
		h.logger.Error("Failed to update template", zap.Error(err))
		utils.InternalServerErrorResponse(c, "Failed to update template")
		return
	}

	utils.SuccessResponse(c, template, "Template updated successfully")
}

// DeleteTemplate handles DELETE /templates/:id
func (h *NotificationHandler) DeleteTemplate(c *gin.Context) {
	_, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	templateID := c.Param("id")
	if templateID == "" {
		utils.BadRequestResponse(c, "Template ID is required")
		return
	}

	err = h.notificationService.DeleteTemplate(tenantID, templateID)
	if err != nil {
		h.logger.Error("Failed to delete template", zap.Error(err))
		utils.InternalServerErrorResponse(c, "Failed to delete template")
		return
	}

	utils.SuccessResponse(c, nil, "Template deleted successfully")
}

// GetUserPreferences handles GET /preferences
func (h *NotificationHandler) GetUserPreferences(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	preferences, err := h.notificationService.GetUserPreferences(userID, tenantID)
	if err != nil {
		h.logger.Error("Failed to get user preferences", zap.Error(err))
		utils.InternalServerErrorResponse(c, "Failed to get preferences")
		return
	}

	utils.SuccessResponse(c, preferences, "Preferences retrieved successfully")
}

// UpdateUserPreferences handles PUT /preferences
func (h *NotificationHandler) UpdateUserPreferences(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	var preferences models.NotificationPreference
	if err := c.ShouldBindJSON(&preferences); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	err = h.notificationService.UpdateUserPreferences(userID, tenantID, &preferences)
	if err != nil {
		h.logger.Error("Failed to update user preferences", zap.Error(err))
		utils.InternalServerErrorResponse(c, "Failed to update preferences")
		return
	}

	utils.SuccessResponse(c, preferences, "Preferences updated successfully")
}

// GetNotificationStats handles GET /stats
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	_, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	stats, err := h.notificationService.GetNotificationStats(tenantID)
	if err != nil {
		h.logger.Error("Failed to get notification stats", zap.Error(err))
		utils.InternalServerErrorResponse(c, "Failed to get statistics")
		return
	}

	utils.SuccessResponse(c, stats, "Statistics retrieved successfully")
}