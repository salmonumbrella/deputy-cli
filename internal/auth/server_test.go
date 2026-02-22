package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/secrets"
)

func TestValidateInstall(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		// Valid cases
		{
			name:    "valid alphanumeric",
			input:   "myinstall",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			input:   "install123",
			wantErr: false,
		},
		{
			name:    "valid with dashes",
			input:   "my-install",
			wantErr: false,
		},
		{
			name:    "valid with underscores",
			input:   "my_install",
			wantErr: false,
		},
		{
			name:    "valid mixed characters",
			input:   "My-Install_123",
			wantErr: false,
		},
		{
			name:    "valid single character",
			input:   "a",
			wantErr: false,
		},
		{
			name:    "valid 64 characters (max length)",
			input:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			wantErr: false,
		},

		// Invalid cases
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errMsg:  "install name cannot be empty",
		},
		{
			name:    "too long (65 characters)",
			input:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			wantErr: true,
			errMsg:  "install name too long (max 64 characters)",
		},
		{
			name:    "contains spaces",
			input:   "my install",
			wantErr: true,
			errMsg:  "install name contains invalid characters",
		},
		{
			name:    "contains special characters",
			input:   "my@install",
			wantErr: true,
			errMsg:  "install name contains invalid characters",
		},
		{
			name:    "contains dot",
			input:   "my.install",
			wantErr: true,
			errMsg:  "install name contains invalid characters",
		},
		{
			name:    "contains slash",
			input:   "my/install",
			wantErr: true,
			errMsg:  "install name contains invalid characters",
		},
		{
			name:    "contains exclamation",
			input:   "install!",
			wantErr: true,
			errMsg:  "install name contains invalid characters",
		},
		{
			name:    "unicode characters",
			input:   "install\u00e9",
			wantErr: true,
			errMsg:  "install name contains invalid characters",
		},
		{
			name:    "leading space",
			input:   " install",
			wantErr: true,
			errMsg:  "install name contains invalid characters",
		},
		{
			name:    "trailing space",
			input:   "install ",
			wantErr: true,
			errMsg:  "install name contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInstall(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInstall(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("ValidateInstall(%q) error = %q, want %q", tt.input, err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateGeo(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		// Valid cases
		{
			name:    "valid au",
			input:   "au",
			wantErr: false,
		},
		{
			name:    "valid uk",
			input:   "uk",
			wantErr: false,
		},
		{
			name:    "valid na",
			input:   "na",
			wantErr: false,
		},

		// Invalid cases
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errMsg:  "invalid region: must be au, uk, or na",
		},
		{
			name:    "uppercase AU",
			input:   "AU",
			wantErr: true,
			errMsg:  "invalid region: must be au, uk, or na",
		},
		{
			name:    "uppercase UK",
			input:   "UK",
			wantErr: true,
			errMsg:  "invalid region: must be au, uk, or na",
		},
		{
			name:    "uppercase NA",
			input:   "NA",
			wantErr: true,
			errMsg:  "invalid region: must be au, uk, or na",
		},
		{
			name:    "mixed case Au",
			input:   "Au",
			wantErr: true,
			errMsg:  "invalid region: must be au, uk, or na",
		},
		{
			name:    "invalid region eu",
			input:   "eu",
			wantErr: true,
			errMsg:  "invalid region: must be au, uk, or na",
		},
		{
			name:    "invalid region us",
			input:   "us",
			wantErr: true,
			errMsg:  "invalid region: must be au, uk, or na",
		},
		{
			name:    "invalid region ap",
			input:   "ap",
			wantErr: true,
			errMsg:  "invalid region: must be au, uk, or na",
		},
		{
			name:    "region with whitespace",
			input:   " au",
			wantErr: true,
			errMsg:  "invalid region: must be au, uk, or na",
		},
		{
			name:    "region with trailing whitespace",
			input:   "au ",
			wantErr: true,
			errMsg:  "invalid region: must be au, uk, or na",
		},
		{
			name:    "random string",
			input:   "xyz",
			wantErr: true,
			errMsg:  "invalid region: must be au, uk, or na",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGeo(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGeo(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("ValidateGeo(%q) error = %q, want %q", tt.input, err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		// Valid cases
		{
			name:    "valid token",
			input:   "abc123xyz",
			wantErr: false,
		},
		{
			name:    "valid single character",
			input:   "a",
			wantErr: false,
		},
		{
			name:    "valid token with special chars",
			input:   "token-with_special.chars@123",
			wantErr: false,
		},
		{
			name:    "valid long token",
			input:   "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			wantErr: false,
		},
		{
			name:    "valid 512 character token (max length)",
			input:   string(make([]byte, 512)),
			wantErr: false,
		},
		{
			name:    "valid token with whitespace in middle",
			input:   "token with spaces",
			wantErr: false,
		},

		// Invalid cases
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errMsg:  "API token cannot be empty",
		},
		{
			name:    "too long token (513 characters)",
			input:   string(make([]byte, 513)),
			wantErr: true,
			errMsg:  "API token too long",
		},
	}

	// Fix the zero-byte tokens - use actual characters
	for i := range tests {
		if tests[i].name == "valid 512 character token (max length)" {
			token := make([]byte, 512)
			for j := range token {
				token[j] = 'a'
			}
			tests[i].input = string(token)
		}
		if tests[i].name == "too long token (513 characters)" {
			token := make([]byte, 513)
			for j := range token {
				token[j] = 'a'
			}
			tests[i].input = string(token)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToken(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("ValidateToken(%q) error = %q, want %q", tt.input, err.Error(), tt.errMsg)
				}
			}
		})
	}
}

