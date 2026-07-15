package models

import "time"

// ================================================================
// DB Models
// ================================================================

type Branch struct {
	ID        string
	Name      string
	Code      string
	IsActive  bool
	CreatedAt time.Time
}

type BranchSemester struct {
	ID             string
	BranchID       string
	SemesterNumber int
	IsActive       bool
	CreatedAt      time.Time
}

type Subject struct {
	ID               string
	BranchSemesterID string
	Name             string
	SubjectCode      string
	IsActive         bool
	CreatedAt        time.Time
}

type Module struct {
	ID           string
	SubjectID    string
	ModuleNumber int
	Title        string
	CreatedAt    time.Time
}

type Note struct {
	ID               string
	ModuleID         string
	UploadedBy       string // stores dept admin ID
	Title            string
	Description      string
	FileKey          string
	FileType         string
	OriginalFilename string
	FileSizeBytes    int
	Status           string
	CreatedAt        time.Time
}

// DeptAdmin matches the dept_admins table.
// Completely separate from the students users table.
type DeptAdmin struct {
	ID           string
	BranchID     string
	Name         string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

// ================================================================
// Admin Request Models
// ================================================================

type InitNoteUploadRequest struct {
	ModuleID         string `json:"module_id"         validate:"required,uuid"`
	Title            string `json:"title"             validate:"required,max=200"`
	Description      string `json:"description"       validate:"max=500"`
	FileType         string `json:"file_type"         validate:"required,oneof=pdf ppt docx image"`
	OriginalFilename string `json:"original_filename" validate:"required,max=255"`
	FileSizeBytes    int    `json:"file_size_bytes"   validate:"required,min=1"`
}

type ConfirmNoteUploadRequest struct {
	NoteID string `json:"note_id" validate:"required,uuid"`
}

type UpdateNoteRequest struct {
	Title       string `json:"title"       validate:"required,max=200"`
	Description string `json:"description" validate:"max=500"`
}

type CreateSubjectRequest struct {
	BranchSemesterID string `json:"branch_semester_id" validate:"required,uuid"`
	Name             string `json:"name"               validate:"required,max=150"`
	SubjectCode      string `json:"subject_code"       validate:"required,max=20"`
}

// ================================================================
// Student Request Models
// ================================================================

//unused for now
type NotePrintItem struct {
	NoteID       string   `json:"note_id"       validate:"required,uuid"`
	Copies       int      `json:"copies"        validate:"required,min=1"`
	PrintingSide string   `json:"printing_side" validate:"required,oneof=single_side double_side"`
	PrintingMode string   `json:"printing_mode" validate:"required,oneof=monochromatic color"`
	PageRange    []string `json:"page_range"    validate:"required,min=1"`
	PageLayout   int      `json:"page_layout"   validate:"required,min=1"`
	NumOfPages   int      `json:"num_of_pages"  validate:"required,min=1"`
}

type NotesPrintInitRequest struct {
	Notes []NotePrintItem `json:"notes" validate:"required,min=1,dive"`
}

// ================================================================
// Admin Response Models
// ================================================================

type InitNoteUploadResponse struct {
	NoteID    string `json:"note_id"`
	UploadURL string `json:"upload_url"`
}

type SubjectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SubjectCode string `json:"subject_code"`
}

// ================================================================
// Student Response Models
// ================================================================

// BranchResponse — shown in branch selection screen
type BranchResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// SemesterResponse — shown in semester selection screen
type SemesterResponse struct {
	ID             string `json:"id"`
	SemesterNumber int    `json:"semester_number"`
}

// SubjectListResponse — shown in subject selection screen
type SubjectListResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SubjectCode string `json:"subject_code"`
}

// ModuleResponse — shown in module selection screen
// module_number 6 = "Additional Resources"
type ModuleResponse struct {
	ID           string `json:"id"`
	ModuleNumber int    `json:"module_number"`
	Title        string `json:"title"`
}

// NoteListItem — one note card shown to the student
// Matches the VTU Circle style — title, description, type, size
// No file_key ever exposed
type NoteListItem struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	FileType         string    `json:"file_type"`
	OriginalFilename string    `json:"original_filename"`
	FileSizeBytes    int       `json:"file_size_bytes"`
	CreatedAt        time.Time `json:"created_at"`
}

// NoteViewResponse — returned when student clicks View button
// Frontend opens view_url directly in browser as a PDF viewer
type NoteViewResponse struct {
	ViewURL   string    `json:"view_url"`
	ExpiresAt time.Time `json:"expires_at"`
}
