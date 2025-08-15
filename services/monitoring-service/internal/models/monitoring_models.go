package models

import (
	"errors"
	"time"
)

var (
	ErrMetricNotFound = errors.New("metric not found")
	ErrAlertNotFound  = errors.New("alert not found")
)

// HealthStatus represents the health status of a service
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus represents the current status of an alert
type AlertStatus string

const (
	AlertStatusActive    AlertStatus = "active"
	AlertStatusResolved  AlertStatus = "resolved"
	AlertStatusSuppressed AlertStatus = "suppressed"
)

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	ID              string                 `json:"id" gorm:"primaryKey"`
	ServiceName     string                 `json:"service_name" gorm:"index"`
	Status          HealthStatus           `json:"status" gorm:"index"`
	Version         string                 `json:"version,omitempty"`
	Uptime          int64                  `json:"uptime_seconds,omitempty"`
	ResponseTime    float64                `json:"response_time_ms,omitempty"`
	LastCheckedAt   time.Time              `json:"last_checked_at" gorm:"index"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	CheckDetails    map[string]interface{} `json:"check_details,omitempty" gorm:"type:jsonb"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// Metric represents a custom metric data point
type Metric struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	TenantID    string                 `json:"tenant_id" gorm:"index"`
	Name        string                 `json:"name" gorm:"index"`
	Type        string                 `json:"type"` // counter, gauge, histogram, summary
	Value       float64                `json:"value"`
	Labels      map[string]string      `json:"labels,omitempty" gorm:"type:jsonb"`
	Tags        []string               `json:"tags,omitempty" gorm:"type:text[]"`
	Unit        string                 `json:"unit,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	Timestamp   time.Time              `json:"timestamp" gorm:"index"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Alert represents an alert/incident
