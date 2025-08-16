package repositories

import (
	"github.com/zen/shared/pkg/database"
	"notification-service/internal/models"
)

type NotificationRepository interface {
	// Notification operations
	CreateNotification(tenantID string, notification *models.Notification) error
	GetNotification(tenantID, notificationID string) (*models.Notification, error)
	GetUserNotifications(tenantID, userID string, limit, offset int) ([]*models.Notification, error)
	UpdateNotificationStatus(tenantID, notificationID, status string) error
	MarkAsRead(tenantID, notificationID, userID string) error
	
	// Template operations
	CreateTemplate(tenantID string, template *models.NotificationTemplate) error
	GetTemplate(tenantID, templateID string) (*models.NotificationTemplate, error)
	GetTemplateByName(tenantID, name string) (*models.NotificationTemplate, error)
	UpdateTemplate(tenantID string, template *models.NotificationTemplate) error
	DeleteTemplate(tenantID, templateID string) error
	
	// Preferences operations
	GetUserPreferences(tenantID, userID string) (*models.NotificationPreference, error)
	CreateOrUpdatePreferences(tenantID string, preferences *models.NotificationPreference) error
	
	// Statistics
	GetUnreadCount(tenantID, userID string) (int64, error)
	GetNotificationStats(tenantID string) (map[string]interface{}, error)
}

type notificationRepository struct {
	tenantDBManager *database.TenantDatabaseManager
}

func NewNotificationRepository(tenantDBManager *database.TenantDatabaseManager) NotificationRepository {
	return &notificationRepository{
		tenantDBManager: tenantDBManager,
	}
}

func (r *notificationRepository) CreateNotification(tenantID string, notification *models.Notification) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		INSERT INTO notifications (id, tenant_id, user_id, type, channel, subject, content, data, status, priority, scheduled_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, notification.ID, notification.TenantID, notification.UserID, notification.Type, notification.Channel, 
		notification.Subject, notification.Content, "{}", notification.Status, notification.Priority, 
		notification.ScheduledAt, notification.CreatedAt, notification.UpdatedAt)

	return result.Error
}

func (r *notificationRepository) GetNotification(tenantID, notificationID string) (*models.Notification, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var notification models.Notification
	err = db.Raw(`
		SELECT id, tenant_id, user_id, type, channel, subject, content, status, priority, scheduled_at, sent_at, read_at, created_at, updated_at
		FROM notifications 
		WHERE id = ? AND tenant_id = ?
	`, notificationID, tenantID).Scan(&notification).Error

	if err != nil {
		return nil, err
	}

	return &notification, nil
}

func (r *notificationRepository) GetUserNotifications(tenantID, userID string, limit, offset int) ([]*models.Notification, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var notifications []*models.Notification
	err = db.Raw(`
		SELECT id, tenant_id, user_id, type, channel, subject, content, status, priority, scheduled_at, sent_at, read_at, created_at, updated_at
		FROM notifications 
		WHERE tenant_id = ? AND user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, tenantID, userID, limit, offset).Scan(&notifications).Error

	return notifications, err
}

func (r *notificationRepository) UpdateNotificationStatus(tenantID, notificationID, status string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	var updateFields string
	if status == "sent" {
		updateFields = "status = ?, sent_at = NOW(), updated_at = NOW()"
	} else {
		updateFields = "status = ?, updated_at = NOW()"
	}

	result := db.Exec(`
		UPDATE notifications SET `+updateFields+`
		WHERE id = ? AND tenant_id = ?
	`, status, notificationID, tenantID)

	return result.Error
}

func (r *notificationRepository) MarkAsRead(tenantID, notificationID, userID string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		UPDATE notifications 
		SET status = 'read', read_at = NOW(), updated_at = NOW()
		WHERE id = ? AND tenant_id = ? AND user_id = ?
	`, notificationID, tenantID, userID)

	return result.Error
}

func (r *notificationRepository) CreateTemplate(tenantID string, template *models.NotificationTemplate) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		INSERT INTO notification_templates (id, tenant_id, name, type, subject, content, variables, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, template.ID, template.TenantID, template.Name, template.Type, template.Subject, 
		template.Content, "{}", template.IsActive, template.CreatedAt, template.UpdatedAt)

	return result.Error
}

func (r *notificationRepository) GetTemplate(tenantID, templateID string) (*models.NotificationTemplate, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var template models.NotificationTemplate
	err = db.Raw(`
		SELECT id, tenant_id, name, type, subject, content, is_active, created_at, updated_at
		FROM notification_templates 
		WHERE id = ? AND tenant_id = ?
	`, templateID, tenantID).Scan(&template).Error

	if err != nil {
		return nil, err
	}

	return &template, nil
}

