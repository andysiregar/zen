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
	"github.com/zen/shared/pkg/auth"
	"github.com/zen/shared/pkg/database"
	"github.com/zen/shared/pkg/routing"
	"api-gateway/internal/config"
	"api-gateway/internal/handlers"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg := config.LoadConfig()
	
	// Initialize JWT service
	jwtService := auth.NewJWTService(
		cfg.JWT.Secret, 
		time.Duration(cfg.JWT.AccessTokenTTL)*time.Hour,
		time.Duration(cfg.JWT.RefreshTokenTTL)*time.Hour,
	)
	
	// Initialize database manager
	masterDBConfig := database.Config{
		Host:     cfg.MasterDatabase.MasterHost,
		User:     cfg.MasterDatabase.MasterUser,
		Password: cfg.MasterDatabase.MasterPassword,
		DBName:   cfg.MasterDatabase.MasterDatabase,
		Port:     5432, // Convert from string to int - could add strconv.Atoi if needed
		SSLMode:  cfg.MasterDatabase.SSLMode,
	}
	dbManager, err := database.NewDatabaseManager(masterDBConfig, cfg.EncryptionKey)
	if err != nil {
		logger.Fatal("Failed to initialize database manager", zap.Error(err))
	}
	
	// Initialize regional router
	regionalRouter := routing.NewRegionalRouter("chilldesk.io")
	
	// Initialize proxy handler
	proxyHandler := handlers.NewProxyHandler(logger, jwtService, dbManager, regionalRouter)

	// Initialize Gin router
	router := gin.New()
	
	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Tenant-Slug"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	router.Use(cors.New(corsConfig))

	// Health check endpoint
	router.GET("/health", proxyHandler.HealthCheck())

	// Custom domain routing - handles requests from tenant custom domains
	router.Any("/", proxyHandler.HandleCustomDomain())

	// API Gateway routes
	v1 := router.Group("/api/v1")
	{
		// Authentication service (no auth required)
		auth := v1.Group("/auth")
		auth.Any("/*path", proxyHandler.AuthProxy())
		
		// Tenant management (auth required)
		tenant := v1.Group("/tenant")
		tenant.Any("/*path", proxyHandler.TenantProxy())
		
		// Database management (auth required)
		database := v1.Group("/database")
		database.Any("/*path", proxyHandler.ServiceProxy("database"))
		
		// Platform admin (auth required)
		platform := v1.Group("/platform")
		platform.Any("/*path", proxyHandler.ServiceProxy("platform"))
		
		// Business services (auth + tenant required)
		ticket := v1.Group("/ticket")
		ticket.Any("/*path", proxyHandler.ServiceProxy("ticket"))
		
		project := v1.Group("/project") 
		project.Any("/*path", proxyHandler.ServiceProxy("project"))
		
		chat := v1.Group("/chat")
		chat.Any("/*path", proxyHandler.ServiceProxy("chat"))
		
		notification := v1.Group("/notification")
		notification.Any("/*path", proxyHandler.ServiceProxy("notification"))
		
		// Regional routing endpoint  
		v1.GET("/regional-redirect", proxyHandler.GetRegionalRedirect())
		
		// Gateway status endpoint
		v1.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "API Gateway v1 is running",
				"version": "1.0.0",
				"services": map[string]string{
					"auth":         "http://localhost:8002",
					"tenant":       "http://localhost:8001",
					"database":     "http://localhost:8003",
					"platform":     "http://localhost:8014",
					"ticket":       "http://localhost:8004", 
					"project":      "http://localhost:8005",
					"chat":         "http://localhost:8006",
					"notification": "http://localhost:8007",
				},
			})
		})
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting API Gateway server", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down API Gateway server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("API Gateway server exited")
}