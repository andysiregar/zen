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
	
	"ticket-service/internal/config"
	"ticket-service/internal/handlers"
	"ticket-service/internal/repositories"
	"ticket-service/internal/services"
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
	ticketRepo := repositories.NewTicketRepository(tenantDBManager)
	ticketService := services.NewTicketService(ticketRepo, logger)
	ticketHandler := handlers.NewTicketHandler(ticketService, logger)

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
			"service":   "ticket-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API versioning
	v1 := router.Group("/api/v1")
	
	// All ticket routes require authentication and tenant context
	tickets := v1.Group("/tickets")
	tickets.Use(middleware.AuthMiddleware(jwtService))
	tickets.Use(middleware.TenantMiddleware(masterDBManager, jwtService))
	{
		// Ticket CRUD operations
		tickets.GET("/", ticketHandler.ListTickets)
		tickets.POST("/", ticketHandler.CreateTicket)
		tickets.GET("/:id", ticketHandler.GetTicket)
		tickets.PUT("/:id", ticketHandler.UpdateTicket)
		tickets.DELETE("/:id", ticketHandler.DeleteTicket)
		
		// Ticket status management
		tickets.PATCH("/:id/assign", ticketHandler.AssignTicket)
		tickets.PATCH("/:id/status", ticketHandler.UpdateTicketStatus)
		tickets.PATCH("/:id/priority", ticketHandler.UpdateTicketPriority)
		
		// Comments
		tickets.GET("/:id/comments", ticketHandler.GetTicketComments)
		tickets.POST("/:id/comments", ticketHandler.CreateTicketComment)
		tickets.PUT("/comments/:comment_id", ticketHandler.UpdateComment)
		tickets.DELETE("/comments/:comment_id", ticketHandler.DeleteComment)
		
		// Attachments
		tickets.GET("/:id/attachments", ticketHandler.GetTicketAttachments)
		tickets.POST("/:id/attachments", ticketHandler.UploadAttachment)
		tickets.DELETE("/attachments/:attachment_id", ticketHandler.DeleteAttachment)
		
		// Search and filtering
		tickets.GET("/search", ticketHandler.SearchTickets)
		tickets.GET("/stats", ticketHandler.GetTicketStats)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Ticket Service", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down Ticket Service...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Ticket Service exited")
}