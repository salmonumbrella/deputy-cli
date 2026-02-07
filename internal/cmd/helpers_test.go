package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
	"github.com/salmonumbrella/deputy-cli/internal/secrets"
)

/*
TESTABILITY NOTES

The helpers.go file provides:
1. ClientFactory interface - for dependency injection in commands
2. DefaultClientFactory - production implementation using keychain
3. WithClientFactory/ClientFactoryFromContext - context-based DI
4. getClientFromContext - uses factory from context
5. getClient - direct keychain access (legacy, used by DefaultClientFactory)

The ClientFactory pattern enables testing commands without real keychain access.
Commands can be refactored to use getClientFromContext(ctx) instead of getClient().
*/

// MockClientFactory implements ClientFactory for testing
type MockClientFactory struct {
	client *api.Client
	err    error
}

func (f *MockClientFactory) NewClient(ctx context.Context) (*api.Client, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.client, nil
}

func TestClientFactory_Interface(t *testing.T) {
	t.Run("DefaultClientFactory implements ClientFactory", func(t *testing.T) {
		var _ ClientFactory = DefaultClientFactory{}
	})

	t.Run("MockClientFactory implements ClientFactory", func(t *testing.T) {
		var _ ClientFactory = &MockClientFactory{}
	})
}

func TestWithClientFactory(t *testing.T) {
	t.Run("injects factory into context", func(t *testing.T) {
		mockFactory := &MockClientFactory{}
		ctx := context.Background()

		ctx = WithClientFactory(ctx, mockFactory)

		// Verify factory is retrievable
		factory := ClientFactoryFromContext(ctx)
		assert.Equal(t, mockFactory, factory)
	})
}

func TestClientFactoryFromContext(t *testing.T) {
	t.Run("returns injected factory when present", func(t *testing.T) {
		mockFactory := &MockClientFactory{}
		ctx := WithClientFactory(context.Background(), mockFactory)

		factory := ClientFactoryFromContext(ctx)

		assert.Equal(t, mockFactory, factory)
	})

	t.Run("returns DefaultClientFactory when no factory in context", func(t *testing.T) {
		oldDefault := defaultClientFactory
		mockFactory := &MockClientFactory{}
		defaultClientFactory = mockFactory
		defer func() { defaultClientFactory = oldDefault }()

		ctx := context.Background()
		factory := ClientFactoryFromContext(ctx)
		assert.Equal(t, mockFactory, factory)
	})

	t.Run("returns DefaultClientFactory for nil context value", func(t *testing.T) {
		oldDefault := defaultClientFactory
		mockFactory := &MockClientFactory{}
		defaultClientFactory = mockFactory
		defer func() { defaultClientFactory = oldDefault }()

		factory := ClientFactoryFromContext(context.Background())
		assert.Equal(t, mockFactory, factory)
	})
}

func TestGetClientFromContext(t *testing.T) {
	t.Run("uses factory from context", func(t *testing.T) {
		expectedClient := &api.Client{}
		mockFactory := &MockClientFactory{client: expectedClient}
		ctx := WithClientFactory(context.Background(), mockFactory)

		client, err := getClientFromContext(ctx)

		require.NoError(t, err)
		assert.Equal(t, expectedClient, client)
	})

	t.Run("returns factory error", func(t *testing.T) {
		expectedErr := errors.New("factory error")
		mockFactory := &MockClientFactory{err: expectedErr}
		ctx := WithClientFactory(context.Background(), mockFactory)

		client, err := getClientFromContext(ctx)

		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, client)
	})

	t.Run("falls back to defaultClientFactory when no factory in context", func(t *testing.T) {
		oldDefault := defaultClientFactory
		expectedClient := &api.Client{}
		mockFactory := &MockClientFactory{client: expectedClient}
		defaultClientFactory = mockFactory
		defer func() { defaultClientFactory = oldDefault }()

		client, err := getClientFromContext(context.Background())

		require.NoError(t, err)
		assert.Equal(t, expectedClient, client)
	})
}

