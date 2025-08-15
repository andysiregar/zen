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
	"background-jobs-service/internal/config"
	"background-jobs-service/internal/handlers"
	"background-jobs-service/internal/repositories"
	"background-jobs-service/internal/services"
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
	jobRepo := repositories.NewJobRepository(tenantDBManager, masterDBManager)
	jobService := services.NewJobService(jobRepo)

	// Initialize handlers
	jobHandler := handlers.NewJobHandler(jobService, logger)

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
			"service":   "background-jobs-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API versioning
	v1 := router.Group("/api/v1")
	{
		// Job routes
		jobs := v1.Group("/jobs")
		jobs.Use(middleware.AuthMiddleware(jwtService))
		jobs.Use(middleware.TenantMiddleware(masterDBManager, jwtService))
		{
			jobs.GET("", jobHandler.GetJobs)
			jobs.POST("", jobHandler.CreateJob)
			jobs.GET("/:id", jobHandler.GetJob)
			jobs.PUT("/:id", jobHandler.UpdateJob)
			jobs.DELETE("/:id", jobHandler.DeleteJob)
			jobs.POST("/:id/retry", jobHandler.RetryJob)
			jobs.POST("/:id/cancel", jobHandler.CancelJob)
			jobs.GET("/metrics", jobHandler.GetJobMetrics)
			jobs.GET("/stats", jobHandler.GetTenantStats)
		}

		// Scheduled job routes
		scheduledJobs := v1.Group("/scheduled-jobs")
		scheduledJobs.Use(middleware.AuthMiddleware(jwtService))
		scheduledJobs.Use(middleware.TenantMiddleware(masterDBManager, jwtService))
		{
			scheduledJobs.GET("", jobHandler.GetScheduledJobs)
			scheduledJobs.POST("", jobHandler.CreateScheduledJob)
			scheduledJobs.GET("/:id", jobHandler.GetScheduledJob)
			scheduledJobs.PUT("/:id", jobHandler.UpdateScheduledJob)
			scheduledJobs.DELETE("/:id", jobHandler.DeleteScheduledJob)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Background Jobs service", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// TODO: Start background job workers here
	// go startJobWorkers(cfg, jobRepo, logger)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down Background Jobs service...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Background Jobs service exited")
}