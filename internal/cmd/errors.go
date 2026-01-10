package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/salmonumbrella/deputy-cli/internal/api"
)

// JSONError is the structured error format for JSON output mode.
type JSONError struct {
	Error JSONErrorDetail `json:"error"`
}

// JSONErrorDetail contains the error details.
type JSONErrorDetail struct {
	Code       string `json:"code"`
	Status     int    `json:"status,omitempty"`
	Message    string `json:"message"`
	Retryable  bool   `json:"retryable"`
	RetryAfter int    `json:"retryAfter,omitempty"`
	Field      string `json:"field,omitempty"`
	Hint       string `json:"hint,omitempty"`
}

// FormatError renders user-friendly errors unless debug is enabled.
func FormatError(err error, debug bool) string {
	if err == nil {
		return ""
	}
	if debug {
		return err.Error()
	}

	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		return formatAPIError(apiErr)
	}

	msg := err.Error()
	if strings.Contains(msg, "invalid jq query") {
		return msg + "\nHint: Check the --query expression or drop it. Use -o json for machine output."
	}
	if strings.Contains(msg, "unknown flag") {
		return msg + "\nHint: Run --help to see valid flags for this command."
	}
	if strings.Contains(msg, "invalid --output") {
		return msg + "\nHint: Use --output text or --output json."
	}

	return msg + "\nHint: Use --debug for full details."
}

func formatAPIError(apiErr *api.APIError) string {
	base := fmt.Sprintf("API error %d: %s", apiErr.StatusCode, apiErr.Message)
	var hint string

	switch apiErr.StatusCode {
	case 401:
		hint = "Run 'deputy auth login' to authenticate."
	case 403:
		hint = "Check role permissions for this endpoint."
	case 404:
		hint = "Resource not found. Try 'deputy resource list' to verify names."
	case 409:
		hint = "Conflict with existing data. Verify the resource state."
	case 422:
		hint = "Validation failed. Check required fields and formats."
	case 429:
		hint = "Rate limited. Wait and retry."
	default:
		if apiErr.StatusCode >= 500 {
			hint = "Server error. Retry or try again later."
		}
	}

	if hint == "" {
		return base + "\nHint: Use --debug for details. Avoid piping stderr into jq (omit 2>&1)."
	}
	return base + "\nHint: " + hint + " Use --debug for details. Avoid piping stderr into jq (omit 2>&1)."
}

// FormatErrorJSON returns a JSON-formatted error for machine parsing.
func FormatErrorJSON(err error) string {
	if err == nil {
		return ""
	}

	detail := JSONErrorDetail{
		Code:    api.ErrCodeInvalidInput,
		Message: err.Error(),
	}

	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		detail.Code = apiErr.Code
		if detail.Code == "" {
			detail.Code = api.CodeFromStatus(apiErr.StatusCode)
		}
		detail.Status = apiErr.StatusCode
		detail.Message = apiErr.Message
		detail.Retryable = apiErr.Retryable
		detail.RetryAfter = apiErr.RetryAfter
		detail.Field = apiErr.Field
		detail.Hint = hintForStatus(apiErr.StatusCode)
	} else {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "invalid jq query"):
			detail.Code = api.ErrCodeInvalidInput
			detail.Hint = "Check the --query expression"
		case strings.Contains(msg, "unknown flag"):
			detail.Code = api.ErrCodeInvalidFlag
			detail.Hint = "Run --help to see valid flags"
		case strings.Contains(msg, "connection refused"), strings.Contains(msg, "no such host"):
			detail.Code = api.ErrCodeNetworkError
			detail.Retryable = true
			detail.Hint = "Check network connection"
		case strings.Contains(msg, "timeout"):
			detail.Code = api.ErrCodeTimeout
			detail.Retryable = true
			detail.Hint = "Request timed out, retry"
		}
	}

	jsonErr := JSONError{Error: detail}
	data, err := json.Marshal(jsonErr)
	if err != nil {
		// Fallback for marshaling failure (should never happen with this struct)
		return fmt.Sprintf(`{"error":{"code":"INTERNAL_ERROR","message":%q}}`, err.Error())
	}
	return string(data)
}

func hintForStatus(status int) string {
	switch status {
	case 401:
		return "Run 'deputy auth login' to authenticate"
	case 403:
		return "Check role permissions for this endpoint"
	case 404:
		return "Resource not found"
	case 409:
		return "Conflict with existing data"
	case 422:
		return "Check required fields and formats"
	case 429:
		return "Rate limited, wait and retry"
	default:
		if status >= 500 {
			return "Server error, retry later"
		}
		return ""
	}
}