func TestGetClient(t *testing.T) {
	oldStoreFactory := newKeychainStore
	defer func() { newKeychainStore = oldStoreFactory }()

	t.Run("returns error when not authenticated", func(t *testing.T) {
		mockStore := secrets.NewMockStore()
		newKeychainStore = func() (secrets.Store, error) { return mockStore, nil }

		client, err := getClient(context.Background())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not authenticated")
		assert.Nil(t, client)
	})

	t.Run("returns error when store init fails", func(t *testing.T) {
		expectedErr := errors.New("store init failed")
		newKeychainStore = func() (secrets.Store, error) { return nil, expectedErr }

		client, err := getClient(context.Background())

		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, client)
	})

	t.Run("returns client when credentials exist", func(t *testing.T) {
		mockStore := secrets.NewMockStore()
		_ = mockStore.Set(&secrets.Credentials{
			Token:   "test-token",
			Install: "test",
			Geo:     "au",
		})
		newKeychainStore = func() (secrets.Store, error) { return mockStore, nil }

		client, err := getClient(context.Background())

		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestDefaultClientFactory_NewClient(t *testing.T) {
	t.Run("delegates to getClient", func(t *testing.T) {
		oldStoreFactory := newKeychainStore
		defer func() { newKeychainStore = oldStoreFactory }()

		mockStore := secrets.NewMockStore()
		_ = mockStore.Set(&secrets.Credentials{
			Token:   "test-token",
			Install: "test",
			Geo:     "au",
		})
		newKeychainStore = func() (secrets.Store, error) { return mockStore, nil }

		factory := DefaultClientFactory{}
		ctx := context.Background()

		client, err := factory.NewClient(ctx)

		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

// TestHelpers_Integration demonstrates the intended usage pattern
func TestHelpers_Integration(t *testing.T) {
	t.Run("mock factory enables command testing", func(t *testing.T) {
		// Create a mock client (would typically have mock HTTP responses)
		mockClient := &api.Client{}
		mockFactory := &MockClientFactory{client: mockClient}

		// Inject factory into context
		ctx := WithClientFactory(context.Background(), mockFactory)

		// Simulate command using getClientFromContext
		client, err := getClientFromContext(ctx)

		require.NoError(t, err)
		assert.Equal(t, mockClient, client)
	})

	t.Run("error factory enables error path testing", func(t *testing.T) {
		mockFactory := &MockClientFactory{
			err: errors.New("not authenticated - run 'deputy auth add' first"),
		}
		ctx := WithClientFactory(context.Background(), mockFactory)

		_, err := getClientFromContext(ctx)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not authenticated")
	})
}

/*
REFACTORING NOTES FOR COMMANDS

To enable full testing of commands using the ClientFactory pattern,
commands should be updated to use getClientFromContext instead of getClient:

BEFORE (not testable):
    func (c *employeesListCmd) runE(cmd *cobra.Command, args []string) error {
        client, err := getClient()  // Direct keychain access
        ...
    }

AFTER (testable):
    func (c *employeesListCmd) runE(cmd *cobra.Command, args []string) error {
        client, err := getClientFromContext(cmd.Context())  // Uses injected factory
        ...
    }

Then tests can inject a mock:
    mockFactory := &MockClientFactory{client: mockClient}
    ctx := WithClientFactory(context.Background(), mockFactory)
    cmd.SetContext(ctx)
    err := cmd.Execute()
*/

func TestValidateDateFormat(t *testing.T) {
	t.Run("accepts valid YYYY-MM-DD format", func(t *testing.T) {
		validDates := []string{
			"2024-01-01",
			"2024-12-31",
			"2023-06-15",
			"2025-02-28",
		}
		for _, date := range validDates {
			err := validateDateFormat(date)
			assert.NoError(t, err, "date %q should be valid", date)
		}
	})

	t.Run("rejects invalid date formats", func(t *testing.T) {
		invalidDates := []struct {
			date string
			desc string
		}{
			{"01-01-2024", "DD-MM-YYYY format"},
			{"2024/01/01", "wrong separator"},
			{"2024-1-1", "single digit month/day"},
			{"24-01-01", "two digit year"},
			{"2024-13-01", "invalid month"},
			{"2024-01-32", "invalid day"},
			{"not-a-date", "non-date string"},
			{"", "empty string"},
			{"2024-01", "missing day"},
			{"2024", "year only"},
		}
		for _, tc := range invalidDates {
			err := validateDateFormat(tc.date)
			assert.Error(t, err, "date %q (%s) should be invalid", tc.date, tc.desc)
			assert.Contains(t, err.Error(), "invalid date format")
			assert.Contains(t, err.Error(), "expected YYYY-MM-DD")
		}
	})

	t.Run("error message includes the invalid date", func(t *testing.T) {
		err := validateDateFormat("bad-date")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `"bad-date"`)
	})
}

func TestWithDebug(t *testing.T) {
	t.Run("stores debug flag in context", func(t *testing.T) {
		ctx := WithDebug(context.Background(), true)
		assert.True(t, DebugFromContext(ctx))
	})

	t.Run("stores false debug flag in context", func(t *testing.T) {
		ctx := WithDebug(context.Background(), false)
		assert.False(t, DebugFromContext(ctx))
	})
}

func TestDebugFromContext(t *testing.T) {
	t.Run("returns true when debug is set", func(t *testing.T) {
		ctx := WithDebug(context.Background(), true)
		assert.True(t, DebugFromContext(ctx))
	})

	t.Run("returns false when debug is not set", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, DebugFromContext(ctx))
	})

	t.Run("returns false as default", func(t *testing.T) {
		// Verify the default is false, not true
		ctx := context.Background()
		debug := DebugFromContext(ctx)
		assert.False(t, debug)
	})
}

func TestConfirmDestructive(t *testing.T) {
	t.Run("auto-confirms with yes flag", func(t *testing.T) {
		ctx := context.Background()

		err := confirmDestructive(ctx, true, "Are you sure?")

		assert.NoError(t, err)
	})

	t.Run("auto-confirms in JSON output mode", func(t *testing.T) {
		ctx := outfmt.WithFormat(context.Background(), "json")

		err := confirmDestructive(ctx, false, "Are you sure?")

		assert.NoError(t, err)
	})

	t.Run("prompts and confirms with y response", func(t *testing.T) {
		inBuf := bytes.NewBufferString("y\n")
		outBuf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{
			In:     inBuf,
			Out:    outBuf,
			ErrOut: outBuf,
		})

		err := confirmDestructive(ctx, false, "Are you sure?")

		assert.NoError(t, err)
		assert.Contains(t, outBuf.String(), "Are you sure? [y/N]:")
	})

	t.Run("prompts and confirms with Y response", func(t *testing.T) {
		inBuf := bytes.NewBufferString("Y\n")
		outBuf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{
			In:     inBuf,
			Out:    outBuf,
			ErrOut: outBuf,
		})

		err := confirmDestructive(ctx, false, "Are you sure?")

		assert.NoError(t, err)
	})

	t.Run("prompts and cancels with n response", func(t *testing.T) {
		inBuf := bytes.NewBufferString("n\n")
		outBuf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{
			In:     inBuf,
			Out:    outBuf,
			ErrOut: outBuf,
		})

		err := confirmDestructive(ctx, false, "Are you sure?")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
	})

	t.Run("prompts and cancels with empty response", func(t *testing.T) {
		inBuf := bytes.NewBufferString("\n")
		outBuf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{
			In:     inBuf,
			Out:    outBuf,
			ErrOut: outBuf,
		})

		err := confirmDestructive(ctx, false, "Are you sure?")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
	})

	t.Run("prompts and cancels with arbitrary text", func(t *testing.T) {
		inBuf := bytes.NewBufferString("no\n")
		outBuf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{
			In:     inBuf,
			Out:    outBuf,
			ErrOut: outBuf,
		})

		err := confirmDestructive(ctx, false, "Are you sure?")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
	})
}

