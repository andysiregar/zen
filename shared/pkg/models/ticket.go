package models

import (
	"time"
	"gorm.io/gorm"
)

type TicketStatus string
type TicketPriority string
type TicketType string

const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusResolved   TicketStatus = "resolved"
	TicketStatusClosed     TicketStatus = "closed"
	TicketStatusOnHold     TicketStatus = "on_hold"
)

const (
	TicketPriorityLow      TicketPriority = "low"
	TicketPriorityMedium   TicketPriority = "medium"
	TicketPriorityHigh     TicketPriority = "high"
	TicketPriorityCritical TicketPriority = "critical"
)

const (
	TicketTypeBug          TicketType = "bug"
	TicketTypeFeature      TicketType = "feature"
	TicketTypeSupport      TicketType = "support"
	TicketTypeQuestion     TicketType = "question"
	TicketTypeIncident     TicketType = "incident"
)

type Ticket struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TenantID    string         `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Title       string         `json:"title" gorm:"not null;size:500"`
	Description string         `json:"description" gorm:"type:text"`
	Status      TicketStatus   `json:"status" gorm:"type:varchar(20);default:'open'"`
	Priority    TicketPriority `json:"priority" gorm:"type:varchar(20);default:'medium'"`
	Type        TicketType     `json:"type" gorm:"type:varchar(20);default:'support'"`
	
	// User relationships
	ReporterID string `json:"reporter_id" gorm:"type:uuid;not null;index"` // User who created ticket
	AssigneeID string `json:"assignee_id" gorm:"type:uuid;index"`          // User assigned to ticket
	
	// Project relationship (optional)
	ProjectID string `json:"project_id" gorm:"type:uuid;index"`
	
	// Categorization
	Category string   `json:"category" gorm:"size:100"`
	Tags     JSONB    `json:"tags" gorm:"type:jsonb;default:'[]'"`
	
	// SLA and timing
	DueDate     *time.Time `json:"due_date"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	ClosedAt    *time.Time `json:"closed_at"`
	
	// Additional metadata
	Metadata    JSONB    `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	
	// Timestamps
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type TicketComment struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TicketID     string    `json:"ticket_id" gorm:"type:uuid;not null;index"`
	AuthorID     string    `json:"author_id" gorm:"type:uuid;not null;index"`
	Content      string    `json:"content" gorm:"type:text;not null"`
	IsInternal   bool      `json:"is_internal" gorm:"default:false"` // Internal notes vs public comments
	Attachments  JSONB     `json:"attachments" gorm:"type:jsonb;default:'[]'"`
	
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type TicketAttachment struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TicketID     string    `json:"ticket_id" gorm:"type:uuid;not null;index"`
	CommentID    string    `json:"comment_id" gorm:"type:uuid;index"` // Optional - can be attached to ticket or comment
	UploadedBy   string    `json:"uploaded_by" gorm:"type:uuid;not null"`
	FileName     string    `json:"file_name" gorm:"not null;size:500"`
	FileSize     int64     `json:"file_size"`
	ContentType  string    `json:"content_type" gorm:"size:100"`
	FilePath     string    `json:"file_path" gorm:"not null;size:1000"`
	
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// Request/Response types
type TicketCreateRequest struct {
	Title       string         `json:"title" binding:"required,min=3,max=500"`
	Description string         `json:"description" binding:"required,min=10"`
	Priority    TicketPriority `json:"priority,omitempty"`
	Type        TicketType     `json:"type,omitempty"`
	Category    string         `json:"category,omitempty"`
	ProjectID   string         `json:"project_id,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	DueDate     *time.Time     `json:"due_date,omitempty"`
}

type TicketUpdateRequest struct {
	Title       *string         `json:"title,omitempty" binding:"omitempty,min=3,max=500"`
	Description *string         `json:"description,omitempty" binding:"omitempty,min=10"`
	Status      *TicketStatus   `json:"status,omitempty"`
	Priority    *TicketPriority `json:"priority,omitempty"`
	Type        *TicketType     `json:"type,omitempty"`
	Category    *string         `json:"category,omitempty"`
	AssigneeID  *string         `json:"assignee_id,omitempty"`
	ProjectID   *string         `json:"project_id,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	DueDate     *time.Time      `json:"due_date,omitempty"`
}

type TicketCommentCreateRequest struct {
	Content    string   `json:"content" binding:"required,min=1"`
	IsInternal bool     `json:"is_internal,omitempty"`
}

type TicketResponse struct {
	ID          string         `json:"id"`
	TenantID    string         `json:"tenant_id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      TicketStatus   `json:"status"`
	Priority    TicketPriority `json:"priority"`
	Type        TicketType     `json:"type"`
	ReporterID  string         `json:"reporter_id"`
	AssigneeID  string         `json:"assignee_id"`
	ProjectID   string         `json:"project_id"`
	Category    string         `json:"category"`
	Tags        []string       `json:"tags"`
	DueDate     *time.Time     `json:"due_date"`
	ResolvedAt  *time.Time     `json:"resolved_at"`
	ClosedAt    *time.Time     `json:"closed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type TicketCommentResponse struct {
	ID          string    `json:"id"`
	TicketID    string    `json:"ticket_id"`
	AuthorID    string    `json:"author_id"`
	Content     string    `json:"content"`
	IsInternal  bool      `json:"is_internal"`
	Attachments []string  `json:"attachments"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Table names
func (Ticket) TableName() string {
	return "tickets"
}

func (TicketComment) TableName() string {
	return "ticket_comments"
}

func (TicketAttachment) TableName() string {
	return "ticket_attachments"
}

// ToResponse converts Ticket to TicketResponse
func (t *Ticket) ToResponse() TicketResponse {
	var tags []string
	if t.Tags != nil {
		// Convert JSONB to string slice
		// In production, you'd parse the JSON properly
		tags = []string{} // TODO: Parse JSONB tags
	}
	
	return TicketResponse{
		ID:          t.ID,
		TenantID:    t.TenantID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		Type:        t.Type,
		ReporterID:  t.ReporterID,
		AssigneeID:  t.AssigneeID,
		ProjectID:   t.ProjectID,
		Category:    t.Category,
		Tags:        tags,
		DueDate:     t.DueDate,
		ResolvedAt:  t.ResolvedAt,
		ClosedAt:    t.ClosedAt,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// ToResponse converts TicketComment to TicketCommentResponse
func (tc *TicketComment) ToResponse() TicketCommentResponse {
	var attachments []string
	if tc.Attachments != nil {
		// Convert JSONB to string slice
		// In production, you'd parse the JSON properly
		attachments = []string{} // TODO: Parse JSONB attachments
	}
	
	return TicketCommentResponse{
		ID:          tc.ID,
		TicketID:    tc.TicketID,
		AuthorID:    tc.AuthorID,
		Content:     tc.Content,
		IsInternal:  tc.IsInternal,
		Attachments: attachments,
		CreatedAt:   tc.CreatedAt,
		UpdatedAt:   tc.UpdatedAt,
	}
}