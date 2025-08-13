package models

import (
	"time"
	"gorm.io/gorm"
)

// PlatformRole defines internal company roles for platform administration
type PlatformRole string

// PlatformPermission defines granular permissions for platform operations
type PlatformPermission string

const (
	// Platform Roles (hierarchical)
	PlatformRoleSuperAdmin    PlatformRole = "platform_super_admin"    // Can do everything
	PlatformRoleAdmin         PlatformRole = "platform_admin"          // Full access except user management
	PlatformRoleCustomerOps   PlatformRole = "platform_customer_ops"   // Customer org/tenant management
	PlatformRoleBillingOps    PlatformRole = "platform_billing_ops"    // Billing and subscription management
	PlatformRoleSupportAgent  PlatformRole = "platform_support_agent"  // Ticket viewing and customer support
	PlatformRoleAnalyst       PlatformRole = "platform_analyst"        // Read-only analytics and reporting
	PlatformRoleReadOnly      PlatformRole = "platform_readonly"       // View-only access
)

const (
	// Organization Management Permissions
	PermOrgView     PlatformPermission = "org:view"     // View customer organizations
	PermOrgCreate   PlatformPermission = "org:create"   // Create new organizations
	PermOrgUpdate   PlatformPermission = "org:update"   // Update organization details
	PermOrgDelete   PlatformPermission = "org:delete"   // Delete organizations
	PermOrgSuspend  PlatformPermission = "org:suspend"  // Suspend/unsuspend organizations

	// Tenant Management Permissions
	PermTenantView     PlatformPermission = "tenant:view"     // View tenant details
	PermTenantCreate   PlatformPermission = "tenant:create"   // Create new tenants
	PermTenantUpdate   PlatformPermission = "tenant:update"   // Update tenant settings
	PermTenantDelete   PlatformPermission = "tenant:delete"   // Delete tenants
	PermTenantSuspend  PlatformPermission = "tenant:suspend"  // Suspend/unsuspend tenants

	// User Management Permissions (cross-tenant)
	PermUserView     PlatformPermission = "user:view"     // View customer users
	PermUserCreate   PlatformPermission = "user:create"   // Create customer users
	PermUserUpdate   PlatformPermission = "user:update"   // Update customer user details
	PermUserDelete   PlatformPermission = "user:delete"   // Delete customer users
	PermUserSuspend  PlatformPermission = "user:suspend"  // Suspend/unsuspend customer users

	// Ticket Management Permissions (cross-tenant)
	PermTicketView     PlatformPermission = "ticket:view"     // View all customer tickets
	PermTicketUpdate   PlatformPermission = "ticket:update"   // Update ticket status/priority
	PermTicketAssign   PlatformPermission = "ticket:assign"   // Assign tickets to agents
	PermTicketClose    PlatformPermission = "ticket:close"    // Close customer tickets
	PermTicketDelete   PlatformPermission = "ticket:delete"   // Delete tickets

	// Billing Management Permissions
	PermBillingView     PlatformPermission = "billing:view"     // View billing information
	PermBillingUpdate   PlatformPermission = "billing:update"   // Update billing settings
	PermBillingInvoice  PlatformPermission = "billing:invoice"  // Generate invoices
	PermBillingRefund   PlatformPermission = "billing:refund"   // Process refunds
	PermBillingReport   PlatformPermission = "billing:report"   // Generate billing reports

	// Analytics & Reporting Permissions
	PermAnalyticsView    PlatformPermission = "analytics:view"    // View analytics dashboards
	PermAnalyticsExport  PlatformPermission = "analytics:export"  // Export analytics data
	PermReportView       PlatformPermission = "report:view"       // View reports
	PermReportGenerate   PlatformPermission = "report:generate"   // Generate custom reports

	// System Management Permissions
	PermSystemView     PlatformPermission = "system:view"     // View system status
	PermSystemUpdate   PlatformPermission = "system:update"   // Update system settings
	PermSystemBackup   PlatformPermission = "system:backup"   // Manage backups
	PermSystemLogs     PlatformPermission = "system:logs"     // View system logs

	// Platform Admin Management (internal)
	PermAdminView     PlatformPermission = "admin:view"     // View platform admin users
	PermAdminCreate   PlatformPermission = "admin:create"   // Create platform admin users
	PermAdminUpdate   PlatformPermission = "admin:update"   // Update platform admin users
	PermAdminDelete   PlatformPermission = "admin:delete"   // Delete platform admin users
)

