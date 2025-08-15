package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zen/shared/pkg/auth"
	"github.com/zen/shared/pkg/database"
	"github.com/zen/shared/pkg/middleware"
	"monitoring-service/internal/config"
	"monitoring-service/internal/handlers"
	"monitoring-service/internal/repositories"
	"monitoring-service/internal/services"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg := config.Load()
	
	// Parse master database port
	port, err := strconv.Atoi(cfg.MasterDatabase.MasterPort)
	if err != nil {
		logger.Fatal("Invalid master database port", zap.Error(err))
	}

	// Initialize database manager for multi-tenant support
	masterDBManager, err := database.NewDatabaseManager(database.Config{
		Host:     cfg.MasterDatabase.MasterHost,
		Port:     port,
		User:     cfg.MasterDatabase.MasterUser,
		Password: cfg.MasterDatabase.MasterPassword,
		DBName:   cfg.MasterDatabase.MasterDatabase,
		SSLMode:  cfg.MasterDatabase.SSLMode,
	}, cfg.EncryptionKey)
	if err != nil {
		logger.Fatal("Failed to create database manager", zap.Error(err))
	}

	// Initialize tenant database manager
	tenantDBManager := database.NewTenantDatabaseManager(masterDBManager.GetMasterDB(), cfg.EncryptionKey)

	// Initialize JWT service
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key-change-in-production"
	}
	jwtService := auth.NewJWTService(
		jwtSecret,
		24*time.Hour,
		7*24*time.Hour,
	)

	// Initialize repositories and services
	monitoringRepo := repositories.NewMonitoringRepository(tenantDBManager, masterDBManager)
	monitoringService := services.NewMonitoringService(monitoringRepo)

	// Initialize handlers
	monitoringHandler := handlers.NewMonitoringHandler(monitoringService, logger)

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
			"service":   "monitoring-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Public health check endpoint (no auth required)
	router.GET("/health-check", monitoringHandler.GetHealthCheck)
	router.GET("/health-check/:service", monitoringHandler.GetServiceHealth)

	// API versioning
	v1 := router.Group("/api/v1")
	{
		// Public log ingestion endpoint (for service logs)
		v1.POST("/logs", monitoringHandler.CreateLogEntry)

		// Protected routes requiring authentication
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtService))
		protected.Use(middleware.TenantMiddleware(masterDBManager, jwtService))
		{
			// Metrics routes
			metrics := protected.Group("/metrics")
			{
				metrics.GET("", monitoringHandler.GetMetrics)
				metrics.POST("", monitoringHandler.CreateMetric)
				metrics.POST("/query", monitoringHandler.QueryMetrics)
			}

			// Alerts routes
			alerts := protected.Group("/alerts")
			{
				alerts.GET("", monitoringHandler.GetAlerts)
				alerts.PUT("/:id", monitoringHandler.UpdateAlert)
				alerts.POST("/:id/acknowledge", monitoringHandler.AcknowledgeAlert)
				alerts.POST("/:id/resolve", monitoringHandler.ResolveAlert)
			}

			// Alert rules routes
			alertRules := protected.Group("/alert-rules")
			{
				alertRules.GET("", monitoringHandler.GetAlertRules)
				alertRules.POST("", monitoringHandler.CreateAlertRule)
				alertRules.PUT("/:id", monitoringHandler.UpdateAlertRule)
				alertRules.DELETE("/:id", monitoringHandler.DeleteAlertRule)
			}

			// Logs routes
			logs := protected.Group("/logs")
			{
				logs.GET("", monitoringHandler.GetLogs)
			}

			// Stats routes
			stats := protected.Group("/stats")
			{
				stats.GET("", monitoringHandler.GetMonitoringStats)
			}
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Monitoring service", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// TODO: Start background monitoring workers here
	// go startHealthCheckers(cfg, monitoringService, logger)
	// go startAlertEvaluators(cfg, monitoringService, logger)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down Monitoring service...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Monitoring service exited")
}