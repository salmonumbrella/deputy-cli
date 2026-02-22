package cmd

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
)

func TestRootCmd_VersionFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	root := NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"--version"})
	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "deputy version")
}

func TestNewRootCmd(t *testing.T) {
	t.Run("creates root command with correct name", func(t *testing.T) {
		cmd := NewRootCmd()
		assert.Equal(t, "deputy", cmd.Use)
	})

	t.Run("has correct short and long description", func(t *testing.T) {
		cmd := NewRootCmd()
		assert.Equal(t, "CLI for Deputy workforce management API", cmd.Short)
		assert.Contains(t, cmd.Long, "command-line interface")
		assert.Contains(t, cmd.Long, "AI agent automation")
	})

	t.Run("has expected global flags", func(t *testing.T) {
		cmd := NewRootCmd()

		// Check persistent flags exist
		outputFlag := cmd.PersistentFlags().Lookup("output")
		require.NotNil(t, outputFlag)
		assert.Equal(t, "o", outputFlag.Shorthand)
		assert.Equal(t, "text", outputFlag.DefValue)

		debugFlag := cmd.PersistentFlags().Lookup("debug")
		require.NotNil(t, debugFlag)
		assert.Equal(t, "false", debugFlag.DefValue)

		queryFlag := cmd.PersistentFlags().Lookup("query")
		require.NotNil(t, queryFlag)
		assert.Equal(t, "q", queryFlag.Shorthand)
		assert.Equal(t, "", queryFlag.DefValue)

		noColorFlag := cmd.PersistentFlags().Lookup("no-color")
		require.NotNil(t, noColorFlag)
		assert.Equal(t, "false", noColorFlag.DefValue)
	})

	t.Run("has expected subcommands", func(t *testing.T) {
		cmd := NewRootCmd()
		subCmds := cmd.Commands()
		names := make([]string, len(subCmds))
		for i, c := range subCmds {
			names[i] = c.Name()
		}

		expectedCmds := []string{
			"version",
			"completion",
			"auth",
			"employees",
			"timesheets",
			"rosters",
			"locations",
			"leave",
			"departments",
			"pay",
			"resource",
			"me",
			"webhooks",
			"sales",
			"management",
		}
		for _, expected := range expectedCmds {
			assert.Contains(t, names, expected, "missing subcommand: %s", expected)
		}
	})

	t.Run("has correct subcommand count", func(t *testing.T) {
		cmd := NewRootCmd()
		// 17 subcommands: version, completion, auth, employees, timesheets, rosters, locations,
		// leave, departments, pay, resource, me, webhooks, sales, management, list, get
		assert.Len(t, cmd.Commands(), 17)
	})

	t.Run("help executes without error", func(t *testing.T) {
		var buf bytes.Buffer
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{
			In:     bytes.NewReader(nil),
			Out:    &buf,
			ErrOut: &bytes.Buffer{},
		})

		cmd := NewRootCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--help"})

		err := cmd.ExecuteContext(ctx)

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "deputy - CLI for Deputy workforce management API")
		assert.Contains(t, output, "Exit codes:")
		assert.Contains(t, output, "Global flags:")
	})

	t.Run("silences usage on error", func(t *testing.T) {
		cmd := NewRootCmd()
		assert.True(t, cmd.SilenceUsage)
		assert.True(t, cmd.SilenceErrors)
	})
}

func TestRootCmd_GlobalFlagsUsage(t *testing.T) {
	t.Run("output flag accepts text", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"--output", "text", "version"})

		err := cmd.Execute()

		require.NoError(t, err)
	})

	t.Run("output flag accepts json", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"--output", "json", "version"})

		err := cmd.Execute()

		require.NoError(t, err)
	})

	t.Run("short output flag works", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"-o", "json", "version"})

		err := cmd.Execute()

		require.NoError(t, err)
	})

	t.Run("debug flag works", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"--debug", "version"})

		err := cmd.Execute()

		require.NoError(t, err)
	})

	t.Run("query flag works with short form", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"-q", ".version", "version"})

		err := cmd.Execute()

		require.NoError(t, err)
	})

	t.Run("no-color flag works", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"--no-color", "version"})

		err := cmd.Execute()

		require.NoError(t, err)
	})

	t.Run("multiple flags can be combined", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"--debug", "--no-color", "-o", "json", "version"})

		err := cmd.Execute()

		require.NoError(t, err)
	})
}

func TestRootCmd_UnknownCommand(t *testing.T) {
	t.Run("returns error on unknown command", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"unknown-command"})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command")
	})

	t.Run("returns error on unknown flag", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"--unknown-flag"})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown flag")
	})
}

func TestExecute(t *testing.T) {
	t.Run("execute function exists and is callable", func(t *testing.T) {
		fn := Execute
		assert.NotNil(t, fn)
	})
}

func TestExecute_UsesArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"deputy", "version"}
	result := Execute()
	require.NoError(t, result.Err)
}

func TestExecute_UnknownCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"deputy", "not-a-command"}
	result := Execute()
	require.Error(t, result.Err)
	assert.Equal(t, ExitGeneral, result.ExitCode)
}

func TestExecute_ReturnsJSONOutput(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"deputy", "--output", "json", "version"}
	result := Execute()
	require.NoError(t, result.Err)
	assert.True(t, result.JSONOutput)
}

func TestExecute_ReturnsDebug(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"deputy", "--debug", "version"}
	result := Execute()
	require.NoError(t, result.Err)
	assert.True(t, result.Debug)
}

