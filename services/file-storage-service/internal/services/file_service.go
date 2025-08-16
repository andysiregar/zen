package services

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"file-storage-service/internal/config"
	"file-storage-service/internal/models"
	"file-storage-service/internal/repositories"
)

type FileService interface {
	// File operations
	UploadFile(tenantID, userID string, fileHeader *multipart.FileHeader, request *models.FileUploadRequest) (*models.FileUploadResponse, error)
	GetFileMetadata(tenantID, fileID string) (*models.FileResponse, error)
	GetFileContent(tenantID, fileID string) (string, error) // Returns file path
	UpdateFile(tenantID, fileID string, request *models.FileUploadRequest) (*models.FileResponse, error)
	DeleteFile(tenantID, fileID string) error
	ShareFile(tenantID, fileID string, request *models.FileShareRequest) (*models.FileResponse, error)
	
	// File listing and search
	ListUserFiles(tenantID, userID string, page, limit int) (*models.FileListResponse, error)
	ListPublicFiles(tenantID string, page, limit int) (*models.FileListResponse, error)
	SearchFiles(tenantID, query string, page, limit int) (*models.FileListResponse, error)
	
	// Statistics and logging
	GetFileStats(tenantID string) (*models.FileStats, error)
	GetUserFileStats(tenantID, userID string) (*models.FileStats, error)
	LogFileAccess(tenantID, fileID, userID, ipAddress, userAgent, action string) error
	GetFileAccessLogs(tenantID, fileID string, page, limit int) ([]*models.FileAccessLog, error)
	
	// Utility methods
	IsFileTypeAllowed(filename string) bool
	CleanupExpiredFiles(tenantID string) error
}

type fileService struct {
	repo   repositories.FileRepository
	config *config.Config
}

func NewFileService(repo repositories.FileRepository, cfg *config.Config) FileService {
	return &fileService{
		repo:   repo,
		config: cfg,
	}
}

func (s *fileService) UploadFile(tenantID, userID string, fileHeader *multipart.FileHeader, request *models.FileUploadRequest) (*models.FileUploadResponse, error) {
	// Validate file type
	if !s.IsFileTypeAllowed(fileHeader.Filename) {
		return nil, fmt.Errorf("file type not allowed")
	}

	// Validate file size
	if fileHeader.Size > s.config.Storage.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size")
	}

	// Open uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Calculate file hash for deduplication
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}
	fileHash := fmt.Sprintf("%x", hasher.Sum(nil))

	// Reset file pointer
	file.Seek(0, io.SeekStart)

	// Check if file already exists (deduplication)
	existingFile, err := s.repo.GetFileByHash(tenantID, fileHash)
	if err == nil && existingFile != nil {
		// File already exists, return existing file metadata
		return &models.FileUploadResponse{
			ID:           existingFile.ID,
			OriginalName: existingFile.OriginalName,
			FileName:     existingFile.FileName,
			FileSize:     existingFile.FileSize,
			MimeType:     existingFile.MimeType,
			PublicURL:    s.generatePublicURL(existingFile),
			PrivateURL:   s.generatePrivateURL(existingFile),
			UploadedAt:   existingFile.UploadedAt,
		}, nil
	}

	// Generate unique file ID and filename
	fileID := uuid.New().String()
	fileExt := filepath.Ext(fileHeader.Filename)
	fileName := fmt.Sprintf("%s_%s%s", fileID, time.Now().Format("20060102_150405"), fileExt)

	// Create tenant directory if it doesn't exist
	tenantDir := filepath.Join(s.config.Storage.BasePath, tenantID)
	if err := os.MkdirAll(tenantDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tenant directory: %w", err)
	}

	// Create full file path
	filePath := filepath.Join(tenantDir, fileName)

	// Save file to disk
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath) // Clean up on failure
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Create file metadata
	now := time.Now()
	fileMetadata := &models.FileMetadata{
		ID:           fileID,
		TenantID:     tenantID,
		UserID:       userID,
		OriginalName: fileHeader.Filename,
		FileName:     fileName,
		FilePath:     filePath,
		FileSize:     fileHeader.Size,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		FileHash:     fileHash,
		IsPublic:     request.IsPublic,
		Tags:         request.Tags,
		Metadata:     request.Metadata,
		UploadedAt:   now,
		UpdatedAt:    now,
		ExpiresAt:    request.ExpiresAt,
	}

	// Save metadata to database
	if err := s.repo.CreateFileMetadata(tenantID, fileMetadata); err != nil {
		os.Remove(filePath) // Clean up on failure
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	// Return upload response
	response := &models.FileUploadResponse{
		ID:           fileID,
		OriginalName: fileHeader.Filename,
		FileName:     fileName,
		FileSize:     fileHeader.Size,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		PrivateURL:   s.generatePrivateURL(fileMetadata),
		UploadedAt:   now,
	}

	if request.IsPublic {
		response.PublicURL = s.generatePublicURL(fileMetadata)
	}

	return response, nil
}

