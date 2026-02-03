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
)

/*
TESTABILITY LIMITATIONS

The rosters commands cannot be fully tested with mock API clients because they
use getClient() which reads credentials from the system keychain:

    client, err := getClient()

GOOD NEWS: Some validations happen BEFORE getClient() is called:
- rosters get: ID parsing validation (strconv.Atoi before getClient)
- rosters create: --employee, --opunit, --start-time, --end-time validation
- rosters copy: --from-date, --to-date, --location validation
- rosters publish: --from-date, --to-date, --location validation
- rosters discard: --from-date, --to-date, --location validation
- rosters swap: ID parsing validation (strconv.Atoi before getClient)

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

// TestRostersCommand_ViaRootCmd verifies the rosters command is properly registered
func TestRostersCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Manage rosters/shifts")
}

// TestRostersCommand_HasAllSubcommands verifies all expected subcommands are registered
func TestRostersCommand_HasAllSubcommands(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()

	expectedSubcommands := []string{
		"list",
		"get",
		"create",
		"copy",
		"publish",
		"discard",
		"swap",
	}

	for _, sub := range expectedSubcommands {
		assert.Contains(t, output, sub, "missing subcommand: %s", sub)
	}
}

// TestRostersCommand_Aliases verifies the command aliases work
func TestRostersCommand_Aliases(t *testing.T) {
	aliases := []string{"roster", "shifts", "shift", "r"}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			root := NewRootCmd()
			buf := &bytes.Buffer{}
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs([]string{alias, "--help"})

			err := root.Execute()

			require.NoError(t, err)
			assert.Contains(t, buf.String(), "Manage rosters/shifts")
		})
	}
}

// TestRostersListCommand verifies the list command is registered
func TestRostersListCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "list", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List rosters")
}

// TestRostersListCommand_RequiresAuth tests that list fails without credentials.
// SKIP: Requires client injection - no pre-client validation in list command.
func TestRostersListCommand_RequiresAuth(t *testing.T) {
	t.Skip("Requires refactoring: list command has no pre-client validation, cannot test without mock client")
}

// TestRostersGetCommand verifies the get command is registered with proper args
func TestRostersGetCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "get", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Get roster details")
	assert.Contains(t, output, "<id>")
}

// TestRostersGetCommand_RequiresIDArgument tests that get requires an ID
func TestRostersGetCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "get"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestRostersGetCommand_InvalidID tests that get validates the ID is numeric.
// This validation happens BEFORE getClient(), so we can test it!
func TestRostersGetCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "get", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid roster ID")
}

// TestRostersCreateCommand verifies the create command has all required flags
func TestRostersCreateCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "create", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Create a new roster/shift")
	assert.Contains(t, output, "--employee")
	assert.Contains(t, output, "--opunit")
	assert.Contains(t, output, "--start-time")
	assert.Contains(t, output, "--end-time")
	assert.Contains(t, output, "--mealbreak")
	assert.Contains(t, output, "--comment")
	assert.Contains(t, output, "--open")
	assert.Contains(t, output, "--publish")
}

// TestRostersCreateCommand_RequiresEmployeeFlag tests that create requires --employee.
// This validation happens BEFORE getClient(), so we can test it!
func TestRostersCreateCommand_RequiresEmployeeFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "create", "--opunit", "1", "--start-time", "1234567890", "--end-time", "1234571490"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--employee is required")
}

// TestRostersCreateCommand_RequiresOpunitFlag tests that create requires --opunit.
// This validation happens BEFORE getClient(), so we can test it!
func TestRostersCreateCommand_RequiresOpunitFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "create", "--employee", "1", "--start-time", "1234567890", "--end-time", "1234571490"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--opunit is required")
}

// TestRostersCreateCommand_RequiresTimeFlags tests that create requires --start-time and --end-time.
// This validation happens BEFORE getClient(), so we can test it!
func TestRostersCreateCommand_RequiresTimeFlags(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "create", "--employee", "1", "--opunit", "1"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--start-time and --end-time are required")
}

// TestRostersCopyCommand verifies the copy command has all required flags
func TestRostersCopyCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "copy", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Copy roster from one week to another")
	assert.Contains(t, output, "--from-date")
	assert.Contains(t, output, "--to-date")
	assert.Contains(t, output, "--location")
}

// TestRostersCopyCommand_RequiresDateFlags tests that copy requires --from-date and --to-date.
// This validation happens BEFORE getClient(), so we can test it!
func TestRostersCopyCommand_RequiresDateFlags(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "copy", "--location", "1"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--from-date and --to-date are required")
}

func TestRostersCopyCommand_InvalidDate(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newRostersCopyCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--from-date", "invalid", "--to-date", "2024-01-08", "--location", "1"})
	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid date format")
}

// TestRostersCopyCommand_RequiresLocationFlag tests that copy requires --location.
// This validation happens BEFORE getClient(), so we can test it!
func TestRostersCopyCommand_RequiresLocationFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "copy", "--from-date", "2024-01-01", "--to-date", "2024-01-08"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--location is required")
}

// TestRostersPublishCommand verifies the publish command has all required flags
func TestRostersPublishCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "publish", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Publish rosters for a date range")
	assert.Contains(t, output, "--from-date")
	assert.Contains(t, output, "--to-date")
	assert.Contains(t, output, "--location")
}

// TestRostersPublishCommand_RequiresDateFlags tests that publish requires --from-date and --to-date.
// This validation happens BEFORE getClient(), so we can test it!
func TestRostersPublishCommand_RequiresDateFlags(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "publish", "--location", "1"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--from-date and --to-date are required")
}

func TestRostersPublishCommand_InvalidDate(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newRostersPublishCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--from-date", "2024-01-01", "--to-date", "invalid", "--location", "1"})
	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid date format")
}

// TestRostersPublishCommand_RequiresLocationFlag tests that publish requires --location.
// This validation happens BEFORE getClient(), so we can test it!
func TestRostersPublishCommand_RequiresLocationFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "publish", "--from-date", "2024-01-01", "--to-date", "2024-01-08"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--location is required")
}

// TestRostersDiscardCommand verifies the discard command has all required flags
func TestRostersDiscardCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "discard", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Discard unpublished roster changes")
	assert.Contains(t, output, "--from-date")
	assert.Contains(t, output, "--to-date")
	assert.Contains(t, output, "--location")
}

// TestRostersDiscardCommand_RequiresDateFlags tests that discard requires --from-date and --to-date.
// This validation happens BEFORE getClient(), so we can test it!
func TestRostersDiscardCommand_RequiresDateFlags(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "discard", "--location", "1"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--from-date and --to-date are required")
}

func TestRostersDiscardCommand_InvalidDate(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newRostersDiscardCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--from-date", "invalid", "--to-date", "2024-01-08", "--location", "1"})
	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid date format")
}

// TestRostersDiscardCommand_RequiresLocationFlag tests that discard requires --location.
// This validation happens BEFORE getClient(), so we can test it!
func TestRostersDiscardCommand_RequiresLocationFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "discard", "--from-date", "2024-01-01", "--to-date", "2024-01-08"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--location is required")
}

// TestRostersSwapCommand verifies the swap command is registered with proper args
func TestRostersSwapCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "swap", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List rosters that can be swapped with the specified roster")
	assert.Contains(t, output, "<roster-id>")
}

// TestRostersSwapCommand_RequiresIDArgument tests that swap requires an ID
func TestRostersSwapCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "swap"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestRostersSwapCommand_InvalidID tests that swap validates the ID is numeric.
// This validation happens BEFORE getClient(), so we can test it!
func TestRostersSwapCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"rosters", "swap", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid roster ID")
}

// TestRostersCreateCommand_AcceptsOptionalFlags verifies optional flags work.
// SKIP: Requires client injection - we can only verify flags exist.
func TestRostersCreateCommand_AcceptsOptionalFlags(t *testing.T) {
	t.Skip("Requires refactoring: cannot verify flag values are passed to API without mock client")

	// Expected behavior when refactored:
	// mockClient := api.NewMockClient()
	// mockClient.Rosters().SetCreateResponse(&api.Roster{Id: 1, Employee: 123})
	//
	// ctx := api.WithClient(context.Background(), mockClient)
	// cmd := newRostersCreateCmd()
	// cmd.SetContext(ctx)
	// cmd.SetArgs([]string{
	//     "--employee", "123",
	//     "--opunit", "456",
	//     "--start-time", "1234567890",
	//     "--end-time", "1234571490",
	//     "--mealbreak", "30",
	//     "--comment", "Test shift",
	//     "--open",
	//     "--publish",
	// })
	// err := cmd.Execute()
	//
	// require.NoError(t, err)
	// input := mockClient.Rosters().LastCreateInput()
	// assert.Equal(t, 123, input.Employee)
	// assert.Equal(t, 456, input.OperationalUnit)
	// assert.Equal(t, "30", input.Mealbreak)
	// assert.Equal(t, "Test shift", input.Comment)
	// assert.True(t, input.Open)
	// assert.True(t, input.Publish)
}

// TestRostersCommand_WithMockClient tests command output using mock HTTP server.
func TestRostersCommand_WithMockClient(t *testing.T) {
	t.Run("list returns rosters table", func(t *testing.T) {
		// Create mock server that returns roster list
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Roster{
				{Id: 1, Employee: 123, Date: "2024-01-15", StartTime: 1705326000, EndTime: 1705354800, Published: true},
				{Id: 2, Employee: 456, Date: "2024-01-16", StartTime: 1705412400, EndTime: 1705441200, Published: false},
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
		cmd := newRostersListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		// Verify
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "2024-01-15")
		assert.Contains(t, output, "2024-01-16")
		assert.Contains(t, output, "123")
		assert.Contains(t, output, "456")
	})
}
