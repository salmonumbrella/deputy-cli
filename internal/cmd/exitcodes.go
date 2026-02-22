package cmd

import (
	"errors"
	"strings"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
)

// Exit codes for machine-readable error handling by agents.
const (
	ExitOK         = 0
	ExitGeneral    = 1
	ExitInputError = 2
	ExitAuthError  = 3
	ExitNotFound   = 4
	ExitRateLimit  = 5
	ExitTempError  = 6
)

// ExitCodeFromError maps an error to a stable exit code.
func ExitCodeFromError(err error) int {
	if err == nil {
		return ExitOK
	}

	// ErrEmptyResult shares ExitNotFound (4): "no results" is treated the
	// same as "resource not found" for agent consumption.
	if errors.Is(err, outfmt.ErrEmptyResult) {
		return ExitNotFound
	}

	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.Code
		if code == "" {
			code = api.CodeFromStatus(apiErr.StatusCode)
		}
		return exitCodeFromAPICode(code)
	}

	// Cobra and custom validation errors
	msg := err.Error()
	if isInputError(msg) {
		return ExitInputError
	}

	if isTempError(msg) {
		return ExitTempError
	}

	return ExitGeneral
}

func exitCodeFromAPICode(code string) int {
	switch code {
	case api.ErrCodeAuthRequired, api.ErrCodeAuthForbidden:
		return ExitAuthError
	case api.ErrCodeNotFound:
		return ExitNotFound
	case api.ErrCodeValidation, api.ErrCodeInvalidInput, api.ErrCodeInvalidFlag, api.ErrCodeConflict:
		return ExitInputError
	case api.ErrCodeRateLimited:
		return ExitRateLimit
	case api.ErrCodeServerError, api.ErrCodeTimeout, api.ErrCodeNetworkError:
		return ExitTempError
	default:
		return ExitGeneral
	}
}

func isTempError(msg string) bool {
	for _, pattern := range []string{
		"connection refused",
		"no such host",
		"timeout",
	} {
		if strings.Contains(msg, pattern) {
			return true
		}
	}
	return false
}

func isInputError(msg string) bool {
	for _, pattern := range []string{
		"unknown flag",
		"required flag",
		"missing required argument",
		"invalid --output",
		"too many arguments",
	} {
		if strings.Contains(msg, pattern) {
			return true
		}
	}
	return false
}
