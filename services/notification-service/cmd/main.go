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
	
	"notification-service/internal/config"
	"notification-service/internal/handlers"
	"notification-service/internal/repositories"
	"notification-service/internal/services"
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
	
	// Parse master database port
	port, err := strconv.Atoi(cfg.MasterDatabase.MasterPort)
	if err != nil {
		logger.Fatal("Invalid master database port", zap.Error(err))
	}

	// Initialize database manager (for master database - tenant metadata)
	masterDBManager, err := database.NewDatabaseManager(database.Config{
		Host:     cfg.MasterDatabase.MasterHost,
		Port:     port,
		User:     cfg.MasterDatabase.MasterUser,
		Password: cfg.MasterDatabase.MasterPassword,
		DBName:   cfg.MasterDatabase.MasterDatabase,
		SSLMode:  cfg.MasterDatabase.SSLMode,
	}, cfg.EncryptionKey)
	if err != nil {
		logger.Fatal("Failed to create master database manager", zap.Error(err))
	}

	// Initialize JWT service - use same secret as auth service
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key-change-in-production"
	}
	jwtService := auth.NewJWTService(
		jwtSecret,    // Use same JWT secret as auth service
		24*time.Hour, // 24 hours access token
		7*24*time.Hour, // 7 days refresh token
	)
	
	// Initialize tenant database manager
	tenantDBManager := database.NewTenantDatabaseManager(masterDBManager.GetMasterDB(), cfg.EncryptionKey)

	// Initialize repository, service, and handler
	notificationRepo := repositories.NewNotificationRepository(tenantDBManager)
	notificationService := services.NewNotificationService(notificationRepo, &cfg.SMTPConfig, logger)
	notificationHandler := handlers.NewNotificationHandler(notificationService, logger)

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
			"service":   "notification-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API versioning
	v1 := router.Group("/api/v1")
	
	// Notification routes (require authentication and tenant context)
	notifications := v1.Group("/notifications")
	notifications.Use(middleware.AuthMiddleware(jwtService))
	notifications.Use(middleware.TenantMiddleware(masterDBManager, jwtService))
	{
		// Notification management
		notifications.POST("", notificationHandler.SendNotification)
		notifications.POST("/bulk", notificationHandler.SendBulkNotification)
		notifications.GET("", notificationHandler.GetUserNotifications)
		notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
		notifications.GET("/unread/count", notificationHandler.GetUnreadCount)
		
		// Template management
		notifications.POST("/templates", notificationHandler.CreateTemplate)
		notifications.GET("/templates/:id", notificationHandler.GetTemplate)
		notifications.PUT("/templates/:id", notificationHandler.UpdateTemplate)
		notifications.DELETE("/templates/:id", notificationHandler.DeleteTemplate)
		
		// User preferences
		notifications.GET("/preferences", notificationHandler.GetUserPreferences)
		notifications.PUT("/preferences", notificationHandler.UpdateUserPreferences)
		
		// Statistics
		notifications.GET("/stats", notificationHandler.GetNotificationStats)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Notification Service", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down Notification Service...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Notification Service exited")
}