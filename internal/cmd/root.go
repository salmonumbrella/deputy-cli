package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
)

var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildDate = "unknown"
)

type globalFlags struct {
	Output  string
	Debug   bool
	Query   string
	NoColor bool
}

var flags globalFlags

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deputy",
		Short: "CLI for Deputy workforce management API",
		Long:  "A command-line interface for interacting with the Deputy API.\nDesigned for both human users and AI agent automation.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			// Preserve existing IO if already set (e.g., in tests), otherwise use defaults
			if !iocontext.HasIO(ctx) {
				ctx = iocontext.WithIO(ctx, iocontext.DefaultIO())
			}
			ctx = outfmt.WithFormat(ctx, flags.Output)
			ctx = outfmt.WithQuery(ctx, flags.Query)
			ctx = WithDebug(ctx, flags.Debug)
			cmd.SetContext(ctx)
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVarP(&flags.Output, "output", "o", "text", "Output format: text or json")
	cmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "Enable debug logging")
	cmd.PersistentFlags().StringVarP(&flags.Query, "query", "q", "", "JQ filter for JSON output")
	cmd.PersistentFlags().BoolVar(&flags.NoColor, "no-color", false, "Disable colored output")

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newCompletionCmd())
	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newEmployeesCmd())
	cmd.AddCommand(newTimesheetsCmd())
	cmd.AddCommand(newRostersCmd())
	cmd.AddCommand(newLocationsCmd())
	cmd.AddCommand(newLeaveCmd())
	cmd.AddCommand(newDepartmentsCmd())
	cmd.AddCommand(newResourceCmd())
	cmd.AddCommand(newMeCmd())
	cmd.AddCommand(newWebhooksCmd())
	cmd.AddCommand(newSalesCmd())
	cmd.AddCommand(newManagementCmd())

	return cmd
}

func Execute() error {
	return NewRootCmd().ExecuteContext(context.Background())
}
