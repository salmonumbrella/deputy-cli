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

// TestLeaveCommand_ViaRootCmd verifies the leave command is properly registered
func TestLeaveCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Manage leave requests")
}

// TestLeaveCommand_HasAllSubcommands verifies all expected subcommands are registered
func TestLeaveCommand_HasAllSubcommands(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()

	expectedSubcommands := []string{
		"list",
		"get",
		"add",
		"approve",
		"decline",
	}

	for _, sub := range expectedSubcommands {
		assert.Contains(t, output, sub, "missing subcommand: %s", sub)
	}
}

// TestLeaveCommand_Aliases verifies the command aliases work
func TestLeaveCommand_Aliases(t *testing.T) {
	aliases := []string{"leaves"}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			root := NewRootCmd()
			buf := &bytes.Buffer{}
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs([]string{alias, "--help"})

			err := root.Execute()

			require.NoError(t, err)
			assert.Contains(t, buf.String(), "Manage leave requests")
		})
	}
}

// TestLeaveListCommand verifies the list command is registered
func TestLeaveListCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "list", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List leave requests")
	assert.Contains(t, output, "--employee")
}

// TestLeaveListCommand_RequiresAuth tests that list fails without credentials.
// SKIP: Requires client injection - no pre-client validation in list command.
func TestLeaveListCommand_RequiresAuth(t *testing.T) {
	t.Skip("Requires refactoring: list command has no pre-client validation, cannot test without mock client")
}

// TestLeaveGetCommand verifies the get command is registered with proper args
func TestLeaveGetCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "get", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Get leave request details")
	assert.Contains(t, output, "<id>")
}

// TestLeaveGetCommand_RequiresIDArgument tests that get requires an ID
func TestLeaveGetCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "get"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestLeaveGetCommand_InvalidID tests that get validates the ID is numeric.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestLeaveGetCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "get", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid leave ID")
}

// TestLeaveAddCommand verifies the add command has all required flags
func TestLeaveAddCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "add", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Add a leave request")
	assert.Contains(t, output, "--employee")
	assert.Contains(t, output, "--start-date")
	assert.Contains(t, output, "--end-date")
	assert.Contains(t, output, "--leave-rule")
	assert.Contains(t, output, "--comment")
}

// TestLeaveAddCommand_RequiresEmployeeFlag tests that add requires --employee.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestLeaveAddCommand_RequiresEmployeeFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "add", "--start-date", "2024-01-15", "--end-date", "2024-01-16"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--employee is required")
}

// TestLeaveAddCommand_RequiresStartDateFlag tests that add requires --start-date.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestLeaveAddCommand_RequiresStartDateFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "add", "--employee", "123", "--end-date", "2024-01-16"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--start-date is required")
}

// TestLeaveAddCommand_RequiresEndDateFlag tests that add requires --end-date.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestLeaveAddCommand_RequiresEndDateFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "add", "--employee", "123", "--start-date", "2024-01-15"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--end-date is required")
}

// TestLeaveApproveCommand verifies the approve command is registered with proper args
func TestLeaveApproveCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "approve", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Approve a leave request")
	assert.Contains(t, output, "<id>")
}

// TestLeaveApproveCommand_RequiresIDArgument tests that approve requires an ID
func TestLeaveApproveCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "approve"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestLeaveApproveCommand_InvalidID tests that approve validates the ID is numeric.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestLeaveApproveCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "approve", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid leave ID")
}

// TestLeaveDeclineCommand verifies the decline command is registered with proper args
func TestLeaveDeclineCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "decline", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Decline a leave request")
	assert.Contains(t, output, "<id>")
	assert.Contains(t, output, "--comment")
}

// TestLeaveDeclineCommand_RequiresIDArgument tests that decline requires an ID
func TestLeaveDeclineCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "decline"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestLeaveDeclineCommand_InvalidID tests that decline validates the ID is numeric.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestLeaveDeclineCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"leave", "decline", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid leave ID")
}

// TestLeaveStatusText tests the leaveStatusText helper function
func TestLeaveStatusText(t *testing.T) {
	tests := []struct {
		status   int
		expected string
	}{
		{0, "Awaiting"},
		{1, "Approved"},
		{2, "Declined"},
		{3, "Cancelled"},
		{4, "Pay Pending"},
		{5, "Pay Approved"},
		{99, "Unknown (99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := leaveStatusText(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestLeaveCommand_WithMockClient tests leave commands using mock HTTP server.
func TestLeaveCommand_WithMockClient(t *testing.T) {
	t.Run("list returns leave table", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Leave{
				{Id: 1, Employee: 10, DateStart: "2024-01-01", DateEnd: "2024-01-05", Days: 4.0, Status: 1},
				{Id: 2, Employee: 20, DateStart: "2024-02-01", DateEnd: "2024-02-03", Days: 2.0, Status: 0},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newLeaveListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "2024-01-01")
		assert.Contains(t, output, "2024-02-01")
		assert.Contains(t, output, "Approved")
		assert.Contains(t, output, "Awaiting")
	})

	t.Run("list with employee filter uses query", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/v1/resource/Leave/QUERY", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Leave{
				{Id: 3, Employee: 99, DateStart: "2024-03-01", DateEnd: "2024-03-02", Days: 1.0, Status: 1},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newLeaveListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--employee", "99"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "2024-03-01")
	})

	t.Run("list returns json output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Leave{
				{Id: 4, Employee: 10, DateStart: "2024-04-01", DateEnd: "2024-04-02", Days: 1.0, Status: 1},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newLeaveListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Id": 4`)
	})

	t.Run("get returns leave details", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Leave{
				Id:        123,
				Employee:  456,
				Company:   1,
				DateStart: "2024-01-15",
				DateEnd:   "2024-01-16",
				Days:      1.0,
				Hours:     8.0,
				Status:    1,
				Comment:   "Vacation",
				LeaveRule: 5,
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newLeaveGetCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "ID:         123")
		assert.Contains(t, output, "Employee:   456")
		assert.Contains(t, output, "Status:     Approved")
		assert.Contains(t, output, "Comment:    Vacation")
	})

	t.Run("add creates leave request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Leave{
				Id:        999,
				Employee:  123,
				DateStart: "2024-01-15",
				DateEnd:   "2024-01-16",
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newLeaveAddCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{
			"--employee", "123",
			"--start-date", "2024-01-15",
			"--end-date", "2024-01-16",
		})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Created leave request 999")
	})

	t.Run("add returns json output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Leave{
				Id:        1001,
				Employee:  123,
				DateStart: "2024-01-15",
				DateEnd:   "2024-01-16",
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newLeaveAddCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{
			"--employee", "123",
			"--start-date", "2024-01-15",
			"--end-date", "2024-01-16",
		})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Id": 1001`)
	})

	t.Run("approve approves leave request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			// Approve returns the updated leave object
			_ = json.NewEncoder(w).Encode(api.Leave{
				Id:     123,
				Status: 1,
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newLeaveApproveCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Leave request 123 approved")
	})

	t.Run("decline declines leave request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			// Decline returns the updated leave object
			_ = json.NewEncoder(w).Encode(api.Leave{
				Id:     123,
				Status: 2,
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newLeaveDeclineCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--comment", "Insufficient notice", "--yes"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Leave request 123 declined")
	})
}
