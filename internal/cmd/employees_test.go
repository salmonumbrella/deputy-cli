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

The employees commands cannot be fully tested with mock API clients because they
use getClient() which reads credentials from the system keychain:

    client, err := getClient()

This function is defined in root.go and:
1. Creates a KeychainStore to read stored credentials
2. Uses those credentials to construct an API client

REFACTORING NEEDED FOR FULL TESTABILITY:

Option 1: Context-based dependency injection (recommended)
Add an API client to the context, similar to iocontext.IO:

    // In api package or new clientcontext package:
    func WithClient(ctx context.Context, client *Client) context.Context
    func FromContext(ctx context.Context) *Client

    // In commands:
    client := api.FromContext(cmd.Context())
    if client == nil {
        client, err = getClient()
        ...
    }

Option 2: Command constructor injection
Pass client factory to command constructors:

    type ClientFactory func() (*api.Client, error)
    func newEmployeesListCmd(getClient ClientFactory) *cobra.Command

Option 3: Package-level factory function
Use a replaceable factory (similar to auth tests suggestion):

    var GetClient = getClient

    // In tests:
    cmd.GetClient = func() (*api.Client, error) {
        return mockClient, nil
    }

Option 4: Interface-based mocking
Define an interface for the API client:

    type EmployeeService interface {
        List(ctx context.Context) ([]Employee, error)
        Get(ctx context.Context, id int) (*Employee, error)
        // ... other methods
    }

Until refactoring is done, these tests verify:
- Command structure and registration
- Subcommand availability
- Flag parsing and definitions
- Argument validation (count, type)
- Pre-API validation (required flags, ID parsing)
- Help text and usage strings
*/

// TestEmployeesCommand_ViaRootCmd verifies the employees command is properly registered
func TestEmployeesCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Manage employees")
}

// TestEmployeesCommand_HasAllSubcommands verifies all expected subcommands are registered
func TestEmployeesCommand_HasAllSubcommands(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()

	expectedSubcommands := []string{
		"list",
		"get",
		"add",
		"update",
		"terminate",
		"invite",
		"assign-location",
		"remove-location",
		"reactivate",
		"delete",
		"add-unavailability",
	}

	for _, sub := range expectedSubcommands {
		assert.Contains(t, output, sub, "missing subcommand: %s", sub)
	}
}

// TestEmployeesCommand_Aliases verifies the command aliases work
func TestEmployeesCommand_Aliases(t *testing.T) {
	aliases := []string{"employee", "emp", "e"}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			root := NewRootCmd()
			buf := &bytes.Buffer{}
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs([]string{alias, "--help"})

			err := root.Execute()

			require.NoError(t, err)
			assert.Contains(t, buf.String(), "Manage employees")
		})
	}
}

// TestEmployeesListCommand verifies the list command is registered
func TestEmployeesListCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "list", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List all employees")
}

// TestEmployeesListCommand_RequiresAuth tests that list fails without credentials.
// SKIP: Requires client injection - see testability notes above.
func TestEmployeesListCommand_RequiresAuth(t *testing.T) {
	t.Skip("Requires refactoring: commands use getClient() which reads from keychain, cannot inject mock client")

	// Expected behavior when refactored:
	// mockClient := api.NewMockClient()
	// mockClient.Employees().SetListResponse([]api.Employee{
	//     {Id: 1, DisplayName: "John Doe", Email: "john@example.com", Active: true},
	// })
	// ctx := api.WithClient(context.Background(), mockClient)
	//
	// cmd := newEmployeesListCmd()
	// cmd.SetContext(ctx)
	// err := cmd.Execute()
	//
	// require.NoError(t, err)
	// assert.Contains(t, buf.String(), "John Doe")
}

// TestEmployeesGetCommand verifies the get command is registered with proper args
func TestEmployeesGetCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "get", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Get employee details")
	assert.Contains(t, output, "<id>")
}

// TestEmployeesGetCommand_RequiresIDArgument tests that get requires an ID
func TestEmployeesGetCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "get"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestEmployeesGetCommand_InvalidID tests that get validates the ID is numeric.
// Note: This validation happens before getClient(), so we can test it.
// However, cobra validates arg count first, so this will fail at keychain access.
func TestEmployeesGetCommand_InvalidID(t *testing.T) {
	t.Skip("ID validation happens after getClient() is called, cannot test without mock client")

	// Expected behavior when refactored:
	// root := NewRootCmd()
	// root.SetArgs([]string{"employees", "get", "not-a-number"})
	//
	// err := root.Execute()
	//
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "invalid employee ID")
}

