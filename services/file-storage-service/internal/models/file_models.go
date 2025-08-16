package models

import (
	"time"
)

type FileMetadata struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	UserID      string    `json:"user_id"`
	OriginalName string   `json:"original_name"`
	FileName    string    `json:"file_name"`
	FilePath    string    `json:"file_path"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	FileHash    string    `json:"file_hash"` // SHA256 hash for deduplication
	IsPublic    bool      `json:"is_public"`
	Tags        []string  `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	UploadedAt  time.Time `json:"uploaded_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	AccessedAt  *time.Time `json:"accessed_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type FileUploadRequest struct {
	IsPublic    bool                   `form:"is_public"`
	Tags        []string               `form:"tags"`
	Metadata    map[string]interface{} `form:"metadata"`
	ExpiresAt   *time.Time             `form:"expires_at"`
}

type FileUploadResponse struct {
	ID           string    `json:"id"`
	OriginalName string    `json:"original_name"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	PublicURL    string    `json:"public_url,omitempty"`
	PrivateURL   string    `json:"private_url"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

type FileListResponse struct {
	Files      []*FileResponse `json:"files"`
	TotalCount int64           `json:"total_count"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
}

type FileResponse struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	UserID       string                 `json:"user_id"`
	OriginalName string                 `json:"original_name"`
	FileName     string                 `json:"file_name"`
	FileSize     int64                  `json:"file_size"`
	MimeType     string                 `json:"mime_type"`
	IsPublic     bool                   `json:"is_public"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
	PublicURL    string                 `json:"public_url,omitempty"`
	PrivateURL   string                 `json:"private_url"`
	UploadedAt   time.Time              `json:"uploaded_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	AccessedAt   *time.Time             `json:"accessed_at,omitempty"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
}

type FileShareRequest struct {
	IsPublic  bool       `json:"is_public" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type FileStats struct {
	TotalFiles    int64 `json:"total_files"`
	TotalSize     int64 `json:"total_size"`
	PublicFiles   int64 `json:"public_files"`
	PrivateFiles  int64 `json:"private_files"`
	FilesToExpire int64 `json:"files_to_expire"`
}

type FileAccessLog struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	FileID     string    `json:"file_id"`
	UserID     string    `json:"user_id"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Action     string    `json:"action"` // "download", "view", "share"
	AccessedAt time.Time `json:"accessed_at"`
}