package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is the single error type used across the entire application.
// Every layer (repository, service, handler) returns or wraps this type.
//
// Code    → machine-readable string, sent to the frontend
// Message → human-readable string, sent to the frontend
// Status  → HTTP status code the handler will use
// Err     → original underlying error, only used for internal logging, never exposed
type AppError struct {
	Code    string
	Message string
	Status  int
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap allows errors.Is / errors.As to work through the chain.
func (e *AppError) Unwrap() error {
	return e.Err
}

// ── Constructors ──────────────────────────────────────────────────────────────

// Internal wraps a low-level error (DB failure, S3 failure, etc.).
// The original err is for logging only — it is never sent to the client.
func Internal(msg string, err error) *AppError {
	return &AppError{
		Code:    "internal_error",
		Message: msg,
		Status:  http.StatusInternalServerError,
		Err:     err,
	}
}

// BadRequest is used when the client sends invalid data.
func BadRequest(code, msg string) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
		Status:  http.StatusBadRequest,
	}
}

// NotFound is used when a requested resource does not exist.
func NotFound(code, msg string) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
		Status:  http.StatusNotFound,
	}
}

// Unauthorized is used when a request is missing or has invalid auth.
func Unauthorized(msg string) *AppError {
	return &AppError{
		Code:    "unauthorized",
		Message: msg,
		Status:  http.StatusUnauthorized,
	}
}

// UnprocessableEntity is used when input is valid JSON but fails business rules.
func UnprocessableEntity(code, msg string) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
		Status:  http.StatusUnprocessableEntity,
	}
}

// ── Helper ────────────────────────────────────────────────────────────────────

// As checks whether err is an *AppError and writes it into target.
// Use this in handlers to decide how to respond.
//
//	var appErr *apperror.AppError
//	if apperror.As(err, &appErr) { ... }
func As(err error, target **AppError) bool {
	return errors.As(err, target)
}