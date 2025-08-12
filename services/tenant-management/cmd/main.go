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
	"shared/pkg/database"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg := config.Load()
	
	// Initialize database connection
	db, err := database.NewPostgresConnection(
		cfg.Database.MasterHost,
		cfg.Database.MasterPort,
		cfg.Database.MasterUser,
		cfg.Database.MasterPassword,
		cfg.Database.MasterDatabase,
		cfg.Database.SSLMode,
	)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	
	// Initialize repository, service, and handler
	tentantRepo := repositories.NewTenantRepository(db)
	tenantService := services.NewTenantService(tentantRepo)
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
		// Tenant management routes
		tenants := v1.Group("/tenants")
		{
			tenants.GET("/", tenantHandler.ListTenants)
			tenants.POST("/", tenantHandler.CreateTenant)
			tenants.GET("/search", tenantHandler.GetTenantByDomain)    // ?domain=example.com
			tenants.GET("/subdomain", tenantHandler.GetTenantBySubdomain) // ?subdomain=demo
			tenants.GET("/:id", tenantHandler.GetTenant)
			tenants.PUT("/:id", tenantHandler.UpdateTenant)
			tenants.DELETE("/:id", tenantHandler.DeleteTenant)
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