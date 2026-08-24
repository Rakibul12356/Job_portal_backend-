package errors

import (
	"fmt"
	"net/http"
)

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type AppError struct {
	HTTPStatus int           `json:"-"`
	Code       string        `json:"code"`
	Message    string        `json:"message"`
	Details    []ErrorDetail `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s (HTTP %d)", e.Code, e.Message, e.HTTPStatus)
}

func NewAppError(httpStatus int, code, message string) *AppError {
	return &AppError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
	}
}

func NewValidationError(message string, details ...ErrorDetail) *AppError {
	return &AppError{
		HTTPStatus: http.StatusBadRequest,
		Code:       "VALIDATION_ERROR",
		Message:    message,
		Details:    details,
	}
}

func NewNotFoundError(message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusNotFound,
		Code:       "NOT_FOUND",
		Message:    message,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    message,
	}
}

func NewForbiddenError(message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusForbidden,
		Code:       "FORBIDDEN",
		Message:    message,
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusConflict,
		Code:       "CONFLICT",
		Message:    message,
	}
}

func NewInternalError(message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusInternalServerError,
		Code:       "INTERNAL_ERROR",
		Message:    message,
	}
}

func NewTooManyRequestsError(message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusTooManyRequests,
		Code:       "RATE_LIMIT_EXCEEDED",
		Message:    message,
	}
}
