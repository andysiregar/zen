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
	
	"project-service/internal/config"
	"project-service/internal/handlers"
	"project-service/internal/repositories"
	"project-service/internal/services"
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
	projectRepo := repositories.NewProjectRepository(tenantDBManager)
	projectService := services.NewProjectService(projectRepo, logger)
	projectHandler := handlers.NewProjectHandler(projectService, logger)

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
			"service":   "project-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API versioning
	v1 := router.Group("/api/v1")
	
	// All project routes require authentication and tenant context
	projects := v1.Group("/projects")
	projects.Use(middleware.AuthMiddleware(jwtService))
	projects.Use(middleware.TenantMiddleware(masterDBManager, jwtService))
	{
		// Project CRUD operations
		projects.GET("/", projectHandler.ListProjects)
		projects.POST("/", projectHandler.CreateProject)
		projects.GET("/:id", projectHandler.GetProject)
		projects.PUT("/:id", projectHandler.UpdateProject)
		projects.DELETE("/:id", projectHandler.DeleteProject)
		
		// Project members
		projects.GET("/:id/members", projectHandler.ListProjectMembers)
		projects.POST("/:id/members", projectHandler.AddProjectMember)
		projects.DELETE("/:id/members/:user_id", projectHandler.RemoveProjectMember)
		projects.PATCH("/:id/members/:user_id/role", projectHandler.UpdateProjectMemberRole)
		
		// Project statistics
		projects.GET("/:id/stats", projectHandler.GetProjectStats)
		projects.GET("/stats/user", projectHandler.GetUserProjectStats)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Project Service", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down Project Service...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Project Service exited")
}