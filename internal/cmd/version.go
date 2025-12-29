package cmd

import (
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("deputy version %s\n", Version)
			cmd.Printf("  commit: %s\n", CommitSHA)
			cmd.Printf("  built:  %s\n", BuildDate)
		},
	}
}