// TestEmployeesAddCommand verifies the add command has all required flags
func TestEmployeesAddCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "add", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Add a new employee")
	assert.Contains(t, output, "--first-name")
	assert.Contains(t, output, "--last-name")
	assert.Contains(t, output, "--email")
	assert.Contains(t, output, "--mobile")
	assert.Contains(t, output, "--start-date")
	assert.Contains(t, output, "--company")
	assert.Contains(t, output, "--role")
}

// TestEmployeesAddCommand_MissingRequiredFlags tests validation for missing flags.
// Note: These validations happen after getClient() is called, so they cannot be
// tested without client injection.
func TestEmployeesAddCommand_MissingRequiredFlags(t *testing.T) {
	t.Skip("Requires refactoring: validation happens after getClient() call")

	// Expected behavior when refactored:
	// tests := []struct {
	//     name        string
	//     args        []string
	//     expectedErr string
	// }{
	//     {
	//         name:        "missing first-name and last-name",
	//         args:        []string{"employees", "add", "--company", "1"},
	//         expectedErr: "--first-name and --last-name are required",
	//     },
	//     {
	//         name:        "missing company",
	//         args:        []string{"employees", "add", "--first-name", "John", "--last-name", "Doe"},
	//         expectedErr: "--company is required",
	//     },
	// }
}

// TestEmployeesUpdateCommand verifies the update command is registered
func TestEmployeesUpdateCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "update", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Update an employee")
	assert.Contains(t, output, "<id>")
	assert.Contains(t, output, "--first-name")
	assert.Contains(t, output, "--last-name")
	assert.Contains(t, output, "--email")
	assert.Contains(t, output, "--mobile")
}

// TestEmployeesUpdateCommand_RequiresIDArgument tests that update requires an ID
func TestEmployeesUpdateCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "update"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestEmployeesTerminateCommand verifies the terminate command is registered
func TestEmployeesTerminateCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "terminate", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Terminate an employee")
	assert.Contains(t, output, "<id>")
	assert.Contains(t, output, "--date")
}

// TestEmployeesTerminateCommand_RequiresIDArgument tests that terminate requires an ID
func TestEmployeesTerminateCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "terminate"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestEmployeesTerminateCommand_MissingDateFlag tests validation for missing date flag.
// SKIP: Validation happens after getClient() call.
func TestEmployeesTerminateCommand_MissingDateFlag(t *testing.T) {
	t.Skip("Requires refactoring: validation happens after getClient() call")

	// Expected behavior when refactored:
	// root := NewRootCmd()
	// root.SetArgs([]string{"employees", "terminate", "123"})
	//
	// err := root.Execute()
	//
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "--date is required")
}

// TestEmployeesInviteCommand verifies the invite command is registered
func TestEmployeesInviteCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "invite", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Send invitation to employee")
	assert.Contains(t, output, "<id>")
}

// TestEmployeesInviteCommand_RequiresIDArgument tests that invite requires an ID
func TestEmployeesInviteCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "invite"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestEmployeesAssignLocationCommand verifies the assign-location command is registered
func TestEmployeesAssignLocationCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "assign-location", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Assign employee to a location")
	assert.Contains(t, output, "<employee-id>")
	assert.Contains(t, output, "--location")
}

// TestEmployeesAssignLocationCommand_RequiresIDArgument tests that assign-location requires an ID
func TestEmployeesAssignLocationCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "assign-location"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestEmployeesRemoveLocationCommand verifies the remove-location command is registered
func TestEmployeesRemoveLocationCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "remove-location", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Remove employee from a location")
	assert.Contains(t, output, "<employee-id>")
	assert.Contains(t, output, "--location")
}

// TestEmployeesReactivateCommand verifies the reactivate command is registered
func TestEmployeesReactivateCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "reactivate", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Reactivate a terminated employee")
	assert.Contains(t, output, "<id>")
}

// TestEmployeesReactivateCommand_RequiresIDArgument tests that reactivate requires an ID
func TestEmployeesReactivateCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "reactivate"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestEmployeesDeleteCommand verifies the delete command is registered
func TestEmployeesDeleteCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "delete", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Delete an employee account")
	assert.Contains(t, output, "<id>")
	assert.Contains(t, output, "--yes")
	assert.Contains(t, output, "-y")
}

// TestEmployeesDeleteCommand_RequiresIDArgument tests that delete requires an ID
func TestEmployeesDeleteCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "delete"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestEmployeesAddUnavailabilityCommand verifies the add-unavailability command is registered
func TestEmployeesAddUnavailabilityCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "add-unavailability", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Add unavailability for employee")
	assert.Contains(t, output, "<employee-id>")
	assert.Contains(t, output, "--start-date")
	assert.Contains(t, output, "--end-date")
	assert.Contains(t, output, "--comment")
}

