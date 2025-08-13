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
	
	"auth-service/internal/config"
	"auth-service/internal/handlers"
	"auth-service/internal/repositories"
	"auth-service/internal/services"
	"shared/pkg/auth"
	"shared/pkg/database"
	"shared/pkg/middleware"
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
		cfg.JWT.Secret,
		time.Duration(cfg.JWT.AccessTokenTTL)*time.Hour,
		time.Duration(cfg.JWT.RefreshTokenTTL)*time.Hour,
	)

	// Initialize repository, service, and handler
	userRepo := repositories.NewUserRepository(db)
	authService := services.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService, jwtService)

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
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}

		// User management routes (require authentication)
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(jwtService))
		{
			users.GET("/profile", authHandler.GetProfile)
			users.PUT("/profile", authHandler.UpdateProfile)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Authentication service", zap.String("port", cfg.Server.Port))
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