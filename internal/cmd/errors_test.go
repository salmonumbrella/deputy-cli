package cmd

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/salmonumbrella/deputy-cli/internal/api"
)

func TestFormatError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		debug    bool
		contains []string // substrings that must appear in output
		equals   string   // exact match (if non-empty, ignores contains)
	}{
		{
			name:   "nil error returns empty string",
			err:    nil,
			equals: "",
		},
		{
			name:     "generic error gets default hint",
			err:      errors.New("something went wrong"),
			contains: []string{"something went wrong", "Hint: Use --debug for full details."},
		},
		{
			name:   "debug mode returns raw error",
			err:    errors.New("raw error message"),
			debug:  true,
			equals: "raw error message",
		},
		{
			name:     "jq query error gets hint",
			err:      errors.New("invalid jq query: syntax error"),
			contains: []string{"invalid jq query", "Check the --query expression", "Use -o json"},
		},
		{
			name:     "unknown flag error gets hint",
			err:      errors.New("unknown flag: --foo"),
			contains: []string{"unknown flag", "Run --help to see valid flags"},
		},
		{
			name:     "invalid output error gets hint",
			err:      errors.New("invalid --output value"),
			contains: []string{"invalid --output", "Use --output text or --output json"},
		},
		{
			name:     "API error 401 unauthorized",
			err:      &api.APIError{StatusCode: 401, Message: "Unauthorized"},
			contains: []string{"API error 401", "Unauthorized", "Run 'deputy auth login'"},
		},
		{
			name:     "API error 403 forbidden",
			err:      &api.APIError{StatusCode: 403, Message: "Forbidden"},
			contains: []string{"API error 403", "Forbidden", "Check role permissions"},
		},
		{
			name:     "API error 404 not found",
			err:      &api.APIError{StatusCode: 404, Message: "Not Found"},
			contains: []string{"API error 404", "Not Found", "Resource not found", "deputy resource list"},
		},
		{
			name:     "API error 409 conflict",
			err:      &api.APIError{StatusCode: 409, Message: "Conflict"},
			contains: []string{"API error 409", "Conflict", "Verify the resource state"},
		},
		{
			name:     "API error 422 validation failed",
			err:      &api.APIError{StatusCode: 422, Message: "Unprocessable Entity"},
			contains: []string{"API error 422", "Unprocessable Entity", "Validation failed", "Check required fields"},
		},
		{
			name:     "API error 429 rate limited",
			err:      &api.APIError{StatusCode: 429, Message: "Too Many Requests"},
			contains: []string{"API error 429", "Too Many Requests", "Rate limited", "Wait and retry"},
		},
		{
			name:     "API error 500 server error",
			err:      &api.APIError{StatusCode: 500, Message: "Internal Server Error"},
			contains: []string{"API error 500", "Internal Server Error", "Server error", "Retry"},
		},
		{
			name:     "API error 502 bad gateway (5xx)",
			err:      &api.APIError{StatusCode: 502, Message: "Bad Gateway"},
			contains: []string{"API error 502", "Bad Gateway", "Server error"},
		},
		{
			name:     "API error 503 service unavailable (5xx)",
			err:      &api.APIError{StatusCode: 503, Message: "Service Unavailable"},
			contains: []string{"API error 503", "Service Unavailable", "Server error"},
		},
		{
			name:     "API error 400 bad request (no specific hint)",
			err:      &api.APIError{StatusCode: 400, Message: "Bad Request"},
			contains: []string{"API error 400", "Bad Request", "Use --debug for details"},
		},
		{
			name:   "API error debug mode returns raw",
			err:    &api.APIError{StatusCode: 401, Message: "Unauthorized"},
			debug:  true,
			equals: "API error 401: Unauthorized",
		},
		{
			name:     "wrapped API error",
			err:      fmt.Errorf("request failed: %w", &api.APIError{StatusCode: 404, Message: "Employee not found"}),
			contains: []string{"API error 404", "Employee not found", "Resource not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore debug flag
			origDebug := flags.Debug
			defer func() { flags.Debug = origDebug }()
			flags.Debug = tt.debug

			got := FormatError(tt.err)

			if tt.equals != "" {
				if got != tt.equals {
					t.Errorf("FormatError() = %q, want %q", got, tt.equals)
				}
				return
			}

			for _, substr := range tt.contains {
				if !strings.Contains(got, substr) {
					t.Errorf("FormatError() = %q, want substring %q", got, substr)
				}
			}
		})
	}
}

func TestFormatAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		wantHint   string
	}{
		{
			name:       "401 has auth hint",
			statusCode: 401,
			message:    "Unauthorized",
			wantHint:   "Run 'deputy auth login' to authenticate.",
		},
		{
			name:       "403 has permission hint",
			statusCode: 403,
			message:    "Forbidden",
			wantHint:   "Check role permissions for this endpoint.",
		},
		{
			name:       "404 has resource hint",
			statusCode: 404,
			message:    "Not Found",
			wantHint:   "Resource not found. Try 'deputy resource list' to verify names.",
		},
		{
			name:       "409 has conflict hint",
			statusCode: 409,
			message:    "Conflict",
			wantHint:   "Conflict with existing data. Verify the resource state.",
		},
		{
			name:       "422 has validation hint",
			statusCode: 422,
			message:    "Unprocessable Entity",
			wantHint:   "Validation failed. Check required fields and formats.",
		},
		{
			name:       "429 has rate limit hint",
			statusCode: 429,
			message:    "Too Many Requests",
			wantHint:   "Rate limited. Wait and retry.",
		},
		{
			name:       "500 has server error hint",
			statusCode: 500,
			message:    "Internal Server Error",
			wantHint:   "Server error. Retry or try again later.",
		},
		{
			name:       "599 has server error hint (5xx range)",
			statusCode: 599,
			message:    "Unknown Server Error",
			wantHint:   "Server error. Retry or try again later.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := &api.APIError{StatusCode: tt.statusCode, Message: tt.message}
			got := formatAPIError(apiErr)

			// Check base message
			wantBase := fmt.Sprintf("API error %d: %s", tt.statusCode, tt.message)
			if !strings.HasPrefix(got, wantBase) {
				t.Errorf("formatAPIError() should start with %q, got %q", wantBase, got)
			}

			// Check hint
			if !strings.Contains(got, tt.wantHint) {
				t.Errorf("formatAPIError() should contain hint %q, got %q", tt.wantHint, got)
			}

			// Check debug suffix
			if !strings.Contains(got, "Use --debug for details.") {
				t.Errorf("formatAPIError() should contain debug suffix, got %q", got)
			}
		})
	}
}

func TestFormatAPIErrorNoSpecificHint(t *testing.T) {
	// Test status codes without specific hints (e.g., 400, 418, 499)
	codes := []int{400, 418, 499}
	for _, code := range codes {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			apiErr := &api.APIError{StatusCode: code, Message: "Error"}
			got := formatAPIError(apiErr)

			// Should NOT contain specific hints for known codes
			specificHints := []string{
				"Run 'deputy auth login'",
				"Check role permissions",
				"Resource not found",
				"Conflict with existing data",
				"Validation failed",
				"Rate limited",
				"Server error",
			}

			for _, hint := range specificHints {
				if strings.Contains(got, hint) {
					t.Errorf("formatAPIError() for code %d should not contain %q, got %q", code, hint, got)
				}
			}

			// Should still have base message and generic debug hint
			wantBase := fmt.Sprintf("API error %d: Error", code)
			if !strings.HasPrefix(got, wantBase) {
				t.Errorf("formatAPIError() should start with %q, got %q", wantBase, got)
			}
			if !strings.Contains(got, "Use --debug for details") {
				t.Errorf("formatAPIError() should contain debug hint, got %q", got)
			}
		})
	}
}
