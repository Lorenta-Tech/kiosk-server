package models

import "time"

// ================================================================
// DB Models
// ================================================================

// DeptAdmin already exists in notes.go — referenced here, not redefined.
// Fields: ID, BranchID, Name, Email, PasswordHash, Role, CreatedAt

// ================================================================
// Super Admin Request Models
// ================================================================

type SuperAdminLoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterDeptAdminRequest struct {
	Name     string `json:"name"      validate:"required,max=150"`
	Email    string `json:"email"     validate:"required,email"`
	Password string `json:"password"  validate:"required"`
	BranchID string `json:"branch_id" validate:"required,uuid"`
}

// ================================================================
// Dept Admin Request Models
// ================================================================

type DeptAdminLoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ================================================================
// Super Admin Response Models
// ================================================================

type SuperAdminLoginResponse struct {
	Token string `json:"token"`
}

type DeptAdminListItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	BranchID  string    `json:"branch_id"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterDeptAdminResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	BranchID string `json:"branch_id"`
}

// ================================================================
// Dept Admin Response Models
// ================================================================

type DeptAdminLoginResponse struct {
	Token string           `json:"token"`
	Admin DeptAdminProfile `json:"admin"`
}

type DeptAdminProfile struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	BranchID string `json:"branch_id"`
}
