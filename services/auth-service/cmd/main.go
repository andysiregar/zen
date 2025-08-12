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
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Get port from environment or use default
	port := os.Getenv("AUTH_SERVICE_PORT")
	if port == "" {
		port = "8002"
	}

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
			"service":   "auth-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API versioning
	v1 := router.Group("/api/v1")
	{
		// Authentication routes - will be implemented later
		auth := v1.Group("/auth")
		{
			auth.POST("/login", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Login endpoint",
				})
			})
			
			auth.POST("/register", func(c *gin.Context) {
				c.JSON(http.StatusCreated, gin.H{
					"message": "Register endpoint",
				})
			})
			
			auth.POST("/refresh", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Token refresh endpoint",
				})
			})
			
			auth.POST("/logout", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Logout endpoint",
				})
			})
		}

		// User management routes
		users := v1.Group("/users")
		{
			users.GET("/profile", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Get user profile endpoint",
				})
			})
			
			users.PUT("/profile", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Update user profile endpoint",
				})
			})
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Authentication service", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down Authentication service...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Authentication service exited")
}