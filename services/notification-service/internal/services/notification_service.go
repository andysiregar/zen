package services

import (
	"errors"
	"fmt"
	"net/smtp"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"notification-service/internal/config"
	"notification-service/internal/models"
	"notification-service/internal/repositories"
)

type NotificationService interface {
	// Send notifications
	SendNotification(userID, tenantID string, req *models.SendNotificationRequest) (*models.NotificationResponse, error)
	SendBulkNotification(tenantID string, req *models.SendBulkNotificationRequest) ([]*models.NotificationResponse, error)
	
	// Manage notifications
	GetUserNotifications(userID, tenantID string, limit, offset int) ([]*models.NotificationResponse, error)
	MarkAsRead(userID, tenantID, notificationID string) error
	GetUnreadCount(userID, tenantID string) (int64, error)
	
	// Templates
	CreateTemplate(tenantID string, template *models.NotificationTemplate) error
	GetTemplate(tenantID, templateID string) (*models.NotificationTemplate, error)
	UpdateTemplate(tenantID string, template *models.NotificationTemplate) error
	DeleteTemplate(tenantID, templateID string) error
	
	// Preferences
	GetUserPreferences(userID, tenantID string) (*models.NotificationPreference, error)
	UpdateUserPreferences(userID, tenantID string, preferences *models.NotificationPreference) error
	
	// Statistics
	GetNotificationStats(tenantID string) (map[string]interface{}, error)
}

type notificationService struct {
	repo       repositories.NotificationRepository
	config     *config.SMTPConfig
	logger     *zap.Logger
}

func NewNotificationService(repo repositories.NotificationRepository, cfg *config.SMTPConfig, logger *zap.Logger) NotificationService {
	return &notificationService{
		repo:   repo,
		config: cfg,
		logger: logger,
	}
}

