package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deputy-cli/internal/config"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
)

var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildDate = "unknown"
)

// ExecuteResult carries information from a completed Execute call so the
// caller can format errors without reading package-level globals.
type ExecuteResult struct {
	Err        error
	JSONOutput bool
	Debug      bool
}

func NewRootCmd() *cobra.Command {
	var fl struct {
		Output     string
		Debug      bool
		Query      string
		Raw        bool
		NoColor    bool
		NoKeychain bool
	}

	cmd := &cobra.Command{
		Use:     "deputy",
		Short:   "CLI for Deputy workforce management API",
		Long:    "A command-line interface for interacting with the Deputy API.\nDesigned for both human users and AI agent automation.",
		Version: Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Load .env once (if present) to support DEPUTY_TOKEN/DEPUTY_INSTALL/etc.
			config.LoadDotenv()

			ctx := cmd.Context()
			// Preserve existing IO if already set (e.g., in tests), otherwise use defaults
			if !iocontext.HasIO(ctx) {
				ctx = iocontext.WithIO(ctx, iocontext.DefaultIO())
			}
			format := strings.ToLower(fl.Output)
			if format != "text" && format != "json" {
				return fmt.Errorf("invalid --output %q (expected text or json)", fl.Output)
			}
			if fl.Raw && format == "text" {
				format = "json"
			}
			ctx = outfmt.WithFormat(ctx, format)
			ctx = outfmt.WithQuery(ctx, fl.Query)
			ctx = outfmt.WithRaw(ctx, fl.Raw)
			ctx = WithDebug(ctx, fl.Debug)
			ctx = WithNoKeychain(ctx, fl.NoKeychain)
			cmd.SetContext(ctx)
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.SetVersionTemplate("deputy version {{.Version}}\n  commit: " + CommitSHA + "\n  built:  " + BuildDate + "\n")

	cmd.PersistentFlags().StringVarP(&fl.Output, "output", "o", "text", "Output format: text or json")
	cmd.PersistentFlags().BoolVar(&fl.Debug, "debug", false, "Enable debug logging")
	cmd.PersistentFlags().StringVarP(&fl.Query, "query", "q", "", "JQ filter for JSON output")
	cmd.PersistentFlags().BoolVar(&fl.Raw, "raw", false, "Output JSON Lines (one object per line)")
	cmd.PersistentFlags().BoolVar(&fl.NoColor, "no-color", false, "Disable colored output")
	cmd.PersistentFlags().BoolVar(&fl.NoKeychain, "no-keychain", false, "Do not read credentials from keychain (use env/.env only)")

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newCompletionCmd())
	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newEmployeesCmd())
	cmd.AddCommand(newTimesheetsCmd())
	cmd.AddCommand(newRostersCmd())
	cmd.AddCommand(newLocationsCmd())
	cmd.AddCommand(newLeaveCmd())
	cmd.AddCommand(newDepartmentsCmd())
	cmd.AddCommand(newPayCmd())
	cmd.AddCommand(newResourceCmd())
	cmd.AddCommand(newMeCmd())
	cmd.AddCommand(newWebhooksCmd())
	cmd.AddCommand(newSalesCmd())
	cmd.AddCommand(newManagementCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newGetCmd())

	return cmd
}

// Execute creates a root command, runs it, and returns the result including
// the resolved output format and debug flag so the caller can format errors
// without consulting package-level globals.
func Execute() ExecuteResult {
	var result ExecuteResult
	root := NewRootCmd()

	// Wrap PersistentPreRunE to capture resolved values after flag parsing.
	origPreRun := root.PersistentPreRunE
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := origPreRun(cmd, args); err != nil {
			return err
		}
		ctx := cmd.Context()
		result.JSONOutput = outfmt.GetFormat(ctx) == "json"
		result.Debug = DebugFromContext(ctx)
		return nil
	}

	result.Err = root.ExecuteContext(context.Background())
	return result
}
