package repositories

import (
	"github.com/zen/shared/pkg/database"
	"file-storage-service/internal/models"
)

type FileRepository interface {
	// File metadata operations
	CreateFileMetadata(tenantID string, file *models.FileMetadata) error
	GetFileMetadata(tenantID, fileID string) (*models.FileMetadata, error)
	GetFileByHash(tenantID, fileHash string) (*models.FileMetadata, error)
	UpdateFileMetadata(tenantID string, file *models.FileMetadata) error
	DeleteFileMetadata(tenantID, fileID string) error
	
	// File listing and search
	ListUserFiles(tenantID, userID string, limit, offset int) ([]*models.FileMetadata, int64, error)
	ListPublicFiles(tenantID string, limit, offset int) ([]*models.FileMetadata, int64, error)
	SearchFiles(tenantID, query string, limit, offset int) ([]*models.FileMetadata, error)
	
	// File operations
	UpdateAccessTime(tenantID, fileID string) error
	GetExpiredFiles(tenantID string) ([]*models.FileMetadata, error)
	
	// Statistics
	GetFileStats(tenantID string) (*models.FileStats, error)
	GetUserFileStats(tenantID, userID string) (*models.FileStats, error)
	
	// Access logging
	LogFileAccess(tenantID string, accessLog *models.FileAccessLog) error
	GetFileAccessLogs(tenantID, fileID string, limit, offset int) ([]*models.FileAccessLog, error)
}

type fileRepository struct {
	tenantDBManager *database.TenantDatabaseManager
}

func NewFileRepository(tenantDBManager *database.TenantDatabaseManager) FileRepository {
	return &fileRepository{
		tenantDBManager: tenantDBManager,
	}
}

func (r *fileRepository) CreateFileMetadata(tenantID string, file *models.FileMetadata) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		INSERT INTO file_metadata (id, tenant_id, user_id, original_name, file_name, file_path, file_size, mime_type, file_hash, is_public, tags, metadata, uploaded_at, updated_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, file.ID, file.TenantID, file.UserID, file.OriginalName, file.FileName, file.FilePath,
		file.FileSize, file.MimeType, file.FileHash, file.IsPublic, "{}", "{}", 
		file.UploadedAt, file.UpdatedAt, file.ExpiresAt)

	return result.Error
}

func (r *fileRepository) GetFileMetadata(tenantID, fileID string) (*models.FileMetadata, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var file models.FileMetadata
	err = db.Raw(`
		SELECT id, tenant_id, user_id, original_name, file_name, file_path, file_size, mime_type, file_hash, is_public, uploaded_at, updated_at, accessed_at, expires_at
		FROM file_metadata 
		WHERE id = ? AND tenant_id = ?
	`, fileID, tenantID).Scan(&file).Error

	if err != nil {
		return nil, err
	}

	return &file, nil
}

func (r *fileRepository) GetFileByHash(tenantID, fileHash string) (*models.FileMetadata, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var file models.FileMetadata
	err = db.Raw(`
		SELECT id, tenant_id, user_id, original_name, file_name, file_path, file_size, mime_type, file_hash, is_public, uploaded_at, updated_at, accessed_at, expires_at
		FROM file_metadata 
		WHERE file_hash = ? AND tenant_id = ?
	`, fileHash, tenantID).Scan(&file).Error

	if err != nil {
		return nil, err
	}

	return &file, nil
}

func (r *fileRepository) UpdateFileMetadata(tenantID string, file *models.FileMetadata) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		UPDATE file_metadata 
		SET original_name = ?, is_public = ?, tags = ?, metadata = ?, updated_at = ?, expires_at = ?
		WHERE id = ? AND tenant_id = ?
	`, file.OriginalName, file.IsPublic, "{}", "{}", file.UpdatedAt, file.ExpiresAt, file.ID, tenantID)

	return result.Error
}

func (r *fileRepository) DeleteFileMetadata(tenantID, fileID string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		DELETE FROM file_metadata 
		WHERE id = ? AND tenant_id = ?
	`, fileID, tenantID)

	return result.Error
}

func (r *fileRepository) ListUserFiles(tenantID, userID string, limit, offset int) ([]*models.FileMetadata, int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	var count int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM file_metadata 
		WHERE tenant_id = ? AND user_id = ?
	`, tenantID, userID).Scan(&count)

	// Get files
	var files []*models.FileMetadata
	err = db.Raw(`
		SELECT id, tenant_id, user_id, original_name, file_name, file_path, file_size, mime_type, file_hash, is_public, uploaded_at, updated_at, accessed_at, expires_at
		FROM file_metadata 
		WHERE tenant_id = ? AND user_id = ?
		ORDER BY uploaded_at DESC
		LIMIT ? OFFSET ?
	`, tenantID, userID, limit, offset).Scan(&files).Error

	return files, count, err
}

func (r *fileRepository) ListPublicFiles(tenantID string, limit, offset int) ([]*models.FileMetadata, int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	var count int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM file_metadata 
		WHERE tenant_id = ? AND is_public = true AND (expires_at IS NULL OR expires_at > NOW())
	`, tenantID).Scan(&count)

	// Get files
	var files []*models.FileMetadata
	err = db.Raw(`
		SELECT id, tenant_id, user_id, original_name, file_name, file_path, file_size, mime_type, file_hash, is_public, uploaded_at, updated_at, accessed_at, expires_at
		FROM file_metadata 
		WHERE tenant_id = ? AND is_public = true AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY uploaded_at DESC
		LIMIT ? OFFSET ?
	`, tenantID, limit, offset).Scan(&files).Error

	return files, count, err
}

