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
TESTABILITY NOTES

Locations commands use getClientFromContext(), so tests can inject a mock client
via WithClientFactory. This file focuses on registration/flag validation; see
locations_commands_test.go for API-backed command execution coverage.
*/

// TestLocationsCommand_ViaRootCmd verifies the locations command is properly registered
func TestLocationsCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Manage locations")
}

// TestLocationsCommand_HasAllSubcommands verifies all expected subcommands are registered
func TestLocationsCommand_HasAllSubcommands(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()

	expectedSubcommands := []string{
		"list",
		"get",
		"add",
		"update",
		"archive",
		"delete",
		"settings",
	}

	for _, sub := range expectedSubcommands {
		assert.Contains(t, output, sub, "missing subcommand: %s", sub)
	}
}

// TestLocationsCommand_Aliases verifies the command aliases work
func TestLocationsCommand_Aliases(t *testing.T) {
	aliases := []string{"location", "loc"}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			root := NewRootCmd()
			buf := &bytes.Buffer{}
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs([]string{alias, "--help"})

			err := root.Execute()

			require.NoError(t, err)
			assert.Contains(t, buf.String(), "Manage locations")
		})
	}
}

// TestLocationsListCommand verifies the list command is registered
func TestLocationsListCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "list", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List all locations")
}

// TestLocationsListCommand_RequiresAuth tests that list fails without credentials.
// SKIP: Requires client injection - no pre-client validation in list command.
func TestLocationsListCommand_RequiresAuth(t *testing.T) {
	t.Skip("Requires refactoring: list command has no pre-client validation, cannot test without mock client")
}

// TestLocationsGetCommand verifies the get command is registered with proper args
func TestLocationsGetCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "get", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Get location details")
	assert.Contains(t, output, "<id>")
}

// TestLocationsGetCommand_RequiresIDArgument tests that get requires an ID
func TestLocationsGetCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "get"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestLocationsGetCommand_InvalidID tests that get validates the ID is numeric.
// This validation happens BEFORE getClient(), so we can test it!
func TestLocationsGetCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "get", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location ID")
}

// TestLocationsAddCommand verifies the add command has all required flags
func TestLocationsAddCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "add", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Add a new location")
	assert.Contains(t, output, "--name")
	assert.Contains(t, output, "--code")
	assert.Contains(t, output, "--address")
	assert.Contains(t, output, "--timezone")
}

// TestLocationsAddCommand_RequiresNameFlag tests that add requires --name.
// This validation happens BEFORE getClient(), so we can test it!
func TestLocationsAddCommand_RequiresNameFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "add"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--name is required")
}

// TestLocationsAddCommand_AcceptsOptionalFlags verifies optional flags exist.
// SKIP: Requires client injection - we can only verify flags exist.
func TestLocationsAddCommand_AcceptsOptionalFlags(t *testing.T) {
	t.Skip("Requires refactoring: cannot verify flag values are passed to API without mock client")

	// Expected behavior when refactored:
	// mockClient := api.NewMockClient()
	// mockClient.Locations().SetCreateResponse(&api.Location{Id: 1, CompanyName: "Test"})
	//
	// ctx := api.WithClient(context.Background(), mockClient)
	// cmd := newLocationsAddCmd()
	// cmd.SetContext(ctx)
	// cmd.SetArgs([]string{
	//     "--name", "Test Location",
	//     "--code", "TEST",
	//     "--address", "123 Main St",
	//     "--timezone", "America/New_York",
	// })
	// err := cmd.Execute()
	//
	// require.NoError(t, err)
	// input := mockClient.Locations().LastCreateInput()
	// assert.Equal(t, "Test Location", input.CompanyName)
	// assert.Equal(t, "TEST", input.Code)
	// assert.Equal(t, "123 Main St", input.Address)
	// assert.Equal(t, "America/New_York", input.Timezone)
}

// TestLocationsUpdateCommand verifies the update command is registered with proper args
func TestLocationsUpdateCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "update", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Update a location")
	assert.Contains(t, output, "<id>")
	assert.Contains(t, output, "--name")
	assert.Contains(t, output, "--code")
	assert.Contains(t, output, "--address")
	assert.Contains(t, output, "--timezone")
}

// TestLocationsUpdateCommand_RequiresIDArgument tests that update requires an ID
func TestLocationsUpdateCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "update"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestLocationsUpdateCommand_InvalidID tests that update validates the ID is numeric.
// This validation happens BEFORE getClient(), so we can test it!
func TestLocationsUpdateCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "update", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location ID")
}

// TestLocationsArchiveCommand verifies the archive command is registered with proper args
func TestLocationsArchiveCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "archive", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Archive a location")
	assert.Contains(t, output, "<id>")
}

// TestLocationsArchiveCommand_RequiresIDArgument tests that archive requires an ID
func TestLocationsArchiveCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "archive"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestLocationsArchiveCommand_InvalidID tests that archive validates the ID is numeric.
// This validation happens BEFORE getClient(), so we can test it!
func TestLocationsArchiveCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "archive", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location ID")
}

// TestLocationsDeleteCommand verifies the delete command is registered with proper args
func TestLocationsDeleteCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "delete", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Delete a location")
	assert.Contains(t, output, "<id>")
}

// TestLocationsDeleteCommand_RequiresIDArgument tests that delete requires an ID
func TestLocationsDeleteCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "delete"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestLocationsDeleteCommand_InvalidID tests that delete validates the ID is numeric.
// This validation happens BEFORE getClient(), so we can test it!
func TestLocationsDeleteCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "delete", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location ID")
}

// TestLocationsSettingsCommand verifies the settings command is registered with proper args
func TestLocationsSettingsCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "settings", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Get location settings")
	assert.Contains(t, output, "<id>")
}

// TestLocationsSettingsCommand_RequiresIDArgument tests that settings requires an ID
func TestLocationsSettingsCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "settings"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestLocationsSettingsCommand_InvalidID tests that settings validates the ID is numeric.
// This validation happens BEFORE getClient(), so we can test it!
func TestLocationsSettingsCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"locations", "settings", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location ID")
}

// TestLocationsCommand_WithMockClient tests command output using mock HTTP server.
func TestLocationsCommand_WithMockClient(t *testing.T) {
	t.Run("list returns locations table", func(t *testing.T) {
		// Create mock server that returns location list
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Location{
				{Id: 1, CompanyName: "Main Office", Code: "MAIN", Active: true},
				{Id: 2, CompanyName: "Branch Office", Code: "BRANCH", Active: false},
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
		cmd := newLocationsListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		// Verify
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Main Office")
		assert.Contains(t, output, "Branch Office")
		assert.Contains(t, output, "MAIN")
		assert.Contains(t, output, "BRANCH")
	})
}
