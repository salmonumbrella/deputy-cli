package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/salmonumbrella/deputy-cli/internal/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testServerTransport redirects all requests to a test server
type testServerTransport struct {
	testServerURL string
	underlying    http.RoundTripper
}

func (t *testServerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Replace the URL with test server URL, keeping the path and query string
	testURL := t.testServerURL + req.URL.Path
	if req.URL.RawQuery != "" {
		testURL += "?" + req.URL.RawQuery
	}
	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, testURL, req.Body)
	if err != nil {
		return nil, err
	}
	// Copy headers
	newReq.Header = req.Header
	return t.underlying.RoundTrip(newReq)
}

// newTestClient creates a client configured to use a test server
func newTestClient(serverURL, token string) *Client {
	creds := &secrets.Credentials{
		Token:   token,
		Install: "test",
		Geo:     "au",
	}
	client := NewClient(creds)
	client.httpClient = &http.Client{
		Transport: &testServerTransport{
			testServerURL: serverURL,
			underlying:    http.DefaultTransport,
		},
	}
	return client
}

func TestNewClient(t *testing.T) {
	creds := &secrets.Credentials{
		Token:   "test-token",
		Install: "mycompany",
		Geo:     "au",
	}

	client := NewClient(creds)

	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, creds, client.creds)
}

func TestNewClient_HTTPClientTimeout(t *testing.T) {
	creds := &secrets.Credentials{
		Token:   "test-token",
		Install: "mycompany",
		Geo:     "au",
	}

	client := NewClient(creds)

	assert.NotNil(t, client.httpClient)
	// Verify timeout is set (30 seconds as per implementation)
	assert.Equal(t, 30*1000*1000*1000, int(client.httpClient.Timeout))
}

func TestClient_SetDebug(t *testing.T) {
	creds := &secrets.Credentials{
		Token:   "test-token",
		Install: "mycompany",
		Geo:     "au",
	}

	client := NewClient(creds)

	assert.False(t, client.debug)
	client.SetDebug(true)
	assert.True(t, client.debug)
	client.SetDebug(false)
	assert.False(t, client.debug)
}