func TestApplyPagination(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	t.Run("returns all items with no pagination", func(t *testing.T) {
		result := applyPagination(items, 0, 0)
		assert.Equal(t, items, result)
	})

	t.Run("applies limit only", func(t *testing.T) {
		result := applyPagination(items, 0, 3)
		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("applies offset only", func(t *testing.T) {
		result := applyPagination(items, 3, 0)
		assert.Equal(t, []int{4, 5, 6, 7, 8, 9, 10}, result)
	})

	t.Run("applies both offset and limit", func(t *testing.T) {
		result := applyPagination(items, 2, 3)
		assert.Equal(t, []int{3, 4, 5}, result)
	})

	t.Run("returns empty slice when offset exceeds length", func(t *testing.T) {
		result := applyPagination(items, 15, 0)
		assert.Equal(t, []int{}, result)
	})

	t.Run("returns remaining items when limit exceeds available", func(t *testing.T) {
		result := applyPagination(items, 8, 10)
		assert.Equal(t, []int{9, 10}, result)
	})

	t.Run("handles empty input slice", func(t *testing.T) {
		result := applyPagination([]int{}, 0, 5)
		assert.Equal(t, []int{}, result)
	})

	t.Run("works with string slices", func(t *testing.T) {
		strings := []string{"a", "b", "c", "d", "e"}
		result := applyPagination(strings, 1, 2)
		assert.Equal(t, []string{"b", "c"}, result)
	})

	t.Run("offset equals length returns empty", func(t *testing.T) {
		result := applyPagination(items, 10, 5)
		assert.Equal(t, []int{}, result)
	})
}

func TestGetClientFromContext_DebugPropagation(t *testing.T) {
	t.Run("propagates debug flag to client", func(t *testing.T) {
		// Create a mock factory that returns a real client we can inspect
		creds := &secrets.Credentials{
			Token:   "test-token",
			Install: "test",
			Geo:     "au",
		}
		mockClient := api.NewClient(creds)
		mockFactory := &MockClientFactory{client: mockClient}

		// Set up context with debug enabled
		ctx := context.Background()
		ctx = WithClientFactory(ctx, mockFactory)
		ctx = WithDebug(ctx, true)

		// Get client from context
		client, err := getClientFromContext(ctx)

		require.NoError(t, err)
		assert.NotNil(t, client)
		// Note: We can't directly verify debug is set on client since it's private,
		// but the call to SetDebug happens in getClientFromContext
	})

	t.Run("debug defaults to false when not set in context", func(t *testing.T) {
		creds := &secrets.Credentials{
			Token:   "test-token",
			Install: "test",
			Geo:     "au",
		}
		mockClient := api.NewClient(creds)
		mockFactory := &MockClientFactory{client: mockClient}

		// Set up context WITHOUT debug flag
		ctx := WithClientFactory(context.Background(), mockFactory)

		client, err := getClientFromContext(ctx)

		require.NoError(t, err)
		assert.NotNil(t, client)
		// SetDebug(false) is called - debug mode is disabled by default
	})
}

func TestRequireArg(t *testing.T) {
	// Helper to create a test command with a given Args validator
	makeCmd := func(args cobra.PositionalArgs) *cobra.Command {
		return &cobra.Command{
			Use:  "test <id>",
			Args: args,
		}
	}

	t.Run("accepts exactly one argument", func(t *testing.T) {
		cmd := makeCmd(RequireArg("id"))
		err := cmd.Args(cmd, []string{"123"})
		assert.NoError(t, err)
	})

	t.Run("rejects zero arguments with descriptive error", func(t *testing.T) {
		cmd := makeCmd(RequireArg("id"))
		err := cmd.Args(cmd, []string{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing required argument: <id>")
		assert.Contains(t, err.Error(), "Hint: Run 'test --help' for usage")
	})

	t.Run("rejects too many arguments with descriptive error", func(t *testing.T) {
		cmd := makeCmd(RequireArg("id"))
		err := cmd.Args(cmd, []string{"123", "456"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "too many arguments, expected <id>")
		assert.Contains(t, err.Error(), "Hint: Run 'test --help' for usage")
	})

	t.Run("uses custom argument name in error", func(t *testing.T) {
		cmd := makeCmd(RequireArg("employee-id"))
		err := cmd.Args(cmd, []string{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing required argument: <employee-id>")
	})

	t.Run("includes full command path in hint", func(t *testing.T) {
		parent := &cobra.Command{Use: "deputy"}
		child := &cobra.Command{
			Use: "employees",
		}
		grandchild := &cobra.Command{
			Use:  "get <id>",
			Args: RequireArg("id"),
		}
		parent.AddCommand(child)
		child.AddCommand(grandchild)

		err := grandchild.Args(grandchild, []string{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Hint: Run 'deputy employees get --help' for usage")
	})
}

func TestRequireArgs(t *testing.T) {
	// Helper to create a test command with a given Args validator
	makeCmd := func(args cobra.PositionalArgs) *cobra.Command {
		return &cobra.Command{
			Use:  "test <resource> <id>",
			Args: args,
		}
	}

	t.Run("accepts exactly the specified number of arguments", func(t *testing.T) {
		cmd := makeCmd(RequireArgs("resource", "id"))
		err := cmd.Args(cmd, []string{"Employee", "123"})
		assert.NoError(t, err)
	})

	t.Run("rejects zero arguments listing all missing", func(t *testing.T) {
		cmd := makeCmd(RequireArgs("resource", "id"))
		err := cmd.Args(cmd, []string{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing required argument(s): <resource> <id>")
		assert.Contains(t, err.Error(), "Hint: Run 'test --help' for usage")
	})

	t.Run("rejects one argument when two required", func(t *testing.T) {
		cmd := makeCmd(RequireArgs("resource", "id"))
		err := cmd.Args(cmd, []string{"Employee"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing required argument(s): <id>")
		assert.Contains(t, err.Error(), "Hint: Run 'test --help' for usage")
	})

	t.Run("rejects too many arguments", func(t *testing.T) {
		cmd := makeCmd(RequireArgs("resource", "id"))
		err := cmd.Args(cmd, []string{"Employee", "123", "extra"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "too many arguments, expected <resource> <id>")
		assert.Contains(t, err.Error(), "Hint: Run 'test --help' for usage")
	})

	t.Run("works with single argument", func(t *testing.T) {
		cmd := makeCmd(RequireArgs("id"))

		// Accepts one arg
		err := cmd.Args(cmd, []string{"123"})
		assert.NoError(t, err)

		// Rejects zero args
		err = cmd.Args(cmd, []string{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing required argument(s): <id>")
	})

	t.Run("works with three arguments", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:  "test <a> <b> <c>",
			Args: RequireArgs("a", "b", "c"),
		}

		// Accepts three args
		err := cmd.Args(cmd, []string{"1", "2", "3"})
		assert.NoError(t, err)

		// Missing last two
		err = cmd.Args(cmd, []string{"1"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing required argument(s): <b> <c>")
	})
}

func TestFormatArgNames(t *testing.T) {
	t.Run("formats single name", func(t *testing.T) {
		result := formatArgNames([]string{"id"})
		assert.Equal(t, "<id>", result)
	})

	t.Run("formats multiple names", func(t *testing.T) {
		result := formatArgNames([]string{"resource", "id"})
		assert.Equal(t, "<resource> <id>", result)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		result := formatArgNames([]string{})
		assert.Equal(t, "", result)
	})
}
