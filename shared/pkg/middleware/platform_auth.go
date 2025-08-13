package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zen/shared/pkg/auth"
	"github.com/zen/shared/pkg/models"
	"github.com/zen/shared/pkg/utils"
)

// PlatformAuthMiddleware creates middleware for platform admin JWT authentication
func PlatformAuthMiddleware(jwtService *auth.PlatformJWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedResponse(c, "Authorization header is required")
			c.Abort()
			return
		}

		// Check Bearer format
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			utils.UnauthorizedResponse(c, "Authorization header must be Bearer token")
			c.Abort()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		if tokenString == "" {
			utils.UnauthorizedResponse(c, "Token is required")
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			utils.UnauthorizedResponse(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Set claims in context for downstream middleware/handlers
		c.Set("platform_admin_id", claims.AdminID)
		c.Set("platform_admin_email", claims.Email)
		c.Set("platform_admin_role", claims.Role)
		c.Set("platform_admin_permissions", claims.Permissions)
		c.Set("platform_jwt_claims", claims)

		c.Next()
	}
}

// OptionalPlatformAuthMiddleware creates middleware for optional platform admin authentication
func OptionalPlatformAuthMiddleware(jwtService *auth.PlatformJWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		if tokenString == "" {
			c.Next()
			return
		}

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			// Invalid token - continue without setting admin context
			c.Next()
			return
		}

		// Set claims in context
		c.Set("platform_admin_id", claims.AdminID)
		c.Set("platform_admin_email", claims.Email)
		c.Set("platform_admin_role", claims.Role)
		c.Set("platform_admin_permissions", claims.Permissions)
		c.Set("platform_jwt_claims", claims)

		c.Next()
	}
}

// GetPlatformAdminID retrieves platform admin ID from context
func GetPlatformAdminID(c *gin.Context) (string, bool) {
	adminID, exists := c.Get("platform_admin_id")
	if !exists {
		return "", false
	}
	return adminID.(string), true
}

// GetPlatformAdminEmail retrieves platform admin email from context
func GetPlatformAdminEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("platform_admin_email")
	if !exists {
		return "", false
	}
	return email.(string), true
}

// GetPlatformAdminRole retrieves platform admin role from context
func GetPlatformAdminRole(c *gin.Context) (models.PlatformRole, bool) {
	role, exists := c.Get("platform_admin_role")
	if !exists {
		return "", false
	}
	return role.(models.PlatformRole), true
}

// GetPlatformAdminPermissions retrieves platform admin permissions from context
func GetPlatformAdminPermissions(c *gin.Context) ([]models.PlatformPermission, bool) {
	permissions, exists := c.Get("platform_admin_permissions")
	if !exists {
		return nil, false
	}
	return permissions.([]models.PlatformPermission), true
}

// GetPlatformJWTClaims retrieves full platform JWT claims from context
func GetPlatformJWTClaims(c *gin.Context) (*auth.PlatformClaims, bool) {
	claims, exists := c.Get("platform_jwt_claims")
	if !exists {
		return nil, false
	}
	return claims.(*auth.PlatformClaims), true
}

// RequirePlatformRole middleware that checks if platform admin has required role
func RequirePlatformRole(requiredRole models.PlatformRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := GetPlatformAdminRole(c)
		if !exists {
			utils.UnauthorizedResponse(c, "Platform admin role not found")
			c.Abort()
			return
		}

		if !hasPlatformRole(role, requiredRole) {
			utils.ForbiddenResponse(c, "Insufficient platform permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePlatformPermission middleware that checks if platform admin has specific permission
func RequirePlatformPermission(requiredPermission models.PlatformPermission) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := GetPlatformAdminPermissions(c)
		if !exists {
			utils.UnauthorizedResponse(c, "Platform admin permissions not found")
			c.Abort()
			return
		}

		if !hasPlatformPermission(permissions, requiredPermission) {
			utils.ForbiddenResponse(c, "Insufficient platform permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPlatformPermission middleware that checks if platform admin has any of the specified permissions
func RequireAnyPlatformPermission(requiredPermissions ...models.PlatformPermission) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := GetPlatformAdminPermissions(c)
		if !exists {
			utils.UnauthorizedResponse(c, "Platform admin permissions not found")
			c.Abort()
			return
		}

		hasAny := false
		for _, requiredPerm := range requiredPermissions {
			if hasPlatformPermission(permissions, requiredPerm) {
				hasAny = true
				break
			}
		}

		if !hasAny {
			utils.ForbiddenResponse(c, "Insufficient platform permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasPlatformRole checks if admin role meets requirement (hierarchical)
func hasPlatformRole(adminRole, requiredRole models.PlatformRole) bool {
	// Define role hierarchy for platform roles
	roleHierarchy := map[models.PlatformRole]int{
		models.PlatformRoleReadOnly:      1,
		models.PlatformRoleAnalyst:       2,
		models.PlatformRoleSupportAgent:  3,
		models.PlatformRoleBillingOps:    4,
		models.PlatformRoleCustomerOps:   5,
		models.PlatformRoleAdmin:         6,
		models.PlatformRoleSuperAdmin:    7,
	}

	adminLevel, exists := roleHierarchy[adminRole]
	if !exists {
		return false
	}

	requiredLevel, exists := roleHierarchy[requiredRole]
	if !exists {
		return false
	}

	return adminLevel >= requiredLevel
}

// hasPlatformPermission checks if admin has specific permission
func hasPlatformPermission(permissions []models.PlatformPermission, requiredPermission models.PlatformPermission) bool {
	for _, perm := range permissions {
		if perm == requiredPermission {
			return true
		}
	}
	return false
}