package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
)

func TestCompletionCommand(t *testing.T) {
	t.Run("bash completion generates output", func(t *testing.T) {
		root := NewRootCmd()
		buf := &bytes.Buffer{}
		ctx := context.Background()
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		root.SetArgs([]string{"completion", "bash"})

		err := root.ExecuteContext(ctx)

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "bash completion")
	})

	t.Run("zsh completion generates output", func(t *testing.T) {
		root := NewRootCmd()
		buf := &bytes.Buffer{}
		ctx := context.Background()
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		root.SetArgs([]string{"completion", "zsh"})

		err := root.ExecuteContext(ctx)

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "compdef")
	})

	t.Run("fish completion generates output", func(t *testing.T) {
		root := NewRootCmd()
		buf := &bytes.Buffer{}
		ctx := context.Background()
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		root.SetArgs([]string{"completion", "fish"})

		err := root.ExecuteContext(ctx)

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "complete")
	})

	t.Run("powershell completion generates output", func(t *testing.T) {
		root := NewRootCmd()
		buf := &bytes.Buffer{}
		ctx := context.Background()
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		root.SetArgs([]string{"completion", "powershell"})

		err := root.ExecuteContext(ctx)

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Register-ArgumentCompleter")
	})

	t.Run("requires shell argument", func(t *testing.T) {
		root := NewRootCmd()
		buf := &bytes.Buffer{}
		ctx := context.Background()
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		root.SetArgs([]string{"completion"})

		err := root.ExecuteContext(ctx)

		require.Error(t, err)
	})

	t.Run("rejects invalid shell", func(t *testing.T) {
		root := NewRootCmd()
		buf := &bytes.Buffer{}
		ctx := context.Background()
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		root.SetArgs([]string{"completion", "invalid"})

		err := root.ExecuteContext(ctx)

		require.Error(t, err)
	})
}

func TestCompletionCommand_ViaRootCmd(t *testing.T) {
	// Test that completion command is properly registered as a subcommand
	root := NewRootCmd()
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Use == "completion [bash|zsh|fish|powershell]" {
			found = true
			break
		}
	}
	assert.True(t, found, "completion command should be registered")
}