func (s *fileService) GetFileMetadata(tenantID, fileID string) (*models.FileResponse, error) {
	file, err := s.repo.GetFileMetadata(tenantID, fileID)
	if err != nil {
		return nil, err
	}

	// Update access time
	s.repo.UpdateAccessTime(tenantID, fileID)

	response := &models.FileResponse{
		ID:           file.ID,
		TenantID:     file.TenantID,
		UserID:       file.UserID,
		OriginalName: file.OriginalName,
		FileName:     file.FileName,
		FileSize:     file.FileSize,
		MimeType:     file.MimeType,
		IsPublic:     file.IsPublic,
		Tags:         file.Tags,
		Metadata:     file.Metadata,
		PrivateURL:   s.generatePrivateURL(file),
		UploadedAt:   file.UploadedAt,
		UpdatedAt:    file.UpdatedAt,
		AccessedAt:   file.AccessedAt,
		ExpiresAt:    file.ExpiresAt,
	}

	if file.IsPublic {
		response.PublicURL = s.generatePublicURL(file)
	}

	return response, nil
}

func (s *fileService) GetFileContent(tenantID, fileID string) (string, error) {
	file, err := s.repo.GetFileMetadata(tenantID, fileID)
	if err != nil {
		return "", err
	}

	// Check if file exists on disk
	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found on disk")
	}

	return file.FilePath, nil
}

func (s *fileService) UpdateFile(tenantID, fileID string, request *models.FileUploadRequest) (*models.FileResponse, error) {
	file, err := s.repo.GetFileMetadata(tenantID, fileID)
	if err != nil {
		return nil, err
	}

	// Update file metadata
	file.IsPublic = request.IsPublic
	file.Tags = request.Tags
	file.Metadata = request.Metadata
	file.UpdatedAt = time.Now()
	file.ExpiresAt = request.ExpiresAt

	if err := s.repo.UpdateFileMetadata(tenantID, file); err != nil {
		return nil, err
	}

	return s.GetFileMetadata(tenantID, fileID)
}

func (s *fileService) DeleteFile(tenantID, fileID string) error {
	file, err := s.repo.GetFileMetadata(tenantID, fileID)
	if err != nil {
		return err
	}

	// Delete file from disk
	if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file from disk: %w", err)
	}

	// Delete metadata from database
	return s.repo.DeleteFileMetadata(tenantID, fileID)
}

func (s *fileService) ShareFile(tenantID, fileID string, request *models.FileShareRequest) (*models.FileResponse, error) {
	updateRequest := &models.FileUploadRequest{
		IsPublic:  request.IsPublic,
		ExpiresAt: request.ExpiresAt,
	}
	return s.UpdateFile(tenantID, fileID, updateRequest)
}

