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
TESTABILITY LIMITATIONS

The timesheets commands cannot be fully tested with mock API clients because they
use getClient() which reads credentials from the system keychain:

    client, err := getClient()

GOOD NEWS: Some validations happen BEFORE getClient() is called:
- timesheets get: ID parsing validation
- timesheets clock-in: --employee required validation
- timesheets clock-out: --employee required validation
- timesheets break: --employee required validation

These can be tested by observing the error messages.

REFACTORING NEEDED FOR FULL TESTABILITY:
See employees_test.go for detailed refactoring options.

Until refactoring is done, these tests verify:
- Command structure and registration
- Subcommand availability
- Flag parsing and definitions
- Argument validation (count, type)
- Pre-API validation (required flags, ID parsing)
- Help text and usage strings
*/

// TestTimesheetsCommand_ViaRootCmd verifies the timesheets command is properly registered
func TestTimesheetsCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"timesheets", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Manage timesheets")
}

// TestTimesheetsCommand_HasAllSubcommands verifies all expected subcommands are registered
func TestTimesheetsCommand_HasAllSubcommands(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"timesheets", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()

	expectedSubcommands := []string{
		"list",
		"get",
		"update",
		"list-pay-rules",
		"select-pay-rule",
		"clock-in",
		"clock-out",
		"start-break",
		"end-break",
	}

	for _, sub := range expectedSubcommands {
		assert.Contains(t, output, sub, "missing subcommand: %s", sub)
	}
}

// TestTimesheetsCommand_Aliases verifies the command aliases work
func TestTimesheetsCommand_Aliases(t *testing.T) {
	aliases := []string{"timesheet", "ts", "t"}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			root := NewRootCmd()
			buf := &bytes.Buffer{}
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs([]string{alias, "--help"})

			err := root.Execute()

			require.NoError(t, err)
			assert.Contains(t, buf.String(), "Manage timesheets")
		})
	}
}

// TestTimesheetsListCommand verifies the list command is registered
func TestTimesheetsListCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"timesheets", "list", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List timesheets")
}

// TestTimesheetsListCommand_RequiresAuth tests that list fails without credentials.
// SKIP: Requires client injection - no pre-client validation in list command.
func TestTimesheetsListCommand_RequiresAuth(t *testing.T) {
	t.Skip("Requires refactoring: list command has no pre-client validation, cannot test without mock client")
}

// TestTimesheetsGetCommand verifies the get command is registered with proper args
func TestTimesheetsGetCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"timesheets", "get", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Get timesheet details")
	assert.Contains(t, output, "<id>")
}

// TestTimesheetsGetCommand_RequiresIDArgument tests that get requires an ID
func TestTimesheetsGetCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"timesheets", "get"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestTimesheetsGetCommand_InvalidID tests that get validates the ID is numeric.
// This validation happens BEFORE getClient(), so we can test it!
func TestTimesheetsGetCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"timesheets", "get", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid timesheet ID")
}

// TestTimesheetsClockInCommand verifies the clock-in command is registered with flags
func TestTimesheetsClockInCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"timesheets", "clock-in", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Clock in an employee")
	assert.Contains(t, output, "--employee")
	assert.Contains(t, output, "--opunit")
	assert.Contains(t, output, "--comment")
}

// TestTimesheetsClockInCommand_RequiresEmployeeFlag tests that clock-in requires --employee.
// This validation happens BEFORE getClient(), so we can test it!
func TestTimesheetsClockInCommand_RequiresEmployeeFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"timesheets", "clock-in"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--employee is required")
}

// TestTimesheetsClockInCommand_AcceptsOptionalFlags verifies optional flags work.
// SKIP: Requires client injection - we can only verify flags exist.
func TestTimesheetsClockInCommand_AcceptsOptionalFlags(t *testing.T) {
	t.Skip("Requires refactoring: cannot verify flag values are passed to API without mock client")

	// Expected behavior when refactored:
	// mockClient := api.NewMockClient()
	// mockClient.Timesheets().SetClockInResponse(&api.Timesheet{Id: 1, Employee: 123})
	//
	// ctx := api.WithClient(context.Background(), mockClient)
	// cmd := newTimesheetsClockInCmd()
	// cmd.SetContext(ctx)
	// cmd.SetArgs([]string{"--employee", "123", "--opunit", "456", "--comment", "Test"})
	// err := cmd.Execute()
	//
	// require.NoError(t, err)
	// input := mockClient.Timesheets().LastClockInInput()
	// assert.Equal(t, 123, input.Employee)
	// assert.Equal(t, 456, input.OperationalUnit)
	// assert.Equal(t, "Test", input.Comment)
}

