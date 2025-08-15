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
	"github.com/zen/shared/pkg/auth"
	"github.com/zen/shared/pkg/database"
	"github.com/zen/shared/pkg/middleware"
	"go.uber.org/zap"

	"file-storage-service/internal/config"
	"file-storage-service/internal/handlers"
	"file-storage-service/internal/repositories"
	"file-storage-service/internal/services"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg := config.Load()

	// Initialize master database connection
	masterDB, err := database.NewPostgresConnection(
		cfg.MasterDatabase.MasterHost,
		cfg.MasterDatabase.MasterPort,
		cfg.MasterDatabase.MasterUser,
		cfg.MasterDatabase.MasterPassword,
		cfg.MasterDatabase.MasterDatabase,
		cfg.MasterDatabase.SSLMode,
	)
	if err != nil {
		logger.Fatal("Failed to connect to master database", zap.Error(err))
	}

	// Initialize tenant database manager
	tenantDBManager := database.NewTenantDatabaseManager(masterDB, cfg.EncryptionKey)

	// Initialize repositories
	fileRepo := repositories.NewFileRepository(tenantDBManager)

	// Initialize JWT service
	jwtService := auth.NewJWTService(
		os.Getenv("JWT_SECRET"), 
		24*time.Hour,  // accessTokenExpiry
		7*24*time.Hour, // refreshTokenExpiry
	)

	// Initialize services
	fileService := services.NewFileService(fileRepo, cfg)

	// Initialize handlers
	fileHandler := handlers.NewFileHandler(fileService, logger)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Add middleware
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(logger))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	}))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "file-storage-service",
			"status":  "healthy",
			"version": "1.0.0",
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// File operations
		files := api.Group("/files")
		{
			files.POST("/upload", middleware.AuthMiddleware(jwtService), fileHandler.UploadFile)
			files.GET("/:fileId", middleware.OptionalAuthMiddleware(jwtService), fileHandler.GetFile)
			files.GET("/:fileId/download", middleware.OptionalAuthMiddleware(jwtService), fileHandler.DownloadFile)
			files.PUT("/:fileId", middleware.AuthMiddleware(jwtService), fileHandler.UpdateFile)
			files.DELETE("/:fileId", middleware.AuthMiddleware(jwtService), fileHandler.DeleteFile)
			files.POST("/:fileId/share", middleware.AuthMiddleware(jwtService), fileHandler.ShareFile)
		}

		// File listing and search
		api.GET("/files", middleware.AuthMiddleware(jwtService), fileHandler.ListFiles)
		api.GET("/files/search", middleware.AuthMiddleware(jwtService), fileHandler.SearchFiles)
		api.GET("/files/public", fileHandler.ListPublicFiles)

		// File statistics
		api.GET("/files/stats", middleware.AuthMiddleware(jwtService), fileHandler.GetFileStats)
		api.GET("/files/:fileId/access-logs", middleware.AuthMiddleware(jwtService), fileHandler.GetFileAccessLogs)
	}

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	// Graceful server startup
	go func() {
		logger.Info("Starting file storage service", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}