func (r *notificationRepository) GetTemplateByName(tenantID, name string) (*models.NotificationTemplate, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var template models.NotificationTemplate
	err = db.Raw(`
		SELECT id, tenant_id, name, type, subject, content, is_active, created_at, updated_at
		FROM notification_templates 
		WHERE name = ? AND tenant_id = ? AND is_active = true
	`, name, tenantID).Scan(&template).Error

	if err != nil {
		return nil, err
	}

	return &template, nil
}

func (r *notificationRepository) UpdateTemplate(tenantID string, template *models.NotificationTemplate) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		UPDATE notification_templates 
		SET name = ?, type = ?, subject = ?, content = ?, variables = ?, is_active = ?, updated_at = ?
		WHERE id = ? AND tenant_id = ?
	`, template.Name, template.Type, template.Subject, template.Content, "{}", 
		template.IsActive, template.UpdatedAt, template.ID, tenantID)

	return result.Error
}

func (r *notificationRepository) DeleteTemplate(tenantID, templateID string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		DELETE FROM notification_templates 
		WHERE id = ? AND tenant_id = ?
	`, templateID, tenantID)

	return result.Error
}

func (r *notificationRepository) GetUserPreferences(tenantID, userID string) (*models.NotificationPreference, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var preferences models.NotificationPreference
	err = db.Raw(`
		SELECT id, tenant_id, user_id, email_enabled, push_enabled, in_app_enabled, sms_enabled, 
			   quiet_hours_start, quiet_hours_end, timezone, created_at, updated_at
		FROM notification_preferences 
		WHERE tenant_id = ? AND user_id = ?
	`, tenantID, userID).Scan(&preferences).Error

	if err != nil {
		// Return default preferences if not found
		return &models.NotificationPreference{
			TenantID:     tenantID,
			UserID:       userID,
			EmailEnabled: true,
			PushEnabled:  true,
			InAppEnabled: true,
			SMSEnabled:   false,
			Timezone:     "UTC",
		}, nil
	}

	return &preferences, nil
}

func (r *notificationRepository) CreateOrUpdatePreferences(tenantID string, preferences *models.NotificationPreference) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		INSERT INTO notification_preferences (id, tenant_id, user_id, email_enabled, push_enabled, in_app_enabled, sms_enabled, quiet_hours_start, quiet_hours_end, timezone, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (tenant_id, user_id) DO UPDATE SET
		email_enabled = EXCLUDED.email_enabled,
		push_enabled = EXCLUDED.push_enabled,
		in_app_enabled = EXCLUDED.in_app_enabled,
		sms_enabled = EXCLUDED.sms_enabled,
		quiet_hours_start = EXCLUDED.quiet_hours_start,
		quiet_hours_end = EXCLUDED.quiet_hours_end,
		timezone = EXCLUDED.timezone,
		updated_at = EXCLUDED.updated_at
	`, preferences.ID, preferences.TenantID, preferences.UserID, preferences.EmailEnabled, 
		preferences.PushEnabled, preferences.InAppEnabled, preferences.SMSEnabled, 
		preferences.QuietHoursStart, preferences.QuietHoursEnd, preferences.Timezone, 
		preferences.CreatedAt, preferences.UpdatedAt)

	return result.Error
}

func (r *notificationRepository) GetUnreadCount(tenantID, userID string) (int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return 0, err
	}

	var count int64
	err = db.Raw(`
		SELECT COUNT(*) 
		FROM notifications 
		WHERE tenant_id = ? AND user_id = ? AND status != 'read'
	`, tenantID, userID).Scan(&count).Error

	return count, err
}

func (r *notificationRepository) GetNotificationStats(tenantID string) (map[string]interface{}, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})

	// Total notifications
	var total int64
	db.Raw("SELECT COUNT(*) FROM notifications WHERE tenant_id = ?", tenantID).Scan(&total)
	stats["total"] = total

	// Status breakdown
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	db.Raw(`
		SELECT status, COUNT(*) as count 
		FROM notifications 
		WHERE tenant_id = ? 
		GROUP BY status
	`, tenantID).Scan(&statusStats)
	stats["by_status"] = statusStats

	// Channel breakdown
	var channelStats []struct {
		Channel string `json:"channel"`
		Count   int64  `json:"count"`
	}
	db.Raw(`
		SELECT channel, COUNT(*) as count 
		FROM notifications 
		WHERE tenant_id = ? 
		GROUP BY channel
	`, tenantID).Scan(&channelStats)
	stats["by_channel"] = channelStats

	return stats, nil
}