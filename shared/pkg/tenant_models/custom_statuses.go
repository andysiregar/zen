package tenant_models

import (
	"time"
	"gorm.io/gorm"
)

type StatusCategory string

const (
	StatusCategoryOpen       StatusCategory = "open"
	StatusCategoryInProgress StatusCategory = "in_progress" 
	StatusCategoryResolved   StatusCategory = "resolved"
	StatusCategoryClosed     StatusCategory = "closed"
)

type CustomStatus struct {
	ID        string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"uniqueIndex;not null;size:100"`
	Category  StatusCategory `json:"category" gorm:"type:varchar(50);not null"`
	Color     *string        `json:"color" gorm:"size:7"` // Hex color code
	SortOrder int            `json:"sort_order" gorm:"default:0"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type CustomStatusCreateRequest struct {
	Name      string         `json:"name" binding:"required,min=1,max=100"`
	Category  StatusCategory `json:"category" binding:"required"`
	Color     *string        `json:"color,omitempty" binding:"omitempty,len=7"` // Must be 7 chars for hex
	SortOrder int            `json:"sort_order,omitempty"`
}

type CustomStatusUpdateRequest struct {
	Name      *string         `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Category  *StatusCategory `json:"category,omitempty"`
	Color     *string         `json:"color,omitempty" binding:"omitempty,len=7"`
	SortOrder *int            `json:"sort_order,omitempty"`
	IsActive  *bool           `json:"is_active,omitempty"`
}

type CustomStatusResponse struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Category  StatusCategory `json:"category"`
	Color     *string        `json:"color"`
	SortOrder int            `json:"sort_order"`
	IsActive  bool           `json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
}

// TableName overrides the table name used by CustomStatus to `custom_statuses`
func (CustomStatus) TableName() string {
	return "custom_statuses"
}

// ToResponse converts a CustomStatus model to CustomStatusResponse
func (cs *CustomStatus) ToResponse() CustomStatusResponse {
	return CustomStatusResponse{
		ID:        cs.ID,
		Name:      cs.Name,
		Category:  cs.Category,
		Color:     cs.Color,
		SortOrder: cs.SortOrder,
		IsActive:  cs.IsActive,
		CreatedAt: cs.CreatedAt,
	}
}