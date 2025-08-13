package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zen/shared/pkg/auth"
	"github.com/zen/shared/pkg/database"
	"github.com/zen/shared/pkg/middleware"
	"github.com/zen/shared/pkg/redis"
	"github.com/zen/services/platform-admin/internal/config"
	"github.com/zen/services/platform-admin/internal/handlers"
	"github.com/zen/services/platform-admin/internal/repositories"
	"github.com/zen/services/platform-admin/internal/services"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg := config.Load()

	// Initialize database manager
	dbManager, err := database.NewDatabaseManager(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to initialize database manager", zap.Error(err))
	}

	// Initialize Redis client
	redisClient, err := redis.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to initialize Redis client", zap.Error(err))
	}

	// Initialize JWT service for platform admins
	platformJWT := auth.NewPlatformJWTService(
		cfg.JWT.SecretKey,
		time.Duration(cfg.JWT.AccessTokenTTL)*time.Hour,
		time.Duration(cfg.JWT.RefreshTokenTTL)*time.Hour,
	)

	// Initialize repositories
	platformAdminRepo := repositories.NewPlatformAdminRepository(dbManager.GetMasterDB())
	organizationRepo := repositories.NewOrganizationRepository(dbManager)
	tenantRepo := repositories.NewTenantRepository(dbManager)
	ticketRepo := repositories.NewTicketRepository(dbManager)

	// Initialize services
	authService := services.NewPlatformAuthService(platformAdminRepo, platformJWT, logger)
	organizationService := services.NewOrganizationService(organizationRepo, tenantRepo, logger)
	tenantService := services.NewTenantService(tenantRepo, dbManager, logger)
	ticketService := services.NewCrossTenantTicketService(ticketRepo, logger)
	billingService := services.NewBillingService(dbManager, logger)

	// Initialize handlers
	authHandler := handlers.NewPlatformAuthHandler(authService, logger)
	organizationHandler := handlers.NewOrganizationHandler(organizationService, logger)
	tenantHandler := handlers.NewTenantHandler(tenantService, logger)
	ticketHandler := handlers.NewCrossTenantTicketHandler(ticketService, logger)
	billingHandler := handlers.NewBillingHandler(billingService, logger)

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://app.yourdomain.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "platform-admin",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	// API routes
	api := router.Group("/api/v1")

	// Public authentication routes
	auth := api.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// Protected platform admin routes
	protected := api.Group("/")
	protected.Use(middleware.PlatformAuthMiddleware(platformJWT))
	{
		// Platform admin management
		adminRoutes := protected.Group("/admins")
		adminRoutes.Use(middleware.RequirePlatformPermission("admin:view"))
		{
			adminRoutes.GET("", authHandler.ListAdmins)
			adminRoutes.GET("/:id", authHandler.GetAdmin)
			
			// Admin creation/modification requires higher permissions
			adminRoutes.POST("", middleware.RequirePlatformPermission("admin:create"), authHandler.CreateAdmin)
			adminRoutes.PUT("/:id", middleware.RequirePlatformPermission("admin:update"), authHandler.UpdateAdmin)
			adminRoutes.DELETE("/:id", middleware.RequirePlatformPermission("admin:delete"), authHandler.DeleteAdmin)
		}

		// Organization management
		orgRoutes := protected.Group("/organizations")
		orgRoutes.Use(middleware.RequirePlatformPermission("org:view"))
		{
			orgRoutes.GET("", organizationHandler.ListOrganizations)
			orgRoutes.GET("/:id", organizationHandler.GetOrganization)
			orgRoutes.GET("/:id/tenants", organizationHandler.GetOrganizationTenants)
			orgRoutes.GET("/:id/users", organizationHandler.GetOrganizationUsers)
			orgRoutes.GET("/:id/stats", organizationHandler.GetOrganizationStats)
			
			orgRoutes.POST("", middleware.RequirePlatformPermission("org:create"), organizationHandler.CreateOrganization)
			orgRoutes.PUT("/:id", middleware.RequirePlatformPermission("org:update"), organizationHandler.UpdateOrganization)
			orgRoutes.PATCH("/:id/suspend", middleware.RequirePlatformPermission("org:suspend"), organizationHandler.SuspendOrganization)
			orgRoutes.PATCH("/:id/activate", middleware.RequirePlatformPermission("org:suspend"), organizationHandler.ActivateOrganization)
			orgRoutes.DELETE("/:id", middleware.RequirePlatformPermission("org:delete"), organizationHandler.DeleteOrganization)
		}

		// Tenant management
		tenantRoutes := protected.Group("/tenants")
		tenantRoutes.Use(middleware.RequirePlatformPermission("tenant:view"))
		{
			tenantRoutes.GET("", tenantHandler.ListTenants)
			tenantRoutes.GET("/:id", tenantHandler.GetTenant)
			tenantRoutes.GET("/:id/users", tenantHandler.GetTenantUsers)
			tenantRoutes.GET("/:id/stats", tenantHandler.GetTenantStats)
			tenantRoutes.GET("/:id/health", tenantHandler.CheckTenantHealth)
			
			tenantRoutes.POST("", middleware.RequirePlatformPermission("tenant:create"), tenantHandler.CreateTenant)
			tenantRoutes.PUT("/:id", middleware.RequirePlatformPermission("tenant:update"), tenantHandler.UpdateTenant)
			tenantRoutes.PATCH("/:id/suspend", middleware.RequirePlatformPermission("tenant:suspend"), tenantHandler.SuspendTenant)
			tenantRoutes.PATCH("/:id/activate", middleware.RequirePlatformPermission("tenant:suspend"), tenantHandler.ActivateTenant)
			tenantRoutes.DELETE("/:id", middleware.RequirePlatformPermission("tenant:delete"), tenantHandler.DeleteTenant)
		}

		// Cross-tenant ticket management
		ticketRoutes := protected.Group("/tickets")
		ticketRoutes.Use(middleware.RequirePlatformPermission("ticket:view"))
		{
			ticketRoutes.GET("", ticketHandler.ListAllTickets)
			ticketRoutes.GET("/search", ticketHandler.SearchTickets)
			ticketRoutes.GET("/stats", ticketHandler.GetTicketStats)
			ticketRoutes.GET("/:id", ticketHandler.GetTicket)
			ticketRoutes.GET("/tenant/:tenant_id", ticketHandler.GetTenantTickets)
			ticketRoutes.GET("/organization/:org_id", ticketHandler.GetOrganizationTickets)
			
			ticketRoutes.PUT("/:id", middleware.RequirePlatformPermission("ticket:update"), ticketHandler.UpdateTicket)
			ticketRoutes.PATCH("/:id/assign", middleware.RequirePlatformPermission("ticket:assign"), ticketHandler.AssignTicket)
			ticketRoutes.PATCH("/:id/close", middleware.RequirePlatformPermission("ticket:close"), ticketHandler.CloseTicket)
			ticketRoutes.DELETE("/:id", middleware.RequirePlatformPermission("ticket:delete"), ticketHandler.DeleteTicket)
		}

		// Billing management
		billingRoutes := protected.Group("/billing")
		billingRoutes.Use(middleware.RequirePlatformPermission("billing:view"))
		{
			billingRoutes.GET("/organizations/:org_id", billingHandler.GetOrganizationBilling)
			billingRoutes.GET("/tenants/:tenant_id", billingHandler.GetTenantBilling)
			billingRoutes.GET("/invoices", billingHandler.ListInvoices)
			billingRoutes.GET("/invoices/:id", billingHandler.GetInvoice)
			billingRoutes.GET("/reports", billingHandler.GetBillingReports)
			
			billingRoutes.POST("/invoices", middleware.RequirePlatformPermission("billing:invoice"), billingHandler.GenerateInvoice)
			billingRoutes.POST("/refunds", middleware.RequirePlatformPermission("billing:refund"), billingHandler.ProcessRefund)
			billingRoutes.PUT("/organizations/:org_id", middleware.RequirePlatformPermission("billing:update"), billingHandler.UpdateOrganizationBilling)
		}

		// Analytics and reporting
		analyticsRoutes := protected.Group("/analytics")
		analyticsRoutes.Use(middleware.RequirePlatformPermission("analytics:view"))
		{
			analyticsRoutes.GET("/dashboard", func(c *gin.Context) {
				// Platform dashboard with key metrics
				c.JSON(http.StatusOK, gin.H{"message": "Platform analytics dashboard"})
			})
			analyticsRoutes.GET("/organizations", func(c *gin.Context) {
				// Organization analytics
				c.JSON(http.StatusOK, gin.H{"message": "Organization analytics"})
			})
			analyticsRoutes.GET("/tenants", func(c *gin.Context) {
				// Tenant analytics
				c.JSON(http.StatusOK, gin.H{"message": "Tenant analytics"})
			})
			analyticsRoutes.GET("/tickets", func(c *gin.Context) {
				// Ticket analytics
				c.JSON(http.StatusOK, gin.H{"message": "Ticket analytics"})
			})
		}
	}

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	// Graceful server startup
	go func() {
		logger.Info("Starting Platform Admin Service", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Platform Admin Service...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Platform Admin Service forced to shutdown", zap.Error(err))
	}

	logger.Info("Platform Admin Service stopped")
}