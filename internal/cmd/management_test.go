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

The management commands use getClientFromContext() which supports dependency injection
via WithClientFactory(). This allows full testing with mock HTTP servers.

Tests verify:
- Command structure and registration
- Subcommand availability
- Flag parsing and definitions
- Pre-API validation (required flags)
- Help text and usage strings
- Full API integration via mock HTTP servers (see TestManagementCommand_WithMockClient)
*/

// TestManagementCommand_ViaRootCmd verifies the management command is properly registered
func TestManagementCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Memos and journals")
}

// TestManagementCommand_HasAllSubcommands verifies all expected subcommands are registered
func TestManagementCommand_HasAllSubcommands(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()

	expectedSubcommands := []string{
		"memo",
		"journal",
	}

	for _, sub := range expectedSubcommands {
		assert.Contains(t, output, sub, "missing subcommand: %s", sub)
	}
}

// ==================== MEMO SUBCOMMAND TESTS ====================

// TestMemoCommand verifies the memo command is registered
func TestMemoCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "memo", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Manage memos")
}

// TestMemoCommand_HasAllSubcommands verifies memo subcommands are registered
func TestMemoCommand_HasAllSubcommands(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "memo", "--help"})

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

// TestMemoListCommand verifies the memo list command is registered
func TestMemoListCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "memo", "list", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List memos")
	assert.Contains(t, output, "--company")
	assert.Contains(t, output, "--limit")
	assert.Contains(t, output, "--offset")
}

// TestMemoListCommand_RequiresCompanyFlag tests that memo list requires --company.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestMemoListCommand_RequiresCompanyFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "memo", "list"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--company is required")
}

// TestMemoAddCommand verifies the memo add command is registered
func TestMemoAddCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "memo", "add", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Create a memo")
	assert.Contains(t, output, "--company")
	assert.Contains(t, output, "--content")
}

// TestMemoAddCommand_RequiresCompanyFlag tests that memo add requires --company.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestMemoAddCommand_RequiresCompanyFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "memo", "add", "--content", "Test memo"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--company is required")
}

// TestMemoAddCommand_RequiresContentFlag tests that memo add requires --content.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestMemoAddCommand_RequiresContentFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "memo", "add", "--company", "1"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--content is required")
}

// ==================== JOURNAL SUBCOMMAND TESTS ====================

// TestJournalCommand verifies the journal command is registered
func TestJournalCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "journal", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Manage journals")
}

// TestJournalCommand_HasAllSubcommands verifies journal subcommands are registered
func TestJournalCommand_HasAllSubcommands(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "journal", "--help"})

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

// TestJournalListCommand verifies the journal list command is registered
func TestJournalListCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "journal", "list", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List journal entries")
	assert.Contains(t, output, "--employee")
	assert.Contains(t, output, "--limit")
	assert.Contains(t, output, "--offset")
}

// TestJournalListCommand_RequiresEmployeeFlag tests that journal list requires --employee.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestJournalListCommand_RequiresEmployeeFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "journal", "list"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--employee is required")
}

// TestJournalAddCommand verifies the journal add command is registered
func TestJournalAddCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "journal", "add", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Post a journal entry")
	assert.Contains(t, output, "--employee")
	assert.Contains(t, output, "--company")
	assert.Contains(t, output, "--comment")
	assert.Contains(t, output, "--category")
}

// TestJournalAddCommand_RequiresEmployeeFlag tests that journal add requires --employee.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestJournalAddCommand_RequiresEmployeeFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "journal", "add", "--company", "1", "--comment", "Test"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--employee is required")
}

// TestJournalAddCommand_RequiresCompanyFlag tests that journal add requires --company.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestJournalAddCommand_RequiresCompanyFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "journal", "add", "--employee", "123", "--comment", "Test"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--company is required")
}

// TestJournalAddCommand_RequiresCommentFlag tests that journal add requires --comment.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestJournalAddCommand_RequiresCommentFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"management", "journal", "add", "--employee", "123", "--company", "1"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--comment is required")
}

// TestManagementCommand_WithMockClient tests command output using mock HTTP server.
func TestManagementCommand_WithMockClient(t *testing.T) {
	t.Run("memo list returns memos table", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Memo{
				{Id: 1, Content: "Important announcement", Created: 1705312800},
				{Id: 2, Content: "Staff meeting tomorrow", Created: 1705226400},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newMemoListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--company", "1"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Important announcement")
		assert.Contains(t, output, "Staff meeting tomorrow")
	})

	t.Run("memo list returns JSON output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Memo{
				{Id: 1, Content: "Important announcement", Created: 1705312800},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newMemoListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--company", "1"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, `"Content": "Important announcement"`)
	})

	t.Run("memo add creates memo", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Memo{
				Id:      999,
				Content: "New memo content",
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newMemoAddCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--company", "1", "--content", "New memo content", "--location", "1"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Created memo 999")
	})

	t.Run("journal list returns journals table", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Journal{
				{Id: 1, Comment: "Performance review", Employee: 123, Created: 1705312800},
				{Id: 2, Comment: "Training completed", Employee: 123, Created: 1705226400},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newJournalListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--employee", "123"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Performance review")
		assert.Contains(t, output, "Training completed")
	})

	t.Run("journal list returns JSON output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Journal{
				{Id: 1, Comment: "Performance review", Employee: 123, Created: 1705312800},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newJournalListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--employee", "123"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, `"Comment": "Performance review"`)
	})

	t.Run("journal add posts journal entry", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Journal{
				Id:       999,
				Employee: 123,
				Comment:  "New journal entry",
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newJournalAddCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{
			"--employee", "123",
			"--company", "1",
			"--comment", "New journal entry",
		})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Posted journal 999 for employee 123")
	})
}
