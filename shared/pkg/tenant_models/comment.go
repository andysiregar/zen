package tenant_models

import (
	"time"
	"gorm.io/gorm"
)

type CommentType string

const (
	CommentTypeComment      CommentType = "comment"
	CommentTypeInternalNote CommentType = "internal_note"
	CommentTypeStatusChange CommentType = "status_change"
	CommentTypeSystemUpdate CommentType = "system_update"
)

type Comment struct {
	ID       string      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TicketID string      `json:"ticket_id" gorm:"type:uuid;not null;index"`
	
	// Content
	Body        string      `json:"body" gorm:"type:text;not null"`
	CommentType CommentType `json:"comment_type" gorm:"type:varchar(50);default:'comment'"`
	
	// Author (References Master DB users.id)
	AuthorID string `json:"author_id" gorm:"type:uuid;not null"`
	
	// Visibility
	IsInternal bool `json:"is_internal" gorm:"default:false"` // Internal notes vs public comments
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type CommentCreateRequest struct {
	TicketID    string      `json:"ticket_id" binding:"required,uuid"`
	Body        string      `json:"body" binding:"required,min=1"`
	CommentType CommentType `json:"comment_type,omitempty"`
	IsInternal  bool        `json:"is_internal,omitempty"`
}

type CommentUpdateRequest struct {
	Body       *string `json:"body,omitempty" binding:"omitempty,min=1"`
	IsInternal *bool   `json:"is_internal,omitempty"`
}

type CommentResponse struct {
	ID          string      `json:"id"`
	TicketID    string      `json:"ticket_id"`
	Body        string      `json:"body"`
	CommentType CommentType `json:"comment_type"`
	AuthorID    string      `json:"author_id"`
	IsInternal  bool        `json:"is_internal"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// TableName overrides the table name used by Comment to `comments`
func (Comment) TableName() string {
	return "comments"
}

// ToResponse converts a Comment model to CommentResponse
func (c *Comment) ToResponse() CommentResponse {
	return CommentResponse{
		ID:          c.ID,
		TicketID:    c.TicketID,
		Body:        c.Body,
		CommentType: c.CommentType,
		AuthorID:    c.AuthorID,
		IsInternal:  c.IsInternal,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// IsPublic checks if the comment is visible to external users
func (c *Comment) IsPublic() bool {
	return !c.IsInternal
}