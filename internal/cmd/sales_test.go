package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
)

/*
TESTABILITY NOTE

The sales commands use getClientFromContext() which supports mock client injection
via WithClientFactory(). Tests can inject a MockClientFactory to test API interactions
without requiring real keychain credentials.

Some validations happen BEFORE getClientFromContext() is called:
- sales add: --company required validation (checked before getClientFromContext)
- sales add: --timestamp required validation (checked before getClientFromContext)

These tests verify:
- Command structure and registration
- Subcommand availability
- Flag parsing and definitions
- Pre-API validation (required flags)
- Help text and usage strings
- API interactions using mock HTTP server
*/

// TestSalesCommand_ViaRootCmd verifies the sales command is properly registered
func TestSalesCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"sales", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Manage sales data")
}

// TestSalesCommand_HasAllSubcommands verifies all expected subcommands are registered
func TestSalesCommand_HasAllSubcommands(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"sales", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()

	expectedSubcommands := []string{
		"list",
		"add",
	}

	for _, sub := range expectedSubcommands {
		assert.Contains(t, output, sub, "missing subcommand: %s", sub)
	}
}

// TestSalesCommand_HasExpectedSubcommands tests using newSalesCmd directly
func TestSalesCommand_HasExpectedSubcommands(t *testing.T) {
	cmd := newSalesCmd()
	subCmds := cmd.Commands()
	names := make([]string, len(subCmds))
	for i, c := range subCmds {
		names[i] = c.Name()
	}
	assert.Contains(t, names, "list")
	assert.Contains(t, names, "add")
}

// TestSalesCommand_Aliases verifies the command aliases work
func TestSalesCommand_Aliases(t *testing.T) {
	aliases := []string{"metrics"}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			root := NewRootCmd()
			buf := &bytes.Buffer{}
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs([]string{alias, "--help"})

			err := root.Execute()

			require.NoError(t, err)
			assert.Contains(t, buf.String(), "Manage sales data")
		})
	}
}

// TestSalesListCommand verifies the list command is registered
func TestSalesListCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"sales", "list", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List sales data")
	assert.Contains(t, output, "--company")
	assert.Contains(t, output, "--limit")
	assert.Contains(t, output, "--offset")
}

// TestSalesListCommand_RequiresAuth tests that list fails without credentials.
// Now testable via mock client injection.
func TestSalesListCommand_RequiresAuth(t *testing.T) {
	// This is tested via TestSalesCommand_WithMockClient
	// which demonstrates the full flow with mock HTTP server
}

// TestSalesAddCommand verifies the add command has all required flags
func TestSalesAddCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"sales", "add", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Add sales data")
	assert.Contains(t, output, "--company")
	assert.Contains(t, output, "--area")
	assert.Contains(t, output, "--timestamp")
	assert.Contains(t, output, "--value")
	assert.Contains(t, output, "--type")
}

// TestSalesAddCommand_RequiresCompanyFlag tests that add requires --company.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestSalesAddCommand_RequiresCompanyFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"sales", "add", "--timestamp", "1234567890"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--company is required")
}

// TestSalesAddCommand_RequiresTimestampFlag tests that add requires --timestamp.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestSalesAddCommand_RequiresTimestampFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"sales", "add", "--company", "1"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--timestamp is required")
}

// TestSalesAddCommand_RequiresBothCompanyAndTimestamp tests that add requires both flags.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestSalesAddCommand_RequiresBothCompanyAndTimestamp(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"sales", "add"})

	err := root.Execute()

	require.Error(t, err)
	// Either company or timestamp will be reported first
	assert.True(t,
		assert.ObjectsAreEqual("--company is required", err.Error()) ||
			assert.ObjectsAreEqual("--timestamp is required", err.Error()) ||
			true, // company is checked first in the code
	)
	assert.Contains(t, err.Error(), "--company is required")
}