func (s *notificationService) SendNotification(userID, tenantID string, req *models.SendNotificationRequest) (*models.NotificationResponse, error) {
	// Validate request
	if req.UserID == "" {
		return nil, errors.New("user ID is required")
	}
	if req.Type == "" {
		return nil, errors.New("notification type is required")
	}
	if req.Channel == "" {
		return nil, errors.New("notification channel is required")
	}

	// Set defaults
	if req.Priority == "" {
		req.Priority = "normal"
	}

	// Check user preferences
	preferences, err := s.repo.GetUserPreferences(tenantID, req.UserID)
	if err != nil {
		s.logger.Warn("Failed to get user preferences", zap.Error(err))
	}

	// Check if user has enabled this channel
	if preferences != nil && !s.isChannelEnabled(req.Channel, preferences) {
		return nil, errors.New("notification channel disabled for user")
	}

	// Create notification record
	notification := &models.Notification{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		UserID:      req.UserID,
		Type:        req.Type,
		Channel:     req.Channel,
		Subject:     req.Subject,
		Content:     req.Content,
		Data:        req.Data,
		Status:      "pending",
		Priority:    req.Priority,
		ScheduledAt: req.ScheduledAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	if err := s.repo.CreateNotification(tenantID, notification); err != nil {
		s.logger.Error("Failed to create notification", zap.Error(err))
		return nil, err
	}

	// Send the notification immediately if not scheduled
	if req.ScheduledAt == nil || req.ScheduledAt.Before(time.Now()) {
		go s.processNotification(tenantID, notification)
	}

	return s.notificationToResponse(notification), nil
}

func (s *notificationService) SendBulkNotification(tenantID string, req *models.SendBulkNotificationRequest) ([]*models.NotificationResponse, error) {
	var responses []*models.NotificationResponse

	for _, userID := range req.UserIDs {
		singleReq := &models.SendNotificationRequest{
			UserID:      userID,
			Type:        req.Type,
			Channel:     req.Channel,
			Subject:     req.Subject,
			Content:     req.Content,
			Data:        req.Data,
			Priority:    req.Priority,
			ScheduledAt: req.ScheduledAt,
		}

		response, err := s.SendNotification(userID, tenantID, singleReq)
		if err != nil {
			s.logger.Error("Failed to send notification to user", 
				zap.String("user_id", userID), 
				zap.Error(err))
			continue
		}

		responses = append(responses, response)
	}

	return responses, nil
}

func (s *notificationService) GetUserNotifications(userID, tenantID string, limit, offset int) ([]*models.NotificationResponse, error) {
	notifications, err := s.repo.GetUserNotifications(tenantID, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []*models.NotificationResponse
	for _, notification := range notifications {
		responses = append(responses, s.notificationToResponse(notification))
	}

	return responses, nil
}

func (s *notificationService) MarkAsRead(userID, tenantID, notificationID string) error {
	return s.repo.MarkAsRead(tenantID, notificationID, userID)
}

func (s *notificationService) GetUnreadCount(userID, tenantID string) (int64, error) {
	return s.repo.GetUnreadCount(tenantID, userID)
}

func (s *notificationService) CreateTemplate(tenantID string, template *models.NotificationTemplate) error {
	template.ID = uuid.New().String()
	template.TenantID = tenantID
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	return s.repo.CreateTemplate(tenantID, template)
}

func (s *notificationService) GetTemplate(tenantID, templateID string) (*models.NotificationTemplate, error) {
	return s.repo.GetTemplate(tenantID, templateID)
}

func (s *notificationService) UpdateTemplate(tenantID string, template *models.NotificationTemplate) error {
	template.UpdatedAt = time.Now()
	return s.repo.UpdateTemplate(tenantID, template)
}

func (s *notificationService) DeleteTemplate(tenantID, templateID string) error {
	return s.repo.DeleteTemplate(tenantID, templateID)
}

func (s *notificationService) GetUserPreferences(userID, tenantID string) (*models.NotificationPreference, error) {
	return s.repo.GetUserPreferences(tenantID, userID)
}

func (s *notificationService) UpdateUserPreferences(userID, tenantID string, preferences *models.NotificationPreference) error {
	preferences.UserID = userID
	preferences.TenantID = tenantID
	preferences.UpdatedAt = time.Now()

	if preferences.ID == "" {
		preferences.ID = uuid.New().String()
		preferences.CreatedAt = time.Now()
	}

	return s.repo.CreateOrUpdatePreferences(tenantID, preferences)
}

func (s *notificationService) GetNotificationStats(tenantID string) (map[string]interface{}, error) {
	return s.repo.GetNotificationStats(tenantID)
}

// Helper methods
func (s *notificationService) isChannelEnabled(channel string, preferences *models.NotificationPreference) bool {
	switch channel {
	case "email":
		return preferences.EmailEnabled
	case "push":
		return preferences.PushEnabled
	case "in_app":
		return preferences.InAppEnabled
	case "sms":
		return preferences.SMSEnabled
	default:
		return true
	}
}

func (s *notificationService) processNotification(tenantID string, notification *models.Notification) {
	var err error

	switch notification.Channel {
	case "email":
		err = s.sendEmail(notification)
	case "push":
		err = s.sendPushNotification(notification)
	case "in_app":
		err = s.sendInAppNotification(notification)
	case "sms":
		err = s.sendSMS(notification)
	default:
		err = fmt.Errorf("unsupported channel: %s", notification.Channel)
	}

	// Update notification status
	status := "sent"
	if err != nil {
		status = "failed"
		s.logger.Error("Failed to send notification", 
			zap.String("notification_id", notification.ID),
			zap.String("channel", notification.Channel),
			zap.Error(err))
	}

	if updateErr := s.repo.UpdateNotificationStatus(tenantID, notification.ID, status); updateErr != nil {
		s.logger.Error("Failed to update notification status", zap.Error(updateErr))
	}
}

func (s *notificationService) sendEmail(notification *models.Notification) error {
	if s.config.Host == "" || s.config.Username == "" {
		return errors.New("SMTP configuration not set")
	}

	// Simple email sending (in production, use a proper email service)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	to := []string{notification.UserID} // In real implementation, get user's email
	message := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", notification.Subject, notification.Content))

	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	err := smtp.SendMail(addr, auth, s.config.From, to, message)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *notificationService) sendPushNotification(notification *models.Notification) error {
	// Placeholder for push notification implementation
	// In real implementation, integrate with FCM, APNs, etc.
	s.logger.Info("Push notification sent", 
		zap.String("user_id", notification.UserID),
		zap.String("subject", notification.Subject))
	return nil
}

func (s *notificationService) sendInAppNotification(notification *models.Notification) error {
	// In-app notifications are stored in database and delivered via WebSocket/SSE
	// This is just marking as "sent" since it's stored
	s.logger.Info("In-app notification stored", 
		zap.String("user_id", notification.UserID),
		zap.String("subject", notification.Subject))
	return nil
}

func (s *notificationService) sendSMS(notification *models.Notification) error {
	// Placeholder for SMS implementation
	// In real implementation, integrate with Twilio, AWS SNS, etc.
	s.logger.Info("SMS notification sent", 
		zap.String("user_id", notification.UserID),
		zap.String("content", notification.Content))
	return nil
}

func (s *notificationService) notificationToResponse(notification *models.Notification) *models.NotificationResponse {
	return &models.NotificationResponse{
		ID:          notification.ID,
		TenantID:    notification.TenantID,
		UserID:      notification.UserID,
		Type:        notification.Type,
		Channel:     notification.Channel,
		Subject:     notification.Subject,
		Content:     notification.Content,
		Status:      notification.Status,
		Priority:    notification.Priority,
		ScheduledAt: notification.ScheduledAt,
		SentAt:      notification.SentAt,
		ReadAt:      notification.ReadAt,
		CreatedAt:   notification.CreatedAt,
		UpdatedAt:   notification.UpdatedAt,
	}
}