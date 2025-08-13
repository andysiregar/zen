package tenant_models

import (
	"time"
	"gorm.io/gorm"
)

type ProjectMemberRole string

const (
	ProjectRoleLead   ProjectMemberRole = "lead"
	ProjectRoleMember ProjectMemberRole = "member"
	ProjectRoleViewer ProjectMemberRole = "viewer"
)

type ProjectMember struct {
	ID        string            `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ProjectID string            `json:"project_id" gorm:"type:uuid;not null;index"`
	UserID    string            `json:"user_id" gorm:"type:uuid;not null"` // FK to master.users.id
	
	// Role within project
	Role ProjectMemberRole `json:"role" gorm:"type:varchar(50);default:'member'"`
	
	// Timestamps
	AddedAt   time.Time      `json:"added_at" gorm:"default:now()"`
	AddedBy   string         `json:"added_by" gorm:"type:uuid;not null"` // FK to master.users.id
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type ProjectMemberCreateRequest struct {
	ProjectID string            `json:"project_id" binding:"required,uuid"`
	UserID    string            `json:"user_id" binding:"required,uuid"`
	Role      ProjectMemberRole `json:"role,omitempty"`
}

type ProjectMemberUpdateRequest struct {
	Role *ProjectMemberRole `json:"role,omitempty"`
}

type ProjectMemberResponse struct {
	ID        string            `json:"id"`
	ProjectID string            `json:"project_id"`
	UserID    string            `json:"user_id"`
	Role      ProjectMemberRole `json:"role"`
	AddedAt   time.Time         `json:"added_at"`
	AddedBy   string            `json:"added_by"`
}

// TableName overrides the table name used by ProjectMember to `project_members`
func (ProjectMember) TableName() string {
	return "project_members"
}

// ToResponse converts a ProjectMember model to ProjectMemberResponse
func (pm *ProjectMember) ToResponse() ProjectMemberResponse {
	return ProjectMemberResponse{
		ID:        pm.ID,
		ProjectID: pm.ProjectID,
		UserID:    pm.UserID,
		Role:      pm.Role,
		AddedAt:   pm.AddedAt,
		AddedBy:   pm.AddedBy,
	}
}