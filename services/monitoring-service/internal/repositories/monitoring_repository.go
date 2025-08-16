package repositories

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/zen/shared/pkg/database"
	"monitoring-service/internal/models"
)

type MonitoringRepository interface {
	// Metrics
	CreateMetric(tenantID string, metric *models.Metric) error
	ListMetrics(tenantID string, filters models.MetricFilters, limit, offset int) ([]*models.Metric, int64, error)
	
	// Service Health
	CreateServiceHealth(health *models.ServiceHealth) error
	GetServiceHealth(serviceName string) (*models.ServiceHealth, error)
	UpdateServiceHealth(health *models.ServiceHealth) error
	GetAllServiceHealth() ([]*models.ServiceHealth, error)
	
	// Alert Rules
	CreateAlertRule(tenantID string, rule *models.AlertRule) error
	GetAlertRule(tenantID, ruleID string) (*models.AlertRule, error)
	UpdateAlertRule(tenantID string, rule *models.AlertRule) error
	DeleteAlertRule(tenantID, ruleID string) error
	GetAlertRules(tenantID string) ([]*models.AlertRule, error)
	
	// Alerts
	CreateAlert(tenantID string, alert *models.Alert) error
	GetAlert(tenantID, alertID string) (*models.Alert, error)
	UpdateAlert(tenantID string, alert *models.Alert) error
	ListAlerts(tenantID string, filters models.AlertFilters, limit, offset int) ([]*models.Alert, int64, error)
	
	// Logs
	CreateLogEntry(entry *models.LogEntry) error
	ListLogs(tenantID string, filters models.LogFilters, limit, offset int) ([]*models.LogEntry, int64, error)
	
	// Stats
	GetMonitoringStats(tenantID string) (*models.MonitoringStats, error)
	CleanupOldData(olderThan time.Time) error
}

type monitoringRepository struct {
	tenantDBManager *database.TenantDatabaseManager
	masterDB        *database.DatabaseManager
}

func NewMonitoringRepository(tenantDBManager *database.TenantDatabaseManager, masterDB *database.DatabaseManager) MonitoringRepository {
	return &monitoringRepository{
		tenantDBManager: tenantDBManager,
		masterDB:        masterDB,
	}
}

func (r *monitoringRepository) CreateMetric(tenantID string, metric *models.Metric) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	labelsJSON, _ := json.Marshal(metric.Labels)
	tagsJSON, _ := json.Marshal(metric.Tags)
	metadataJSON, _ := json.Marshal(metric.Metadata)

	result := db.Exec(`
		INSERT INTO metrics (id, tenant_id, name, type, value, labels, tags, unit, description, metadata, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, metric.ID, metric.TenantID, metric.Name, metric.Type, metric.Value, string(labelsJSON),
		string(tagsJSON), metric.Unit, metric.Description, string(metadataJSON),
		metric.Timestamp, metric.CreatedAt)

	return result.Error
}

func (r *monitoringRepository) ListMetrics(tenantID string, filters models.MetricFilters, limit, offset int) ([]*models.Metric, int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, 0, err
	}

	whereClause := "WHERE tenant_id = ?"
	args := []interface{}{tenantID}

	if filters.Name != "" {
		whereClause += " AND name = ?"
		args = append(args, filters.Name)
	}
	if filters.Type != "" {
		whereClause += " AND type = ?"
		args = append(args, filters.Type)
	}
	if filters.DateFrom != nil {
		whereClause += " AND timestamp >= ?"
		args = append(args, filters.DateFrom)
	}
	if filters.DateTo != nil {
		whereClause += " AND timestamp <= ?"
		args = append(args, filters.DateTo)
	}

	// Get total count
	var count int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM metrics %s", whereClause)
	db.Raw(countQuery, args...).Scan(&count)

	// Get metrics
	query := fmt.Sprintf(`
		SELECT id, tenant_id, name, type, value, labels, tags, unit, description, metadata, timestamp, created_at
		FROM metrics %s
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, limit, offset)
	rows, err := db.Raw(query, args...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var metrics []*models.Metric
	for rows.Next() {
		var metric models.Metric
		var labelsJSON, tagsJSON, metadataJSON string
		
		rows.Scan(
			&metric.ID, &metric.TenantID, &metric.Name, &metric.Type, &metric.Value,
			&labelsJSON, &tagsJSON, &metric.Unit, &metric.Description, &metadataJSON,
			&metric.Timestamp, &metric.CreatedAt,
		)

		json.Unmarshal([]byte(labelsJSON), &metric.Labels)
		json.Unmarshal([]byte(tagsJSON), &metric.Tags)
		json.Unmarshal([]byte(metadataJSON), &metric.Metadata)

		metrics = append(metrics, &metric)
	}

	return metrics, count, nil
}

