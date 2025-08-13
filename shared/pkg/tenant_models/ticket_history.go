package tenant_models

import (
	"time"
)

type ChangeType string

const (
	ChangeTypeCreate     ChangeType = "create"
	ChangeTypeUpdate     ChangeType = "update"
	ChangeTypeDelete     ChangeType = "delete"
	ChangeTypeAssign     ChangeType = "assign"
	ChangeTypeResolve    ChangeType = "resolve"
	ChangeTypeReopen     ChangeType = "reopen"
	ChangeTypeComment    ChangeType = "comment"
	ChangeTypeAttachment ChangeType = "attachment"
)

type TicketHistory struct {
	ID       string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TicketID string `json:"ticket_id" gorm:"type:uuid;not null;index"`
	
	// Change Details
	FieldName  string     `json:"field_name" gorm:"not null;size:100"`  // status, assignee, priority, etc.
	OldValue   *string    `json:"old_value" gorm:"type:text"`
	NewValue   *string    `json:"new_value" gorm:"type:text"`
	ChangeType ChangeType `json:"change_type" gorm:"type:varchar(50);not null"`
	
	// Actor (References Master DB users.id)
	ChangedBy string `json:"changed_by" gorm:"type:uuid;not null"`
	
	// Additional context
	Comment *string `json:"comment" gorm:"type:text"` // Optional comment about the change
	
	// Timestamp
	ChangedAt time.Time `json:"changed_at" gorm:"default:now()"`
}

type TicketHistoryCreateRequest struct {
	TicketID   string     `json:"ticket_id" binding:"required,uuid"`
	FieldName  string     `json:"field_name" binding:"required,min=1,max=100"`
	OldValue   *string    `json:"old_value,omitempty"`
	NewValue   *string    `json:"new_value,omitempty"`
	ChangeType ChangeType `json:"change_type" binding:"required"`
	Comment    *string    `json:"comment,omitempty"`
}

type TicketHistoryResponse struct {
	ID         string     `json:"id"`
	TicketID   string     `json:"ticket_id"`
	FieldName  string     `json:"field_name"`
	OldValue   *string    `json:"old_value"`
	NewValue   *string    `json:"new_value"`
	ChangeType ChangeType `json:"change_type"`
	ChangedBy  string     `json:"changed_by"`
	Comment    *string    `json:"comment"`
	ChangedAt  time.Time  `json:"changed_at"`
}

// TableName overrides the table name used by TicketHistory to `ticket_history`
func (TicketHistory) TableName() string {
	return "ticket_history"
}

// ToResponse converts a TicketHistory model to TicketHistoryResponse
func (th *TicketHistory) ToResponse() TicketHistoryResponse {
	return TicketHistoryResponse{
		ID:         th.ID,
		TicketID:   th.TicketID,
		FieldName:  th.FieldName,
		OldValue:   th.OldValue,
		NewValue:   th.NewValue,
		ChangeType: th.ChangeType,
		ChangedBy:  th.ChangedBy,
		Comment:    th.Comment,
		ChangedAt:  th.ChangedAt,
	}
}

// IsFieldChange checks if this history entry represents a field change
func (th *TicketHistory) IsFieldChange() bool {
	return th.ChangeType == ChangeTypeUpdate
}

// IsStatusChange checks if this history entry represents a status change
func (th *TicketHistory) IsStatusChange() bool {
	return th.FieldName == "status" && th.ChangeType == ChangeTypeUpdate
}