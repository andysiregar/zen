package tenant_models

import (
	"time"
	"gorm.io/gorm"
	"github.com/zen/shared/pkg/models"
)

type TicketType string
type TicketPriority string
type TicketStatus string
type TicketResolution string
type TicketChannel string

const (
	TicketTypeTask    TicketType = "task"
	TicketTypeBug     TicketType = "bug"
	TicketTypeFeature TicketType = "feature"
	TicketTypeSupport TicketType = "support"
)

const (
	PriorityLow      TicketPriority = "low"
	PriorityMedium   TicketPriority = "medium"
	PriorityHigh     TicketPriority = "high"
	PriorityCritical TicketPriority = "critical"
	PriorityBlocker  TicketPriority = "blocker"
)

const (
	StatusOpen       TicketStatus = "open"
	StatusInProgress TicketStatus = "in_progress"
	StatusResolved   TicketStatus = "resolved"
	StatusClosed     TicketStatus = "closed"
)

const (
	ResolutionFixed    TicketResolution = "fixed"
	ResolutionWontFix  TicketResolution = "won't_fix"
	ResolutionDuplicate TicketResolution = "duplicate"
	ResolutionInvalid  TicketResolution = "invalid"
)

const (
	ChannelEmail TicketChannel = "email"
	ChannelWeb   TicketChannel = "web"
	ChannelChat  TicketChannel = "chat"
	ChannelPhone TicketChannel = "phone"
	ChannelAPI   TicketChannel = "api"
)

