package tenant_models

import (
	"time"
	"gorm.io/gorm"
)

type Attachment struct {
	ID       string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	
	// Relationships - Either ticket_id or comment_id must be set, but not both
	TicketID  *string `json:"ticket_id" gorm:"type:uuid;index"`
	CommentID *string `json:"comment_id" gorm:"type:uuid;index"`
	
	// File Details
	Filename         string `json:"filename" gorm:"not null;size:255"`          // Stored filename
	OriginalFilename string `json:"original_filename" gorm:"not null;size:255"` // User's original filename
	FileSize         int64  `json:"file_size" gorm:"not null"`                  // Size in bytes
	MimeType         string `json:"mime_type" gorm:"not null;size:100"`         // MIME type
	FilePath         string `json:"file_path" gorm:"not null;size:500"`         // Storage path/URL
	
	// Uploader (References Master DB users.id)
	UploadedBy string `json:"uploaded_by" gorm:"type:uuid;not null"`
	
	// Timestamps
	UploadedAt time.Time      `json:"uploaded_at" gorm:"default:now()"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

type AttachmentCreateRequest struct {
	TicketID         *string `json:"ticket_id,omitempty" binding:"omitempty,uuid"`
	CommentID        *string `json:"comment_id,omitempty" binding:"omitempty,uuid"`
	OriginalFilename string  `json:"original_filename" binding:"required,min=1,max=255"`
	MimeType         string  `json:"mime_type" binding:"required"`
}

type AttachmentResponse struct {
	ID               string    `json:"id"`
	TicketID         *string   `json:"ticket_id"`
	CommentID        *string   `json:"comment_id"`
	Filename         string    `json:"filename"`
	OriginalFilename string    `json:"original_filename"`
	FileSize         int64     `json:"file_size"`
	MimeType         string    `json:"mime_type"`
	FilePath         string    `json:"file_path"`
	UploadedBy       string    `json:"uploaded_by"`
	UploadedAt       time.Time `json:"uploaded_at"`
}

// TableName overrides the table name used by Attachment to `attachments`
func (Attachment) TableName() string {
	return "attachments"
}

// ToResponse converts an Attachment model to AttachmentResponse
func (a *Attachment) ToResponse() AttachmentResponse {
	return AttachmentResponse{
		ID:               a.ID,
		TicketID:         a.TicketID,
		CommentID:        a.CommentID,
		Filename:         a.Filename,
		OriginalFilename: a.OriginalFilename,
		FileSize:         a.FileSize,
		MimeType:         a.MimeType,
		FilePath:         a.FilePath,
		UploadedBy:       a.UploadedBy,
		UploadedAt:       a.UploadedAt,
	}
}

// IsImage checks if the attachment is an image file
func (a *Attachment) IsImage() bool {
	imageTypes := []string{
		"image/jpeg",
		"image/jpg", 
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
	}
	
	for _, imgType := range imageTypes {
		if a.MimeType == imgType {
			return true
		}
	}
	return false
}

// GetFileExtension returns the file extension based on MIME type
func (a *Attachment) GetFileExtension() string {
	extensions := map[string]string{
		"image/jpeg":      ".jpg",
		"image/jpg":       ".jpg",
		"image/png":       ".png",
		"image/gif":       ".gif",
		"image/webp":      ".webp",
		"image/svg+xml":   ".svg",
		"application/pdf": ".pdf",
		"text/plain":      ".txt",
		"text/csv":        ".csv",
		"application/json": ".json",
		"application/zip": ".zip",
		"application/x-rar-compressed": ".rar",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": ".xlsx",
	}
	
	if ext, exists := extensions[a.MimeType]; exists {
		return ext
	}
	return ""
}