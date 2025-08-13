package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zen/shared/pkg/auth"
	"github.com/zen/shared/pkg/database"
	"github.com/zen/shared/pkg/models"
	"github.com/zen/shared/pkg/utils"
	"gorm.io/gorm"
)

// TenantContext holds tenant-specific information
type TenantContext struct {
	TenantID         string
	OrganizationID   string
	UserID           string
	UserRole         models.MembershipRole
	MasterDB         *gorm.DB
	TenantDB         *gorm.DB
	TenantInfo       *models.Tenant
	MembershipInfo   *models.UserTenantMembership
}

// TenantMiddleware creates middleware for tenant database switching
func TenantMiddleware(dbManager *database.DatabaseManager, jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract tenant from subdomain or header
		tenantSlug := extractTenantSlug(c)
		if tenantSlug == "" {
			utils.BadRequestResponse(c, "Tenant not specified")
			c.Abort()
			return
		}

		// Get user info from JWT
		userID, exists := c.Get("user_id")
		if !exists {
			utils.UnauthorizedResponse(c, "User not authenticated")
			c.Abort()
			return
		}

		organizationID, exists := c.Get("organization_id")
		if !exists {
			utils.UnauthorizedResponse(c, "Organization not found")
			c.Abort()
			return
		}

		// Get master database connection
		masterDB := dbManager.GetMasterDB()

		// Find tenant by slug and organization
		var tenant models.Tenant
		err := masterDB.Where("slug = ? AND organization_id = ?", tenantSlug, organizationID).First(&tenant).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.NotFoundResponse(c, "Tenant not found")
			} else {
				utils.InternalServerErrorResponse(c, "Database error")
			}
			c.Abort()
			return
		}

		// Check if tenant is active
		if tenant.Status != models.TenantStatusActive {
			utils.ForbiddenResponse(c, "Tenant is not active")
			c.Abort()
			return
		}

		// Check user membership in this tenant
		var membership models.UserTenantMembership
		err = masterDB.Where("user_id = ? AND tenant_id = ?", userID, tenant.ID).First(&membership).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.ForbiddenResponse(c, "Access denied to this tenant")
			} else {
				utils.InternalServerErrorResponse(c, "Database error")
			}
			c.Abort()
			return
		}

		// Check if membership is active
		if !membership.IsActive() {
			utils.ForbiddenResponse(c, "Membership is not active")
			c.Abort()
			return
		}

		// Get tenant database connection
		tenantConnInfo := database.TenantConnectionInfo{
			TenantID:          tenant.ID,
			Host:              tenant.DbHost,
			Port:              tenant.DbPort,
			User:              tenant.DbUser,
			EncryptedPassword: tenant.DbPasswordEncrypted,
			DBName:            tenant.DbName,
			SSLMode:           tenant.DbSslMode,
		}

		tenantDB, err := dbManager.GetTenantDB(tenantConnInfo)
		if err != nil {
			utils.InternalServerErrorResponse(c, "Failed to connect to tenant database")
			c.Abort()
			return
		}

		// Create tenant context
		tenantContext := &TenantContext{
			TenantID:       tenant.ID,
			OrganizationID: tenant.OrganizationID,
			UserID:         userID.(string),
			UserRole:       membership.Role,
			MasterDB:       masterDB,
			TenantDB:       tenantDB,
			TenantInfo:     &tenant,
			MembershipInfo: &membership,
		}

		// Set context in Gin
		c.Set("tenant_context", tenantContext)
		c.Set("tenant_id", tenant.ID)
		c.Set("tenant_db", tenantDB)
		c.Set("master_db", masterDB)
		c.Set("user_role", membership.Role)

		c.Next()
	}
}

// extractTenantSlug extracts tenant slug from subdomain or header
func extractTenantSlug(c *gin.Context) string {
	// First try to get from header (for API clients)
	if tenantSlug := c.GetHeader("X-Tenant-Slug"); tenantSlug != "" {
		return tenantSlug
	}

	// Try to get from subdomain
	host := c.Request.Host
	if strings.Contains(host, ".") {
		parts := strings.Split(host, ".")
		if len(parts) >= 2 && parts[0] != "www" && parts[0] != "api" {
			return parts[0] // Return subdomain as tenant slug
		}
	}

	// Try to get from URL path parameter
	if tenantSlug := c.Param("tenant_slug"); tenantSlug != "" {
		return tenantSlug
	}

	// Try to get from query parameter
	if tenantSlug := c.Query("tenant"); tenantSlug != "" {
		return tenantSlug
	}

	return ""
}

// GetTenantContext retrieves tenant context from Gin context
func GetTenantContext(c *gin.Context) (*TenantContext, error) {
	context, exists := c.Get("tenant_context")
	if !exists {
		return nil, fmt.Errorf("tenant context not found")
	}

	tenantContext, ok := context.(*TenantContext)
	if !ok {
		return nil, fmt.Errorf("invalid tenant context type")
	}

	return tenantContext, nil
}

// RequireTenantRole middleware that checks if user has required tenant role or higher
func RequireTenantRole(requiredRole models.MembershipRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantContext, err := GetTenantContext(c)
		if err != nil {
			utils.InternalServerErrorResponse(c, "Tenant context not found")
			c.Abort()
			return
		}

		if !tenantContext.MembershipInfo.HasRole(requiredRole) {
			utils.ForbiddenResponse(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOwnerOrAdmin middleware that requires owner or admin role
func RequireOwnerOrAdmin() gin.HandlerFunc {
	return RequireTenantRole(models.MembershipRoleAdmin)
}

// RequireManager middleware that requires manager role or higher
func RequireManager() gin.HandlerFunc {
	return RequireTenantRole(models.MembershipRoleManager)
}

// RequireAgent middleware that requires agent role or higher  
func RequireAgent() gin.HandlerFunc {
	return RequireTenantRole(models.MembershipRoleAgent)
}

// TenantHealthCheck checks if tenant database is accessible
func TenantHealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantContext, err := GetTenantContext(c)
		if err != nil {
			utils.InternalServerErrorResponse(c, "Tenant context not found")
			c.Abort()
			return
		}

		// Perform health check on tenant database
		if err := database.HealthCheck(tenantContext.TenantDB); err != nil {
			c.JSON(503, gin.H{"error": "Tenant database unavailable", "status": "error"})
			c.Abort()
			return
		}

		c.Next()
	}
}