func (r *monitoringRepository) CreateServiceHealth(health *models.ServiceHealth) error {
	masterDB := r.masterDB.GetMasterDB()

	metadataJSON, _ := json.Marshal(health.Metadata)
	checkDetailsJSON, _ := json.Marshal(health.CheckDetails)

	result := masterDB.Exec(`
		INSERT INTO service_health (id, service_name, status, version, uptime, response_time, last_checked_at, error_message, metadata, check_details, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, health.ID, health.ServiceName, health.Status, health.Version, health.Uptime,
		health.ResponseTime, health.LastCheckedAt, health.ErrorMessage,
		string(metadataJSON), string(checkDetailsJSON), health.CreatedAt, health.UpdatedAt)

	return result.Error
}

func (r *monitoringRepository) GetServiceHealth(serviceName string) (*models.ServiceHealth, error) {
	masterDB := r.masterDB.GetMasterDB()

	var health models.ServiceHealth
	var metadataJSON, checkDetailsJSON string
	
	err := masterDB.Raw(`
		SELECT id, service_name, status, version, uptime, response_time, last_checked_at, error_message, metadata, check_details, created_at, updated_at
		FROM service_health 
		WHERE service_name = ?
		ORDER BY last_checked_at DESC
		LIMIT 1
	`, serviceName).Row().Scan(
		&health.ID, &health.ServiceName, &health.Status, &health.Version, &health.Uptime,
		&health.ResponseTime, &health.LastCheckedAt, &health.ErrorMessage,
		&metadataJSON, &checkDetailsJSON, &health.CreatedAt, &health.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(metadataJSON), &health.Metadata)
	json.Unmarshal([]byte(checkDetailsJSON), &health.CheckDetails)

	return &health, nil
}

func (r *monitoringRepository) UpdateServiceHealth(health *models.ServiceHealth) error {
	masterDB := r.masterDB.GetMasterDB()

	metadataJSON, _ := json.Marshal(health.Metadata)
	checkDetailsJSON, _ := json.Marshal(health.CheckDetails)

	result := masterDB.Exec(`
		UPDATE service_health 
		SET status = ?, version = ?, uptime = ?, response_time = ?, last_checked_at = ?, error_message = ?, metadata = ?, check_details = ?, updated_at = ?
		WHERE id = ?
	`, health.Status, health.Version, health.Uptime, health.ResponseTime, health.LastCheckedAt,
		health.ErrorMessage, string(metadataJSON), string(checkDetailsJSON), health.UpdatedAt, health.ID)

	return result.Error
}

func (r *monitoringRepository) GetAllServiceHealth() ([]*models.ServiceHealth, error) {
	masterDB := r.masterDB.GetMasterDB()

	rows, err := masterDB.Raw(`
		SELECT DISTINCT ON (service_name) id, service_name, status, version, uptime, response_time, last_checked_at, error_message, metadata, check_details, created_at, updated_at
		FROM service_health 
		ORDER BY service_name, last_checked_at DESC
	`).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*models.ServiceHealth
	for rows.Next() {
		var health models.ServiceHealth
		var metadataJSON, checkDetailsJSON string
		
		rows.Scan(
			&health.ID, &health.ServiceName, &health.Status, &health.Version, &health.Uptime,
			&health.ResponseTime, &health.LastCheckedAt, &health.ErrorMessage,
			&metadataJSON, &checkDetailsJSON, &health.CreatedAt, &health.UpdatedAt,
		)

		json.Unmarshal([]byte(metadataJSON), &health.Metadata)
		json.Unmarshal([]byte(checkDetailsJSON), &health.CheckDetails)

		services = append(services, &health)
	}

	return services, nil
}

func (r *monitoringRepository) CreateAlertRule(tenantID string, rule *models.AlertRule) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	labelsJSON, _ := json.Marshal(rule.Labels)
	annotationsJSON, _ := json.Marshal(rule.Annotations)
	channelsJSON, _ := json.Marshal(rule.NotificationChannels)

	result := db.Exec(`
		INSERT INTO alert_rules (id, tenant_id, name, description, severity, condition, metric_query, threshold, duration, is_active, labels, annotations, notification_channels, last_evaluated, last_triggered, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, rule.ID, rule.TenantID, rule.Name, rule.Description, rule.Severity, rule.Condition,
		rule.MetricQuery, rule.Threshold, rule.Duration, rule.IsActive, string(labelsJSON),
		string(annotationsJSON), string(channelsJSON), rule.LastEvaluated, rule.LastTriggered,
		rule.CreatedBy, rule.CreatedAt, rule.UpdatedAt)

	return result.Error
}

func (r *monitoringRepository) GetAlertRule(tenantID, ruleID string) (*models.AlertRule, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var rule models.AlertRule
	var labelsJSON, annotationsJSON, channelsJSON string
	
	err = db.Raw(`
		SELECT id, tenant_id, name, description, severity, condition, metric_query, threshold, duration, is_active, labels, annotations, notification_channels, last_evaluated, last_triggered, created_by, created_at, updated_at
		FROM alert_rules 
		WHERE id = ? AND tenant_id = ?
	`, ruleID, tenantID).Row().Scan(
		&rule.ID, &rule.TenantID, &rule.Name, &rule.Description, &rule.Severity, &rule.Condition,
		&rule.MetricQuery, &rule.Threshold, &rule.Duration, &rule.IsActive, &labelsJSON,
		&annotationsJSON, &channelsJSON, &rule.LastEvaluated, &rule.LastTriggered,
		&rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(labelsJSON), &rule.Labels)
	json.Unmarshal([]byte(annotationsJSON), &rule.Annotations)
	json.Unmarshal([]byte(channelsJSON), &rule.NotificationChannels)

	return &rule, nil
}

func (r *monitoringRepository) UpdateAlertRule(tenantID string, rule *models.AlertRule) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	labelsJSON, _ := json.Marshal(rule.Labels)
	annotationsJSON, _ := json.Marshal(rule.Annotations)
	channelsJSON, _ := json.Marshal(rule.NotificationChannels)

	result := db.Exec(`
		UPDATE alert_rules 
		SET name = ?, description = ?, severity = ?, condition = ?, metric_query = ?, threshold = ?, duration = ?, is_active = ?, labels = ?, annotations = ?, notification_channels = ?, last_evaluated = ?, last_triggered = ?, updated_at = ?
		WHERE id = ? AND tenant_id = ?
	`, rule.Name, rule.Description, rule.Severity, rule.Condition, rule.MetricQuery, rule.Threshold,
		rule.Duration, rule.IsActive, string(labelsJSON), string(annotationsJSON), string(channelsJSON),
		rule.LastEvaluated, rule.LastTriggered, rule.UpdatedAt, rule.ID, tenantID)

	return result.Error
}

