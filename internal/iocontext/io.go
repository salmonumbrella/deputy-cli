package iocontext

import (
	"context"
	"io"
	"os"
)

type contextKey string

const ioKey contextKey = "io"

type IO struct {
	Out    io.Writer
	ErrOut io.Writer
	In     io.Reader
}

func DefaultIO() *IO {
	return &IO{
		Out:    os.Stdout,
		ErrOut: os.Stderr,
		In:     os.Stdin,
	}
}

func WithIO(ctx context.Context, io *IO) context.Context {
	return context.WithValue(ctx, ioKey, io)
}

func FromContext(ctx context.Context) *IO {
	if io, ok := ctx.Value(ioKey).(*IO); ok {
		return io
	}
	return DefaultIO()
}

// HasIO returns true if the context has an IO value set
func HasIO(ctx context.Context) bool {
	_, ok := ctx.Value(ioKey).(*IO)
	return ok
}