type Ticket struct {
	ID           string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TicketNumber int            `json:"ticket_number" gorm:"autoIncrement;uniqueIndex"` // Auto-incrementing ticket number per tenant
	Title        string         `json:"title" gorm:"not null;size:500"`
	Description  string         `json:"description" gorm:"type:text"`
	
	// Classification
	TicketType TicketType     `json:"ticket_type" gorm:"type:varchar(50);default:'task'"`
	Priority   TicketPriority `json:"priority" gorm:"type:varchar(20);default:'medium'"`
	Status     TicketStatus   `json:"status" gorm:"type:varchar(50);default:'open'"`
	
	// Relationships
	ProjectID      *string `json:"project_id" gorm:"type:uuid;index"`
	ParentTicketID *string `json:"parent_ticket_id" gorm:"type:uuid"`
	
	// Assignment (All reference Master DB users.id)
	ReporterID *string `json:"reporter_id" gorm:"type:uuid;not null"` // Who created the ticket
	AssigneeID *string `json:"assignee_id" gorm:"type:uuid"`          // Who is assigned to work on it
	
	// Customer Support Fields
	CustomerEmail string        `json:"customer_email" gorm:"size:255"`
	CustomerName  string        `json:"customer_name" gorm:"size:255"`
	Channel       TicketChannel `json:"channel" gorm:"type:varchar(50)"`
	
	// Resolution
	Resolution TicketResolution `json:"resolution" gorm:"type:varchar(100)"`
	ResolvedAt *time.Time       `json:"resolved_at"`
	ResolvedBy *string          `json:"resolved_by" gorm:"type:uuid"` // FK to master.users.id
	
	// Time Tracking
	EstimatedHours *float64 `json:"estimated_hours" gorm:"type:decimal(8,2)"`
	ActualHours    *float64 `json:"actual_hours" gorm:"type:decimal(8,2)"`
	
	// Metadata
	Labels       models.JSONB `json:"labels" gorm:"type:jsonb;default:'[]'"`         // Flexible tagging system
	CustomFields models.JSONB `json:"custom_fields" gorm:"type:jsonb;default:'{}'"`  // Tenant-specific custom fields
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type TicketCreateRequest struct {
	Title       string         `json:"title" binding:"required,min=5,max=500"`
	Description string         `json:"description,omitempty"`
	TicketType  TicketType     `json:"ticket_type,omitempty"`
	Priority    TicketPriority `json:"priority,omitempty"`
	ProjectID   *string        `json:"project_id,omitempty" binding:"omitempty,uuid"`
	AssigneeID  *string        `json:"assignee_id,omitempty" binding:"omitempty,uuid"`
	
	// Customer Support Fields
	CustomerEmail string        `json:"customer_email,omitempty" binding:"omitempty,email"`
	CustomerName  string        `json:"customer_name,omitempty"`
	Channel       TicketChannel `json:"channel,omitempty"`
	
	// Time Tracking
	EstimatedHours *float64 `json:"estimated_hours,omitempty" binding:"omitempty,gt=0"`
	
	// Metadata
	Labels       models.JSONB `json:"labels,omitempty"`
	CustomFields models.JSONB `json:"custom_fields,omitempty"`
}

type TicketUpdateRequest struct {
	Title        *string           `json:"title,omitempty" binding:"omitempty,min=5,max=500"`
	Description  *string           `json:"description,omitempty"`
	TicketType   *TicketType       `json:"ticket_type,omitempty"`
	Priority     *TicketPriority   `json:"priority,omitempty"`
	Status       *TicketStatus     `json:"status,omitempty"`
	AssigneeID   *string           `json:"assignee_id,omitempty" binding:"omitempty,uuid"`
	Resolution   *TicketResolution `json:"resolution,omitempty"`
	
	// Time Tracking
	EstimatedHours *float64 `json:"estimated_hours,omitempty" binding:"omitempty,gt=0"`
	ActualHours    *float64 `json:"actual_hours,omitempty" binding:"omitempty,gt=0"`
	
	// Metadata
	Labels       *models.JSONB `json:"labels,omitempty"`
	CustomFields *models.JSONB `json:"custom_fields,omitempty"`
}

type TicketResponse struct {
	ID           string           `json:"id"`
	TicketNumber int              `json:"ticket_number"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	TicketType   TicketType       `json:"ticket_type"`
	Priority     TicketPriority   `json:"priority"`
	Status       TicketStatus     `json:"status"`
	ProjectID    *string          `json:"project_id"`
	ReporterID   *string          `json:"reporter_id"`
	AssigneeID   *string          `json:"assignee_id"`
	CustomerEmail string          `json:"customer_email"`
	CustomerName string           `json:"customer_name"`
	Channel      TicketChannel    `json:"channel"`
	Resolution   TicketResolution `json:"resolution"`
	ResolvedAt   *time.Time       `json:"resolved_at"`
	ResolvedBy   *string          `json:"resolved_by"`
	EstimatedHours *float64       `json:"estimated_hours"`
	ActualHours    *float64       `json:"actual_hours"`
	Labels       models.JSONB     `json:"labels"`
	CustomFields models.JSONB     `json:"custom_fields"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// TableName overrides the table name used by Ticket to `tickets`
func (Ticket) TableName() string {
	return "tickets"
}

// ToResponse converts a Ticket model to TicketResponse
func (t *Ticket) ToResponse() TicketResponse {
	return TicketResponse{
		ID:           t.ID,
		TicketNumber: t.TicketNumber,
		Title:        t.Title,
		Description:  t.Description,
		TicketType:   t.TicketType,
		Priority:     t.Priority,
		Status:       t.Status,
		ProjectID:    t.ProjectID,
		ReporterID:   t.ReporterID,
		AssigneeID:   t.AssigneeID,
		CustomerEmail: t.CustomerEmail,
		CustomerName: t.CustomerName,
		Channel:      t.Channel,
		Resolution:   t.Resolution,
		ResolvedAt:   t.ResolvedAt,
		ResolvedBy:   t.ResolvedBy,
		EstimatedHours: t.EstimatedHours,
		ActualHours:    t.ActualHours,
		Labels:       t.Labels,
		CustomFields: t.CustomFields,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}

// IsOpen checks if the ticket is in an open state
func (t *Ticket) IsOpen() bool {
	return t.Status == StatusOpen || t.Status == StatusInProgress
}

// IsResolved checks if the ticket is resolved
func (t *Ticket) IsResolved() bool {
	return t.Status == StatusResolved || t.Status == StatusClosed
}