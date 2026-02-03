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
TESTABILITY NOTES

The departments commands use getClientFromContext() which supports dependency injection
via the ClientFactory pattern. This allows full testing with mock HTTP servers.

Tests verify:
- Command structure and registration
- Subcommand availability
- Flag parsing and definitions
- Argument validation (count, type)
- Pre-API validation (required flags, ID parsing)
- Help text and usage strings
- Full API integration with mock HTTP servers (see TestDepartmentsCommand_WithMockClient)
*/

// TestDepartmentsCommand_ViaRootCmd verifies the departments command is properly registered
func TestDepartmentsCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Manage departments")
}

// TestDepartmentsCommand_HasAllSubcommands verifies all expected subcommands are registered
func TestDepartmentsCommand_HasAllSubcommands(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()

	expectedSubcommands := []string{
		"list",
		"get",
		"add",
		"update",
		"delete",
	}

	for _, sub := range expectedSubcommands {
		assert.Contains(t, output, sub, "missing subcommand: %s", sub)
	}
}

// TestDepartmentsCommand_Aliases verifies the command aliases work
func TestDepartmentsCommand_Aliases(t *testing.T) {
	aliases := []string{"department", "dept", "opunit", "d", "areas", "area"}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			root := NewRootCmd()
			buf := &bytes.Buffer{}
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs([]string{alias, "--help"})

			err := root.Execute()

			require.NoError(t, err)
			assert.Contains(t, buf.String(), "Manage departments")
		})
	}
}

// TestDepartmentsListCommand verifies the list command is registered
func TestDepartmentsListCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "list", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List all departments")
}

// TestDepartmentsGetCommand verifies the get command is registered with proper args
func TestDepartmentsGetCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "get", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Get department details")
	assert.Contains(t, output, "<id>")
}

// TestDepartmentsGetCommand_RequiresIDArgument tests that get requires an ID
func TestDepartmentsGetCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "get"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestDepartmentsGetCommand_InvalidID tests that get validates the ID is numeric.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestDepartmentsGetCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "get", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid department ID")
}

// TestDepartmentsAddCommand verifies the add command has all required flags
func TestDepartmentsAddCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "add", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Add a new department")
	assert.Contains(t, output, "--name")
	assert.Contains(t, output, "--company")
	assert.Contains(t, output, "--code")
	assert.Contains(t, output, "--parent")
	assert.Contains(t, output, "--sort-order")
}

// TestDepartmentsAddCommand_RequiresNameFlag tests that add requires --name.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestDepartmentsAddCommand_RequiresNameFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "add", "--company", "1"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--name is required")
}

// TestDepartmentsAddCommand_RequiresCompanyFlag tests that add requires --company.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestDepartmentsAddCommand_RequiresCompanyFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "add", "--name", "Test Department"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--company is required")
}

// TestDepartmentsUpdateCommand verifies the update command is registered with proper args
func TestDepartmentsUpdateCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "update", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Update a department")
	assert.Contains(t, output, "<id>")
	assert.Contains(t, output, "--name")
	assert.Contains(t, output, "--code")
	assert.Contains(t, output, "--sort-order")
	assert.Contains(t, output, "--active")
	assert.Contains(t, output, "--set-active")
}

// TestDepartmentsUpdateCommand_RequiresIDArgument tests that update requires an ID
func TestDepartmentsUpdateCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "update"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestDepartmentsUpdateCommand_InvalidID tests that update validates the ID is numeric.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestDepartmentsUpdateCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "update", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid department ID")
}

// TestDepartmentsDeleteCommand verifies the delete command is registered with proper args
func TestDepartmentsDeleteCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "delete", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Delete a department")
	assert.Contains(t, output, "<id>")
}

// TestDepartmentsDeleteCommand_RequiresIDArgument tests that delete requires an ID
func TestDepartmentsDeleteCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "delete"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestDepartmentsDeleteCommand_InvalidID tests that delete validates the ID is numeric.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestDepartmentsDeleteCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"departments", "delete", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid department ID")
}

// TestDepartmentsCommand_WithMockClient tests command output using mock HTTP server.
func TestDepartmentsCommand_WithMockClient(t *testing.T) {
	t.Run("list returns departments table", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Department{
				{Id: 1, CompanyName: "Sales", CompanyCode: "SALES", Company: 1, Active: true},
				{Id: 2, CompanyName: "Engineering", CompanyCode: "ENG", Company: 1, Active: false},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newDepartmentsListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Sales")
		assert.Contains(t, output, "SALES")
		assert.Contains(t, output, "Yes") // Active
		assert.Contains(t, output, "Engineering")
		assert.Contains(t, output, "No") // Inactive
	})

	t.Run("list returns JSON output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Department{
				{Id: 1, CompanyName: "Sales", CompanyCode: "SALES", Company: 1, Active: true},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newDepartmentsListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, `"CompanyName": "Sales"`)
		assert.Contains(t, output, `"CompanyCode": "SALES"`)
	})

	t.Run("get returns department details", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Department{
				Id:          123,
				CompanyName: "Test Department",
				CompanyCode: "TEST",
				Company:     1,
				ParentId:    0,
				SortOrder:   5,
				Active:      true,
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newDepartmentsGetCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "ID:         123")
		assert.Contains(t, output, "Name:       Test Department")
		assert.Contains(t, output, "Code:       TEST")
		assert.Contains(t, output, "Company:    1")
		assert.Contains(t, output, "Active:     true")
	})

	t.Run("add creates department", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Department{
				Id:          999,
				CompanyName: "New Department",
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newDepartmentsAddCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--name", "New Department", "--company", "1"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Created department 999: New Department")
	})

	t.Run("update updates department", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Department{
				Id:          123,
				CompanyName: "Updated Name",
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newDepartmentsUpdateCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--name", "Updated Name"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Updated department 123: Updated Name")
	})

	t.Run("update returns json output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Department{
				Id:          123,
				CompanyName: "Updated Name",
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newDepartmentsUpdateCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--name", "Updated Name"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Id": 123`)
	})

	t.Run("delete deletes department", func(t *testing.T) {
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

		cmd := newDepartmentsDeleteCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--yes"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Deleted department 123")
	})
}

func TestDepartmentsGetCommand_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.Department{
			Id:          123,
			CompanyName: "Engineering",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
	ctx = outfmt.WithFormat(ctx, "json")

	cmd := newDepartmentsGetCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"123"})
	err := cmd.Execute()

	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"Id": 123`)
}

func TestDepartmentsDeleteCommand_Cancelled(t *testing.T) {
	inBuf := bytes.NewBufferString("n\n")
	outBuf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = iocontext.WithIO(ctx, &iocontext.IO{In: inBuf, Out: outBuf, ErrOut: outBuf})

	cmd := newDepartmentsDeleteCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(outBuf)
	cmd.SetArgs([]string{"123"})
	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}
