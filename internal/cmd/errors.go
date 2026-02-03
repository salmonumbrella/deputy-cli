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
	hint := hintForStatus(apiErr.StatusCode)

	if hint == "" {
		return base + "\nHint: Use --debug for details. Avoid piping stderr into jq (omit 2>&1)."
	}
	return base + "\nHint: " + hint + ". Use --debug for details. Avoid piping stderr into jq (omit 2>&1)."
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

// hintForStatus returns a human-readable hint for the given HTTP status code.
// Used by both text and JSON error formatting.
func hintForStatus(status int) string {
	switch status {
	case 400:
		return "Check field names match the resource schema (use 'deputy resource info <Resource>')"
	case 401:
		return "Run 'deputy auth login' to authenticate"
	case 403:
		return "Check role permissions for this endpoint"
	case 404:
		return "Resource not found, try 'deputy resource list' to verify names"
	case 409:
		return "Conflict with existing data, verify the resource state"
	case 412:
		return "Precondition failed. Try 'deputy auth test' to verify credentials"
	case 417:
		return "Data format error. Check JSON structure (arrays vs objects)"
	case 422:
		return "Validation failed, check required fields and formats"
	case 429:
		return "Rate limited, wait and retry"
	default:
		if status >= 500 {
			return "Server error, retry later"
		}
		return ""
	}
}
