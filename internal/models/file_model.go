package models

import "time"

type InitFileRequest struct {
	FileName    string `json:"file_name"    validate:"required"`
	ContentType string `json:"content_type" validate:"required,oneof=application/pdf"`
}

type InitUploadRequest struct {
	Files []InitFileRequest `json:"files" validate:"required,min=1,dive"`
}

type UploadSession struct {
	ID        string
	UserID    string
	UserEmail string
	Status    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type UploadFile struct {
	ID         string
	SessionID  string
	FileName   string
	StagingKey string
	FinalKey   *string
	FileStatus string
	CreatedAt  time.Time
}

type InitFileResponse struct {
	FileID     string `json:"file_id"`
	FileName   string `json:"file_name"`
	UploadURL  string `json:"upload_url"`
	StagingKey string `json:"staging_key"`
}

type InitUploadResponse struct {
	SessionID string             `json:"session_id"`
	ExpiresAt time.Time          `json:"expires_at"`
	Files     []InitFileResponse `json:"files"`
}