// =============================================================================
// HTTP Handler Tests
// =============================================================================

// createTestServer creates a SetupServer for testing with controlled parameters
func createTestServer(t *testing.T) *SetupServer {
	t.Helper()
	store := secrets.NewMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create test server: %v", err)
	}
	return server
}

// createTestServerWithLimiter creates a SetupServer with a custom rate limiter
func createTestServerWithLimiter(t *testing.T, maxAttempts int, window time.Duration) *SetupServer {
	t.Helper()
	store := secrets.NewMockStore()

	// Create server normally, then replace the limiter
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create test server: %v", err)
	}

	// Replace with custom limiter for testing
	server.limiter = newRateLimiter(maxAttempts, window)

	return server
}

// TestRateLimiter tests the rate limiter directly
func TestRateLimiter(t *testing.T) {
	t.Run("allows requests within limit", func(t *testing.T) {
		rl := newRateLimiter(3, time.Minute)

		for i := 0; i < 3; i++ {
			if err := rl.check("192.168.1.1", "/test"); err != nil {
				t.Errorf("request %d should succeed, got error: %v", i+1, err)
			}
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		rl := newRateLimiter(3, time.Minute)

		// Use up the limit
		for i := 0; i < 3; i++ {
			_ = rl.check("192.168.1.1", "/test")
		}

		// Fourth request should fail
		if err := rl.check("192.168.1.1", "/test"); err == nil {
			t.Error("expected rate limit error, got nil")
		}
	})

	t.Run("tracks different IPs separately", func(t *testing.T) {
		rl := newRateLimiter(2, time.Minute)

		// Use up limit for IP1
		_ = rl.check("192.168.1.1", "/test")
		_ = rl.check("192.168.1.1", "/test")

		// IP2 should still work
		if err := rl.check("192.168.1.2", "/test"); err != nil {
			t.Errorf("different IP should not be rate limited: %v", err)
		}
	})

	t.Run("tracks different endpoints separately", func(t *testing.T) {
		rl := newRateLimiter(2, time.Minute)

		// Use up limit for /validate
		_ = rl.check("192.168.1.1", "/validate")
		_ = rl.check("192.168.1.1", "/validate")

		// /submit should still work
		if err := rl.check("192.168.1.1", "/submit"); err != nil {
			t.Errorf("different endpoint should not be rate limited: %v", err)
		}
	})

	t.Run("resets after window expires", func(t *testing.T) {
		rl := newRateLimiter(2, 10*time.Millisecond)

		// Use up limit
		_ = rl.check("192.168.1.1", "/test")
		_ = rl.check("192.168.1.1", "/test")

		// Should be blocked
		if err := rl.check("192.168.1.1", "/test"); err == nil {
			t.Error("should be rate limited")
		}

		// Wait for window to expire
		time.Sleep(15 * time.Millisecond)

		// Should work again
		if err := rl.check("192.168.1.1", "/test"); err != nil {
			t.Errorf("should be allowed after window reset: %v", err)
		}
	})
}

// TestCSRFProtection tests CSRF token validation on handlers
func TestCSRFProtection(t *testing.T) {
	server := createTestServer(t)

	tests := []struct {
		name           string
		endpoint       string
		csrfToken      string
		expectedStatus int
	}{
		{
			name:           "validate with valid CSRF",
			endpoint:       "/validate",
			csrfToken:      server.csrfToken,
			expectedStatus: http.StatusOK, // Will fail validation but pass CSRF
		},
		{
			name:           "validate with invalid CSRF",
			endpoint:       "/validate",
			csrfToken:      "invalid-token",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "validate with empty CSRF",
			endpoint:       "/validate",
			csrfToken:      "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "submit with valid CSRF",
			endpoint:       "/submit",
			csrfToken:      server.csrfToken,
			expectedStatus: http.StatusOK, // Will fail validation but pass CSRF
		},
		{
			name:           "submit with invalid CSRF",
			endpoint:       "/submit",
			csrfToken:      "wrong-token",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "complete with valid CSRF",
			endpoint:       "/complete",
			csrfToken:      server.csrfToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "complete with invalid CSRF",
			endpoint:       "/complete",
			csrfToken:      "bad-token",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh server for /complete tests since they modify state
			testServer := server
			if tt.endpoint == "/complete" {
				testServer = createTestServer(t)
				if tt.csrfToken == server.csrfToken {
					tt.csrfToken = testServer.csrfToken
				}
			}

			body := bytes.NewBufferString(`{"install":"test","geo":"au","token":"testtoken"}`)
			req := httptest.NewRequest(http.MethodPost, tt.endpoint, body)
			req.Header.Set("Content-Type", "application/json")
			if tt.csrfToken != "" {
				req.Header.Set("X-CSRF-Token", tt.csrfToken)
			}

			rec := httptest.NewRecorder()

			// Route to appropriate handler
			switch tt.endpoint {
			case "/validate":
				testServer.handleValidate(rec, req)
			case "/submit":
				testServer.handleSubmit(rec, req)
			case "/complete":
				testServer.handleComplete(rec, req)
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d; body: %s",
					tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

// TestValidateHandler tests the /validate endpoint
func TestValidateHandler(t *testing.T) {
	t.Run("rejects non-POST methods", func(t *testing.T) {
		server := createTestServer(t)

		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}
		for _, method := range methods {
			req := httptest.NewRequest(method, "/validate", nil)
			rec := httptest.NewRecorder()

			server.handleValidate(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s: expected 405, got %d", method, rec.Code)
			}
		}
	})

	t.Run("rejects invalid JSON body", func(t *testing.T) {
		server := createTestServer(t)

		body := bytes.NewBufferString(`{invalid json}`)
		req := httptest.NewRequest(http.MethodPost, "/validate", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		rec := httptest.NewRecorder()

		server.handleValidate(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["success"] != false {
			t.Error("expected success=false")
		}
		if resp["error"] != "Invalid request body" {
			t.Errorf("unexpected error message: %v", resp["error"])
		}
	})

	t.Run("validates install name", func(t *testing.T) {
		server := createTestServer(t)

		tests := []struct {
			install     string
			expectError bool
			errorMsg    string
		}{
			{"", true, "install name cannot be empty"},
			{"my@install", true, "install name contains invalid characters"},
			{"validinstall", false, ""}, // Will proceed to API validation
		}

		for _, tc := range tests {
			body, _ := json.Marshal(map[string]string{
				"install": tc.install,
				"geo":     "au",
				"token":   "testtoken",
			})

			req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-CSRF-Token", server.csrfToken)
			rec := httptest.NewRecorder()

			server.handleValidate(rec, req)

			var resp map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response for install=%q: %v", tc.install, err)
			}

			if tc.expectError {
				if resp["success"] != false {
					t.Errorf("install=%q: expected success=false", tc.install)
				}
				if resp["error"] != tc.errorMsg {
					t.Errorf("install=%q: expected error %q, got %q", tc.install, tc.errorMsg, resp["error"])
				}
			}
		}
	})

	t.Run("validates geo region", func(t *testing.T) {
		server := createTestServer(t)

		tests := []struct {
			geo         string
			expectError bool
		}{
			{"", true},
			{"invalid", true},
			{"AU", true}, // Case-sensitive
			{"au", false},
			{"uk", false},
			{"na", false},
		}

		for _, tc := range tests {
			body, _ := json.Marshal(map[string]string{
				"install": "validinstall",
				"geo":     tc.geo,
				"token":   "testtoken",
			})

			req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-CSRF-Token", server.csrfToken)
			rec := httptest.NewRecorder()

			server.handleValidate(rec, req)

			var resp map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response for geo=%q: %v", tc.geo, err)
			}

			if tc.expectError {
				if resp["success"] != false {
					t.Errorf("geo=%q: expected success=false", tc.geo)
				}
				errMsg, ok := resp["error"].(string)
				if !ok || errMsg == "" {
					t.Errorf("geo=%q: expected error message", tc.geo)
				}
			}
		}
	})

	t.Run("validates token presence", func(t *testing.T) {
		server := createTestServer(t)

		body, _ := json.Marshal(map[string]string{
			"install": "validinstall",
			"geo":     "au",
			"token":   "",
		})

		req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		rec := httptest.NewRecorder()

		server.handleValidate(rec, req)

		var resp map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["success"] != false {
			t.Error("expected success=false for empty token")
		}
		if resp["error"] != "API token cannot be empty" {
			t.Errorf("unexpected error: %v", resp["error"])
		}
	})

	t.Run("trims whitespace from input", func(t *testing.T) {
		server := createTestServer(t)

		// Whitespace around invalid install should still fail
		body, _ := json.Marshal(map[string]string{
			"install": "  ",
			"geo":     "au",
			"token":   "testtoken",
		})

		req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		rec := httptest.NewRecorder()

		server.handleValidate(rec, req)

		var resp map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["success"] != false {
			t.Error("expected success=false for whitespace-only install")
		}
	})
}

// TestValidateHandlerRateLimiting tests rate limiting on the validate endpoint
func TestValidateHandlerRateLimiting(t *testing.T) {
	// Create server with very low rate limit for testing
	server := createTestServerWithLimiter(t, 3, time.Minute)

	body, _ := json.Marshal(map[string]string{
		"install": "test",
		"geo":     "au",
		"token":   "testtoken",
	})

	// Make requests up to the limit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		req.RemoteAddr = "192.168.1.100:12345"
		rec := httptest.NewRecorder()

		server.handleValidate(rec, req)

		if rec.Code == http.StatusTooManyRequests {
			t.Errorf("request %d should not be rate limited", i+1)
		}
	}

	// Fourth request should be rate limited
	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	req.RemoteAddr = "192.168.1.100:12345"
	rec := httptest.NewRecorder()

	server.handleValidate(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["success"] != false {
		t.Error("expected success=false")
	}
}

// TestSubmitHandlerRateLimiting tests rate limiting on the submit endpoint
func TestSubmitHandlerRateLimiting(t *testing.T) {
	server := createTestServerWithLimiter(t, 2, time.Minute)

	body, _ := json.Marshal(map[string]string{
		"install": "test",
		"geo":     "au",
		"token":   "testtoken",
	})

	// Exhaust the rate limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		req.RemoteAddr = "10.0.0.1:8080"
		rec := httptest.NewRecorder()

		server.handleSubmit(rec, req)
	}

	// Next request should fail
	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	req.RemoteAddr = "10.0.0.1:8080"
	rec := httptest.NewRecorder()

	server.handleSubmit(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}
}

// TestSetupHandler tests the root setup page handler
func TestSetupHandler(t *testing.T) {
	server := createTestServer(t)

	t.Run("returns 404 for non-root paths", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/other", nil)
		rec := httptest.NewRecorder()

		server.handleSetup(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("returns HTML for root path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		server.handleSetup(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "text/html; charset=utf-8" {
			t.Errorf("expected text/html content type, got %s", contentType)
		}
	})

	t.Run("sets security headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		server.handleSetup(rec, req)

		headers := map[string]string{
			"X-Content-Type-Options": "nosniff",
			"X-Frame-Options":        "DENY",
		}

		for header, expected := range headers {
			if got := rec.Header().Get(header); got != expected {
				t.Errorf("%s: expected %q, got %q", header, expected, got)
			}
		}

		if csp := rec.Header().Get("Content-Security-Policy"); csp == "" {
			t.Error("expected Content-Security-Policy header to be set")
		}
	})

	t.Run("includes CSRF token in response", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		server.handleSetup(rec, req)

		body := rec.Body.String()
		if !bytes.Contains([]byte(body), []byte(server.csrfToken)) {
			t.Error("expected CSRF token to be included in HTML response")
		}
	})
}

// TestSuccessHandler tests the success page handler
func TestSuccessHandler(t *testing.T) {
	t.Run("returns HTML page", func(t *testing.T) {
		server := createTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/success", nil)
		rec := httptest.NewRecorder()

		server.handleSuccess(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "text/html; charset=utf-8" {
			t.Errorf("expected text/html, got %s", contentType)
		}
	})

	t.Run("displays pending result info", func(t *testing.T) {
		server := createTestServer(t)

		// Set pending result
		server.pendingMu.Lock()
		server.pendingResult = &SetupResult{
			Install: "mycompany",
			Geo:     "au",
		}
		server.pendingMu.Unlock()

		req := httptest.NewRequest(http.MethodGet, "/success", nil)
		rec := httptest.NewRecorder()

		server.handleSuccess(rec, req)

		body := rec.Body.String()
		if !bytes.Contains([]byte(body), []byte("mycompany")) {
			t.Error("expected install name in success page")
		}
		if !bytes.Contains([]byte(body), []byte("AU")) {
			t.Error("expected geo (uppercased) in success page")
		}
	})
}

// TestCompleteHandler tests the complete endpoint
func TestCompleteHandler(t *testing.T) {
	t.Run("rejects non-POST methods", func(t *testing.T) {
		server := createTestServer(t)

		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
		for _, method := range methods {
			req := httptest.NewRequest(method, "/complete", nil)
			rec := httptest.NewRecorder()

			server.handleComplete(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s: expected 405, got %d", method, rec.Code)
			}
		}
	})

	t.Run("returns success with valid CSRF", func(t *testing.T) {
		server := createTestServer(t)

		req := httptest.NewRequest(http.MethodPost, "/complete", nil)
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		rec := httptest.NewRecorder()

		server.handleComplete(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["success"] != true {
			t.Error("expected success=true")
		}
	})
}

// TestRateLimiterCleanup tests the cleanup method of rate limiter
func TestRateLimiterCleanup(t *testing.T) {
	t.Run("removes expired entries", func(t *testing.T) {
		rl := newRateLimiter(5, 10*time.Millisecond)

		// Add some entries
		_ = rl.check("192.168.1.1", "/test")
		_ = rl.check("192.168.1.2", "/test")
		_ = rl.check("192.168.1.3", "/test")

		// Verify entries exist
		rl.mu.Lock()
		entriesBeforeCleanup := len(rl.attempts)
		rl.mu.Unlock()
		if entriesBeforeCleanup != 3 {
			t.Errorf("expected 3 entries before cleanup, got %d", entriesBeforeCleanup)
		}

		// Wait for entries to expire
		time.Sleep(15 * time.Millisecond)

		// Run cleanup
		rl.cleanup()

		// Verify entries are removed
		rl.mu.Lock()
		entriesAfterCleanup := len(rl.attempts)
		rl.mu.Unlock()
		if entriesAfterCleanup != 0 {
			t.Errorf("expected 0 entries after cleanup, got %d", entriesAfterCleanup)
		}
	})

	t.Run("keeps non-expired entries", func(t *testing.T) {
		rl := newRateLimiter(5, 1*time.Hour)

		// Add an entry that won't expire soon
		_ = rl.check("192.168.1.1", "/test")

		// Run cleanup immediately
		rl.cleanup()

		// Verify entry still exists
		rl.mu.Lock()
		entriesCount := len(rl.attempts)
		rl.mu.Unlock()
		if entriesCount != 1 {
			t.Errorf("expected 1 entry after cleanup (non-expired), got %d", entriesCount)
		}
	})

	t.Run("mixed expired and non-expired entries", func(t *testing.T) {
		rl := newRateLimiter(5, 10*time.Millisecond)

		// Add first entry that will expire
		_ = rl.check("192.168.1.1", "/test")

		// Wait for first entry to expire
		time.Sleep(15 * time.Millisecond)

		// Add second entry that won't be expired yet
		rl.mu.Lock()
		rl.attempts["192.168.1.2:/test"] = &clientLimit{
			count:   1,
			resetAt: time.Now().Add(1 * time.Hour),
		}
		rl.mu.Unlock()

		// Run cleanup
		rl.cleanup()

		// Verify only non-expired entry remains
		rl.mu.Lock()
		entriesCount := len(rl.attempts)
		_, exists := rl.attempts["192.168.1.2:/test"]
		rl.mu.Unlock()

		if entriesCount != 1 {
			t.Errorf("expected 1 entry after cleanup, got %d", entriesCount)
		}
		if !exists {
			t.Error("expected non-expired entry to still exist")
		}
	})
}

// TestSubmitHandler tests additional submit handler paths
func TestSubmitHandler(t *testing.T) {
	t.Run("rejects non-POST methods", func(t *testing.T) {
		server := createTestServer(t)

		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}
		for _, method := range methods {
			req := httptest.NewRequest(method, "/submit", nil)
			rec := httptest.NewRecorder()

			server.handleSubmit(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s: expected 405, got %d", method, rec.Code)
			}
		}
	})

	t.Run("rejects invalid JSON body", func(t *testing.T) {
		server := createTestServer(t)

		body := bytes.NewBufferString(`{invalid json}`)
		req := httptest.NewRequest(http.MethodPost, "/submit", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		rec := httptest.NewRecorder()

		server.handleSubmit(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["success"] != false {
			t.Error("expected success=false")
		}
		if resp["error"] != "Invalid request body" {
			t.Errorf("unexpected error message: %v", resp["error"])
		}
	})

	t.Run("validates install name", func(t *testing.T) {
		server := createTestServer(t)

		body, _ := json.Marshal(map[string]string{
			"install": "",
			"geo":     "au",
			"token":   "testtoken",
		})

		req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		rec := httptest.NewRecorder()

		server.handleSubmit(rec, req)

		var resp map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["success"] != false {
			t.Error("expected success=false for empty install")
		}
		if resp["error"] != "install name cannot be empty" {
			t.Errorf("unexpected error: %v", resp["error"])
		}
	})

	t.Run("validates geo region", func(t *testing.T) {
		server := createTestServer(t)

		body, _ := json.Marshal(map[string]string{
			"install": "valid",
			"geo":     "invalid",
			"token":   "testtoken",
		})

		req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		rec := httptest.NewRecorder()

		server.handleSubmit(rec, req)

		var resp map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["success"] != false {
			t.Error("expected success=false for invalid geo")
		}
	})

	t.Run("validates token presence", func(t *testing.T) {
		server := createTestServer(t)

		body, _ := json.Marshal(map[string]string{
			"install": "valid",
			"geo":     "au",
			"token":   "",
		})

		req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		rec := httptest.NewRecorder()

		server.handleSubmit(rec, req)

		var resp map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["success"] != false {
			t.Error("expected success=false for empty token")
		}
	})

	t.Run("trims whitespace from input", func(t *testing.T) {
		server := createTestServer(t)

		body, _ := json.Marshal(map[string]string{
			"install": "  ",
			"geo":     "au",
			"token":   "testtoken",
		})

		req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		rec := httptest.NewRecorder()

		server.handleSubmit(rec, req)

		var resp map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["success"] != false {
			t.Error("expected success=false for whitespace-only install")
		}
	})
}

// TestCompleteHandler_WithPendingResult tests complete handler with a pending result
func TestCompleteHandler_WithPendingResult(t *testing.T) {
	t.Run("sends pending result through channel", func(t *testing.T) {
		server := createTestServer(t)

		// Set a pending result
		server.pendingMu.Lock()
		server.pendingResult = &SetupResult{
			Install: "testinstall",
			Geo:     "au",
		}
		server.pendingMu.Unlock()

		// Create a channel to capture the result
		go func() {
			req := httptest.NewRequest(http.MethodPost, "/complete", nil)
			req.Header.Set("X-CSRF-Token", server.csrfToken)
			rec := httptest.NewRecorder()

			server.handleComplete(rec, req)
		}()

		// Read from result channel
		select {
		case result := <-server.result:
			if result.Install != "testinstall" {
				t.Errorf("expected install 'testinstall', got %q", result.Install)
			}
			if result.Geo != "au" {
				t.Errorf("expected geo 'au', got %q", result.Geo)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("timeout waiting for result on channel")
		}
	})
}

// TestGetClientIP tests client IP extraction
func TestGetClientIP(t *testing.T) {
	t.Run("extracts IP from RemoteAddr with port", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.100:54321"

		ip := getClientIP(req)

		if ip != "192.168.1.100" {
			t.Errorf("expected '192.168.1.100', got %q", ip)
		}
	})

	t.Run("handles IPv6 address", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "[::1]:54321"

		ip := getClientIP(req)

		if ip != "::1" {
			t.Errorf("expected '::1', got %q", ip)
		}
	})
}

// TestNewSetupServer tests server creation
func TestNewSetupServer(t *testing.T) {
	t.Run("creates server with CSRF token", func(t *testing.T) {
		store := secrets.NewMockStore()
		server, err := NewSetupServer(store)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(server.csrfToken) != 64 { // 32 bytes hex-encoded = 64 chars
			t.Errorf("expected CSRF token of 64 chars, got %d", len(server.csrfToken))
		}
	})

	t.Run("creates server with initialized channels", func(t *testing.T) {
		store := secrets.NewMockStore()
		server, err := NewSetupServer(store)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if server.result == nil {
			t.Error("result channel should not be nil")
		}
		if server.shutdown == nil {
			t.Error("shutdown channel should not be nil")
		}
		if server.stopCleanup == nil {
			t.Error("stopCleanup channel should not be nil")
		}
	})

	t.Run("creates server with rate limiter", func(t *testing.T) {
		store := secrets.NewMockStore()
		server, err := NewSetupServer(store)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if server.limiter == nil {
			t.Error("rate limiter should not be nil")
		}
	})
}

func TestSetupServer_Start_CompleteFlow(t *testing.T) {
	// Track when browser func is called to ensure goroutine completes
	browserCalled := make(chan struct{})
	urlCh := make(chan string, 1)

	// Set up mock BEFORE any server operations to avoid race
	origOpen := openBrowserFunc
	openBrowserFunc = func(url string) error {
		urlCh <- url
		close(browserCalled)
		return nil
	}
	t.Cleanup(func() {
		// Wait for browser goroutine to complete before restoring
		select {
		case <-browserCalled:
		case <-time.After(100 * time.Millisecond):
		}
		openBrowserFunc = origOpen
	})

	store := secrets.NewMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	server.SetValidator(func(ctx context.Context, install, geo, token string) error {
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resultCh := make(chan *SetupResult, 1)
	errCh := make(chan error, 1)
	go func() {
		res, startErr := server.Start(ctx)
		if startErr != nil {
			errCh <- startErr
			return
		}
		resultCh <- res
	}()

	var baseURL string
	select {
	case baseURL = <-urlCh:
	case <-ctx.Done():
		t.Fatal("timed out waiting for server url")
	}

	submitBody := `{"install":"acme","geo":"au","token":"test-token"}`
	req, err := http.NewRequest(http.MethodPost, baseURL+"/submit", strings.NewReader(submitBody))
	if err != nil {
		t.Fatalf("failed to create submit request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("submit request failed: %v", err)
	}
	_ = resp.Body.Close()

	completeReq, err := http.NewRequest(http.MethodPost, baseURL+"/complete", nil)
	if err != nil {
		t.Fatalf("failed to create complete request: %v", err)
	}
	completeReq.Header.Set("X-CSRF-Token", server.csrfToken)
	completeResp, err := http.DefaultClient.Do(completeReq)
	if err != nil {
		t.Fatalf("complete request failed: %v", err)
	}
	_ = completeResp.Body.Close()

	select {
	case res := <-resultCh:
		if res == nil {
			t.Fatal("expected result, got nil")
		}
		if res.Install != "acme" || res.Geo != "au" {
			t.Fatalf("unexpected result: %+v", res)
		}
	case startErr := <-errCh:
		t.Fatalf("start failed: %v", startErr)
	case <-ctx.Done():
		t.Fatal("timed out waiting for result")
	}

	creds, err := store.Get()
	if err != nil {
		t.Fatalf("expected creds stored: %v", err)
	}
	if creds.Install != "acme" || creds.Geo != "au" || creds.Token != "test-token" {
		t.Fatalf("unexpected creds: %+v", creds)
	}
}

func TestOpenBrowser_UsesStartCommand(t *testing.T) {
	expected := map[string]string{
		"darwin":  "open",
		"linux":   "xdg-open",
		"windows": "rundll32",
	}

	var gotName string
	var gotArgs []string
	origStart := startCommand
	startCommand = func(name string, args ...string) error {
		gotName = name
		gotArgs = append([]string{}, args...)
		return nil
	}
	defer func() { startCommand = origStart }()

	origGOOS := goos
	defer func() { goos = origGOOS }()

	for osName, cmd := range expected {
		gotName = ""
		gotArgs = nil
		goos = osName

		if err := openBrowser("http://example.com"); err != nil {
			t.Fatalf("openBrowser error for %s: %v", osName, err)
		}
		if gotName != cmd {
			t.Fatalf("expected command %q for %s, got %q", cmd, osName, gotName)
		}
		if len(gotArgs) == 0 || !strings.Contains(gotArgs[len(gotArgs)-1], "http://example.com") {
			t.Fatalf("expected url in args for %s, got %v", osName, gotArgs)
		}
	}

	goos = "plan9"
	if err := openBrowser("http://example.com"); err == nil {
		t.Fatalf("expected error for unsupported platform")
	}
}

func TestSetupServer_Start_ContextCanceled(t *testing.T) {
	// Track when browser func is called to ensure goroutine completes
	browserCalled := make(chan struct{})

	// Set up mock BEFORE any server operations to avoid race
	origOpen := openBrowserFunc
	openBrowserFunc = func(url string) error {
		close(browserCalled)
		return nil
	}
	t.Cleanup(func() {
		// Wait for browser goroutine to complete before restoring
		select {
		case <-browserCalled:
		case <-time.After(100 * time.Millisecond):
		}
		openBrowserFunc = origOpen
	})

	store := secrets.NewMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	server.SetValidator(func(ctx context.Context, install, geo, token string) error {
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = server.Start(ctx)
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestSetupServer_Start_ShutdownPendingResult(t *testing.T) {
	// Track when browser func is called to ensure goroutine completes
	browserCalled := make(chan struct{})

	// Set up mock BEFORE any server operations to avoid race
	origOpen := openBrowserFunc
	openBrowserFunc = func(url string) error {
		close(browserCalled)
		return nil
	}
	t.Cleanup(func() {
		// Wait for browser goroutine to complete before restoring
		select {
		case <-browserCalled:
		case <-time.After(100 * time.Millisecond):
		}
		openBrowserFunc = origOpen
	})

	store := secrets.NewMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	server.SetValidator(func(ctx context.Context, install, geo, token string) error {
		return nil
	})

	resultCh := make(chan *SetupResult, 1)
	errCh := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		res, startErr := server.Start(ctx)
		if startErr != nil {
			errCh <- startErr
			return
		}
		resultCh <- res
	}()

	// Wait for browser goroutine to have read openBrowserFunc before triggering shutdown
	select {
	case <-browserCalled:
	case <-time.After(100 * time.Millisecond):
	}

	server.pendingMu.Lock()
	server.pendingResult = &SetupResult{Install: "acme", Geo: "uk"}
	server.pendingMu.Unlock()
	close(server.shutdown)

	select {
	case res := <-resultCh:
		if res == nil || res.Install != "acme" || res.Geo != "uk" {
			t.Fatalf("unexpected result: %+v", res)
		}
	case startErr := <-errCh:
		t.Fatalf("start failed: %v", startErr)
	case <-ctx.Done():
		t.Fatal("timed out waiting for result")
	}
}

type errResponseWriter struct {
	header http.Header
}

func (w *errResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = http.Header{}
	}
	return w.header
}

func (w *errResponseWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("write failed")
}

func (w *errResponseWriter) WriteHeader(statusCode int) {}

func TestWriteJSON_ErrorPath(t *testing.T) {
	w := &errResponseWriter{}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type authTestServerTransport struct {
	testServerURL string
	underlying    http.RoundTripper
}

func (t *authTestServerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	testURL := t.testServerURL + req.URL.Path
	if req.URL.RawQuery != "" {
		testURL += "?" + req.URL.RawQuery
	}
	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, testURL, req.Body)
	if err != nil {
		return nil, err
	}
	newReq.Header = req.Header
	return t.underlying.RoundTrip(newReq)
}

func newAuthTestClient(serverURL string, creds *secrets.Credentials) *api.Client {
	client := api.NewClient(creds)
	client.SetHTTPClient(&http.Client{
		Transport: &authTestServerTransport{
			testServerURL: serverURL,
			underlying:    http.DefaultTransport,
		},
	})
	return client
}

func TestValidateCredentials_SuccessAndFailure(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"Name": "Test User"})
		}))
		defer server.Close()

		oldNewClient := newAPIClient
		newAPIClient = func(creds *secrets.Credentials) *api.Client {
			return newAuthTestClient(server.URL, creds)
		}
		defer func() { newAPIClient = oldNewClient }()

		store := secrets.NewMockStore()
		setup, err := NewSetupServer(store)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = setup.validateCredentials(context.Background(), "acme", "au", "token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		oldNewClient := newAPIClient
		newAPIClient = func(creds *secrets.Credentials) *api.Client {
			return newAuthTestClient(server.URL, creds)
		}
		defer func() { newAPIClient = oldNewClient }()

		store := secrets.NewMockStore()
		setup, err := NewSetupServer(store)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = setup.validateCredentials(context.Background(), "acme", "au", "token")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("missing values", func(t *testing.T) {
		store := secrets.NewMockStore()
		setup, err := NewSetupServer(store)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = setup.validateCredentials(context.Background(), "", "au", "")
		if err == nil {
			t.Fatal("expected error for missing values")
		}
	})
}

// =============================================================================
// Testability Notes
// =============================================================================
//
// The following handlers have limited testability due to external dependencies:
//
// 1. handleValidate and handleSubmit:
//    - Both call s.validateCredentials() which makes a real HTTP request
//      to the Deputy API via api.Client.Me()
//    - To fully test these endpoints (including successful API validation),
//      we would need to:
//      a) Extract the API validation into an interface that can be mocked
//      b) Inject a mock API client into SetupServer
//      c) Or use a test HTTP server to mock the Deputy API
//
// 2. Start():
//    - Opens a browser via openBrowser() which uses platform-specific commands
//    - Creates a real TCP listener
//    - Would require refactoring to accept interfaces for browser opening
//      and listener creation
//
// Current tests cover:
// - CSRF token validation
// - Rate limiting (directly and via handlers)
// - Input validation (install, geo, token)
// - HTTP method validation
// - JSON parsing errors
// - Security headers
// - HTML template rendering
//