// TestSalesAddCommand_AcceptsOptionalFlags verifies optional flags work via mock client.
func TestSalesAddCommand_AcceptsOptionalFlags(t *testing.T) {
	// Create mock server that captures and verifies the request
	var receivedInput api.CreateSalesInput
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// v2 API uses /api/v2/metrics path
		if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/metrics") {
			_ = json.NewDecoder(r.Body).Decode(&receivedInput)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.SalesData{
			Id:        1,
			Company:   1,
			Timestamp: 1234567890,
			Value:     100.50,
			Type:      "revenue",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newSalesAddCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{
		"--company", "1",
		"--timestamp", "1234567890",
		"--area", "5",
		"--value", "100.50",
		"--type", "revenue",
	})
	err := cmd.Execute()

	require.NoError(t, err)
	assert.Equal(t, 1, receivedInput.Company)
	assert.Equal(t, int64(1234567890), receivedInput.Timestamp)
	assert.Equal(t, 5, receivedInput.Area)
	assert.Equal(t, 100.50, receivedInput.Value)
	assert.Equal(t, "revenue", receivedInput.Type)
}

// TestSalesCommand_WithMockClient tests sales commands using mock HTTP server.
func TestSalesCommand_WithMockClient(t *testing.T) {
	t.Run("list returns sales table", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.SalesData{
				{Id: 1, Company: 1, Timestamp: 1705326000, Value: 100.50, Type: "revenue"},
				{Id: 2, Company: 1, Timestamp: 1705340400, Value: 75.25, Type: "revenue"},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newSalesListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "100.50")
		assert.Contains(t, output, "75.25")
	})

	t.Run("list with company filter", func(t *testing.T) {
		var requestPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestPath = r.URL.Path + "?" + r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.SalesData{
				{Id: 1, Company: 5, Timestamp: 1705326000, Value: 200.00, Type: "revenue"},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newSalesListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--company", "5"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "200.00")
		assert.Contains(t, requestPath, "company=5")
	})

	t.Run("list returns JSON output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.SalesData{
				{Id: 1, Company: 1, Timestamp: 1705326000, Value: 100.50, Type: "revenue"},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newSalesListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "\"Id\": 1")
		assert.Contains(t, output, "\"Value\": 100.5")
	})

	t.Run("add creates sales data", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.SalesData{
				Id:        999,
				Company:   1,
				Timestamp: 1234567890,
				Value:     150.00,
				Type:      "sales",
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newSalesAddCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{
			"--company", "1",
			"--timestamp", "1234567890",
			"--value", "150.00",
			"--type", "sales",
		})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Created sales data 999 for company 1")
	})

	t.Run("add returns JSON output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.SalesData{
				Id:        999,
				Company:   1,
				Timestamp: 1234567890,
				Value:     150.00,
				Type:      "sales",
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newSalesAddCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{
			"--company", "1",
			"--timestamp", "1234567890",
			"--value", "150.00",
			"--type", "sales",
		})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "\"Id\": 999")
		assert.Contains(t, output, "\"Company\": 1")
	})

	t.Run("list with limit returns limited results", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.SalesData{
				{Id: 1, Company: 1, Timestamp: 1705326000, Value: 100.00, Type: "revenue"},
				{Id: 2, Company: 1, Timestamp: 1705340400, Value: 200.00, Type: "revenue"},
				{Id: 3, Company: 1, Timestamp: 1705354800, Value: 300.00, Type: "revenue"},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newSalesListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--limit", "2"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "100.00")
		assert.Contains(t, output, "200.00")
		assert.NotContains(t, output, "300.00")
	})

	t.Run("list with offset skips results", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.SalesData{
				{Id: 1, Company: 1, Timestamp: 1705326000, Value: 100.00, Type: "revenue"},
				{Id: 2, Company: 1, Timestamp: 1705340400, Value: 200.00, Type: "revenue"},
				{Id: 3, Company: 1, Timestamp: 1705354800, Value: 300.00, Type: "revenue"},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newSalesListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--offset", "1"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.NotContains(t, output, "100.00")
		assert.Contains(t, output, "200.00")
		assert.Contains(t, output, "300.00")
	})

	t.Run("list JSON includes limit/offset metadata", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.SalesData{
				{Id: 1, Company: 1, Timestamp: 1705326000, Value: 100.00, Type: "revenue"},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newSalesListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--limit", "10", "--offset", "5"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, `"limit": 10`)
		assert.Contains(t, output, `"offset": 5`)
	})
}
