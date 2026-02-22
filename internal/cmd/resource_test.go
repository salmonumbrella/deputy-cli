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

The resource commands use getClientFromContext() which supports dependency injection
via WithClientFactory(). Tests can inject a MockClientFactory with a mock HTTP server.

Commands tested with mock client:
- resource info: Uses Resource().Info() API
- resource query: Uses Resource().Query() API
- resource get: Uses Resource().Get() API

Note: resource list does NOT require credentials - it only returns hardcoded
known resources from api.KnownResources().

See TestResourceCommand_WithMockClient for full mock client tests.
*/

// TestResourceCommand_ViaRootCmd verifies the resource command is properly registered
func TestResourceCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"resource", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Generic commands for querying any Deputy resource")
}

// TestResourceCommand_HasAllSubcommands verifies all expected subcommands are registered
func TestResourceCommand_HasAllSubcommands(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"resource", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()

	expectedSubcommands := []string{
		"list",
		"info",
		"query",
		"get",
	}

	for _, sub := range expectedSubcommands {
		assert.Contains(t, output, sub, "missing subcommand: %s", sub)
	}
}

// TestResourceCommand_HasExpectedSubcommands tests using newResourceCmd directly
func TestResourceCommand_HasExpectedSubcommands(t *testing.T) {
	cmd := newResourceCmd()
	subCmds := cmd.Commands()
	names := make([]string, len(subCmds))
	for i, c := range subCmds {
		names[i] = c.Name()
	}
	assert.Contains(t, names, "list")
	assert.Contains(t, names, "info")
	assert.Contains(t, names, "query")
	assert.Contains(t, names, "get")
}

// TestResourceCommand_Aliases verifies the command aliases work
func TestResourceCommand_Aliases(t *testing.T) {
	aliases := []string{"res"}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			root := NewRootCmd()
			buf := &bytes.Buffer{}
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs([]string{alias, "--help"})

			err := root.Execute()

			require.NoError(t, err)
			assert.Contains(t, buf.String(), "Generic commands for querying any Deputy resource")
		})
	}
}

// TestResourceListCommand verifies the list command is registered
func TestResourceListCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"resource", "list", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List known resource types")
}

// TestResourceListCommand_NoAuthRequired tests that list works without credentials.
// This command only returns hardcoded known resources.
func TestResourceListCommand_NoAuthRequired(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"resource", "list"})

	err := root.Execute()

	// This should succeed without credentials since it uses api.KnownResources()
	require.NoError(t, err)
	// Note: The output goes to outfmt which uses io.Out from context, not cmd.Out
	// Just verify no error for now - the command runs successfully
}

// TestResourceInfoCommand verifies the info command is registered with proper args
func TestResourceInfoCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"resource", "info", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Get schema information for a resource")
	assert.Contains(t, output, "<ResourceName>")
}

// TestResourceInfoCommand_RequiresResourceNameArgument tests that info requires a ResourceName
func TestResourceInfoCommand_RequiresResourceNameArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"resource", "info"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestResourceInfoCommand_WithMockClient tests that info works with mock client.
func TestResourceInfoCommand_WithMockClient(t *testing.T) {
	// Create mock server that returns resource info
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.ResourceInfo{
			Name: "Employee",
			Fields: map[string]interface{}{
				"Id":        "integer",
				"FirstName": "string",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newResourceInfoCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"Employee"})
	err := cmd.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Resource: Employee")
	assert.Contains(t, output, "Fields:")
}

