package iocontext

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIO(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	in := strings.NewReader("input")

	io := &IO{Out: out, ErrOut: errOut, In: in}

	assert.Equal(t, out, io.Out)
	assert.Equal(t, errOut, io.ErrOut)
	assert.Equal(t, in, io.In)
}

func TestIOContext_RoundTrip(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	in := strings.NewReader("test input")

	io := &IO{Out: out, ErrOut: errOut, In: in}
	ctx := WithIO(context.Background(), io)

	retrieved := FromContext(ctx)
	require.NotNil(t, retrieved)
	assert.Equal(t, io, retrieved)
}

func TestFromContext_NoIO_ReturnsDefault(t *testing.T) {
	ctx := context.Background()

	retrieved := FromContext(ctx)

	// FromContext returns DefaultIO when no IO is set in context
	require.NotNil(t, retrieved)
	assert.Equal(t, os.Stdout, retrieved.Out)
	assert.Equal(t, os.Stderr, retrieved.ErrOut)
	assert.Equal(t, os.Stdin, retrieved.In)
}

func TestDefaultIO(t *testing.T) {
	io := DefaultIO()

	require.NotNil(t, io)
	assert.Equal(t, os.Stdout, io.Out)
	assert.Equal(t, os.Stderr, io.ErrOut)
	assert.Equal(t, os.Stdin, io.In)
}

func TestHasIO(t *testing.T) {
	ctx := context.Background()
	assert.False(t, HasIO(ctx))

	ctx = WithIO(ctx, &IO{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}, In: strings.NewReader("input")})
	assert.True(t, HasIO(ctx))
}