// TestEmployeesAddUnavailabilityCommand_RequiresIDArgument tests that add-unavailability requires an ID
func TestEmployeesAddUnavailabilityCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"employees", "add-unavailability"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestEmployeesCommand_WithMockClient tests command output using mock HTTP server.
func TestEmployeesCommand_WithMockClient(t *testing.T) {
	t.Run("list returns employees table", func(t *testing.T) {
		// Create mock server that returns employee list
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Employee{
				{Id: 1, DisplayName: "John Doe", Email: "john@example.com", Active: true},
				{Id: 2, DisplayName: "Jane Smith", Email: "jane@example.com", Active: false},
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
		cmd := newEmployeesListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		// Verify
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "John Doe")
		assert.Contains(t, output, "jane@example.com")
		assert.Contains(t, output, "Yes") // Active
		assert.Contains(t, output, "No")  // Inactive
	})

	t.Run("list returns JSON output", func(t *testing.T) {
		// Create mock server that returns employee list
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Employee{
				{Id: 1, DisplayName: "John Doe", Email: "john@example.com", Active: true},
			})
		}))
		defer server.Close()

		// Create client and factory
		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		// Set up context with factory, IO, and JSON format
		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		// Import outfmt to set format
		ctx = outfmt.WithFormat(ctx, "json")

		// Create and execute command
		cmd := newEmployeesListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		// Verify
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, `"DisplayName": "John Doe"`)
		assert.Contains(t, output, `"Email": "john@example.com"`)
	})
}

// TestEmployeesDeleteCommand_WithConfirmation tests the --yes flag behavior
func TestEmployeesDeleteCommand_WithConfirmation(t *testing.T) {
	t.Run("delete with --yes skips confirmation", func(t *testing.T) {
		// Create mock server that handles delete
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "DELETE" && r.URL.Path == "/api/v1/supervise/employee/123" {
				w.WriteHeader(http.StatusOK)
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		// Create client and factory
		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		// Set up context with factory and IO
		outBuf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: outBuf, ErrOut: outBuf})

		// Create and execute command with --yes flag
		cmd := newEmployeesDeleteCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(outBuf)
		cmd.SetArgs([]string{"123", "--yes"})
		err := cmd.Execute()

		// Verify
		require.NoError(t, err)
		assert.Contains(t, outBuf.String(), "Employee 123 deleted")
		// Should not contain the confirmation prompt
		assert.NotContains(t, outBuf.String(), "[y/N]")
	})

	t.Run("delete without --yes prompts for confirmation", func(t *testing.T) {
		// Set up context with IO that provides "n" response
		inBuf := bytes.NewBufferString("n\n")
		outBuf := &bytes.Buffer{}
		ctx := context.Background()
		ctx = iocontext.WithIO(ctx, &iocontext.IO{In: inBuf, Out: outBuf, ErrOut: outBuf})

		// Create and execute command without --yes flag
		cmd := newEmployeesDeleteCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(outBuf)
		cmd.SetArgs([]string{"123"})
		err := cmd.Execute()

		// Verify - should fail because we answered "n"
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
		assert.Contains(t, outBuf.String(), "Are you sure you want to delete employee 123?")
		assert.Contains(t, outBuf.String(), "[y/N]")
	})

	t.Run("delete in JSON mode auto-confirms", func(t *testing.T) {
		// Create mock server that handles delete
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "DELETE" && r.URL.Path == "/api/v1/supervise/employee/123" {
				w.WriteHeader(http.StatusOK)
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		// Create client and factory
		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		// Set up context with factory, IO, and JSON format (no input provided)
		outBuf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: outBuf, ErrOut: outBuf})
		ctx = outfmt.WithFormat(ctx, "json")

		// Create and execute command WITHOUT --yes flag
		cmd := newEmployeesDeleteCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(outBuf)
		cmd.SetArgs([]string{"123"})
		err := cmd.Execute()

		// Verify - should succeed because JSON mode auto-confirms
		require.NoError(t, err)
		// Should not contain the confirmation prompt
		assert.NotContains(t, outBuf.String(), "[y/N]")
	})

	t.Run("delete with y confirmation proceeds", func(t *testing.T) {
		// Create mock server that handles delete
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "DELETE" && r.URL.Path == "/api/v1/supervise/employee/123" {
				w.WriteHeader(http.StatusOK)
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		// Create client and factory
		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		// Set up context with IO that provides "y" response
		inBuf := bytes.NewBufferString("y\n")
		outBuf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{In: inBuf, Out: outBuf, ErrOut: outBuf})

		// Create and execute command without --yes flag
		cmd := newEmployeesDeleteCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(outBuf)
		cmd.SetArgs([]string{"123"})
		err := cmd.Execute()

		// Verify - should succeed because we answered "y"
		require.NoError(t, err)
		assert.Contains(t, outBuf.String(), "Employee 123 deleted")
	})
}

