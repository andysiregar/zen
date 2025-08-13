package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zen/shared/pkg/auth"
	"github.com/zen/shared/pkg/utils"
)

// AuthMiddleware creates middleware for JWT authentication
func AuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
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
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("organization_id", claims.OrganizationID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("permissions", claims.Permissions)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// OptionalAuthMiddleware creates middleware for optional JWT authentication
// Sets user context if token is present and valid, but doesn't require it
func OptionalAuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
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
			// Invalid token - continue without setting user context
			c.Next()
			return
		}

		// Set claims in context
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("organization_id", claims.OrganizationID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("permissions", claims.Permissions)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}


// GetUserID retrieves user ID from context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetTenantID retrieves tenant ID from context
func GetTenantID(c *gin.Context) (string, bool) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return "", false
	}
	return tenantID.(string), true
}

// GetOrganizationID retrieves organization ID from context
func GetOrganizationID(c *gin.Context) (string, bool) {
	orgID, exists := c.Get("organization_id")
	if !exists {
		return "", false
	}
	return orgID.(string), true
}

// GetJWTClaims retrieves full JWT claims from context
func GetJWTClaims(c *gin.Context) (*auth.Claims, bool) {
	claims, exists := c.Get("jwt_claims")
	if !exists {
		return nil, false
	}
	return claims.(*auth.Claims), true
}

// RequireRole middleware that checks if user has required role or permission
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			utils.UnauthorizedResponse(c, "User role not found")
			c.Abort()
			return
		}

		userRole := role.(string)
		if !hasRequiredRole(userRole, requiredRole) {
			utils.ForbiddenResponse(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasRequiredRole checks if user role meets requirement
func hasRequiredRole(userRole, requiredRole string) bool {
	// Define role hierarchy
	roleHierarchy := map[string]int{
		"guest":       1,
		"user":        2,
		"agent":       3,
		"manager":     4,
		"admin":       5,
		"super_admin": 6,
	}

	userLevel, exists := roleHierarchy[userRole]
	if !exists {
		return false
	}

	requiredLevel, exists := roleHierarchy[requiredRole]
	if !exists {
		return false
	}

	return userLevel >= requiredLevel
}