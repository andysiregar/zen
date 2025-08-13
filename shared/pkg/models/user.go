package models

import (
	"time"
	"gorm.io/gorm"
)

type UserRole string
type UserStatus string

const (
	RoleSuperAdmin UserRole = "super_admin"
	RoleAdmin      UserRole = "admin"
	RoleUser       UserRole = "user"
	RoleGuest      UserRole = "guest"
)

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
	StatusPending   UserStatus = "pending"
)

type User struct {
	ID             string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrganizationID string     `json:"organization_id" gorm:"type:uuid;not null;index"`
	Email          string     `json:"email" gorm:"uniqueIndex;not null;size:255"`
	Password       string     `json:"-" gorm:"not null;size:255"`
	FirstName      string     `json:"first_name" gorm:"not null;size:100"`
	LastName       string     `json:"last_name" gorm:"not null;size:100"`
	Role           UserRole   `json:"role" gorm:"type:varchar(20);default:'user'"`
	Status         UserStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	
	// Profile information
	Avatar       string    `json:"avatar" gorm:"size:500"`
	Phone        string    `json:"phone" gorm:"size:50"`
	Timezone     string    `json:"timezone" gorm:"size:50;default:'UTC'"`
	Language     string    `json:"language" gorm:"size:10;default:'en'"`
	
	// Authentication
	EmailVerified      bool      `json:"email_verified" gorm:"default:false"`
	EmailVerifiedAt    *time.Time `json:"email_verified_at"`
	PasswordResetToken string    `json:"-" gorm:"size:255"`
	PasswordResetAt    *time.Time `json:"-"`
	
	// Access control
	Permissions   string `json:"permissions" gorm:"type:text"` // JSON array of permissions
	LastLoginAt   *time.Time `json:"last_login_at"`
	LastLoginIP   string     `json:"last_login_ip" gorm:"size:45"`
	LoginAttempts int        `json:"login_attempts" gorm:"default:0"`
	LockedAt      *time.Time `json:"locked_at"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type UserCreateRequest struct {
	OrganizationID string   `json:"organization_id" binding:"required,uuid"`
	Email          string   `json:"email" binding:"required,email"`
	Password       string   `json:"password" binding:"required,min=8,max=128"`
	FirstName      string   `json:"first_name" binding:"required,min=2,max=100"`
	LastName       string   `json:"last_name" binding:"required,min=2,max=100"`
	Role           UserRole `json:"role,omitempty"`
	Phone          string   `json:"phone,omitempty"`
	Timezone       string   `json:"timezone,omitempty"`
	Language       string   `json:"language,omitempty"`
}

type UserUpdateRequest struct {
	FirstName *string    `json:"first_name,omitempty" binding:"omitempty,min=2,max=100"`
	LastName  *string    `json:"last_name,omitempty" binding:"omitempty,min=2,max=100"`
	Phone     *string    `json:"phone,omitempty"`
	Avatar    *string    `json:"avatar,omitempty"`
	Timezone  *string    `json:"timezone,omitempty"`
	Language  *string    `json:"language,omitempty"`
	Role      *UserRole  `json:"role,omitempty"`
	Status    *UserStatus `json:"status,omitempty"`
}

type UserLoginRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required"`
	OrganizationID string `json:"organization_id,omitempty"`
}

type UserResponse struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	Email          string     `json:"email"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Role           UserRole   `json:"role"`
	Status         UserStatus `json:"status"`
	Avatar         string     `json:"avatar"`
	Phone          string     `json:"phone"`
	Timezone       string     `json:"timezone"`
	Language       string     `json:"language"`
	EmailVerified  bool       `json:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	LastLoginAt    *time.Time `json:"last_login_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"` // seconds
}

// TableName overrides the table name used by User to `users`
func (User) TableName() string {
	return "users"
}

// ToResponse converts a User model to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:             u.ID,
		OrganizationID: u.OrganizationID,
		Email:          u.Email,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		Role:           u.Role,
		Status:         u.Status,
		Avatar:         u.Avatar,
		Phone:          u.Phone,
		Timezone:       u.Timezone,
		Language:       u.Language,
		EmailVerified:  u.EmailVerified,
		EmailVerifiedAt: u.EmailVerifiedAt,
		LastLoginAt:    u.LastLoginAt,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

// FullName returns the user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// IsLocked checks if the user account is locked
func (u *User) IsLocked() bool {
	return u.LockedAt != nil && time.Since(*u.LockedAt) < 30*time.Minute
}