func (s *fileService) ListUserFiles(tenantID, userID string, page, limit int) (*models.FileListResponse, error) {
	offset := (page - 1) * limit
	files, total, err := s.repo.ListUserFiles(tenantID, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	fileResponses := make([]*models.FileResponse, len(files))
	for i, file := range files {
		fileResponses[i] = s.convertToFileResponse(file)
	}

	return &models.FileListResponse{
		Files:      fileResponses,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (s *fileService) ListPublicFiles(tenantID string, page, limit int) (*models.FileListResponse, error) {
	offset := (page - 1) * limit
	files, total, err := s.repo.ListPublicFiles(tenantID, limit, offset)
	if err != nil {
		return nil, err
	}

	fileResponses := make([]*models.FileResponse, len(files))
	for i, file := range files {
		fileResponses[i] = s.convertToFileResponse(file)
	}

	return &models.FileListResponse{
		Files:      fileResponses,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (s *fileService) SearchFiles(tenantID, query string, page, limit int) (*models.FileListResponse, error) {
	offset := (page - 1) * limit
	files, err := s.repo.SearchFiles(tenantID, query, limit, offset)
	if err != nil {
		return nil, err
	}

	fileResponses := make([]*models.FileResponse, len(files))
	for i, file := range files {
		fileResponses[i] = s.convertToFileResponse(file)
	}

	return &models.FileListResponse{
		Files:      fileResponses,
		TotalCount: int64(len(files)),
		Page:       page,
		Limit:      limit,
	}, nil
}

func (s *fileService) GetFileStats(tenantID string) (*models.FileStats, error) {
	return s.repo.GetFileStats(tenantID)
}

func (s *fileService) GetUserFileStats(tenantID, userID string) (*models.FileStats, error) {
	return s.repo.GetUserFileStats(tenantID, userID)
}

func (s *fileService) LogFileAccess(tenantID, fileID, userID, ipAddress, userAgent, action string) error {
	accessLog := &models.FileAccessLog{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		FileID:     fileID,
		UserID:     userID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Action:     action,
		AccessedAt: time.Now(),
	}
	return s.repo.LogFileAccess(tenantID, accessLog)
}

func (s *fileService) GetFileAccessLogs(tenantID, fileID string, page, limit int) ([]*models.FileAccessLog, error) {
	offset := (page - 1) * limit
	return s.repo.GetFileAccessLogs(tenantID, fileID, limit, offset)
}

func (s *fileService) IsFileTypeAllowed(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowedType := range s.config.Storage.AllowedTypes {
		if ext == strings.ToLower(allowedType) {
			return true
		}
	}
	return false
}

func (s *fileService) CleanupExpiredFiles(tenantID string) error {
	expiredFiles, err := s.repo.GetExpiredFiles(tenantID)
	if err != nil {
		return err
	}

	for _, file := range expiredFiles {
		// Delete file from disk
		if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
			continue // Log error but continue with other files
		}
		
		// Delete metadata from database
		s.repo.DeleteFileMetadata(tenantID, file.ID)
	}

	return nil
}

// Helper methods
func (s *fileService) generatePublicURL(file *models.FileMetadata) string {
	return fmt.Sprintf("%s/public/%s", s.config.Storage.CDNBaseURL, file.ID)
}

func (s *fileService) generatePrivateURL(file *models.FileMetadata) string {
	return fmt.Sprintf("%s/%s", s.config.Storage.CDNBaseURL, file.ID)
}

func (s *fileService) convertToFileResponse(file *models.FileMetadata) *models.FileResponse {
	response := &models.FileResponse{
		ID:           file.ID,
		TenantID:     file.TenantID,
		UserID:       file.UserID,
		OriginalName: file.OriginalName,
		FileName:     file.FileName,
		FileSize:     file.FileSize,
		MimeType:     file.MimeType,
		IsPublic:     file.IsPublic,
		Tags:         file.Tags,
		Metadata:     file.Metadata,
		PrivateURL:   s.generatePrivateURL(file),
		UploadedAt:   file.UploadedAt,
		UpdatedAt:    file.UpdatedAt,
		AccessedAt:   file.AccessedAt,
		ExpiresAt:    file.ExpiresAt,
	}

	if file.IsPublic {
		response.PublicURL = s.generatePublicURL(file)
	}

	return response
}