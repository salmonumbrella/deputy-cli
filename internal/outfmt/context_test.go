package outfmt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatContext_RoundTrip(t *testing.T) {
	ctx := WithFormat(context.Background(), "json")

	format := GetFormat(ctx)
	assert.Equal(t, "json", format)
}

func TestQueryContext_RoundTrip(t *testing.T) {
	ctx := WithQuery(context.Background(), ".items[0].name")

	query := GetQuery(ctx)
	assert.Equal(t, ".items[0].name", query)
}

func TestFormatContext_Default(t *testing.T) {
	ctx := context.Background()

	format := GetFormat(ctx)
	assert.Equal(t, "text", format) // Default should be text
}

func TestQueryContext_Default(t *testing.T) {
	ctx := context.Background()

	query := GetQuery(ctx)
	assert.Equal(t, "", query) // Default should be empty
}
