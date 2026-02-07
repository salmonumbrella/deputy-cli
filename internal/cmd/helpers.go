package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
	"github.com/salmonumbrella/deputy-cli/internal/secrets"
)

// validateDateFormat validates that a date string is in YYYY-MM-DD format.
func validateDateFormat(dateStr string) error {
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format %q: expected YYYY-MM-DD", dateStr)
	}
	return nil
}

// RequireArg creates a Cobra Args validator that requires exactly one argument
// with a descriptive error message.
func RequireArg(argName string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing required argument: <%s>\nHint: Run '%s --help' for usage", argName, cmd.CommandPath())
		}
		if len(args) > 1 {
			return fmt.Errorf("too many arguments, expected <%s>\nHint: Run '%s --help' for usage", argName, cmd.CommandPath())
		}
		return nil
	}
}

// RequireArgs creates a Cobra Args validator that requires exactly the specified
// number of arguments with descriptive error messages.
func RequireArgs(argNames ...string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		expected := len(argNames)
		if len(args) < expected {
			var missing []string
			for i := len(args); i < expected; i++ {
				missing = append(missing, "<"+argNames[i]+">")
			}
			return fmt.Errorf("missing required argument(s): %s\nHint: Run '%s --help' for usage",
				strings.Join(missing, " "), cmd.CommandPath())
		}
		if len(args) > expected {
			return fmt.Errorf("too many arguments, expected %s\nHint: Run '%s --help' for usage",
				formatArgNames(argNames), cmd.CommandPath())
		}
		return nil
	}
}

// formatArgNames formats argument names for display.
func formatArgNames(names []string) string {
	var formatted []string
	for _, name := range names {
		formatted = append(formatted, "<"+name+">")
	}
	return strings.Join(formatted, " ")
}

// Context key for debug flag
type debugKey struct{}

// WithDebug stores the debug flag in context
func WithDebug(ctx context.Context, debug bool) context.Context {
	return context.WithValue(ctx, debugKey{}, debug)
}

// DebugFromContext retrieves the debug flag from context
func DebugFromContext(ctx context.Context) bool {
	if debug, ok := ctx.Value(debugKey{}).(bool); ok {
		return debug
	}
	return false
}

// Context key for "no keychain" mode.
type noKeychainKey struct{}

func WithNoKeychain(ctx context.Context, noKeychain bool) context.Context {
	return context.WithValue(ctx, noKeychainKey{}, noKeychain)
}

func NoKeychainFromContext(ctx context.Context) bool {
	if v, ok := ctx.Value(noKeychainKey{}).(bool); ok {
		return v
	}
	if s := strings.TrimSpace(os.Getenv("DEPUTY_NO_KEYCHAIN")); s != "" && s != "0" && strings.ToLower(s) != "false" {
		return true
	}
	return false
}

// ClientFactory creates API clients - allows injection for testing
type ClientFactory interface {
	NewClient(ctx context.Context) (*api.Client, error)
}

// DefaultClientFactory creates clients from keychain credentials
type DefaultClientFactory struct{}

// NewClient creates an API client using keychain credentials
func (f DefaultClientFactory) NewClient(ctx context.Context) (*api.Client, error) {
	return getClient(ctx)
}

// Context key for injecting client factory
type clientFactoryKey struct{}

// WithClientFactory injects a client factory into context
func WithClientFactory(ctx context.Context, factory ClientFactory) context.Context {
	return context.WithValue(ctx, clientFactoryKey{}, factory)
}

// ClientFactoryFromContext retrieves client factory from context
func ClientFactoryFromContext(ctx context.Context) ClientFactory {
	if factory, ok := ctx.Value(clientFactoryKey{}).(ClientFactory); ok {
		return factory
	}
	return defaultClientFactory
}

var defaultClientFactory ClientFactory = DefaultClientFactory{}

// getClientFromContext retrieves an API client using the factory from context
func getClientFromContext(ctx context.Context) (*api.Client, error) {
	factory := ClientFactoryFromContext(ctx)
	client, err := factory.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	client.SetDebug(DebugFromContext(ctx))
	return client, nil
}

func getClient(ctx context.Context) (*api.Client, error) {
	creds, err := loadCredentialsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return api.NewClient(creds), nil
}

var newKeychainStore = func() (secrets.Store, error) {
	return secrets.NewKeychainStore()
}

func loadCredentialsFromContext(ctx context.Context) (*secrets.Credentials, error) {
	// 1) If tests injected a store via WithStore(ctx, store), respect it.
	if store, ok := ctx.Value(storeKey{}).(secrets.Store); ok {
		creds, err := store.Get()
		if err != nil {
			if errors.Is(err, secrets.ErrNotFound) {
				return nil, errors.New("not authenticated - set DEPUTY_TOKEN or run 'deputy auth add' first")
			}
			return nil, err
		}
		return creds, nil
	}

	// 2) Environment (including .env) avoids keychain prompts.
	if creds, ok, err := secrets.FromEnv(); err != nil {
		return nil, err
	} else if ok {
		return creds, nil
	}

	// 3) Keychain fallback unless explicitly disabled.
	if NoKeychainFromContext(ctx) {
		return nil, errors.New("not authenticated - keychain disabled (set DEPUTY_TOKEN in env/.env)")
	}

	store, err := newKeychainStore()
	if err != nil {
		return nil, err
	}

	creds, err := store.Get()
	if err != nil {
		if errors.Is(err, secrets.ErrNotFound) {
			return nil, errors.New("not authenticated - set DEPUTY_TOKEN or run 'deputy auth add' first")
		}
		return nil, err
	}
	return creds, nil
}

// applyPagination applies offset and limit to a slice.
// It returns the original slice if no pagination is needed.
// This is used for client-side pagination when the API doesn't support it.
func applyPagination[T any](items []T, offset, limit int) []T {
	if offset > 0 {
		if offset >= len(items) {
			return []T{}
		}
		items = items[offset:]
	}
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}
	return items
}

// confirmDestructive prompts the user for confirmation before a destructive operation.
// It auto-confirms (returns nil) if:
// - yes flag is true (--yes/-y was passed)
// - output format is JSON (programmatic/AI agent mode)
// Otherwise, it prompts the user and returns an error if they don't confirm.
func confirmDestructive(ctx context.Context, yes bool, promptMsg string) error {
	// Auto-confirm in JSON output mode (AI agent automation)
	format := outfmt.GetFormat(ctx)
	if format == "json" {
		return nil
	}

	// Auto-confirm if --yes flag is set
	if yes {
		return nil
	}

	// Prompt the user for confirmation
	io := iocontext.FromContext(ctx)
	_, _ = fmt.Fprintf(io.Out, "%s [y/N]: ", promptMsg)
	var response string
	_, _ = fmt.Fscanln(io.In, &response)
	if response != "y" && response != "Y" {
		return errors.New("operation cancelled")
	}
	return nil
}