func TestClient_Me(t *testing.T) {
	// Create a test server that simulates the Deputy API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request path ends with /me
		assert.True(t, strings.HasSuffix(r.URL.Path, "/me"), "Expected path to end with /me, got %s", r.URL.Path)

		// Verify authorization header
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		w.Header().Set("Content-Type", "application/json")
		// Simulate Deputy API response with PascalCase field names
		_, _ = w.Write([]byte(`{
			"UserId": 1,
			"Name": "Test User",
			"PrimaryEmail": "test@example.com",
			"Company": 42
		}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	user, err := client.Me().Info(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, user.UserId)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "test@example.com", user.PrimaryEmail)
	assert.Equal(t, 42, user.Company)
}

func TestClient_Me_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")

	_, err := client.Me().Info(context.Background())
	assert.Error(t, err)
	// In non-debug mode, should return sanitized error without body content
	assert.Equal(t, "API error 401: unauthorized", err.Error())
}

func TestClient_Me_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Me().Info(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 500")
}

func TestClient_Me_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Me().Info(context.Background())
	assert.Error(t, err)
}

func TestClient_Me_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Me().Info(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestClient_Me_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Access denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Me().Info(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestSanitizeErrorResponse_NonDebugMode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		expected   string
	}{
		{
			name:       "401 unauthorized",
			statusCode: 401,
			body:       []byte(`{"error": "Invalid token", "details": "Token xyz123 has expired"}`),
			expected:   "API error 401: unauthorized",
		},
		{
			name:       "404 not found",
			statusCode: 404,
			body:       []byte(`{"error": "Resource not found", "id": 12345}`),
			expected:   "API error 404: not found",
		},
		{
			name:       "500 server error",
			statusCode: 500,
			body:       []byte(`{"error": "Internal server error", "stack": "at someFunction..."}`),
			expected:   "API error 500: server error",
		},
		{
			name:       "400 bad request",
			statusCode: 400,
			body:       []byte(`{"error": "Validation failed", "fields": {"email": "invalid format"}}`),
			expected:   "API error 400: bad request",
		},
		{
			name:       "403 forbidden",
			statusCode: 403,
			body:       []byte(`{"error": "Access denied"}`),
			expected:   "API error 403: forbidden",
		},
		{
			name:       "429 too many requests",
			statusCode: 429,
			body:       []byte(`{"error": "Rate limit exceeded"}`),
			expected:   "API error 429: too many requests",
		},
		{
			name:       "unmapped 5xx error",
			statusCode: 599,
			body:       []byte(`{"error": "Unknown server error"}`),
			expected:   "API error 599: server error",
		},
		{
			name:       "unmapped 4xx error",
			statusCode: 418,
			body:       []byte(`{"error": "I'm a teapot"}`),
			expected:   "API error 418: request failed",
		},
		{
			name:       "empty body",
			statusCode: 401,
			body:       []byte{},
			expected:   "API error 401: unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizeErrorResponse(tt.statusCode, tt.body, false)
			assert.EqualError(t, err, tt.expected)
		})
	}
}

func TestSanitizeErrorResponse_DebugMode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		contains   []string
	}{
		{
			name:       "includes full body in debug mode",
			statusCode: 401,
			body:       []byte(`{"error": "Invalid token", "details": "Token xyz123 has expired"}`),
			contains:   []string{"API error 401:", "Invalid token", "xyz123"},
		},
		{
			name:       "includes sensitive data in debug mode",
			statusCode: 500,
			body:       []byte(`{"error": "Database error", "query": "SELECT * FROM users WHERE id=123"}`),
			contains:   []string{"API error 500:", "Database error", "SELECT * FROM users"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizeErrorResponse(tt.statusCode, tt.body, true)
			for _, substr := range tt.contains {
				assert.Contains(t, err.Error(), substr)
			}
		})
	}
}

func TestSanitizeErrorResponse_DebugModeTruncation(t *testing.T) {
	// Create a body longer than maxErrorBodyLen (500 chars)
	longBody := make([]byte, 600)
	for i := range longBody {
		longBody[i] = 'x'
	}

	err := sanitizeErrorResponse(500, longBody, true)

	// Should be truncated and end with "..."
	errStr := err.Error()
	assert.Contains(t, errStr, "API error 500:")
	assert.Contains(t, errStr, "...")
	// The error message should contain the prefix + 500 chars + "..."
	// "API error 500: " = 15 chars, body = 500 chars, "..." = 3 chars = 518 total
	assert.LessOrEqual(t, len(errStr), 520)
}

func TestSanitizeErrorResponse_DebugModeEmptyBody(t *testing.T) {
	// Empty body in debug mode should fall back to generic message
	err := sanitizeErrorResponse(401, []byte{}, true)
	assert.EqualError(t, err, "API error 401: unauthorized")
}

func TestClient_Me_Error_SanitizedInNonDebugMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token", "token": "secret-token-value"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	// Ensure debug mode is off (default)
	client.SetDebug(false)

	_, err := client.Me().Info(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "API error 401: unauthorized", err.Error())
	// Should NOT contain the sensitive token
	assert.NotContains(t, err.Error(), "secret-token-value")
}

func TestClient_Me_Error_FullBodyInDebugMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token", "details": "Token abc123 expired"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	client.SetDebug(true)

	_, err := client.Me().Info(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401:")
	assert.Contains(t, err.Error(), "Invalid token")
	assert.Contains(t, err.Error(), "abc123")
}

func TestSanitizeErrorResponse_PopulatesCodeAndRetryable(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedCode  string
		expectedRetry bool
	}{
		{"401 sets auth_required and not retryable", 401, ErrCodeAuthRequired, false},
		{"404 sets not_found and not retryable", 404, ErrCodeNotFound, false},
		{"429 sets rate_limited and retryable", 429, ErrCodeRateLimited, true},
		{"500 sets server_error and retryable", 500, ErrCodeServerError, true},
		{"503 sets server_error and retryable", 503, ErrCodeServerError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizeErrorResponse(tt.statusCode, nil, false)
			apiErr, ok := err.(*APIError)
			require.True(t, ok, "expected *APIError")
			assert.Equal(t, tt.expectedCode, apiErr.Code)
			assert.Equal(t, tt.expectedRetry, apiErr.Retryable)
		})
	}
}

func TestSanitizeErrorResponse_ParsesStructuredError(t *testing.T) {
	body := []byte(`{"error":{"code":400,"message":"Invalid search field Employee"}}`)
	err := sanitizeErrorResponse(400, body, false)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Contains(t, apiErr.Message, "Invalid search field Employee")
}

func TestSanitizeErrorResponse_StructuredErrorVariants(t *testing.T) {
	tests := []struct {
		name            string
		statusCode      int
		body            []byte
		debug           bool
		expectedMessage string
	}{
		{
			name:            "structured error takes priority over debug mode",
			statusCode:      400,
			body:            []byte(`{"error":{"code":400,"message":"Field validation failed"}}`),
			debug:           true,
			expectedMessage: "Field validation failed",
		},
		{
			name:            "structured error with different status code",
			statusCode:      422,
			body:            []byte(`{"error":{"code":422,"message":"Missing required parameter: employee_id"}}`),
			debug:           false,
			expectedMessage: "Missing required parameter: employee_id",
		},
		{
			name:            "falls back to generic when error message is empty",
			statusCode:      400,
			body:            []byte(`{"error":{"code":400,"message":""}}`),
			debug:           false,
			expectedMessage: "bad request",
		},
		{
			name:            "falls back to generic on malformed JSON",
			statusCode:      400,
			body:            []byte(`{malformed json`),
			debug:           false,
			expectedMessage: "bad request",
		},
		{
			name:            "falls back to generic when error object missing",
			statusCode:      401,
			body:            []byte(`{"status":"error","reason":"unauthorized"}`),
			debug:           false,
			expectedMessage: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizeErrorResponse(tt.statusCode, tt.body, tt.debug)
			var apiErr *APIError
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, tt.expectedMessage, apiErr.Message)
		})
	}
}
