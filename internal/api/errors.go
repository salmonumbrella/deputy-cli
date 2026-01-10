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
	ErrCodeAuthExpired   = "AUTH_EXPIRED"
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
	StatusCode int
	Message    string
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
