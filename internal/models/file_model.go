package models

import "time"

type InitFileRequest struct {
	FileName    string `json:"file_name"    validate:"required"`
	ContentType string `json:"content_type" validate:"required,oneof=application/pdf"`
}

type InitUploadRequest struct {
	Files []InitFileRequest `json:"files" validate:"required,min=1,dive"`
}

type ConfirmFileRequest struct {
	FileID       string   `json:"file_id"       validate:"required,uuid4"`
	Copies       int      `json:"copies"        validate:"required,min=1"`
	PrintingSide string   `json:"printing_side" validate:"required,oneof=single_side double_side"`
	PrintingMode string   `json:"printing_mode" validate:"required,oneof=monochromatic color"`
	PageRange    []string `json:"page_range"    validate:"required"`
	PageLayout   int      `json:"page_layout"   validate:"required,min=1"`
	NumOfPages   int      `json:"num_of_pages"  validate:"required,min=1"`
}

type ConfirmUploadRequest struct {
	SessionID string               `json:"session_id" validate:"required,uuid4"`
	Files     []ConfirmFileRequest `json:"files"      validate:"required,min=1,dive"`
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

// Print option fields are pointer types because they start as NULL
type UploadFile struct {
	ID            string
	SessionID     string
	FileName      string
	StagingKey    string
	FinalKey      *string  // NULL until promoted after payment
	PrintingMode  *string  // set during confirm
	PrintingSide  *string  // set during confirm
	PageRange     *[]string  // set during confirm
	PageLayout    *int     // set during confirm
	Copies        *int     // set during confirm
	NumberOfPages *int     // set during confirm
	Price         *float64 // calculated during confirm
	FileStatus    string
	CreatedAt     time.Time
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

type ConfirmFileResponse struct {
	FileID     string  `json:"file_id"`
	FileName   string  `json:"file_name"`
	NumOfPages int     `json:"num_of_pages"`
	Copies     int     `json:"copies"`
	Price      float64 `json:"price"`
}

type ConfirmUploadResponse struct {
	SessionID   string                `json:"session_id"`
	Status      string                `json:"status"`
	Files       []ConfirmFileResponse `json:"files"`
	TotalSheets int                   `json:"total_sheets"`
	TotalAmount float64               `json:"total_amount"`
}
