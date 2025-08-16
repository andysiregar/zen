package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"github.com/zen/shared/pkg/auth"
	"github.com/zen/shared/pkg/database"
	"github.com/zen/shared/pkg/middleware"
	"github.com/zen/shared/pkg/models"
	"github.com/zen/shared/pkg/routing"
	"github.com/zen/shared/pkg/utils"
)

type ProxyHandler struct {
	logger         *zap.Logger
	jwtService     *auth.JWTService
	dbManager      *database.DatabaseManager
	regionalRouter *routing.RegionalRouter
	serviceURLs    map[string]string
}

func NewProxyHandler(logger *zap.Logger, jwtService *auth.JWTService, dbManager *database.DatabaseManager, regionalRouter *routing.RegionalRouter) *ProxyHandler {
	return &ProxyHandler{
		logger:         logger,
		jwtService:     jwtService,
		dbManager:      dbManager,
		regionalRouter: regionalRouter,
		serviceURLs: map[string]string{
			"auth":         "http://localhost:8002",
			"tenant":       "http://localhost:8001", 
			"database":     "http://localhost:8003",
			"platform":     "http://localhost:8014",
			"ticket":       "http://localhost:8004",
			"project":      "http://localhost:8005",
			"chat":         "http://localhost:8006",
			"notification": "http://localhost:8007",
			"reporting":    "http://localhost:8012",
			"billing":      "http://localhost:8013",
		},
	}
}

// ProxyToService routes requests to appropriate microservice
func (h *ProxyHandler) ProxyToService(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceURL, exists := h.serviceURLs[serviceName]
		if !exists {
			utils.NotFoundResponse(c, "Service not found")
			return
		}

		// Parse service URL
		target, err := url.Parse(serviceURL)
		if err != nil {
			h.logger.Error("Invalid service URL", zap.String("service", serviceName), zap.Error(err))
			utils.InternalServerErrorResponse(c, "Service configuration error")
			return
		}

		// Create reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(target)
		
		// Customize the proxy director to modify the request
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Host = target.Host
			req.URL.Host = target.Host
			req.URL.Scheme = target.Scheme
			
			// For auth service, keep the full path as-is since it expects /api/v1/auth/*
			// For other services, we may need different path handling
			// Currently just pass through the full path
			
			// Forward headers
			req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
			req.Header.Set("X-Gateway", "api-gateway")
		}

		// Handle errors
		proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
			h.logger.Error("Proxy error", 
				zap.String("service", serviceName),
				zap.String("path", req.URL.Path),
				zap.Error(err))
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`{"success": false, "error": "Service temporarily unavailable"}`))
		}

		// Log the request
		h.logger.Info("Proxying request",
			zap.String("service", serviceName),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("target", target.String()))

		// Execute proxy
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// AuthProxy handles authentication service requests (no auth required)
func (h *ProxyHandler) AuthProxy() gin.HandlerFunc {
	return h.ProxyToService("auth")
}

// TenantProxy handles tenant management requests (requires auth)
func (h *ProxyHandler) TenantProxy() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Apply auth middleware first
		middleware.AuthMiddleware(h.jwtService)(c)
		if c.IsAborted() {
			return
		}
		
		// Then proxy to tenant service
		h.ProxyToService("tenant")(c)
	})
}

// ServiceProxy handles other service requests (requires auth + tenant)
func (h *ProxyHandler) ServiceProxy(serviceName string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Apply auth middleware
		middleware.AuthMiddleware(h.jwtService)(c)
		if c.IsAborted() {
			return
		}
		
		// Apply tenant middleware if tenant context is needed
		if h.requiresTenantContext(c.Request.URL.Path) {
			middleware.TenantMiddleware(h.dbManager, h.jwtService)(c)
			if c.IsAborted() {
				return
			}
		}
		
		// Proxy to service
		h.ProxyToService(serviceName)(c)
	})
}

// requiresTenantContext determines if a request needs tenant context
func (h *ProxyHandler) requiresTenantContext(path string) bool {
	// Skip tenant middleware for certain paths
	skipTenantPaths := []string{
		"/api/v1/auth/",
		"/api/v1/tenant/",
		"/health",
		"/api/v1/user/profile",
	}
	
	for _, skipPath := range skipTenantPaths {
		if strings.HasPrefix(path, skipPath) {
			return false
		}
	}
	
	return true
}

// HealthCheck provides gateway health status
func (h *ProxyHandler) HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check connectivity to all services
		healthStatus := map[string]string{}
		allHealthy := true
		
		for serviceName, serviceURL := range h.serviceURLs {
			if h.checkServiceHealth(serviceURL) {
				healthStatus[serviceName] = "healthy"
			} else {
				healthStatus[serviceName] = "unhealthy"
				allHealthy = false
			}
		}
		
		status := "healthy"
		httpStatus := http.StatusOK
		if !allHealthy {
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
		}
		
		c.JSON(httpStatus, gin.H{
			"status":     status,
			"service":    "api-gateway",
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"services":   healthStatus,
		})
	}
}

// checkServiceHealth performs health check on a service
func (h *ProxyHandler) checkServiceHealth(serviceURL string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	resp, err := client.Get(serviceURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

// HandleCustomDomain handles requests from tenant custom domains
func (h *ProxyHandler) HandleCustomDomain() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host
		
		// Check if it's a custom domain
		isCustom, domain := h.regionalRouter.ParseCustomDomain(host)
		if !isCustom {
			// Not a custom domain, handle normally
			c.Next()
			return
		}

		// For custom domains, redirect to authentication
		utils.SuccessResponse(c, gin.H{
			"custom_domain": domain,
			"auth_url":     "https://chilldesk.io/auth/login",
			"message":      "Please authenticate through the main portal",
		}, "Custom domain detected")
	}
}

// GetRegionalRedirect provides regional routing information after authentication
func (h *ProxyHandler) GetRegionalRedirect() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantSlug := c.Query("tenant")
		if tenantSlug == "" {
			utils.BadRequestResponse(c, "tenant parameter required")
			return
		}

		// TODO: Look up tenant in database to get region
		// For now, return US region by default
		region := "us"
		
		regionalEndpoint := h.regionalRouter.GetRegionalEndpoint(models.TenantRegionUS)
		
		utils.SuccessResponse(c, gin.H{
			"tenant":            tenantSlug,
			"region":           region,
			"regional_endpoint": regionalEndpoint,
			"portal_url":       fmt.Sprintf("%s/tenant/%s", regionalEndpoint, tenantSlug),
		}, "Regional routing information")
	}
}