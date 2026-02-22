package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
)

/*
TESTABILITY

The me commands use getClientFromContext() which supports dependency injection
via WithClientFactory(). This allows full testing with mock HTTP servers.

Tests verify:
- Command structure and registration
- Subcommand availability
- Help text and usage strings
- API responses via mock client (see TestMeCommand_WithMockClient)
*/

// TestMeCommand_ViaRootCmd verifies the me command is properly registered
func TestMeCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"me", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Commands for the current user")
}

// TestMeCommand_HasAllSubcommands verifies all expected subcommands are registered
func TestMeCommand_HasAllSubcommands(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"me", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()

	expectedSubcommands := []string{
		"info",
		"timesheets",
		"rosters",
		"leave",
	}

	for _, sub := range expectedSubcommands {
		assert.Contains(t, output, sub, "missing subcommand: %s", sub)
	}
}

// TestMeCommand_HasExpectedSubcommands tests using newMeCmd directly
func TestMeCommand_HasExpectedSubcommands(t *testing.T) {
	cmd := newMeCmd()
	subCmds := cmd.Commands()
	names := make([]string, len(subCmds))
	for i, c := range subCmds {
		names[i] = c.Name()
	}
	assert.Contains(t, names, "info")
	assert.Contains(t, names, "timesheets")
	assert.Contains(t, names, "rosters")
	assert.Contains(t, names, "leave")
}

// TestMeInfoCommand verifies the info command is registered
func TestMeInfoCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"me", "info", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Show current user info")
}

// TestMeTimesheetsCommand verifies the timesheets command is registered
func TestMeTimesheetsCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"me", "timesheets", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List my timesheets")
	assert.Contains(t, output, "--limit")
	assert.Contains(t, output, "--offset")
}

// TestMeRostersCommand verifies the rosters command is registered
func TestMeRostersCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"me", "rosters", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List my rosters")
	assert.Contains(t, output, "--limit")
	assert.Contains(t, output, "--offset")
}

// TestMeLeaveCommand verifies the leave command is registered
func TestMeLeaveCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"me", "leave", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List my leave requests")
	assert.Contains(t, output, "--limit")
	assert.Contains(t, output, "--offset")
}

// TestMeCommand_WithMockClient tests command output using mock HTTP server.
func TestMeCommand_WithMockClient(t *testing.T) {
	t.Run("info returns user details", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			// Simulate Deputy API response with PascalCase field names
			_, _ = w.Write([]byte(`{
				"UserId": 1,
				"EmployeeId": 2,
				"Login": "testuser",
				"Name": "Test User",
				"FirstName": "Test",
				"LastName": "User",
				"PrimaryEmail": "test@example.com",
				"PrimaryPhone": "+1234567890",
				"Company": 100,
				"Portfolio": "Test Portfolio",
				"Role": 2
			}`))
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newMeInfoCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Test User")
		assert.Contains(t, output, "test@example.com")
		assert.Contains(t, output, "100") // Company
	})

	t.Run("info returns JSON output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			// Simulate Deputy API response with PascalCase field names
			_, _ = w.Write([]byte(`{
				"UserId": 1,
				"Name": "Test User",
				"PrimaryEmail": "test@example.com"
			}`))
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newMeInfoCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		// JSON output now uses snake_case field names
		assert.Contains(t, output, `"name": "Test User"`)
		assert.Contains(t, output, `"primary_email": "test@example.com"`)
		assert.Contains(t, output, `"id": 1`)
	})

	t.Run("timesheets returns timesheet table", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Timesheet{
				{Id: 1, Date: "2024-01-15", StartTime: 1705326000, EndTime: 1705354800, TotalTimeStr: "8:00"},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newMeTimesheetsCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "2024-01-15")
		assert.Contains(t, output, "8:00")
	})

	t.Run("timesheets returns json output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Timesheet{
				{Id: 2, Date: "2024-02-01"},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newMeTimesheetsCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Id": 2`)
	})

	t.Run("rosters returns roster table", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Roster{
				{Id: 1, Date: "2024-01-15", StartTime: 1705326000, EndTime: 1705354800},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newMeRostersCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "2024-01-15")
	})

	t.Run("rosters returns json output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Roster{
				{Id: 2, Date: "2024-02-01"},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newMeRostersCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Id": 2`)
	})

	t.Run("leave returns leave table", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Leave{
				{Id: 1, DateStart: "2024-01-15", DateEnd: "2024-01-16", Status: 1, Hours: 16.0},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newMeLeaveCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "2024-01-15")
		assert.Contains(t, output, "16.0")
	})

	t.Run("leave returns json output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Leave{
				{Id: 2, DateStart: "2024-02-01", DateEnd: "2024-02-02"},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newMeLeaveCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Id": 2`)
	})
}
