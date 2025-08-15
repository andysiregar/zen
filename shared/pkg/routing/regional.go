package routing

import (
	"fmt"
	"net/url"
	"strings"
	"github.com/zen/shared/pkg/models"
)

// RegionalRouter handles routing logic for multi-region deployments
type RegionalRouter struct {
	masterDomain string
}

// NewRegionalRouter creates a new regional router
func NewRegionalRouter(masterDomain string) *RegionalRouter {
	return &RegionalRouter{
		masterDomain: masterDomain,
	}
}

// GetRegionalEndpoint returns the regional API endpoint for a given region
func (r *RegionalRouter) GetRegionalEndpoint(region models.TenantRegion) string {
	switch region {
	case models.TenantRegionAsia:
		return "https://asia.chilldesk.io"
	case models.TenantRegionUS:
		return "https://us.chilldesk.io"
	case models.TenantRegionEurope:
		return "https://eu.chilldesk.io"
	default:
		return "https://us.chilldesk.io"
	}
}

// GetAuthRedirectURL creates the authentication URL for regional routing
// When user visits custom domain, they get redirected to master auth with return URL
func (r *RegionalRouter) GetAuthRedirectURL(customDomain, tenantSlug string, region models.TenantRegion) string {
	regionalEndpoint := r.GetRegionalEndpoint(region)
	
	// Build the return URL (where user goes after successful auth)
	returnURL := fmt.Sprintf("%s/tenant/%s/dashboard", regionalEndpoint, tenantSlug)
	
	// Build the auth URL (master domain auth with return URL)
	authURL := fmt.Sprintf("https://%s/auth/login?return_url=%s&custom_domain=%s", 
		r.masterDomain, 
		url.QueryEscape(returnURL),
		url.QueryEscape(customDomain))
	
	return authURL
}

// GetTenantPortalURL returns the full portal URL for a tenant after authentication
func (r *RegionalRouter) GetTenantPortalURL(tenant *models.Tenant) string {
	regionalEndpoint := r.GetRegionalEndpoint(tenant.Region)
	return fmt.Sprintf("%s/tenant/%s", regionalEndpoint, tenant.Slug)
}

// ParseCustomDomain extracts tenant information from a custom domain request
// This is used when a user visits their custom domain (e.g., mycorp.com)
func (r *RegionalRouter) ParseCustomDomain(host string) (isCustomDomain bool, domain string) {
	// Check if it's not one of our regional domains
	chillDeskDomains := []string{
		"chilldesk.io",
		"asia.chilldesk.io", 
		"us.chilldesk.io",
		"eu.chilldesk.io",
		"localhost",
	}
	
	hostLower := strings.ToLower(host)
	for _, domain := range chillDeskDomains {
		if strings.Contains(hostLower, domain) {
			return false, ""
		}
	}
	
	// If it's not our domain, it's a custom domain
	return true, host
}

// Regional deployment configuration
type RegionalConfig struct {
	Region     models.TenantRegion `json:"region"`
	APIGateway string              `json:"api_gateway"`
	Database   DatabaseConfig      `json:"database"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	SSLMode  string `json:"ssl_mode"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// GetRegionalConfig returns configuration for a specific region
func GetRegionalConfig(region models.TenantRegion) RegionalConfig {
	configs := map[models.TenantRegion]RegionalConfig{
		models.TenantRegionAsia: {
			Region:     models.TenantRegionAsia,
			APIGateway: "https://asia.chilldesk.io",
			Database: DatabaseConfig{
				Host:     "asia-db.chilldesk.internal",
				Port:     5432,
				SSLMode:  "require",
				Username: "tenant_user",
				Password: "encrypted_password",
			},
		},
		models.TenantRegionUS: {
			Region:     models.TenantRegionUS,
			APIGateway: "https://us.chilldesk.io",
			Database: DatabaseConfig{
				Host:     "us-db.chilldesk.internal",
				Port:     5432,
				SSLMode:  "require",
				Username: "tenant_user",
				Password: "encrypted_password",
			},
		},
		models.TenantRegionEurope: {
			Region:     models.TenantRegionEurope,
			APIGateway: "https://eu.chilldesk.io",
			Database: DatabaseConfig{
				Host:     "eu-db.chilldesk.internal",
				Port:     5432,
				SSLMode:  "require",
				Username: "tenant_user",
				Password: "encrypted_password",
			},
		},
	}
	
	if config, exists := configs[region]; exists {
		return config
	}
	
	// Default to US region
	return configs[models.TenantRegionUS]
}