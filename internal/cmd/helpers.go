package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

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

// ClientFactory creates API clients - allows injection for testing
type ClientFactory interface {
	NewClient(ctx context.Context) (*api.Client, error)
}

// DefaultClientFactory creates clients from keychain credentials
type DefaultClientFactory struct{}

// NewClient creates an API client using keychain credentials
func (f DefaultClientFactory) NewClient(ctx context.Context) (*api.Client, error) {
	return getClient()
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

func getClient() (*api.Client, error) {
	store, err := newKeychainStore()
	if err != nil {
		return nil, err
	}

	creds, err := store.Get()
	if err != nil {
		if errors.Is(err, secrets.ErrNotFound) {
			return nil, errors.New("not authenticated - run 'deputy auth add' first")
		}
		return nil, err
	}

	return api.NewClient(creds), nil
}

var newKeychainStore = func() (secrets.Store, error) {
	return secrets.NewKeychainStore()
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