func (r *monitoringRepository) DeleteAlertRule(tenantID, ruleID string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		DELETE FROM alert_rules 
		WHERE id = ? AND tenant_id = ?
	`, ruleID, tenantID)

	return result.Error
}

func (r *monitoringRepository) GetAlertRules(tenantID string) ([]*models.AlertRule, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	rows, err := db.Raw(`
		SELECT id, tenant_id, name, description, severity, condition, metric_query, threshold, duration, is_active, labels, annotations, notification_channels, last_evaluated, last_triggered, created_by, created_at, updated_at
		FROM alert_rules 
		WHERE tenant_id = ?
		ORDER BY created_at DESC
	`, tenantID).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.AlertRule
	for rows.Next() {
		var rule models.AlertRule
		var labelsJSON, annotationsJSON, channelsJSON string
		
		rows.Scan(
			&rule.ID, &rule.TenantID, &rule.Name, &rule.Description, &rule.Severity, &rule.Condition,
			&rule.MetricQuery, &rule.Threshold, &rule.Duration, &rule.IsActive, &labelsJSON,
			&annotationsJSON, &channelsJSON, &rule.LastEvaluated, &rule.LastTriggered,
			&rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt,
		)

		json.Unmarshal([]byte(labelsJSON), &rule.Labels)
		json.Unmarshal([]byte(annotationsJSON), &rule.Annotations)
		json.Unmarshal([]byte(channelsJSON), &rule.NotificationChannels)

		rules = append(rules, &rule)
	}

	return rules, nil
}

func (r *monitoringRepository) CreateAlert(tenantID string, alert *models.Alert) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	labelsJSON, _ := json.Marshal(alert.Labels)
	annotationsJSON, _ := json.Marshal(alert.Annotations)

	result := db.Exec(`
		INSERT INTO alerts (id, tenant_id, name, description, severity, status, service_name, metric_name, condition, value, threshold, message, labels, annotations, triggered_at, resolved_at, acknowledged_at, acknowledged_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, alert.ID, alert.TenantID, alert.Name, alert.Description, alert.Severity, alert.Status,
		alert.ServiceName, alert.MetricName, alert.Condition, alert.Value, alert.Threshold,
		alert.Message, string(labelsJSON), string(annotationsJSON), alert.TriggeredAt,
		alert.ResolvedAt, alert.AcknowledgedAt, alert.AcknowledgedBy, alert.CreatedAt, alert.UpdatedAt)

	return result.Error
}

