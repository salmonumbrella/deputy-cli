package cmd

import (
	"encoding/json"
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
			contains: []string{"API error 409", "Conflict", "verify the resource state"},
		},
		{
			name:     "API error 422 validation failed",
			err:      &api.APIError{StatusCode: 422, Message: "Unprocessable Entity"},
			contains: []string{"API error 422", "Unprocessable Entity", "Validation failed", "check required fields"},
		},
		{
			name:     "API error 429 rate limited",
			err:      &api.APIError{StatusCode: 429, Message: "Too Many Requests"},
			contains: []string{"API error 429", "Too Many Requests", "Rate limited", "wait and retry"},
		},
		{
			name:     "API error 500 server error",
			err:      &api.APIError{StatusCode: 500, Message: "Internal Server Error"},
			contains: []string{"API error 500", "Internal Server Error", "Server error", "retry"},
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
			name:     "API error 400 bad request",
			err:      &api.APIError{StatusCode: 400, Message: "Bad Request"},
			contains: []string{"API error 400", "Bad Request", "field names", "deputy resource info"},
		},
		{
			name:     "API error 412 precondition failed",
			err:      &api.APIError{StatusCode: 412, Message: "Precondition Failed"},
			contains: []string{"API error 412", "Precondition Failed", "deputy auth test"},
		},
		{
			name:     "API error 417 expectation failed",
			err:      &api.APIError{StatusCode: 417, Message: "Expectation Failed"},
			contains: []string{"API error 417", "Expectation Failed", "JSON structure"},
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
			got := FormatError(tt.err, tt.debug)

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
			name:       "400 has field names hint",
			statusCode: 400,
			message:    "Bad Request",
			wantHint:   "Check field names match the resource schema (use 'deputy resource info <Resource>').",
		},
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
			wantHint:   "Resource not found, try 'deputy resource list' to verify names.",
		},
		{
			name:       "409 has conflict hint",
			statusCode: 409,
			message:    "Conflict",
			wantHint:   "Conflict with existing data, verify the resource state.",
		},
		{
			name:       "412 has precondition hint",
			statusCode: 412,
			message:    "Precondition Failed",
			wantHint:   "Precondition failed. Try 'deputy auth test' to verify credentials.",
		},
		{
			name:       "417 has data format hint",
			statusCode: 417,
			message:    "Expectation Failed",
			wantHint:   "Data format error. Check JSON structure (arrays vs objects).",
		},
		{
			name:       "422 has validation hint",
			statusCode: 422,
			message:    "Unprocessable Entity",
			wantHint:   "Validation failed, check required fields and formats.",
		},
		{
			name:       "429 has rate limit hint",
			statusCode: 429,
			message:    "Too Many Requests",
			wantHint:   "Rate limited, wait and retry.",
		},
		{
			name:       "500 has server error hint",
			statusCode: 500,
			message:    "Internal Server Error",
			wantHint:   "Server error, retry later.",
		},
		{
			name:       "599 has server error hint (5xx range)",
			statusCode: 599,
			message:    "Unknown Server Error",
			wantHint:   "Server error, retry later.",
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
	// Test status codes without specific hints (e.g., 418, 499)
	codes := []int{418, 499}
	for _, code := range codes {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			apiErr := &api.APIError{StatusCode: code, Message: "Error"}
			got := formatAPIError(apiErr)

			// Should NOT contain specific hints for known codes
			specificHints := []string{
				"Check field names",
				"Run 'deputy auth login'",
				"Check role permissions",
				"Resource not found",
				"Conflict with existing data",
				"Precondition failed",
				"Data format error",
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

func TestFormatErrorJSON(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantCode   string
		wantStatus int
		wantRetry  bool
	}{
		{
			name:     "nil error returns empty",
			err:      nil,
			wantCode: "",
		},
		{
			name:       "API error 401",
			err:        &api.APIError{Code: api.ErrCodeAuthRequired, StatusCode: 401, Message: "unauthorized"},
			wantCode:   "AUTH_REQUIRED",
			wantStatus: 401,
			wantRetry:  false,
		},
		{
			name:       "API error 429 retryable",
			err:        &api.APIError{Code: api.ErrCodeRateLimited, StatusCode: 429, Message: "too many requests", Retryable: true, RetryAfter: 30},
			wantCode:   "RATE_LIMITED",
			wantStatus: 429,
			wantRetry:  true,
		},
		{
			name:       "API error 500 retryable",
			err:        &api.APIError{Code: api.ErrCodeServerError, StatusCode: 500, Message: "server error", Retryable: true},
			wantCode:   "SERVER_ERROR",
			wantStatus: 500,
			wantRetry:  true,
		},
		{
			name:      "generic error",
			err:       errors.New("something went wrong"),
			wantCode:  "INVALID_INPUT",
			wantRetry: false,
		},
		{
			name:      "network error",
			err:       errors.New("connection refused"),
			wantCode:  "NETWORK_ERROR",
			wantRetry: true,
		},
		{
			name:      "timeout error",
			err:       errors.New("request timeout"),
			wantCode:  "TIMEOUT",
			wantRetry: true,
		},
		{
			name:      "invalid flag error",
			err:       errors.New("unknown flag: --foo"),
			wantCode:  "INVALID_FLAG",
			wantRetry: false,
		},
		{
			name:       "wrapped API error",
			err:        fmt.Errorf("request failed: %w", &api.APIError{Code: api.ErrCodeNotFound, StatusCode: 404, Message: "not found"}),
			wantCode:   "NOT_FOUND",
			wantStatus: 404,
			wantRetry:  false,
		},
		{
			name:      "jq query error",
			err:       errors.New("invalid jq query: syntax error"),
			wantCode:  "INVALID_INPUT",
			wantRetry: false,
		},
		{
			name:       "API error without code uses CodeFromStatus",
			err:        &api.APIError{StatusCode: 403, Message: "forbidden"},
			wantCode:   "AUTH_FORBIDDEN",
			wantStatus: 403,
			wantRetry:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatErrorJSON(tt.err)

			if tt.wantCode == "" {
				if got != "" {
					t.Errorf("FormatErrorJSON() = %q, want empty", got)
				}
				return
			}

			// Parse the JSON
			var jsonErr JSONError
			if err := json.Unmarshal([]byte(got), &jsonErr); err != nil {
				t.Fatalf("FormatErrorJSON() returned invalid JSON: %v\nGot: %s", err, got)
			}

			if jsonErr.Error.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", jsonErr.Error.Code, tt.wantCode)
			}
			if tt.wantStatus > 0 && jsonErr.Error.Status != tt.wantStatus {
				t.Errorf("status = %d, want %d", jsonErr.Error.Status, tt.wantStatus)
			}
			if jsonErr.Error.Retryable != tt.wantRetry {
				t.Errorf("retryable = %v, want %v", jsonErr.Error.Retryable, tt.wantRetry)
			}
		})
	}
}

func TestHintForStatus_400(t *testing.T) {
	hint := hintForStatus(400)
	if !strings.Contains(hint, "field names") {
		t.Errorf("hintForStatus(400) = %q, want to contain 'field names'", hint)
	}
	if !strings.Contains(hint, "deputy resource info") {
		t.Errorf("hintForStatus(400) = %q, want to contain 'deputy resource info'", hint)
	}
}

func TestHintForStatus_412(t *testing.T) {
	hint := hintForStatus(412)
	if !strings.Contains(hint, "auth test") {
		t.Errorf("hintForStatus(412) = %q, want to contain 'auth test'", hint)
	}
}

func TestHintForStatus_417(t *testing.T) {
	hint := hintForStatus(417)
	if !strings.Contains(hint, "JSON structure") {
		t.Errorf("hintForStatus(417) = %q, want to contain 'JSON structure'", hint)
	}
}
