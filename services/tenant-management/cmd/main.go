package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	
	"tenant-management/internal/config"
	"tenant-management/internal/handlers"
	"tenant-management/internal/repositories"
	"tenant-management/internal/services"
	"github.com/zen/shared/pkg/auth"
	"github.com/zen/shared/pkg/database"
	"github.com/zen/shared/pkg/middleware"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg := config.Load()
	
	// Initialize database connection
	db, err := database.NewPostgresConnection(
		cfg.MasterDatabase.MasterHost,
		cfg.MasterDatabase.MasterPort,
		cfg.MasterDatabase.MasterUser,
		cfg.MasterDatabase.MasterPassword,
		cfg.MasterDatabase.MasterDatabase,
		cfg.MasterDatabase.SSLMode,
	)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize JWT service
	jwtService := auth.NewJWTService(
		cfg.EncryptionKey, // Use encryption key for JWT secret for now
		24*time.Hour,      // 24 hours access token
		7*24*time.Hour,    // 7 days refresh token
	)

	// Initialize database manager for multi-tenant support
	dbManager, err := database.NewDatabaseManager(database.Config{
		Host:     cfg.MasterDatabase.MasterHost,
		Port:     5432,
		User:     cfg.MasterDatabase.MasterUser,
		Password: cfg.MasterDatabase.MasterPassword,
		DBName:   cfg.MasterDatabase.MasterDatabase,
		SSLMode:  cfg.MasterDatabase.SSLMode,
	}, cfg.EncryptionKey)
	if err != nil {
		logger.Fatal("Failed to create database manager", zap.Error(err))
	}
	
	// Initialize repository, service, and handler
	tenantRepo := repositories.NewTenantRepository(db)
	tenantService := services.NewTenantService(tenantRepo)
	tenantHandler := handlers.NewTenantHandler(tenantService)

	// Initialize Gin router
	router := gin.New()
	
	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "tenant-management",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API versioning
	v1 := router.Group("/api/v1")
	{
		// Tenant management routes (require authentication)
		tenants := v1.Group("/tenants")
		tenants.Use(middleware.AuthMiddleware(jwtService))
		{
			tenants.GET("/", tenantHandler.ListTenants)
			tenants.POST("/", tenantHandler.CreateTenant)
			tenants.GET("/by-slug", tenantHandler.GetTenantBySlug) // ?slug=demo
			tenants.GET("/:id", tenantHandler.GetTenant)
			tenants.PUT("/:id", tenantHandler.UpdateTenant)
			tenants.DELETE("/:id", tenantHandler.DeleteTenant)
		}

		// Multi-tenant routes (require authentication + tenant context)
		tenantRoutes := v1.Group("/tenant")
		tenantRoutes.Use(middleware.AuthMiddleware(jwtService))
		tenantRoutes.Use(middleware.TenantMiddleware(dbManager, jwtService))
		{
			// This is where tenant-specific operations would go
			// For example: tickets, projects, etc.
			tenantRoutes.GET("/info", func(c *gin.Context) {
				tenantContext, _ := middleware.GetTenantContext(c)
				c.JSON(http.StatusOK, gin.H{
					"tenant_id":   tenantContext.TenantID,
					"tenant_info": tenantContext.TenantInfo.ToResponse(),
				})
			})
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Tenant Management service", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down Tenant Management service...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Tenant Management service exited")
}