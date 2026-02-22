package api

import (
	"errors"
	"fmt"
	"net/http"
)

// Error codes for machine-readable error handling
const (
	// Authentication
	ErrCodeAuthRequired  = "AUTH_REQUIRED"
	ErrCodeAuthForbidden = "AUTH_FORBIDDEN"

	// Validation
	ErrCodeValidation   = "VALIDATION_FAILED"
	ErrCodeInvalidInput = "INVALID_INPUT"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeConflict     = "CONFLICT"

	// Rate limiting
	ErrCodeRateLimited = "RATE_LIMITED"

	// Server
	ErrCodeServerError = "SERVER_ERROR"
	ErrCodeTimeout     = "TIMEOUT"

	// Client
	ErrCodeNetworkError = "NETWORK_ERROR"
	ErrCodeInvalidFlag  = "INVALID_FLAG"
)

// APIError represents an error response from the Deputy API.
type APIError struct {
	Code       string `json:"code"`
	StatusCode int    `json:"status"`
	Message    string `json:"message"`
	Retryable  bool   `json:"retryable"`
	RetryAfter int    `json:"retryAfter,omitempty"`
	Field      string `json:"field,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// IsStatus returns true when err is an APIError with the given status code.
func IsStatus(err error, code int) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == code
	}
	return false
}

// IsNotFound checks for HTTP 404 API errors.
func IsNotFound(err error) bool {
	return IsStatus(err, http.StatusNotFound)
}

// IsForbidden checks for HTTP 403 API errors.
func IsForbidden(err error) bool {
	return IsStatus(err, http.StatusForbidden)
}

// CodeFromStatus returns the error code for an HTTP status code.
func CodeFromStatus(status int) string {
	switch status {
	case 401:
		return ErrCodeAuthRequired
	case 403:
		return ErrCodeAuthForbidden
	case 404:
		return ErrCodeNotFound
	case 409:
		return ErrCodeConflict
	case 422:
		return ErrCodeValidation
	case 429:
		return ErrCodeRateLimited
	default:
		if status >= 500 {
			return ErrCodeServerError
		}
		return ErrCodeInvalidInput
	}
}

// IsRetryable returns true if the status code indicates a retryable error.
func IsRetryable(status int) bool {
	return status == 429 || status >= 500
}
