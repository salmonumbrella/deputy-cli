package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/salmonumbrella/deputy-cli/internal/api"
)

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