func (r *monitoringRepository) GetAlert(tenantID, alertID string) (*models.Alert, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var alert models.Alert
	var labelsJSON, annotationsJSON string
	
	err = db.Raw(`
		SELECT id, tenant_id, name, description, severity, status, service_name, metric_name, condition, value, threshold, message, labels, annotations, triggered_at, resolved_at, acknowledged_at, acknowledged_by, created_at, updated_at
		FROM alerts 
		WHERE id = ? AND tenant_id = ?
	`, alertID, tenantID).Row().Scan(
		&alert.ID, &alert.TenantID, &alert.Name, &alert.Description, &alert.Severity, &alert.Status,
		&alert.ServiceName, &alert.MetricName, &alert.Condition, &alert.Value, &alert.Threshold,
		&alert.Message, &labelsJSON, &annotationsJSON, &alert.TriggeredAt, &alert.ResolvedAt,
		&alert.AcknowledgedAt, &alert.AcknowledgedBy, &alert.CreatedAt, &alert.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(labelsJSON), &alert.Labels)
	json.Unmarshal([]byte(annotationsJSON), &alert.Annotations)

	return &alert, nil
}

func (r *monitoringRepository) UpdateAlert(tenantID string, alert *models.Alert) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	labelsJSON, _ := json.Marshal(alert.Labels)
	annotationsJSON, _ := json.Marshal(alert.Annotations)

	result := db.Exec(`
		UPDATE alerts 
		SET name = ?, description = ?, severity = ?, status = ?, service_name = ?, metric_name = ?, condition = ?, value = ?, threshold = ?, message = ?, labels = ?, annotations = ?, triggered_at = ?, resolved_at = ?, acknowledged_at = ?, acknowledged_by = ?, updated_at = ?
		WHERE id = ? AND tenant_id = ?
	`, alert.Name, alert.Description, alert.Severity, alert.Status, alert.ServiceName, alert.MetricName,
		alert.Condition, alert.Value, alert.Threshold, alert.Message, string(labelsJSON),
		string(annotationsJSON), alert.TriggeredAt, alert.ResolvedAt, alert.AcknowledgedAt,
		alert.AcknowledgedBy, alert.UpdatedAt, alert.ID, tenantID)

	return result.Error
}