// PlatformAdmin represents internal company users who manage the platform
type PlatformAdmin struct {
	ID          string        `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email       string        `json:"email" gorm:"uniqueIndex;not null;size:255"`
	Password    string        `json:"-" gorm:"not null;size:255"`
	FirstName   string        `json:"first_name" gorm:"not null;size:100"`
	LastName    string        `json:"last_name" gorm:"not null;size:100"`
	Role        PlatformRole  `json:"role" gorm:"type:varchar(30);not null"`
	Status      UserStatus    `json:"status" gorm:"type:varchar(20);default:'pending'"`
	
	// Profile
	Avatar      string        `json:"avatar" gorm:"size:500"`
	Phone       string        `json:"phone" gorm:"size:50"`
	
	// Authentication
	EmailVerified      bool      `json:"email_verified" gorm:"default:false"`
	EmailVerifiedAt    *time.Time `json:"email_verified_at"`
	PasswordResetToken string    `json:"-" gorm:"size:255"`
	PasswordResetAt    *time.Time `json:"-"`
	
	// Access control
	CustomPermissions  string    `json:"custom_permissions" gorm:"type:text"` // JSON array for role overrides
	LastLoginAt        *time.Time `json:"last_login_at"`
	LastLoginIP        string     `json:"last_login_ip" gorm:"size:45"`
	LoginAttempts      int        `json:"login_attempts" gorm:"default:0"`
	LockedAt           *time.Time `json:"locked_at"`
	
	// Audit trail
	CreatedBy     string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy     string        `json:"updated_by" gorm:"type:uuid"`
	
	// Timestamps
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// PlatformAdminCreateRequest for creating new platform admin users
type PlatformAdminCreateRequest struct {
	Email             string        `json:"email" binding:"required,email"`
	Password          string        `json:"password" binding:"required,min=8,max=128"`
	FirstName         string        `json:"first_name" binding:"required,min=2,max=100"`
	LastName          string        `json:"last_name" binding:"required,min=2,max=100"`
	Role              PlatformRole  `json:"role" binding:"required"`
	Phone             string        `json:"phone,omitempty"`
	CustomPermissions []PlatformPermission `json:"custom_permissions,omitempty"`
}

// PlatformAdminUpdateRequest for updating platform admin users
type PlatformAdminUpdateRequest struct {
	FirstName         *string       `json:"first_name,omitempty" binding:"omitempty,min=2,max=100"`
	LastName          *string       `json:"last_name,omitempty" binding:"omitempty,min=2,max=100"`
	Phone             *string       `json:"phone,omitempty"`
	Avatar            *string       `json:"avatar,omitempty"`
	Role              *PlatformRole `json:"role,omitempty"`
	Status            *UserStatus   `json:"status,omitempty"`
	CustomPermissions []PlatformPermission `json:"custom_permissions,omitempty"`
}

// PlatformAdminResponse for API responses
type PlatformAdminResponse struct {
	ID                string        `json:"id"`
	Email             string        `json:"email"`
	FirstName         string        `json:"first_name"`
	LastName          string        `json:"last_name"`
	Role              PlatformRole  `json:"role"`
	Status            UserStatus    `json:"status"`
	Avatar            string        `json:"avatar"`
	Phone             string        `json:"phone"`
	EmailVerified     bool          `json:"email_verified"`
	EmailVerifiedAt   *time.Time    `json:"email_verified_at"`
	LastLoginAt       *time.Time    `json:"last_login_at"`
	Permissions       []PlatformPermission `json:"permissions"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

// PlatformAdminLoginRequest for platform admin authentication
type PlatformAdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// PlatformAuthResponse for platform admin login
type PlatformAuthResponse struct {
	Admin        PlatformAdminResponse `json:"admin"`
	AccessToken  string                `json:"access_token"`
	RefreshToken string                `json:"refresh_token"`
	ExpiresIn    int                   `json:"expires_in"`
}

// TableName overrides the table name
func (PlatformAdmin) TableName() string {
	return "platform_admins"
}

// ToResponse converts PlatformAdmin to response format
func (pa *PlatformAdmin) ToResponse() PlatformAdminResponse {
	return PlatformAdminResponse{
		ID:              pa.ID,
		Email:           pa.Email,
		FirstName:       pa.FirstName,
		LastName:        pa.LastName,
		Role:            pa.Role,
		Status:          pa.Status,
		Avatar:          pa.Avatar,
		Phone:           pa.Phone,
		EmailVerified:   pa.EmailVerified,
		EmailVerifiedAt: pa.EmailVerifiedAt,
		LastLoginAt:     pa.LastLoginAt,
		Permissions:     pa.GetAllPermissions(),
		CreatedAt:       pa.CreatedAt,
		UpdatedAt:       pa.UpdatedAt,
	}
}

// FullName returns the admin's full name
func (pa *PlatformAdmin) FullName() string {
	return pa.FirstName + " " + pa.LastName
}

// IsActive checks if the admin account is active
func (pa *PlatformAdmin) IsActive() bool {
	return pa.Status == StatusActive
}

// IsLocked checks if the admin account is locked
func (pa *PlatformAdmin) IsLocked() bool {
	return pa.LockedAt != nil && time.Since(*pa.LockedAt) < 30*time.Minute
}

// GetRolePermissions returns default permissions for the admin's role
func (pa *PlatformAdmin) GetRolePermissions() []PlatformPermission {
	return GetRolePermissions(pa.Role)
}

// GetAllPermissions returns all permissions (role + custom)
func (pa *PlatformAdmin) GetAllPermissions() []PlatformPermission {
	rolePerms := pa.GetRolePermissions()
	
	// TODO: Parse CustomPermissions JSON and merge with role permissions
	// For now, just return role permissions
	return rolePerms
}

// HasPermission checks if admin has specific permission
func (pa *PlatformAdmin) HasPermission(permission PlatformPermission) bool {
	permissions := pa.GetAllPermissions()
	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// GetRolePermissions returns default permissions for a given role
func GetRolePermissions(role PlatformRole) []PlatformPermission {
	switch role {
	case PlatformRoleSuperAdmin:
		return []PlatformPermission{
			// All permissions - super admin can do everything
			PermOrgView, PermOrgCreate, PermOrgUpdate, PermOrgDelete, PermOrgSuspend,
			PermTenantView, PermTenantCreate, PermTenantUpdate, PermTenantDelete, PermTenantSuspend,
			PermUserView, PermUserCreate, PermUserUpdate, PermUserDelete, PermUserSuspend,
			PermTicketView, PermTicketUpdate, PermTicketAssign, PermTicketClose, PermTicketDelete,
			PermBillingView, PermBillingUpdate, PermBillingInvoice, PermBillingRefund, PermBillingReport,
			PermAnalyticsView, PermAnalyticsExport, PermReportView, PermReportGenerate,
			PermSystemView, PermSystemUpdate, PermSystemBackup, PermSystemLogs,
			PermAdminView, PermAdminCreate, PermAdminUpdate, PermAdminDelete,
		}
	
	case PlatformRoleAdmin:
		return []PlatformPermission{
			// Full access except admin user management
			PermOrgView, PermOrgCreate, PermOrgUpdate, PermOrgDelete, PermOrgSuspend,
			PermTenantView, PermTenantCreate, PermTenantUpdate, PermTenantDelete, PermTenantSuspend,
			PermUserView, PermUserCreate, PermUserUpdate, PermUserDelete, PermUserSuspend,
			PermTicketView, PermTicketUpdate, PermTicketAssign, PermTicketClose, PermTicketDelete,
			PermBillingView, PermBillingUpdate, PermBillingInvoice, PermBillingRefund, PermBillingReport,
			PermAnalyticsView, PermAnalyticsExport, PermReportView, PermReportGenerate,
			PermSystemView, PermSystemUpdate, PermSystemBackup, PermSystemLogs,
			PermAdminView,
		}
	
	case PlatformRoleCustomerOps:
		return []PlatformPermission{
			// Customer organization and tenant management
			PermOrgView, PermOrgCreate, PermOrgUpdate, PermOrgSuspend,
			PermTenantView, PermTenantCreate, PermTenantUpdate, PermTenantSuspend,
			PermUserView, PermUserCreate, PermUserUpdate, PermUserSuspend,
			PermAnalyticsView, PermReportView,
		}
	
	case PlatformRoleBillingOps:
		return []PlatformPermission{
			// Billing and subscription management
			PermOrgView, PermTenantView, PermUserView,
			PermBillingView, PermBillingUpdate, PermBillingInvoice, PermBillingRefund, PermBillingReport,
			PermAnalyticsView, PermReportView,
		}
	
	case PlatformRoleSupportAgent:
		return []PlatformPermission{
			// Ticket viewing and customer support
			PermOrgView, PermTenantView, PermUserView,
			PermTicketView, PermTicketUpdate, PermTicketAssign, PermTicketClose,
			PermAnalyticsView, PermReportView,
		}
	
	case PlatformRoleAnalyst:
		return []PlatformPermission{
			// Read-only analytics and reporting
			PermOrgView, PermTenantView, PermUserView, PermTicketView, PermBillingView,
			PermAnalyticsView, PermAnalyticsExport, PermReportView, PermReportGenerate,
		}
	
	case PlatformRoleReadOnly:
		return []PlatformPermission{
			// View-only access
			PermOrgView, PermTenantView, PermUserView, PermTicketView, PermBillingView,
			PermAnalyticsView, PermReportView,
		}
	
	default:
		return []PlatformPermission{}
	}
}