func TestResourceInfoCommand_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.ResourceInfo{
			Name: "Employee",
			Fields: map[string]interface{}{
				"Id": "integer",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
	ctx = outfmt.WithFormat(ctx, "json")

	cmd := newResourceInfoCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"Employee"})
	err := cmd.Execute()

	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"name": "Employee"`)
}

func TestResourceInfoCommand_AssociationsText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.ResourceInfo{
			Name: "Roster",
			Fields: map[string]interface{}{
				"Id": "integer",
			},
			Assocs: map[string]interface{}{
				"Employee": "Employee",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newResourceInfoCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"Roster"})
	err := cmd.Execute()

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Associations:")
	assert.Contains(t, buf.String(), "Employee")
}

// TestResourceQueryCommand verifies the query command is registered with proper args
func TestResourceQueryCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"resource", "query", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	// Long description starts with "Query any Deputy resource with filters"
	assert.Contains(t, output, "Query any Deputy resource with filters")
	assert.Contains(t, output, "<ResourceName>")
	assert.Contains(t, output, "--filter")
	assert.Contains(t, output, "--join")
	assert.Contains(t, output, "--sort")
	assert.Contains(t, output, "--limit")
	assert.Contains(t, output, "--start")
}

// TestResourceQueryCommand_RequiresResourceNameArgument tests that query requires a ResourceName
func TestResourceQueryCommand_RequiresResourceNameArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"resource", "query"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestResourceQueryCommand_HelpShowsFilterSyntax tests that query help shows filter syntax examples
func TestResourceQueryCommand_HelpShowsFilterSyntax(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"resource", "query", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	// Check filter syntax documentation is present
	assert.Contains(t, output, "field=value")
	assert.Contains(t, output, "field>value")
}

// TestResourceQueryCommand_WithMockClient tests that query works with mock client.
func TestResourceQueryCommand_WithMockClient(t *testing.T) {
	// Create mock server that returns query results
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{"Id": 1, "FirstName": "John", "Active": true},
			{"Id": 2, "FirstName": "Jane", "Active": true},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newResourceQueryCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"Employee", "--filter", "Active=1"})
	err := cmd.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Found 2 result(s)")
	assert.Contains(t, output, "John")
	assert.Contains(t, output, "Jane")
}

// TestResourceGetCommand verifies the get command is registered with proper args
func TestResourceGetCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"resource", "get", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Get a specific resource by ID")
	assert.Contains(t, output, "<ResourceName>")
	assert.Contains(t, output, "<id>")
}

// TestResourceGetCommand_RequiresArguments tests that get requires ResourceName and id
func TestResourceGetCommand_RequiresArguments(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"resource", "get"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument(s): <ResourceName> <id>")
}

// TestResourceGetCommand_RequiresBothArguments tests that get requires both arguments
func TestResourceGetCommand_RequiresBothArguments(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"resource", "get", "Employee"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument(s): <id>")
}

// TestResourceGetCommand_InvalidID tests that get validates the ID is numeric.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestResourceGetCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"resource", "get", "Employee", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ID")
}

// TestResourceGetCommand_WithMockClient tests that get works with mock client.
func TestResourceGetCommand_WithMockClient(t *testing.T) {
	// Create mock server that returns a single resource
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"Id":        123,
			"FirstName": "John",
			"LastName":  "Doe",
			"Active":    true,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newResourceGetCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"Employee", "123"})
	err := cmd.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Id: 123")
	assert.Contains(t, output, "FirstName: John")
}

// TestParseFilters tests the filter parsing function
func TestParseFilters(t *testing.T) {
	tests := []struct {
		name        string
		filters     []string
		expectErr   bool
		errContains string
	}{
		{
			name:      "empty filters returns nil",
			filters:   nil,
			expectErr: false,
		},
		{
			name:      "empty slice returns nil",
			filters:   []string{},
			expectErr: false,
		},
		{
			name:      "equals operator",
			filters:   []string{"Active=1"},
			expectErr: false,
		},
		{
			name:      "greater than operator",
			filters:   []string{"StartTime>1234567890"},
			expectErr: false,
		},
		{
			name:      "less than operator",
			filters:   []string{"EndTime<1234567890"},
			expectErr: false,
		},
		{
			name:      "greater than or equal operator",
			filters:   []string{"Date>=2024-01-01"},
			expectErr: false,
		},
		{
			name:      "less than or equal operator",
			filters:   []string{"Date<=2024-12-31"},
			expectErr: false,
		},
		{
			name:      "multiple filters",
			filters:   []string{"Active=1", "Date>=2024-01-01"},
			expectErr: false,
		},
		{
			name:        "invalid filter syntax - no operator",
			filters:     []string{"InvalidFilter"},
			expectErr:   true,
			errContains: "invalid filter syntax",
		},
		{
			name:        "invalid filter syntax - missing field",
			filters:     []string{"=value"},
			expectErr:   true,
			errContains: "invalid filter syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFilters(tt.filters)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				if len(tt.filters) == 0 {
					assert.Nil(t, result)
				}
			}
		})
	}
}

// TestParseFilters_CorrectFormat tests that parseFilters creates correct Deputy format
func TestParseFilters_CorrectFormat(t *testing.T) {
	filters := []string{"Active=1", "Date>=2024-01-01"}
	result, err := parseFilters(filters)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Check f1 (first filter)
	f1, ok := result["f1"].(map[string]interface{})
	require.True(t, ok, "f1 should be a map")
	assert.Equal(t, "Active", f1["field"])
	assert.Equal(t, "eq", f1["type"])
	assert.Equal(t, "1", f1["data"])

	// Check f2 (second filter)
	f2, ok := result["f2"].(map[string]interface{})
	require.True(t, ok, "f2 should be a map")
	assert.Equal(t, "Date", f2["field"])
	assert.Equal(t, "ge", f2["type"])
	assert.Equal(t, "2024-01-01", f2["data"])
}

// TestResourceCommand_WithMockClient tests resource commands using mock HTTP server.
func TestResourceCommand_WithMockClient(t *testing.T) {
	t.Run("info returns resource schema", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.ResourceInfo{
				Name: "Employee",
				Fields: map[string]interface{}{
					"Id":        "integer",
					"FirstName": "string",
					"LastName":  "string",
				},
				Assocs: map[string]interface{}{
					"Company": "Company",
				},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newResourceInfoCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"Employee"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Resource: Employee")
		assert.Contains(t, output, "Fields:")
		assert.Contains(t, output, "Id")
	})

	t.Run("info returns JSON output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.ResourceInfo{
				Name: "Timesheet",
				Fields: map[string]interface{}{
					"Id":        "integer",
					"StartTime": "datetime",
				},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newResourceInfoCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"Timesheet"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, `"name": "Timesheet"`)
	})

	t.Run("query returns filtered results", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{"Id": 1, "FirstName": "John", "Active": true},
				{"Id": 2, "FirstName": "Jane", "Active": true},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newResourceQueryCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"Employee", "--filter", "Active=1"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Found 2 result(s)")
		assert.Contains(t, output, "John")
		assert.Contains(t, output, "Jane")
	})

	t.Run("query returns JSON output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{"Id": 1, "FirstName": "John"},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newResourceQueryCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"Employee"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, `"FirstName": "John"`)
	})

	t.Run("query with no results", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newResourceQueryCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"Employee", "--filter", "Id=999999"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "No results found")
	})

	t.Run("get returns single resource", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"Id":        123,
				"FirstName": "John",
				"LastName":  "Doe",
				"Active":    true,
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newResourceGetCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"Employee", "123"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Id: 123")
		assert.Contains(t, output, "FirstName: John")
	})

	t.Run("get returns JSON output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"Id":        456,
				"FirstName": "Jane",
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newResourceGetCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"Timesheet", "456"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, `"Id": 456`)
		assert.Contains(t, output, `"FirstName": "Jane"`)
	})
}
