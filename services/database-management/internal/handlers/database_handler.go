package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"database-management/internal/services"
	"github.com/zen/shared/pkg/utils"
)

type DatabaseHandler struct {
	databaseService *services.DatabaseService
}

type CreateDatabaseRequest struct {
	Name string `json:"name" binding:"required"`
}

type DropDatabaseRequest struct {
	Name string `json:"name" binding:"required"`
}

type BackupDatabaseRequest struct {
	Name string `json:"name" binding:"required"`
}

func NewDatabaseHandler(databaseService *services.DatabaseService) *DatabaseHandler {
	return &DatabaseHandler{
		databaseService: databaseService,
	}
}

// CreateDatabase handles database creation requests
func (h *DatabaseHandler) CreateDatabase(c *gin.Context) {
	var req CreateDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.databaseService.CreateDatabase(req.Name)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create database")
		return
	}

	utils.CreatedResponse(c, gin.H{
		"database": req.Name,
	}, "Database created successfully")
}

// DropDatabase handles database deletion requests
func (h *DatabaseHandler) DropDatabase(c *gin.Context) {
	var req DropDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.databaseService.DropDatabase(req.Name)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to drop database")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"database": req.Name,
	}, "Database dropped successfully")
}

// ListDatabases handles listing all databases
func (h *DatabaseHandler) ListDatabases(c *gin.Context) {
	databases, err := h.databaseService.ListDatabases()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list databases")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"databases": databases,
	}, "Databases retrieved successfully")
}

// BackupDatabase handles database backup requests
func (h *DatabaseHandler) BackupDatabase(c *gin.Context) {
	var req BackupDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.databaseService.BackupDatabase(req.Name)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to backup database")
		return
	}

	if !result.Success {
		utils.ErrorResponse(c, http.StatusInternalServerError, result.Message)
		return
	}

	utils.SuccessResponse(c, result, "Database backup created successfully")
}

// GetDatabaseStats handles database statistics requests
func (h *DatabaseHandler) GetDatabaseStats(c *gin.Context) {
	dbName := c.Param("name")
	if dbName == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Database name is required")
		return
	}

	stats, err := h.databaseService.GetDatabaseStats(dbName)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get database stats")
		return
	}

	utils.SuccessResponse(c, stats, "Database statistics retrieved successfully")
}

// RunMigrations handles database migration requests
func (h *DatabaseHandler) RunMigrations(c *gin.Context) {
	dbName := c.Param("name")
	if dbName == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Database name is required")
		return
	}

	err := h.databaseService.RunMigrations(dbName)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to run migrations")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"database": dbName,
	}, "Migrations completed successfully")
}