package models

import (
	"time"
	"gorm.io/gorm"
)

type Organization struct {
	ID   string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name string `json:"name" gorm:"not null;size:255"`
	Slug string `json:"slug" gorm:"uniqueIndex;not null;size:100"` // URL-friendly identifier
	
	// Domain configuration
	Domain string `json:"domain" gorm:"uniqueIndex;size:255"` // Custom domain (optional)
	
	// Billing & Subscription
	SubscriptionPlan   string `json:"subscription_plan" gorm:"size:50;default:'basic'"`
	SubscriptionStatus string `json:"subscription_status" gorm:"size:20;default:'active'"`
	BillingEmail       string `json:"billing_email" gorm:"not null;size:255"`
	
	// Limits & Configuration
	MaxTenants         int `json:"max_tenants" gorm:"default:5"`
	MaxUsersPerTenant  int `json:"max_users_per_tenant" gorm:"default:50"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type OrganizationCreateRequest struct {
	Name         string `json:"name" binding:"required,min=2,max=255"`
	Slug         string `json:"slug" binding:"required,min=2,max=100,alphanum"`
	BillingEmail string `json:"billing_email" binding:"required,email"`
	Domain       string `json:"domain,omitempty"`
}

type OrganizationUpdateRequest struct {
	Name                *string `json:"name,omitempty" binding:"omitempty,min=2,max=255"`
	Domain              *string `json:"domain,omitempty" binding:"omitempty,min=3,max=255"`
	BillingEmail        *string `json:"billing_email,omitempty" binding:"omitempty,email"`
	SubscriptionPlan    *string `json:"subscription_plan,omitempty"`
	SubscriptionStatus  *string `json:"subscription_status,omitempty"`
	MaxTenants          *int    `json:"max_tenants,omitempty" binding:"omitempty,min=1"`
	MaxUsersPerTenant   *int    `json:"max_users_per_tenant,omitempty" binding:"omitempty,min=1"`
}

type OrganizationResponse struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Slug               string    `json:"slug"`
	Domain             string    `json:"domain"`
	SubscriptionPlan   string    `json:"subscription_plan"`
	SubscriptionStatus string    `json:"subscription_status"`
	BillingEmail       string    `json:"billing_email"`
	MaxTenants         int       `json:"max_tenants"`
	MaxUsersPerTenant  int       `json:"max_users_per_tenant"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// TableName overrides the table name used by Organization to `organizations`
func (Organization) TableName() string {
	return "organizations"
}

// ToResponse converts an Organization model to OrganizationResponse
func (o *Organization) ToResponse() OrganizationResponse {
	return OrganizationResponse{
		ID:                 o.ID,
		Name:               o.Name,
		Slug:               o.Slug,
		Domain:             o.Domain,
		SubscriptionPlan:   o.SubscriptionPlan,
		SubscriptionStatus: o.SubscriptionStatus,
		BillingEmail:       o.BillingEmail,
		MaxTenants:         o.MaxTenants,
		MaxUsersPerTenant:  o.MaxUsersPerTenant,
		CreatedAt:          o.CreatedAt,
		UpdatedAt:          o.UpdatedAt,
	}
}