func TestRootCmd_HelpForSubcommands(t *testing.T) {
	subcommands := []struct {
		name        string
		description string
	}{
		{"version", "Print version information"},
		{"completion", "Generate shell completion"},
		{"auth", "Authentication"},
		{"employees", "Manage employees"},
		{"timesheets", "Manage timesheets"},
		{"rosters", "Manage rosters"},
		{"locations", "Manage locations"},
		{"leave", "Manage leave"},
		{"departments", "Manage departments"},
		{"pay", "Manage pay rates and agreements"},
		{"resource", "Generic resource"},
		{"me", "Current user"},
		{"webhooks", "Manage webhooks"},
		{"sales", "Manage sales"},
		{"management", "Management"},
	}

	for _, sc := range subcommands {
		t.Run(sc.name+" help", func(t *testing.T) {
			cmd := NewRootCmd()
			buf := &bytes.Buffer{}
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs([]string{sc.name, "--help"})

			err := cmd.Execute()

			require.NoError(t, err)
			output := buf.String()
			assert.NotEmpty(t, output, "help output should not be empty for %s", sc.name)
		})
	}
}

func TestAutoJSON_NonTTY(t *testing.T) {
	root := NewRootCmd()
	var resolvedFormat string
	root.AddCommand(&cobra.Command{
		Use: "probe",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedFormat = outfmt.GetFormat(cmd.Context())
			return nil
		},
	})

	ctx := iocontext.WithIO(context.Background(), &iocontext.IO{
		In:     bytes.NewReader(nil),
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	})
	root.SetContext(ctx)
	root.SetArgs([]string{"probe"})
	err := root.ExecuteContext(ctx)
	require.NoError(t, err)
	// In tests, stdout is a buffer (not a TTY), so format should be "json"
	assert.Equal(t, "json", resolvedFormat)
}

func TestAutoJSON_ExplicitTextOverride(t *testing.T) {
	root := NewRootCmd()
	var resolvedFormat string
	root.AddCommand(&cobra.Command{
		Use: "probe",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedFormat = outfmt.GetFormat(cmd.Context())
			return nil
		},
	})

	ctx := iocontext.WithIO(context.Background(), &iocontext.IO{
		In:     bytes.NewReader(nil),
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	})
	root.SetContext(ctx)
	root.SetArgs([]string{"--output", "text", "probe"})
	err := root.ExecuteContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, "text", resolvedFormat)
}

func TestAutoJSON_EnvOverride(t *testing.T) {
	t.Setenv("DEPUTY_OUTPUT", "text")
	root := NewRootCmd()
	var resolvedFormat string
	root.AddCommand(&cobra.Command{
		Use: "probe",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedFormat = outfmt.GetFormat(cmd.Context())
			return nil
		},
	})

	ctx := iocontext.WithIO(context.Background(), &iocontext.IO{
		In:     bytes.NewReader(nil),
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	})
	root.SetContext(ctx)
	root.SetArgs([]string{"probe"})
	err := root.ExecuteContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, "text", resolvedFormat)
}

func TestAutoJSON_RawOverridesEnv(t *testing.T) {
	t.Setenv("DEPUTY_OUTPUT", "text")
	root := NewRootCmd()
	var resolvedFormat string
	var resolvedRaw bool
	root.AddCommand(&cobra.Command{
		Use: "probe",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedFormat = outfmt.GetFormat(cmd.Context())
			resolvedRaw = outfmt.IsRaw(cmd.Context())
			return nil
		},
	})

	ctx := iocontext.WithIO(context.Background(), &iocontext.IO{
		In:     bytes.NewReader(nil),
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	})
	root.SetContext(ctx)
	root.SetArgs([]string{"--raw", "probe"})
	err := root.ExecuteContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, "json", resolvedFormat)
	assert.True(t, resolvedRaw)
}

func TestAutoJSON_EnvInvalid(t *testing.T) {
	t.Setenv("DEPUTY_OUTPUT", "xml")
	root := NewRootCmd()
	root.AddCommand(&cobra.Command{
		Use:  "probe",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	})

	ctx := iocontext.WithIO(context.Background(), &iocontext.IO{
		In:     bytes.NewReader(nil),
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	})
	root.SetContext(ctx)
	root.SetArgs([]string{"probe"})
	err := root.ExecuteContext(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid DEPUTY_OUTPUT")
}

func TestRootHelp_UsesEmbeddedHelpText(t *testing.T) {
	var buf bytes.Buffer
	ctx := iocontext.WithIO(context.Background(), &iocontext.IO{
		In:     bytes.NewReader(nil),
		Out:    &buf,
		ErrOut: &bytes.Buffer{},
	})

	root := NewRootCmd()
	root.SetContext(ctx)
	root.SetOut(&buf)
	root.SetArgs([]string{"--help"})
	err := root.ExecuteContext(ctx)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "deputy - CLI for Deputy workforce management API")
	assert.Contains(t, output, "Exit codes:")
	assert.Contains(t, output, "Global flags:")
	assert.Contains(t, output, "Environment:")
	// Should NOT contain Cobra boilerplate
	assert.NotContains(t, output, "Available Commands:")
}

func TestSubcommandHelp_UsesCobra(t *testing.T) {
	var buf bytes.Buffer
	ctx := iocontext.WithIO(context.Background(), &iocontext.IO{
		In:     bytes.NewReader(nil),
		Out:    &buf,
		ErrOut: &bytes.Buffer{},
	})

	root := NewRootCmd()
	root.SetContext(ctx)
	root.SetOut(&buf)
	root.SetArgs([]string{"employees", "--help"})
	err := root.ExecuteContext(ctx)
	require.NoError(t, err)

	output := buf.String()
	// Subcommands should use Cobra's default help
	assert.Contains(t, output, "Available Commands:")
}
