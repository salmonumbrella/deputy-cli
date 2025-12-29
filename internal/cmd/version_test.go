package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommitSHA := CommitSHA
	origBuildDate := BuildDate
	defer func() {
		Version = origVersion
		CommitSHA = origCommitSHA
		BuildDate = origBuildDate
	}()

	t.Run("executes without error", func(t *testing.T) {
		cmd := newVersionCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()

		require.NoError(t, err)
	})

	t.Run("output contains version information", func(t *testing.T) {
		Version = "1.2.3"
		CommitSHA = "abc123"
		BuildDate = "2024-01-15"

		cmd := newVersionCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "deputy version 1.2.3")
		assert.Contains(t, output, "commit: abc123")
		assert.Contains(t, output, "built:  2024-01-15")
	})

	t.Run("output contains default dev values", func(t *testing.T) {
		Version = "dev"
		CommitSHA = "unknown"
		BuildDate = "unknown"

		cmd := newVersionCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "deputy version dev")
		assert.Contains(t, output, "commit: unknown")
		assert.Contains(t, output, "built:  unknown")
	})
}

func TestVersionCommand_ViaRootCmd(t *testing.T) {
	// Test that version command is properly registered as a subcommand
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"version"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "deputy version")
}