func (r *monitoringRepository) ListAlerts(tenantID string, filters models.AlertFilters, limit, offset int) ([]*models.Alert, int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, 0, err
	}

	whereClause := "WHERE tenant_id = ?"
	args := []interface{}{tenantID}

	if filters.Status != "" {
		whereClause += " AND status = ?"
		args = append(args, filters.Status)
	}
	if filters.Severity != "" {
		whereClause += " AND severity = ?"
		args = append(args, filters.Severity)
	}
	if filters.ServiceName != "" {
		whereClause += " AND service_name = ?"
		args = append(args, filters.ServiceName)
	}
	if filters.DateFrom != nil {
		whereClause += " AND triggered_at >= ?"
		args = append(args, filters.DateFrom)
	}
	if filters.DateTo != nil {
		whereClause += " AND triggered_at <= ?"
		args = append(args, filters.DateTo)
	}

	// Get total count
	var count int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM alerts %s", whereClause)
	db.Raw(countQuery, args...).Scan(&count)

	// Get alerts
	query := fmt.Sprintf(`
		SELECT id, tenant_id, name, description, severity, status, service_name, metric_name, condition, value, threshold, message, labels, annotations, triggered_at, resolved_at, acknowledged_at, acknowledged_by, created_at, updated_at
		FROM alerts %s
		ORDER BY triggered_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, limit, offset)
	rows, err := db.Raw(query, args...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var alerts []*models.Alert
	for rows.Next() {
		var alert models.Alert
		var labelsJSON, annotationsJSON string
		
		rows.Scan(
			&alert.ID, &alert.TenantID, &alert.Name, &alert.Description, &alert.Severity, &alert.Status,
			&alert.ServiceName, &alert.MetricName, &alert.Condition, &alert.Value, &alert.Threshold,
			&alert.Message, &labelsJSON, &annotationsJSON, &alert.TriggeredAt, &alert.ResolvedAt,
			&alert.AcknowledgedAt, &alert.AcknowledgedBy, &alert.CreatedAt, &alert.UpdatedAt,
		)

		json.Unmarshal([]byte(labelsJSON), &alert.Labels)
		json.Unmarshal([]byte(annotationsJSON), &alert.Annotations)

		alerts = append(alerts, &alert)
	}

	return alerts, count, nil
}

func (r *monitoringRepository) CreateLogEntry(entry *models.LogEntry) error {
	// For simplicity, storing logs in the master database
	// In production, you might want a separate logging database
	masterDB := r.masterDB.GetMasterDB()

	fieldsJSON, _ := json.Marshal(entry.Fields)

	result := masterDB.Exec(`
		INSERT INTO log_entries (id, tenant_id, service_name, level, message, fields, trace_id, span_id, source, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, entry.ID, entry.TenantID, entry.ServiceName, entry.Level, entry.Message,
		string(fieldsJSON), entry.TraceID, entry.SpanID, entry.Source, entry.Timestamp, entry.CreatedAt)

	return result.Error
}

