package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/salmonumbrella/deputy-cli/internal/secrets"
)

// ListOptions contains pagination options for list operations.
// The Deputy API uses max and start query parameters.
type ListOptions struct {
	Limit  int // max results to return (0 = unlimited)
	Offset int // start position for pagination
}

type Client struct {
	httpClient *http.Client
	creds      *secrets.Credentials
	debug      bool
}

func NewClient(creds *secrets.Credentials) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		creds: creds,
	}
}

func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

// SetHTTPClient sets a custom HTTP client (useful for testing)
func (c *Client) SetHTTPClient(httpClient *http.Client) {
	if httpClient == nil {
		return
	}
	c.httpClient = httpClient
}

// maxErrorBodyLen is the maximum number of characters to include from an error response body in debug mode.
const maxErrorBodyLen = 500

// sanitizeErrorResponse creates an error message from an API error response.
// It first attempts to parse structured API errors with the format {"error":{"code":N,"message":"..."}}.
// In debug mode with non-structured errors, it includes the response body (truncated to maxErrorBodyLen characters).
// Otherwise, it returns a generic message based on the status code.
func sanitizeErrorResponse(statusCode int, body []byte, debug bool) error {
	code := CodeFromStatus(statusCode)
	retryable := IsRetryable(statusCode)

	var message string

	// Try to parse common JSON error shapes without being strict about types.
	// Deputy (and proxies) can vary response schemas across endpoints.
	var structuredAny map[string]any
	if err := json.Unmarshal(body, &structuredAny); err == nil {
		if errObj, ok := structuredAny["error"].(map[string]any); ok {
			if msg, ok := errObj["message"].(string); ok && strings.TrimSpace(msg) != "" {
				message = msg
			} else if msg, ok := structuredAny["message"].(string); ok && strings.TrimSpace(msg) != "" {
				message = msg
			}
		} else if msg, ok := structuredAny["message"].(string); ok && strings.TrimSpace(msg) != "" {
			message = msg
		}
	}

	if message == "" && debug && len(body) > 0 {
		bodyStr := string(body)
		if len(bodyStr) > maxErrorBodyLen {
			bodyStr = bodyStr[:maxErrorBodyLen] + "..."
		}
		message = bodyStr
	}

	if message == "" {
		// Fall back to generic messages
		genericMessages := map[int]string{
			400: "bad request",
			401: "unauthorized",
			403: "forbidden",
			404: "not found",
			405: "method not allowed",
			409: "conflict",
			422: "unprocessable entity",
			429: "too many requests",
			500: "server error",
			502: "bad gateway",
			503: "service unavailable",
			504: "gateway timeout",
		}
		if msg, ok := genericMessages[statusCode]; ok {
			message = msg
		} else if statusCode >= 500 {
			message = "server error"
		} else {
			message = "request failed"
		}
	}

	return &APIError{
		Code:       code,
		StatusCode: statusCode,
		Message:    message,
		Retryable:  retryable,
	}
}

// buildURL constructs a URL with optional query parameters for pagination.
func (c *Client) buildURL(path string, opts *ListOptions) string {
	url := c.creds.BaseURL() + path
	if opts == nil {
		return url
	}

	params := make([]string, 0, 2)
	if opts.Limit > 0 {
		params = append(params, fmt.Sprintf("max=%d", opts.Limit))
	}
	if opts.Offset > 0 {
		params = append(params, fmt.Sprintf("start=%d", opts.Offset))
	}

	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}
	return url
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, result any) error {
	return c.doWithOpts(ctx, method, path, body, result, nil)
}

func (c *Client) doWithOpts(ctx context.Context, method, path string, body io.Reader, result any, opts *ListOptions) error {
	return c.doRequest(ctx, method, c.buildURL(path, opts), body, result)
}

func (c *Client) doV2(ctx context.Context, method, path string, body io.Reader, result any) error {
	return c.doRequest(ctx, method, c.creds.BaseURLV2()+path, body, result)
}

// doRequest executes an HTTP request and decodes the response.
// Shared by doWithOpts (v1) and doV2 to keep error handling and header logic in one place.
func (c *Client) doRequest(ctx context.Context, method, url string, body io.Reader, result any) error {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", c.creds.AuthorizationHeaderValue())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		apiErr := sanitizeErrorResponse(resp.StatusCode, respBody, c.debug)
		if c.debug {
			return fmt.Errorf("%s %s: %w", method, url, apiErr)
		}
		return apiErr
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}
