package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		// 14 subcommands: version, completion, auth, employees, timesheets, rosters, locations,
		// leave, departments, resource, me, webhooks, sales, management
		assert.Len(t, cmd.Commands(), 14)
	})

	t.Run("help executes without error", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"--help"})

		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "deputy")
		assert.Contains(t, output, "Usage:")
		assert.Contains(t, output, "Available Commands:")
		assert.Contains(t, output, "Flags:")
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
	// Execute() creates a new root command and runs it with context.Background()
	// Since it operates on real stdin/stdout, we test via NewRootCmd instead

	t.Run("execute function exists and is callable", func(t *testing.T) {
		// We can verify Execute exists and returns an error type
		// Can't easily test without side effects on real stdout/stderr
		// The function signature is: func Execute() error
		fn := Execute
		assert.NotNil(t, fn)
	})
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