type Alert struct {
	ID            string                 `json:"id" gorm:"primaryKey"`
	TenantID      string                 `json:"tenant_id" gorm:"index"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Severity      AlertSeverity          `json:"severity" gorm:"index"`
	Status        AlertStatus            `json:"status" gorm:"index"`
	ServiceName   string                 `json:"service_name,omitempty" gorm:"index"`
	MetricName    string                 `json:"metric_name,omitempty"`
	Condition     string                 `json:"condition"` // e.g., "value > 100"
	Value         float64                `json:"value,omitempty"`
	Threshold     float64                `json:"threshold,omitempty"`
	Message       string                 `json:"message"`
	Labels        map[string]string      `json:"labels,omitempty" gorm:"type:jsonb"`
	Annotations   map[string]interface{} `json:"annotations,omitempty" gorm:"type:jsonb"`
	TriggeredAt   time.Time              `json:"triggered_at" gorm:"index"`
	ResolvedAt    *time.Time             `json:"resolved_at,omitempty"`
	AcknowledgedAt *time.Time            `json:"acknowledged_at,omitempty"`
	AcknowledgedBy string                `json:"acknowledged_by,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// AlertRule represents a rule for generating alerts
type AlertRule struct {
	ID             string                 `json:"id" gorm:"primaryKey"`
	TenantID       string                 `json:"tenant_id" gorm:"index"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Severity       AlertSeverity          `json:"severity"`
	Condition      string                 `json:"condition"` // expression like "cpu_usage > 80"
	MetricQuery    string                 `json:"metric_query"`
	Threshold      float64                `json:"threshold"`
	Duration       int                    `json:"duration_seconds"` // how long condition must be true
	IsActive       bool                   `json:"is_active" gorm:"default:true"`
	Labels         map[string]string      `json:"labels,omitempty" gorm:"type:jsonb"`
	Annotations    map[string]interface{} `json:"annotations,omitempty" gorm:"type:jsonb"`
	NotificationChannels []string         `json:"notification_channels,omitempty" gorm:"type:text[]"`
	LastEvaluated  *time.Time             `json:"last_evaluated,omitempty"`
	LastTriggered  *time.Time             `json:"last_triggered,omitempty"`
	CreatedBy      string                 `json:"created_by,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// LogEntry represents a log entry
type LogEntry struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	TenantID    string                 `json:"tenant_id" gorm:"index"`
	ServiceName string                 `json:"service_name" gorm:"index"`
	Level       string                 `json:"level" gorm:"index"` // debug, info, warn, error, fatal
	Message     string                 `json:"message"`
	Fields      map[string]interface{} `json:"fields,omitempty" gorm:"type:jsonb"`
	TraceID     string                 `json:"trace_id,omitempty" gorm:"index"`
	SpanID      string                 `json:"span_id,omitempty"`
	Source      string                 `json:"source,omitempty"` // file:line
	Timestamp   time.Time              `json:"timestamp" gorm:"index"`
	CreatedAt   time.Time              `json:"created_at"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	TenantID           string    `json:"tenant_id"`
	ServiceName        string    `json:"service_name"`
	CPUUsagePercent    float64   `json:"cpu_usage_percent"`
	MemoryUsagePercent float64   `json:"memory_usage_percent"`
	DiskUsagePercent   float64   `json:"disk_usage_percent"`
	RequestCount       int64     `json:"request_count"`
	ErrorCount         int64     `json:"error_count"`
	ResponseTime       float64   `json:"avg_response_time_ms"`
	ActiveConnections  int64     `json:"active_connections"`
	Timestamp          time.Time `json:"timestamp"`
}

// MonitoringDashboard represents a monitoring dashboard
type MonitoringDashboard struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	TenantID    string                 `json:"tenant_id" gorm:"index"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config" gorm:"type:jsonb"`
	Widgets     []DashboardWidget      `json:"widgets" gorm:"type:jsonb"`
	IsPublic    bool                   `json:"is_public" gorm:"default:false"`
	Tags        []string               `json:"tags,omitempty" gorm:"type:text[]"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// DashboardWidget represents a widget in a dashboard
type DashboardWidget struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // chart, table, stat, log
	Title       string                 `json:"title"`
	Query       string                 `json:"query"`
	Config      map[string]interface{} `json:"config"`
	Position    WidgetPosition         `json:"position"`
}

type WidgetPosition struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// API Request/Response Models

type CreateMetricRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Type        string                 `json:"type" binding:"required"`
	Value       float64                `json:"value" binding:"required"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Unit        string                 `json:"unit,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type CreateAlertRuleRequest struct {
	Name                 string                 `json:"name" binding:"required"`
	Description          string                 `json:"description"`
	Severity             AlertSeverity          `json:"severity" binding:"required"`
	Condition            string                 `json:"condition" binding:"required"`
	MetricQuery          string                 `json:"metric_query" binding:"required"`
	Threshold            float64                `json:"threshold" binding:"required"`
	Duration             int                    `json:"duration_seconds"`
	Labels               map[string]string      `json:"labels,omitempty"`
	Annotations          map[string]interface{} `json:"annotations,omitempty"`
	NotificationChannels []string               `json:"notification_channels,omitempty"`
}

type UpdateAlertRequest struct {
	Status         AlertStatus `json:"status,omitempty"`
	AcknowledgedBy string      `json:"acknowledged_by,omitempty"`
}

type MetricQueryRequest struct {
	Query     string     `json:"query" binding:"required"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Step      string     `json:"step,omitempty"` // 1m, 5m, 1h, etc.
}

type HealthCheckResponse struct {
	Services []*ServiceHealth `json:"services"`
	Overall  HealthStatus     `json:"overall_status"`
	Summary  HealthSummary    `json:"summary"`
}

type HealthSummary struct {
	TotalServices   int `json:"total_services"`
	HealthyServices int `json:"healthy_services"`
	UnhealthyServices int `json:"unhealthy_services"`
	DegradedServices  int `json:"degraded_services"`
}

type MetricListResponse struct {
	Metrics    []*Metric `json:"metrics"`
	TotalCount int64     `json:"total_count"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
}

type AlertListResponse struct {
	Alerts     []*Alert `json:"alerts"`
	TotalCount int64    `json:"total_count"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
}

type LogListResponse struct {
	Logs       []*LogEntry `json:"logs"`
	TotalCount int64       `json:"total_count"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
}

type MonitoringStats struct {
	TotalMetrics        int64   `json:"total_metrics"`
	ActiveAlerts        int64   `json:"active_alerts"`
	ResolvedAlerts      int64   `json:"resolved_alerts"`
	HealthyServices     int     `json:"healthy_services"`
	UnhealthyServices   int     `json:"unhealthy_services"`
	AverageResponseTime float64 `json:"average_response_time_ms"`
	ErrorRate           float64 `json:"error_rate_percent"`
}

type MetricFilters struct {
	Name        string
	Type        string
	ServiceName string
	Tags        []string
	DateFrom    *time.Time
	DateTo      *time.Time
}

type AlertFilters struct {
	Status      AlertStatus
	Severity    AlertSeverity
	ServiceName string
	DateFrom    *time.Time
	DateTo      *time.Time
}

type LogFilters struct {
	Level       string
	ServiceName string
	TraceID     string
	DateFrom    *time.Time
	DateTo      *time.Time
}