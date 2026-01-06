package outfmt

import "context"

type contextKey string

const (
	formatKey contextKey = "format"
	queryKey  contextKey = "query"
	rawKey    contextKey = "raw"
)

func WithFormat(ctx context.Context, format string) context.Context {
	return context.WithValue(ctx, formatKey, format)
}

func GetFormat(ctx context.Context) string {
	if f, ok := ctx.Value(formatKey).(string); ok {
		return f
	}
	return "text"
}

func WithQuery(ctx context.Context, query string) context.Context {
	return context.WithValue(ctx, queryKey, query)
}

func GetQuery(ctx context.Context) string {
	if q, ok := ctx.Value(queryKey).(string); ok {
		return q
	}
	return ""
}

func WithRaw(ctx context.Context, raw bool) context.Context {
	return context.WithValue(ctx, rawKey, raw)
}

func IsRaw(ctx context.Context) bool {
	if raw, ok := ctx.Value(rawKey).(bool); ok {
		return raw
	}
	return false
}
