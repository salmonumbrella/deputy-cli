package outfmt

import "context"

type contextKey string

const (
	formatKey    contextKey = "format"
	queryKey     contextKey = "query"
	rawKey       contextKey = "raw"
	limitKey     contextKey = "limit"
	offsetKey    contextKey = "offset"
	failEmptyKey contextKey = "failEmpty"
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

func WithLimit(ctx context.Context, limit int) context.Context {
	return context.WithValue(ctx, limitKey, limit)
}

func GetLimit(ctx context.Context) (int, bool) {
	if l, ok := ctx.Value(limitKey).(int); ok {
		return l, true
	}
	return 0, false
}

func WithOffset(ctx context.Context, offset int) context.Context {
	return context.WithValue(ctx, offsetKey, offset)
}

func GetOffset(ctx context.Context) (int, bool) {
	if o, ok := ctx.Value(offsetKey).(int); ok {
		return o, true
	}
	return 0, false
}

func WithFailEmpty(ctx context.Context, failEmpty bool) context.Context {
	return context.WithValue(ctx, failEmptyKey, failEmpty)
}

func GetFailEmpty(ctx context.Context) bool {
	if v, ok := ctx.Value(failEmptyKey).(bool); ok {
		return v
	}
	return false
}
