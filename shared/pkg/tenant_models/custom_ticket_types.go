package tenant_models

import (
	"time"
	"gorm.io/gorm"
)

type CustomTicketType struct {
	ID          string  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string  `json:"name" gorm:"uniqueIndex;not null;size:100"`
	Description *string `json:"description" gorm:"type:text"`
	Icon        *string `json:"icon" gorm:"size:50"`  // Icon identifier
	Color       *string `json:"color" gorm:"size:7"`  // Hex color code
	IsActive    bool    `json:"is_active" gorm:"default:true"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type CustomTicketTypeCreateRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty" binding:"omitempty,max=50"`
	Color       *string `json:"color,omitempty" binding:"omitempty,len=7"` // Must be 7 chars for hex
}

type CustomTicketTypeUpdateRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty" binding:"omitempty,max=50"`
	Color       *string `json:"color,omitempty" binding:"omitempty,len=7"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type CustomTicketTypeResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Icon        *string   `json:"icon"`
	Color       *string   `json:"color"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName overrides the table name used by CustomTicketType to `custom_ticket_types`
func (CustomTicketType) TableName() string {
	return "custom_ticket_types"
}

// ToResponse converts a CustomTicketType model to CustomTicketTypeResponse
func (ctt *CustomTicketType) ToResponse() CustomTicketTypeResponse {
	return CustomTicketTypeResponse{
		ID:          ctt.ID,
		Name:        ctt.Name,
		Description: ctt.Description,
		Icon:        ctt.Icon,
		Color:       ctt.Color,
		IsActive:    ctt.IsActive,
		CreatedAt:   ctt.CreatedAt,
	}
}