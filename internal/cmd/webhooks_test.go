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
WEBHOOKS COMMAND TESTS

These tests use mock HTTP servers via the ClientFactory pattern:
- getClientFromContext() retrieves a ClientFactory from context
- Tests inject MockClientFactory with a test HTTP server client
- This allows full API response testing without keychain access

Test coverage:
- Command structure and registration
- Subcommand availability
- Flag parsing and definitions
- Argument validation (count, type)
- Pre-API validation (required flags, ID parsing)
- API response handling via mock HTTP servers
- Table and JSON output formatting
*/

// TestWebhooksCommand_ViaRootCmd verifies the webhooks command is properly registered
func TestWebhooksCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"webhooks", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Manage webhooks")
}

// TestWebhooksCommand_HasAllSubcommands verifies all expected subcommands are registered
func TestWebhooksCommand_HasAllSubcommands(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"webhooks", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()

	expectedSubcommands := []string{
		"list",
		"get",
		"add",
		"delete",
	}

	for _, sub := range expectedSubcommands {
		assert.Contains(t, output, sub, "missing subcommand: %s", sub)
	}
}

// TestWebhooksCommand_HasExpectedSubcommands tests using newWebhooksCmd directly
func TestWebhooksCommand_HasExpectedSubcommands(t *testing.T) {
	cmd := newWebhooksCmd()
	subCmds := cmd.Commands()
	names := make([]string, len(subCmds))
	for i, c := range subCmds {
		names[i] = c.Name()
	}
	assert.Contains(t, names, "list")
	assert.Contains(t, names, "get")
	assert.Contains(t, names, "add")
	assert.Contains(t, names, "delete")
}

// TestWebhooksCommand_Aliases verifies the command aliases work
func TestWebhooksCommand_Aliases(t *testing.T) {
	aliases := []string{"webhook", "wh"}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			root := NewRootCmd()
			buf := &bytes.Buffer{}
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs([]string{alias, "--help"})

			err := root.Execute()

			require.NoError(t, err)
			assert.Contains(t, buf.String(), "Manage webhooks")
		})
	}
}

// TestWebhooksListCommand verifies the list command is registered
func TestWebhooksListCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"webhooks", "list", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "List all webhooks")
}

// TestWebhooksGetCommand verifies the get command is registered with proper args
func TestWebhooksGetCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"webhooks", "get", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Get webhook details")
	assert.Contains(t, output, "<id>")
}

// TestWebhooksGetCommand_RequiresIDArgument tests that get requires an ID
func TestWebhooksGetCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"webhooks", "get"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestWebhooksGetCommand_InvalidID tests that get validates the ID is numeric.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestWebhooksGetCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"webhooks", "get", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid webhook ID")
}

// TestWebhooksAddCommand verifies the add command has all required flags
func TestWebhooksAddCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"webhooks", "add", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Add a new webhook")
	assert.Contains(t, output, "--topic")
	assert.Contains(t, output, "--url")
	assert.Contains(t, output, "--type")
	assert.Contains(t, output, "--enabled")
}

// TestWebhooksAddCommand_RequiresTopicFlag tests that add requires --topic.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestWebhooksAddCommand_RequiresTopicFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"webhooks", "add", "--url", "https://example.com/webhook"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--topic is required")
}

// TestWebhooksAddCommand_RequiresURLFlag tests that add requires --url.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestWebhooksAddCommand_RequiresURLFlag(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"webhooks", "add", "--topic", "roster.created"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--url is required")
}

// TestWebhooksAddCommand_RequiresBothTopicAndURL tests that add requires both flags.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestWebhooksAddCommand_RequiresBothTopicAndURL(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"webhooks", "add"})

	err := root.Execute()

	require.Error(t, err)
	// topic is checked first in the code
	assert.Contains(t, err.Error(), "--topic is required")
}

// TestWebhooksDeleteCommand verifies the delete command is registered with proper args
func TestWebhooksDeleteCommand(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"webhooks", "delete", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Delete a webhook")
	assert.Contains(t, output, "<id>")
}

// TestWebhooksDeleteCommand_RequiresIDArgument tests that delete requires an ID
func TestWebhooksDeleteCommand_RequiresIDArgument(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"webhooks", "delete"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

// TestWebhooksDeleteCommand_InvalidID tests that delete validates the ID is numeric.
// This validation happens BEFORE getClientFromContext(), so we can test it!
func TestWebhooksDeleteCommand_InvalidID(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"webhooks", "delete", "not-a-number"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid webhook ID")
}

// TestWebhooksCommand_WithMockClient tests command output using mock HTTP server.
func TestWebhooksCommand_WithMockClient(t *testing.T) {
	t.Run("list returns webhooks table", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Webhook{
				{Id: 1, Topic: "Timesheet.Insert", Url: "https://example.com/hook1", Enabled: true},
				{Id: 2, Topic: "Roster.Update", Url: "https://example.com/hook2", Enabled: false},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newWebhooksListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Timesheet.Insert")
		assert.Contains(t, output, "Roster.Update")
		assert.Contains(t, output, "Yes") // Enabled
		assert.Contains(t, output, "No")  // Disabled
	})

	t.Run("list returns JSON output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.Webhook{
				{Id: 1, Topic: "Timesheet.Insert", Url: "https://example.com/hook", Enabled: true},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newWebhooksListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, `"Topic": "Timesheet.Insert"`)
		assert.Contains(t, output, `"Address": "https://example.com/hook"`)
	})

	t.Run("get returns webhook details", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Webhook{
				Id:      123,
				Topic:   "Roster.Created",
				Url:     "https://example.com/webhook",
				Type:    "json",
				Enabled: true,
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newWebhooksGetCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "ID:      123")
		assert.Contains(t, output, "Topic:   Roster.Created")
		assert.Contains(t, output, "URL:     https://example.com/webhook")
		assert.Contains(t, output, "Type:    json")
		assert.Contains(t, output, "Enabled: true")
	})

	t.Run("get returns JSON output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Webhook{
				Id:      123,
				Topic:   "Roster.Created",
				Url:     "https://example.com/webhook",
				Type:    "json",
				Enabled: true,
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newWebhooksGetCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, `"Id": 123`)
		assert.Contains(t, output, `"Topic": "Roster.Created"`)
	})

	t.Run("add creates webhook", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Webhook{
				Id:      999,
				Topic:   "Timesheet.Created",
				Url:     "https://example.com/new-webhook",
				Enabled: true,
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newWebhooksAddCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{
			"--topic", "Timesheet.Created",
			"--url", "https://example.com/new-webhook",
		})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Created webhook 999 for topic Timesheet.Created")
	})

	t.Run("add returns JSON output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Webhook{
				Id:      999,
				Topic:   "Timesheet.Created",
				Url:     "https://example.com/new-webhook",
				Enabled: true,
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newWebhooksAddCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{
			"--topic", "Timesheet.Created",
			"--url", "https://example.com/new-webhook",
		})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, `"Id": 999`)
		assert.Contains(t, output, `"Topic": "Timesheet.Created"`)
	})

	t.Run("delete removes webhook", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newWebhooksDeleteCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--yes"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Deleted webhook 123")
	})

	t.Run("delete cancelled without confirmation", func(t *testing.T) {
		inBuf := bytes.NewBufferString("n\n")
		outBuf := &bytes.Buffer{}
		ctx := context.Background()
		ctx = iocontext.WithIO(ctx, &iocontext.IO{In: inBuf, Out: outBuf, ErrOut: outBuf})

		cmd := newWebhooksDeleteCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(outBuf)
		cmd.SetArgs([]string{"123"})
		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
	})
}
