package tenant_models

import (
	"time"
	"gorm.io/gorm"
)

type ProjectStatus string
type ProjectType string
type ProjectMethodology string

const (
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusArchived  ProjectStatus = "archived"
	ProjectStatusCompleted ProjectStatus = "completed"
)

const (
	ProjectTypeSoftware ProjectType = "software"
	ProjectTypeSupport  ProjectType = "support"
	ProjectTypeBusiness ProjectType = "business"
)

const (
	MethodologyKanban    ProjectMethodology = "kanban"
	MethodologyScrum     ProjectMethodology = "scrum"
	MethodologyWaterfall ProjectMethodology = "waterfall"
)

type Project struct {
	ID          string             `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Key         string             `json:"key" gorm:"uniqueIndex;not null;size:20"`        // e.g., 'PROJ', 'MOBILE'
	Name        string             `json:"name" gorm:"not null;size:255"`
	Description string             `json:"description" gorm:"type:text"`
	
	// Configuration
	ProjectType  ProjectType        `json:"project_type" gorm:"type:varchar(50);default:'software'"`
	Methodology  ProjectMethodology `json:"methodology" gorm:"type:varchar(50);default:'kanban'"`
	
	// Status & Workflow
	Status ProjectStatus `json:"status" gorm:"type:varchar(50);default:'active'"`
	
	// Ownership (References Master DB users.id)
	LeadID    string `json:"lead_id" gorm:"type:uuid;not null"`    // FK to master.users.id
	CreatedBy string `json:"created_by" gorm:"type:uuid;not null"` // FK to master.users.id
	
	// Timestamps
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	ArchivedAt *time.Time     `json:"archived_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

type ProjectCreateRequest struct {
	Key         string             `json:"key" binding:"required,min=2,max=20,alphanum"`
	Name        string             `json:"name" binding:"required,min=2,max=255"`
	Description string             `json:"description,omitempty"`
	ProjectType ProjectType        `json:"project_type,omitempty"`
	Methodology ProjectMethodology `json:"methodology,omitempty"`
	LeadID      string             `json:"lead_id" binding:"required,uuid"`
}

type ProjectUpdateRequest struct {
	Name        *string             `json:"name,omitempty" binding:"omitempty,min=2,max=255"`
	Description *string             `json:"description,omitempty"`
	ProjectType *ProjectType        `json:"project_type,omitempty"`
	Methodology *ProjectMethodology `json:"methodology,omitempty"`
	Status      *ProjectStatus      `json:"status,omitempty"`
	LeadID      *string             `json:"lead_id,omitempty" binding:"omitempty,uuid"`
}

type ProjectResponse struct {
	ID          string             `json:"id"`
	Key         string             `json:"key"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	ProjectType ProjectType        `json:"project_type"`
	Methodology ProjectMethodology `json:"methodology"`
	Status      ProjectStatus      `json:"status"`
	LeadID      string             `json:"lead_id"`
	CreatedBy   string             `json:"created_by"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	ArchivedAt  *time.Time         `json:"archived_at"`
}

// TableName overrides the table name used by Project to `projects`
func (Project) TableName() string {
	return "projects"
}

// ToResponse converts a Project model to ProjectResponse
func (p *Project) ToResponse() ProjectResponse {
	return ProjectResponse{
		ID:          p.ID,
		Key:         p.Key,
		Name:        p.Name,
		Description: p.Description,
		ProjectType: p.ProjectType,
		Methodology: p.Methodology,
		Status:      p.Status,
		LeadID:      p.LeadID,
		CreatedBy:   p.CreatedBy,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		ArchivedAt:  p.ArchivedAt,
	}
}

// IsActive checks if the project is active
func (p *Project) IsActive() bool {
	return p.Status == ProjectStatusActive
}