func (r *fileRepository) SearchFiles(tenantID, query string, limit, offset int) ([]*models.FileMetadata, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var files []*models.FileMetadata
	searchQuery := "%" + query + "%"
	err = db.Raw(`
		SELECT id, tenant_id, user_id, original_name, file_name, file_path, file_size, mime_type, file_hash, is_public, uploaded_at, updated_at, accessed_at, expires_at
		FROM file_metadata 
		WHERE tenant_id = ? AND (original_name ILIKE ? OR file_name ILIKE ?)
		ORDER BY uploaded_at DESC
		LIMIT ? OFFSET ?
	`, tenantID, searchQuery, searchQuery, limit, offset).Scan(&files).Error

	return files, err
}

func (r *fileRepository) UpdateAccessTime(tenantID, fileID string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		UPDATE file_metadata 
		SET accessed_at = NOW()
		WHERE id = ? AND tenant_id = ?
	`, fileID, tenantID)

	return result.Error
}

func (r *fileRepository) GetExpiredFiles(tenantID string) ([]*models.FileMetadata, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var files []*models.FileMetadata
	err = db.Raw(`
		SELECT id, tenant_id, user_id, original_name, file_name, file_path, file_size, mime_type, file_hash, is_public, uploaded_at, updated_at, accessed_at, expires_at
		FROM file_metadata 
		WHERE tenant_id = ? AND expires_at IS NOT NULL AND expires_at <= NOW()
	`, tenantID).Scan(&files).Error

	return files, err
}

func (r *fileRepository) GetFileStats(tenantID string) (*models.FileStats, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	stats := &models.FileStats{}

	// Total files and size
	db.Raw(`
		SELECT COUNT(*) as total_files, COALESCE(SUM(file_size), 0) as total_size
		FROM file_metadata 
		WHERE tenant_id = ?
	`, tenantID).Scan(stats)

	// Public files count
	db.Raw(`
		SELECT COUNT(*) 
		FROM file_metadata 
		WHERE tenant_id = ? AND is_public = true
	`, tenantID).Scan(&stats.PublicFiles)

	stats.PrivateFiles = stats.TotalFiles - stats.PublicFiles

	// Files to expire
	db.Raw(`
		SELECT COUNT(*) 
		FROM file_metadata 
		WHERE tenant_id = ? AND expires_at IS NOT NULL AND expires_at > NOW()
	`, tenantID).Scan(&stats.FilesToExpire)

	return stats, nil
}

func (r *fileRepository) GetUserFileStats(tenantID, userID string) (*models.FileStats, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	stats := &models.FileStats{}

	// User's total files and size
	db.Raw(`
		SELECT COUNT(*) as total_files, COALESCE(SUM(file_size), 0) as total_size
		FROM file_metadata 
		WHERE tenant_id = ? AND user_id = ?
	`, tenantID, userID).Scan(stats)

	// User's public files count
	db.Raw(`
		SELECT COUNT(*) 
		FROM file_metadata 
		WHERE tenant_id = ? AND user_id = ? AND is_public = true
	`, tenantID, userID).Scan(&stats.PublicFiles)

	stats.PrivateFiles = stats.TotalFiles - stats.PublicFiles

	return stats, nil
}

func (r *fileRepository) LogFileAccess(tenantID string, accessLog *models.FileAccessLog) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		INSERT INTO file_access_logs (id, tenant_id, file_id, user_id, ip_address, user_agent, action, accessed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, accessLog.ID, accessLog.TenantID, accessLog.FileID, accessLog.UserID, 
		accessLog.IPAddress, accessLog.UserAgent, accessLog.Action, accessLog.AccessedAt)

	return result.Error
}

func (r *fileRepository) GetFileAccessLogs(tenantID, fileID string, limit, offset int) ([]*models.FileAccessLog, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var logs []*models.FileAccessLog
	err = db.Raw(`
		SELECT id, tenant_id, file_id, user_id, ip_address, user_agent, action, accessed_at
		FROM file_access_logs 
		WHERE tenant_id = ? AND file_id = ?
		ORDER BY accessed_at DESC
		LIMIT ? OFFSET ?
	`, tenantID, fileID, limit, offset).Scan(&logs).Error

	return logs, err
}