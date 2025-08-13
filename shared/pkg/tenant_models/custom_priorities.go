package tenant_models

import (
	"time"
	"gorm.io/gorm"
)

type CustomPriority struct {
	ID       string  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name     string  `json:"name" gorm:"uniqueIndex;not null;size:100"`
	Level    int     `json:"level" gorm:"uniqueIndex;not null"` // Numeric level for sorting
	Color    *string `json:"color" gorm:"size:7"`              // Hex color code
	IsActive bool    `json:"is_active" gorm:"default:true"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type CustomPriorityCreateRequest struct {
	Name  string  `json:"name" binding:"required,min=1,max=100"`
	Level int     `json:"level" binding:"required,min=1"`
	Color *string `json:"color,omitempty" binding:"omitempty,len=7"` // Must be 7 chars for hex
}

type CustomPriorityUpdateRequest struct {
	Name     *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Level    *int    `json:"level,omitempty" binding:"omitempty,min=1"`
	Color    *string `json:"color,omitempty" binding:"omitempty,len=7"`
	IsActive *bool   `json:"is_active,omitempty"`
}

type CustomPriorityResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Level     int       `json:"level"`
	Color     *string   `json:"color"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName overrides the table name used by CustomPriority to `custom_priorities`
func (CustomPriority) TableName() string {
	return "custom_priorities"
}

// ToResponse converts a CustomPriority model to CustomPriorityResponse
func (cp *CustomPriority) ToResponse() CustomPriorityResponse {
	return CustomPriorityResponse{
		ID:        cp.ID,
		Name:      cp.Name,
		Level:     cp.Level,
		Color:     cp.Color,
		IsActive:  cp.IsActive,
		CreatedAt: cp.CreatedAt,
	}
}