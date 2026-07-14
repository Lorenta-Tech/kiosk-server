package models

import (
	"time"
)

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
	PageRange    []string `json:"page_range"    validate:"required,min=1"`
	PageLayout   int      `json:"page_layout"   validate:"required,min=1"`
	NumOfPages   int      `json:"num_of_pages"  validate:"required,min=1"`
}

type ConfirmUploadRequest struct {
	SessionID string               `json:"session_id" validate:"required,uuid4"`
	Files     []ConfirmFileRequest `json:"files"      validate:"required,min=1,dive"`
}

type GetJobByTokenRequest struct {
	Token int `json:"token" validate:"required,min=100000,max=999999"`
}

type ExpireSessionRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid4"`
}

type GetJobBySessionIDRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid4"`
}

type ErrorRequestFromPrinter struct {
	SessionID string       `json:"session_id" validate:"required,uuid4"`
	Error     PrinterError `json:"error" validate:"required"`
	PrinterID string       `json:"printer_id"`
}

//DB rows

type UploadSession struct {
	ID          string
	UserID      string
	UserEmail   string
	Status      string
	Token       string
	TotalAmount *float64
	TotalSheets *int
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

type UploadFile struct {
	ID            string
	SessionID     string
	FileName      string
	StagingKey    string
	FinalKey      *string
	PrintingMode  *string
	PrintingSide  *string
	PageRange     []string
	PageLayout    *int
	Copies        *int
	NumberOfPages *int
	Price         *float64
	FileStatus    string
	CreatedAt     time.Time
}

// Init response

type InitFileResponse struct {
	FileID     string `json:"file_id"`
	FileName   string `json:"file_name"`
	UploadURL  string `json:"upload_url"`
	StagingKey string `json:"staging_key"`
}

type InitUploadResponse struct {
	SessionID string             `json:"session_id"`
	Token     int                `json:"token"`
	ExpiresAt time.Time          `json:"expires_at"`
	Files     []InitFileResponse `json:"files"`
}

// Confirm response

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

// Shared file shape

// PrintJobFile is the safe public shape of a file.
// DownloadURL is only populated for the token lookup route —

type PrintJobFile struct {
	FileID        string   `json:"file_id"`
	FileName      string   `json:"file_name"`
	PrintingMode  *string  `json:"printing_mode"`
	PrintingSide  *string  `json:"printing_side"`
	PageRange     []string `json:"page_range"`
	PageLayout    *int     `json:"page_layout"`
	Copies        *int     `json:"copies"`
	NumberOfPages *int     `json:"number_of_pages"`
	Price         *float64 `json:"price"`
	FileStatus    string   `json:"file_status"`
	DownloadURL   *string  `json:"download_url,omitempty"` // only set for token lookup
}

// PrintJob is one session with its files.
type PrintJob struct {
	SessionID   string         `json:"session_id"`
	Status      string         `json:"status"`
	Token       string         `json:"token"`
	TotalAmount *float64       `json:"total_amount"`
	TotalSheets *int           `json:"total_sheets"`
	CreatedAt   time.Time      `json:"created_at"`
	Files       []PrintJobFile `json:"files"`
}

//Recent print jobs response

type RecentPrintJobsResponse struct {
	Jobs  []PrintJob `json:"jobs"`
	Total int        `json:"total"`
}

// Token lookup response
type TokenJobResponse struct {
	Job PrintJob `json:"job"`
}

type AdminPrintHistoryResponse struct {
	History []PrintJob `json:"history"`
}

// NotesCreateSessionRequest
type NotesCreateSessionRequest struct {
	Id string `json:"id"`
}

type NotesUploadCreateSessionRequest struct {
	Files []NotesCreateSessionRequest `json:"files"      validate:"required,min=1,dive"`
}

type NotesUploadCreateSessionResponse struct {
	SessionID string    `json:"session_id"`
	Token     int       `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type NotesConfigs struct {
	FileID       string   `json:"file_id"       validate:"required,uuid4"`
	Copies       int      `json:"copies"        validate:"required,min=1"`
	PrintingSide string   `json:"printing_side" validate:"required,oneof=single_side double_side"`
	PrintingMode string   `json:"printing_mode" validate:"required,oneof=monochromatic color"`
	PageRange    []string `json:"page_range"    validate:"required,min=1"`
	PageLayout   int      `json:"page_layout"   validate:"required,min=1"`
	NumOfPages   int      `json:"num_of_pages"  validate:"required,min=1"`
}

type NotesUploadConfirmSessionRequest struct {
	SessionID string         `json:"session_id" validate:"required,uuid4"`
	Notes     []NotesConfigs `json:"notes"      validate:"required,min=1,dive"`
}

type ConfirmNotesResponse struct {
	FileID     string  `json:"file_id"`
	FileName   string  `json:"file_name"`
	NumOfPages int     `json:"num_of_pages"`
	Copies     int     `json:"copies"`
	Price      float64 `json:"price"`
}

type ConfirmNotesUploadResponse struct {
	SessionID   string                 `json:"session_id"`
	Status      string                 `json:"status"`
	Files       []ConfirmNotesResponse `json:"files"`
	TotalSheets int                    `json:"total_sheets"`
	TotalAmount float64                `json:"total_amount"`
}