func (r *monitoringRepository) ListLogs(tenantID string, filters models.LogFilters, limit, offset int) ([]*models.LogEntry, int64, error) {
	masterDB := r.masterDB.GetMasterDB()

	whereClause := "WHERE tenant_id = ?"
	args := []interface{}{tenantID}

	if filters.Level != "" {
		whereClause += " AND level = ?"
		args = append(args, filters.Level)
	}
	if filters.ServiceName != "" {
		whereClause += " AND service_name = ?"
		args = append(args, filters.ServiceName)
	}
	if filters.TraceID != "" {
		whereClause += " AND trace_id = ?"
		args = append(args, filters.TraceID)
	}
	if filters.DateFrom != nil {
		whereClause += " AND timestamp >= ?"
		args = append(args, filters.DateFrom)
	}
	if filters.DateTo != nil {
		whereClause += " AND timestamp <= ?"
		args = append(args, filters.DateTo)
	}

	// Get total count
	var count int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM log_entries %s", whereClause)
	masterDB.Raw(countQuery, args...).Scan(&count)

	// Get logs
	query := fmt.Sprintf(`
		SELECT id, tenant_id, service_name, level, message, fields, trace_id, span_id, source, timestamp, created_at
		FROM log_entries %s
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, limit, offset)
	rows, err := masterDB.Raw(query, args...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*models.LogEntry
	for rows.Next() {
		var entry models.LogEntry
		var fieldsJSON string
		
		rows.Scan(
			&entry.ID, &entry.TenantID, &entry.ServiceName, &entry.Level, &entry.Message,
			&fieldsJSON, &entry.TraceID, &entry.SpanID, &entry.Source, &entry.Timestamp, &entry.CreatedAt,
		)

		json.Unmarshal([]byte(fieldsJSON), &entry.Fields)
		logs = append(logs, &entry)
	}

	return logs, count, nil
}

func (r *monitoringRepository) GetMonitoringStats(tenantID string) (*models.MonitoringStats, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	stats := &models.MonitoringStats{}

	// Get metrics count
	db.Raw("SELECT COUNT(*) FROM metrics WHERE tenant_id = ?", tenantID).Scan(&stats.TotalMetrics)

	// Get alert counts
	db.Raw("SELECT COUNT(*) FROM alerts WHERE tenant_id = ? AND status = ?", tenantID, models.AlertStatusActive).Scan(&stats.ActiveAlerts)
	db.Raw("SELECT COUNT(*) FROM alerts WHERE tenant_id = ? AND status = ?", tenantID, models.AlertStatusResolved).Scan(&stats.ResolvedAlerts)

	// Get service health counts from master DB
	masterDB := r.masterDB.GetMasterDB()
	masterDB.Raw("SELECT COUNT(*) FROM service_health WHERE status = ?", models.HealthStatusHealthy).Scan(&stats.HealthyServices)
	masterDB.Raw("SELECT COUNT(*) FROM service_health WHERE status = ?", models.HealthStatusUnhealthy).Scan(&stats.UnhealthyServices)

	// Calculate averages (simplified)
	var avgResponseTime float64
	masterDB.Raw("SELECT AVG(response_time) FROM service_health WHERE response_time > 0").Scan(&avgResponseTime)
	stats.AverageResponseTime = avgResponseTime

	// Calculate error rate (simplified - would need more complex logic in production)
	stats.ErrorRate = 5.0 // Placeholder

	return stats, nil
}

func (r *monitoringRepository) CleanupOldData(olderThan time.Time) error {
	masterDB := r.masterDB.GetMasterDB()
	
	// Cleanup old logs
	masterDB.Exec("DELETE FROM log_entries WHERE created_at < ?", olderThan)
	
	// Cleanup old service health records (keep only latest per service)
	masterDB.Exec(`
		DELETE FROM service_health 
		WHERE id NOT IN (
			SELECT DISTINCT ON (service_name) id 
			FROM service_health 
			ORDER BY service_name, last_checked_at DESC
		) AND created_at < ?
	`, olderThan.Add(24*time.Hour)) // Keep 24 hours of history
	
	return nil
}