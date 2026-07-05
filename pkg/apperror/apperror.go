package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

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

func Internal(msg string, err error) *AppError {
	return &AppError{
		Code:    "internal_error",
		Message: msg,
		Status:  http.StatusInternalServerError,
		Err:     err,
	}
}

func BadRequest(code, msg string) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
		Status:  http.StatusBadRequest,
	}
}

func Conflict(code, msg string) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
		Status:  http.StatusConflict, // 409
	}
}

func NotFound(code, msg string) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
		Status:  http.StatusNotFound,
	}
}

func Unauthorized(msg string) *AppError {
	return &AppError{
		Code:    "unauthorized",
		Message: msg,
		Status:  http.StatusUnauthorized,
	}
}

func UnprocessableEntity(code, msg string) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
		Status:  http.StatusUnprocessableEntity,
	}
}

func As(err error, target **AppError) bool {
	return errors.As(err, target)
}

func Forbidden(msg string) *AppError {
	return &AppError{
		Code:    "forbidden",
		Message: msg,
		Status:  http.StatusForbidden,
	}
}