// TestEmployeesAddCommand_WithMockClient tests the add command with mock API
func TestEmployeesAddCommand_WithMockClient(t *testing.T) {
	t.Run("add creates employee", func(t *testing.T) {
		// Create mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && r.URL.Path == "/api/v1/supervise/employee" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(api.Employee{
					Id:          456,
					DisplayName: "John Doe",
				})
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesAddCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--first-name", "John", "--last-name", "Doe", "--company", "1"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Created employee 456: John Doe")
	})

	t.Run("add requires first name and last name", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesAddCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--company", "1"})
		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--first-name and --last-name are required")
	})

	t.Run("add requires company", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesAddCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--first-name", "John", "--last-name", "Doe"})
		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--company is required")
	})

	t.Run("add validates date format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesAddCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--first-name", "John", "--last-name", "Doe", "--company", "1", "--start-date", "invalid"})
		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid date format")
	})

	t.Run("add returns JSON in json mode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Employee{
				Id:          456,
				DisplayName: "John Doe",
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newEmployeesAddCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--first-name", "John", "--last-name", "Doe", "--company", "1"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Id": 456`)
	})
}

// TestEmployeesUpdateCommand_WithMockClient tests the update command with mock API
func TestEmployeesUpdateCommand_WithMockClient(t *testing.T) {
	t.Run("update updates employee", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Employee{
				Id:          123,
				DisplayName: "Updated Name",
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesUpdateCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--first-name", "Updated"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Updated employee 123: Updated Name")
	})

	t.Run("update returns JSON in json mode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Employee{
				Id:          123,
				DisplayName: "Updated Name",
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newEmployeesUpdateCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--first-name", "Updated"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Id": 123`)
	})

	t.Run("update handles invalid ID", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesUpdateCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"invalid", "--first-name", "Test"})
		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid employee ID")
	})
}

// TestEmployeesTerminateCommand_WithMockClient tests the terminate command with mock API
func TestEmployeesTerminateCommand_WithMockClient(t *testing.T) {
	t.Run("terminate terminates employee", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && r.URL.Path == "/api/v1/supervise/employee/123/terminate" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(api.Employee{
					Id:          123,
					DisplayName: "Terminated Employee",
					Active:      false,
				})
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesTerminateCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--date", "2024-01-01", "--yes"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Employee 123 terminated")
	})

	t.Run("terminate requires date", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesTerminateCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--yes"})
		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--date is required")
	})

	t.Run("terminate validates date format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesTerminateCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--date", "invalid", "--yes"})
		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid date format")
	})
}

// TestEmployeesAssignLocationCommand_WithMockClient tests the assign-location command
func TestEmployeesAssignLocationCommand_WithMockClient(t *testing.T) {
	t.Run("assign-location assigns employee to location", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesAssignLocationCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--location", "456"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Employee 123 assigned to location 456")
	})

	t.Run("assign-location requires location flag", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesAssignLocationCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})
		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--location is required")
	})

	t.Run("assign-location handles invalid employee ID", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesAssignLocationCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"invalid", "--location", "456"})
		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid employee ID")
	})
}

// TestEmployeesRemoveLocationCommand_WithMockClient tests the remove-location command
func TestEmployeesRemoveLocationCommand_WithMockClient(t *testing.T) {
	t.Run("remove-location removes employee from location", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesRemoveLocationCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--location", "456"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Employee 123 removed from location 456")
	})

	t.Run("remove-location requires location flag", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesRemoveLocationCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})
		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--location is required")
	})
}

// TestEmployeesAddUnavailabilityCommand_WithMockClient tests the add-unavailability command
func TestEmployeesAddUnavailabilityCommand_WithMockClient(t *testing.T) {
	t.Run("add-unavailability adds unavailability", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Unavailability{
				Id:       789,
				Employee: 123,
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesAddUnavailabilityCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--start-date", "2024-01-01", "--end-date", "2024-01-05"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Added unavailability 789 for employee 123")
	})

	t.Run("add-unavailability requires dates", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesAddUnavailabilityCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})
		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--start-date and --end-date are required")
	})

	t.Run("add-unavailability validates start date format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesAddUnavailabilityCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--start-date", "invalid", "--end-date", "2024-01-05"})
		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid date format")
	})

	t.Run("add-unavailability validates end date format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesAddUnavailabilityCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--start-date", "2024-01-01", "--end-date", "invalid"})
		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid date format")
	})

	t.Run("add-unavailability handles invalid employee ID", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesAddUnavailabilityCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"invalid", "--start-date", "2024-01-01", "--end-date", "2024-01-05"})
		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid employee ID")
	})
}
