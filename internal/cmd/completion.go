package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate completion script for your shell.

To load completions:

Bash:
  $ source <(deputy completion bash)
  # To load completions for each session, add to ~/.bashrc:
  # source <(deputy completion bash)

Zsh:
  $ deputy completion zsh > "${fpath[1]}/_deputy"
  # You may need to start a new shell for this to take effect.

Fish:
  $ deputy completion fish | source
  # To load completions for each session, add to ~/.config/fish/config.fish:
  # deputy completion fish | source

PowerShell:
  PS> deputy completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> deputy completion powershell >> $PROFILE
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := iocontext.FromContext(cmd.Context()).Out
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(out)
			case "zsh":
				return cmd.Root().GenZshCompletion(out)
			case "fish":
				return cmd.Root().GenFishCompletion(out, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(out)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
	return cmd
}
