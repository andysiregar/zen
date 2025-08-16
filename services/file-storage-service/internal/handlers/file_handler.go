package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zen/shared/pkg/utils"
	"go.uber.org/zap"

	"file-storage-service/internal/models"
	"file-storage-service/internal/services"
)

type FileHandler struct {
	fileService services.FileService
	logger      *zap.Logger
}

func NewFileHandler(fileService services.FileService, logger *zap.Logger) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		logger:      logger,
	}
}

// UploadFile uploads a new file
func (h *FileHandler) UploadFile(c *gin.Context) {
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	if tenantID == "" || userID == "" {
		h.logger.Error("Missing tenant_id or user_id in request context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing tenant or user information")
		return
	}

	// Parse multipart form
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Error("Failed to get file from request", zap.Error(err))
		utils.ErrorResponse(c, http.StatusBadRequest, "No file provided")
		return
	}
	file.Close()

	// Parse upload request
	var request models.FileUploadRequest
	if err := c.ShouldBind(&request); err != nil {
		h.logger.Error("Failed to bind upload request", zap.Error(err))
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Upload file
	response, err := h.fileService.UploadFile(tenantID, userID, fileHeader, &request)
	if err != nil {
		h.logger.Error("Failed to upload file", zap.Error(err), zap.String("tenant_id", tenantID), zap.String("user_id", userID))
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Log file upload
	h.fileService.LogFileAccess(tenantID, response.ID, userID, c.ClientIP(), c.GetHeader("User-Agent"), "upload")

	h.logger.Info("File uploaded successfully", 
		zap.String("file_id", response.ID), 
		zap.String("tenant_id", tenantID), 
		zap.String("user_id", userID))

	utils.SuccessResponse(c, response, "File uploaded successfully")
}

// GetFile retrieves file metadata
func (h *FileHandler) GetFile(c *gin.Context) {
	fileID := c.Param("fileId")
	tenantID := c.GetString("tenant_id")

	if tenantID == "" {
		tenantID = c.Query("tenant_id") // Allow public access with tenant_id in query
	}

	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	file, err := h.fileService.GetFileMetadata(tenantID, fileID)
	if err != nil {
		h.logger.Error("Failed to get file metadata", zap.Error(err), zap.String("file_id", fileID))
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	// Log file access
	h.fileService.LogFileAccess(tenantID, fileID, userID, c.ClientIP(), c.GetHeader("User-Agent"), "view")

	utils.SuccessResponse(c, file, "File retrieved successfully")
}

// DownloadFile serves the actual file content
func (h *FileHandler) DownloadFile(c *gin.Context) {
	fileID := c.Param("fileId")
	tenantID := c.GetString("tenant_id")

	if tenantID == "" {
		tenantID = c.Query("tenant_id") // Allow public access with tenant_id in query
	}

	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	// Get file metadata
	file, err := h.fileService.GetFileMetadata(tenantID, fileID)
	if err != nil {
		h.logger.Error("Failed to get file for download", zap.Error(err), zap.String("file_id", fileID))
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	// Get file path
	filePath, err := h.fileService.GetFileContent(tenantID, fileID)
	if err != nil {
		h.logger.Error("Failed to get file content", zap.Error(err), zap.String("file_id", fileID))
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	// Log file download
	h.fileService.LogFileAccess(tenantID, fileID, userID, c.ClientIP(), c.GetHeader("User-Agent"), "download")

	h.logger.Info("File downloaded", 
		zap.String("file_id", fileID), 
		zap.String("tenant_id", tenantID), 
		zap.String("user_id", userID))

	// Set appropriate headers
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+file.OriginalName)
	c.Header("Content-Type", file.MimeType)

	// Serve the file
	c.File(filePath)
}

// UpdateFile updates file metadata
func (h *FileHandler) UpdateFile(c *gin.Context) {
	fileID := c.Param("fileId")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	if tenantID == "" || userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing tenant or user information")
		return
	}

	var request models.FileUploadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to bind update request", zap.Error(err))
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	file, err := h.fileService.UpdateFile(tenantID, fileID, &request)
	if err != nil {
		h.logger.Error("Failed to update file", zap.Error(err), zap.String("file_id", fileID))
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("File updated successfully", 
		zap.String("file_id", fileID), 
		zap.String("tenant_id", tenantID), 
		zap.String("user_id", userID))

	utils.SuccessResponse(c, file, "File updated successfully")
}

// DeleteFile deletes a file
func (h *FileHandler) DeleteFile(c *gin.Context) {
	fileID := c.Param("fileId")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	if tenantID == "" || userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing tenant or user information")
		return
	}

	if err := h.fileService.DeleteFile(tenantID, fileID); err != nil {
		h.logger.Error("Failed to delete file", zap.Error(err), zap.String("file_id", fileID))
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("File deleted successfully", 
		zap.String("file_id", fileID), 
		zap.String("tenant_id", tenantID), 
		zap.String("user_id", userID))

	utils.SuccessResponse(c, nil, "File deleted successfully")
}

// ShareFile updates file sharing settings
func (h *FileHandler) ShareFile(c *gin.Context) {
	fileID := c.Param("fileId")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	if tenantID == "" || userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing tenant or user information")
		return
	}

	var request models.FileShareRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to bind share request", zap.Error(err))
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	file, err := h.fileService.ShareFile(tenantID, fileID, &request)
	if err != nil {
		h.logger.Error("Failed to share file", zap.Error(err), zap.String("file_id", fileID))
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Log file share action
	h.fileService.LogFileAccess(tenantID, fileID, userID, c.ClientIP(), c.GetHeader("User-Agent"), "share")

	h.logger.Info("File shared successfully", 
		zap.String("file_id", fileID), 
		zap.String("tenant_id", tenantID), 
		zap.String("user_id", userID))

	utils.SuccessResponse(c, file, "File sharing updated successfully")
}

// ListFiles lists user's files
func (h *FileHandler) ListFiles(c *gin.Context) {
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	if tenantID == "" || userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing tenant or user information")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	files, err := h.fileService.ListUserFiles(tenantID, userID, page, limit)
	if err != nil {
		h.logger.Error("Failed to list user files", zap.Error(err), zap.String("user_id", userID))
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, files, "Files retrieved successfully")
}

// ListPublicFiles lists public files for a tenant
func (h *FileHandler) ListPublicFiles(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	files, err := h.fileService.ListPublicFiles(tenantID, page, limit)
	if err != nil {
		h.logger.Error("Failed to list public files", zap.Error(err), zap.String("tenant_id", tenantID))
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, files, "Public files retrieved successfully")
}

// SearchFiles searches for files
func (h *FileHandler) SearchFiles(c *gin.Context) {
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	if tenantID == "" || userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing tenant or user information")
		return
	}

	query := c.Query("q")
	if query == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Search query required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	files, err := h.fileService.SearchFiles(tenantID, query, page, limit)
	if err != nil {
		h.logger.Error("Failed to search files", zap.Error(err), zap.String("query", query))
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, files, "Search completed successfully")
}

// GetFileStats returns file statistics
func (h *FileHandler) GetFileStats(c *gin.Context) {
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	if tenantID == "" || userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing tenant or user information")
		return
	}

	// Check if requesting user stats or tenant stats
	if c.Query("user_stats") == "true" {
		stats, err := h.fileService.GetUserFileStats(tenantID, userID)
		if err != nil {
			h.logger.Error("Failed to get user file stats", zap.Error(err), zap.String("user_id", userID))
			utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
		utils.SuccessResponse(c, stats, "User file stats retrieved successfully")
	} else {
		stats, err := h.fileService.GetFileStats(tenantID)
		if err != nil {
			h.logger.Error("Failed to get file stats", zap.Error(err), zap.String("tenant_id", tenantID))
			utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
		utils.SuccessResponse(c, stats, "File stats retrieved successfully")
	}
}

// GetFileAccessLogs returns file access logs
func (h *FileHandler) GetFileAccessLogs(c *gin.Context) {
	fileID := c.Param("fileId")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	if tenantID == "" || userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing tenant or user information")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	logs, err := h.fileService.GetFileAccessLogs(tenantID, fileID, page, limit)
	if err != nil {
		h.logger.Error("Failed to get file access logs", zap.Error(err), zap.String("file_id", fileID))
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, logs, "File access logs retrieved successfully")
}