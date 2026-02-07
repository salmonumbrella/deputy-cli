package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/auth"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
	"github.com/salmonumbrella/deputy-cli/internal/secrets"
	"github.com/spf13/cobra"
)

var validGeos = []string{"au", "uk", "na"}

// storeKey is the context key for store injection
type storeKey struct{}

// WithStore injects a credential store into context for testing
func WithStore(ctx context.Context, store secrets.Store) context.Context {
	return context.WithValue(ctx, storeKey{}, store)
}

// getStore retrieves store from context or creates default KeychainStore
func getStore(ctx context.Context) (secrets.Store, error) {
	if store, ok := ctx.Value(storeKey{}).(secrets.Store); ok {
		return store, nil
	}
	return secrets.NewKeychainStore()
}

type setupServer interface {
	Start(ctx context.Context) (*auth.SetupResult, error)
}

type setupServerFactory func(store secrets.Store) (setupServer, error)

type setupServerFactoryKey struct{}

// WithSetupServerFactory injects a setup server factory into context for testing
func WithSetupServerFactory(ctx context.Context, factory setupServerFactory) context.Context {
	return context.WithValue(ctx, setupServerFactoryKey{}, factory)
}

func setupServerFactoryFromContext(ctx context.Context) setupServerFactory {
	if factory, ok := ctx.Value(setupServerFactoryKey{}).(setupServerFactory); ok && factory != nil {
		return factory
	}
	return func(store secrets.Store) (setupServer, error) {
		return auth.NewSetupServer(store)
	}
}

type authClientFactory func(creds *secrets.Credentials) (*api.Client, error)

type authClientFactoryKey struct{}

// WithAuthClientFactory injects an auth client factory into context for testing
func WithAuthClientFactory(ctx context.Context, factory authClientFactory) context.Context {
	return context.WithValue(ctx, authClientFactoryKey{}, factory)
}

func authClientFactoryFromContext(ctx context.Context) authClientFactory {
	if factory, ok := ctx.Value(authClientFactoryKey{}).(authClientFactory); ok && factory != nil {
		return factory
	}
	return func(creds *secrets.Credentials) (*api.Client, error) {
		return api.NewClient(creds), nil
	}
}

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Commands for managing Deputy API authentication credentials.",
	}

	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthAddCmd())
	cmd.AddCommand(newAuthStatusCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthTestCmd())

	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate via browser",
		Long:  "Opens a browser window to authenticate with your Deputy account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getStore(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to open keychain: %w", err)
			}

			factory := setupServerFactoryFromContext(cmd.Context())
			server, err := factory(store)
			if err != nil {
				return fmt.Errorf("failed to start auth server: %w", err)
			}

			// Handle interrupt gracefully
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigCh
				cancel()
			}()

			result, err := server.Start(ctx)
			if err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "\nAuthenticated successfully!\n")
			_, _ = fmt.Fprintf(io.Out, "Install: %s.%s.deputy.com\n", result.Install, result.Geo)
			return nil
		},
	}
}

func newAuthAddCmd() *cobra.Command {
	var token, install, geo string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add credentials via CLI",
		Long:  "Add Deputy API credentials directly via command-line flags.",
		Example: `  deputy auth add --token YOUR_TOKEN --install mycompany --geo au
  deputy auth add -t YOUR_TOKEN -i mycompany -g na`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" {
				return errors.New("--token is required")
			}
			if install == "" {
				return errors.New("--install is required")
			}
			if geo == "" {
				return errors.New("--geo is required")
			}

			geo = strings.ToLower(geo)
			if !isValidGeo(geo) {
				return fmt.Errorf("invalid geo %q: must be one of %v", geo, validGeos)
			}

			store, err := getStore(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to open keychain: %w", err)
			}

			creds := &secrets.Credentials{
				Token:     token,
				Install:   strings.ToLower(install),
				Geo:       geo,
				CreatedAt: time.Now(),
			}

			if err := store.Set(creds); err != nil {
				return fmt.Errorf("failed to save credentials: %w", err)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Credentials saved for %s.%s.deputy.com\n", creds.Install, creds.Geo)
			return nil
		},
	}

	cmd.Flags().StringVarP(&token, "token", "t", "", "Deputy permanent API token (required)")
	cmd.Flags().StringVarP(&install, "install", "i", "", "Deputy install name (required)")
	cmd.Flags().StringVarP(&geo, "geo", "g", "", "Geographic region: au, uk, or na (required)")

	return cmd
}

// authStatus represents the authentication status for JSON output
type authStatus struct {
	Install     string `json:"install"`
	Region      string `json:"region"`
	BaseURL     string `json:"base_url"`
	TokenMasked string `json:"token_masked"`
	Added       string `json:"added"`
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := loadCredentialsFromContext(cmd.Context())
			if err != nil {
				if strings.Contains(err.Error(), "not authenticated") {
					io := iocontext.FromContext(cmd.Context())
					_, _ = fmt.Fprintln(io.Out, "Not authenticated. Set DEPUTY_TOKEN (env/.env) or run 'deputy auth add' to configure.")
					return nil
				}
				return err
			}

			// Mask token
			maskedToken := "****"
			if len(creds.Token) > 8 {
				maskedToken = creds.Token[:4] + "..." + creds.Token[len(creds.Token)-4:]
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				status := authStatus{
					Install:     creds.Install,
					Region:      strings.ToUpper(creds.Geo),
					BaseURL:     creds.BaseURL(),
					TokenMasked: maskedToken,
					Added:       creds.CreatedAt.Format(time.RFC3339),
				}
				f := outfmt.New(cmd.Context())
				return f.Output(status)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Install:  %s\n", creds.Install)
			_, _ = fmt.Fprintf(io.Out, "Region:   %s\n", strings.ToUpper(creds.Geo))
			_, _ = fmt.Fprintf(io.Out, "Base URL: %s\n", creds.BaseURL())
			_, _ = fmt.Fprintf(io.Out, "Token:    %s\n", maskedToken)
			_, _ = fmt.Fprintf(io.Out, "Added:    %s\n", creds.CreatedAt.Format(time.RFC3339))
			return nil
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getStore(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to open keychain: %w", err)
			}

			if err := store.Delete(); err != nil {
				if errors.Is(err, secrets.ErrNotFound) {
					io := iocontext.FromContext(cmd.Context())
					_, _ = fmt.Fprintln(io.Out, "No credentials to remove.")
					return nil
				}
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintln(io.Out, "Credentials removed.")
			return nil
		},
	}
}

func isValidGeo(geo string) bool {
	for _, g := range validGeos {
		if g == geo {
			return true
		}
	}
	return false
}

func newAuthTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test authentication by calling /me endpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := loadCredentialsFromContext(cmd.Context())
			if err != nil {
				return err
			}

			factory := authClientFactoryFromContext(cmd.Context())
			client, err := factory(creds)
			if err != nil {
				return err
			}
			me, err := client.Me().Info(cmd.Context())
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Authentication successful!\n")
			_, _ = fmt.Fprintf(io.Out, "User: %s (%s)\n", me.Name, me.PrimaryEmail)
			_, _ = fmt.Fprintf(io.Out, "ID:   %d\n", me.EmployeeId)
			return nil
		},
	}
}