// TestTimesheetsClockOutCommand verifies the clock-out command is registered
func TestTimesheetsClockOutCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"timesheets", "clock-out", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Clock out an employee")
	assert.Contains(t, output, "--employee")
	assert.Contains(t, output, "--comment")
}

// TestTimesheetsClockOutCommand_RequiresEmployeeFlag tests that clock-out requires --employee.
// This validation happens BEFORE getClient(), so we can test it!
func TestTimesheetsClockOutCommand_RequiresEmployeeFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"timesheets", "clock-out"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--employee is required")
}

// TestTimesheetsStartBreakCommand verifies the start-break command is registered
func TestTimesheetsStartBreakCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"timesheets", "start-break", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Start a break on an active timesheet")
	assert.Contains(t, output, "--timesheet")
	assert.Contains(t, output, "--employee")
}

// TestTimesheetsStartBreakCommand_RequiresTimesheetOrEmployee tests that start-break requires --timesheet or --employee.
// This validation happens BEFORE getClient(), so we can test it!
func TestTimesheetsStartBreakCommand_RequiresTimesheetOrEmployee(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"timesheets", "start-break"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--timesheet or --employee is required")
}

// TestTimesheetsEndBreakCommand verifies the end-break command is registered
func TestTimesheetsEndBreakCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"timesheets", "end-break", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "End a break on an active timesheet")
	assert.Contains(t, output, "--timesheet")
	assert.Contains(t, output, "--employee")
}

// TestTimesheetsEndBreakCommand_RequiresTimesheetOrEmployee tests that end-break requires --timesheet or --employee.
// This validation happens BEFORE getClient(), so we can test it!
func TestTimesheetsEndBreakCommand_RequiresTimesheetOrEmployee(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"timesheets", "end-break"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--timesheet or --employee is required")
}

// TestTimesheetsCommand_WithMockClient tests command output using mock HTTP server.
func TestTimesheetsCommand_WithMockClient(t *testing.T) {
	t.Run("list returns timesheets table", func(t *testing.T) {
		// Create mock server that returns timesheet list
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Timesheet{
				{Id: 1, Employee: 123, Date: "2024-01-15", StartTime: 1705326000, EndTime: 1705354800, TotalTimeStr: "8:00"},
				{Id: 2, Employee: 456, Date: "2024-01-16", StartTime: 1705412400, EndTime: 1705441200, TotalTimeStr: "8:00"},
			})
		}))
		defer server.Close()

		// Create client and factory
		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		// Set up context with factory and IO
		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		// Create and execute command
		cmd := newTimesheetsListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		// Verify
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "2024-01-15")
		assert.Contains(t, output, "2024-01-16")
		assert.Contains(t, output, "8:00")     // Total time
		assert.Contains(t, output, "Complete") // Status when EndTime > 0
	})

	t.Run("start-break outputs success message", func(t *testing.T) {
		// Create mock server that accepts pause request
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/supervise/timesheet/pause", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Create client and factory
		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		// Set up context with factory and IO
		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		// Create and execute command
		cmd := newTimesheetsStartBreakCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--employee", "123"})
		err := cmd.Execute()

		// Verify
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Break started for employee 123")
	})

	t.Run("start-break outputs JSON", func(t *testing.T) {
		// Create mock server that accepts pause request
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Create client and factory
		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		// Set up context with factory, IO, and JSON format
		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		// Create and execute command
		cmd := newTimesheetsStartBreakCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--employee", "123"})
		err := cmd.Execute()

		// Verify
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, `"status"`)
		assert.Contains(t, output, `"break_started"`)
		assert.Contains(t, output, `"employee"`)
	})

	t.Run("end-break outputs success message", func(t *testing.T) {
		// Create mock server that accepts resume request
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/supervise/timesheet/resume", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Create client and factory
		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		// Set up context with factory and IO
		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		// Create and execute command
		cmd := newTimesheetsEndBreakCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--employee", "456"})
		err := cmd.Execute()

		// Verify
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Break ended for employee 456")
	})

	t.Run("end-break outputs JSON", func(t *testing.T) {
		// Create mock server that accepts resume request
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Create client and factory
		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		// Set up context with factory, IO, and JSON format
		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		// Create and execute command
		cmd := newTimesheetsEndBreakCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--employee", "456"})
		err := cmd.Execute()

		// Verify
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, `"status"`)
		assert.Contains(t, output, `"break_ended"`)
		assert.Contains(t, output, `"employee"`)
	})
}
