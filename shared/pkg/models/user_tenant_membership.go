package models

import (
	"time"
	"gorm.io/gorm"
)

type MembershipStatus string
type MembershipRole string

const (
	MembershipStatusActive    MembershipStatus = "active"
	MembershipStatusSuspended MembershipStatus = "suspended" 
	MembershipStatusPending   MembershipStatus = "pending"
)

const (
	MembershipRoleOwner   MembershipRole = "Owner"
	MembershipRoleAdmin   MembershipRole = "Admin"
	MembershipRoleManager MembershipRole = "Manager"
	MembershipRoleAgent   MembershipRole = "Agent"
	MembershipRoleViewer  MembershipRole = "Viewer"
)

type UserTenantMembership struct {
	ID       string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID   string `json:"user_id" gorm:"type:uuid;not null;index"`
	TenantID string `json:"tenant_id" gorm:"type:uuid;not null;index"`
	
	// Role & Permissions
	Role        MembershipRole `json:"role" gorm:"type:varchar(50);not null"`
	Permissions JSONB          `json:"permissions" gorm:"type:jsonb;default:'[]'"` // Additional granular permissions
	
	// Status
	Status    MembershipStatus `json:"status" gorm:"type:varchar(20);default:'active'"`
	InvitedBy *string          `json:"invited_by" gorm:"type:uuid"`
	InvitedAt time.Time        `json:"invited_at" gorm:"default:now()"`
	JoinedAt  *time.Time       `json:"joined_at"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type UserTenantMembershipCreateRequest struct {
	UserID   string         `json:"user_id" binding:"required,uuid"`
	TenantID string         `json:"tenant_id" binding:"required,uuid"`
	Role     MembershipRole `json:"role" binding:"required"`
}

type UserTenantMembershipUpdateRequest struct {
	Role        *MembershipRole   `json:"role,omitempty"`
	Status      *MembershipStatus `json:"status,omitempty"`
	Permissions *JSONB            `json:"permissions,omitempty"`
}

type UserTenantMembershipResponse struct {
	ID          string           `json:"id"`
	UserID      string           `json:"user_id"`
	TenantID    string           `json:"tenant_id"`
	Role        MembershipRole   `json:"role"`
	Permissions JSONB            `json:"permissions"`
	Status      MembershipStatus `json:"status"`
	InvitedBy   *string          `json:"invited_by"`
	InvitedAt   time.Time        `json:"invited_at"`
	JoinedAt    *time.Time       `json:"joined_at"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// TableName overrides the table name used by UserTenantMembership
func (UserTenantMembership) TableName() string {
	return "user_tenant_memberships"
}

// ToResponse converts a UserTenantMembership model to UserTenantMembershipResponse
func (utm *UserTenantMembership) ToResponse() UserTenantMembershipResponse {
	return UserTenantMembershipResponse{
		ID:          utm.ID,
		UserID:      utm.UserID,
		TenantID:    utm.TenantID,
		Role:        utm.Role,
		Permissions: utm.Permissions,
		Status:      utm.Status,
		InvitedBy:   utm.InvitedBy,
		InvitedAt:   utm.InvitedAt,
		JoinedAt:    utm.JoinedAt,
		CreatedAt:   utm.CreatedAt,
		UpdatedAt:   utm.UpdatedAt,
	}
}

// IsActive checks if the membership is active
func (utm *UserTenantMembership) IsActive() bool {
	return utm.Status == MembershipStatusActive
}

// HasRole checks if the membership has a specific role or higher
func (utm *UserTenantMembership) HasRole(role MembershipRole) bool {
	roles := map[MembershipRole]int{
		MembershipRoleViewer:  1,
		MembershipRoleAgent:   2,
		MembershipRoleManager: 3,
		MembershipRoleAdmin:   4,
		MembershipRoleOwner:   5,
	}
	
	userLevel, exists := roles[utm.Role]
	if !exists {
		return false
	}
	
	requiredLevel, exists := roles[role]
	if !exists {
		return false
	}
	
	return userLevel >= requiredLevel
}