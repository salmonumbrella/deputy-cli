package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"text/tabwriter"

	"github.com/itchyny/gojq"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
)

type Formatter struct {
	ctx       context.Context
	out       io.Writer
	errOut    io.Writer
	tabWriter *tabwriter.Writer
}

func New(ctx context.Context) *Formatter {
	io := iocontext.FromContext(ctx)
	return &Formatter{
		ctx:    ctx,
		out:    io.Out,
		errOut: io.ErrOut,
	}
}

func (f *Formatter) Output(data any) error {
	format := GetFormat(f.ctx)
	if format == "json" {
		return f.outputJSON(data)
	}
	return fmt.Errorf("use table methods for text output")
}

// OutputWithMeta outputs data wrapped with metadata for agent consumption.
// In raw mode, outputs data without wrapper (raw mode is for JSON Lines).
func (f *Formatter) OutputWithMeta(data any, meta map[string]any) error {
	if IsRaw(f.ctx) {
		return f.Output(data) // Raw mode doesn't get meta wrapper
	}

	wrapped := map[string]any{
		"items": coerceNilSlice(data),
		"meta":  meta,
	}
	return f.Output(wrapped)
}

// OutputList outputs a list/array with standard metadata wrapper.
// Automatically includes count from data length, plus limit/offset from context.
// Use this for all list commands to ensure consistent agent-friendly output.
//
// When fail-empty is set via context, OutputList writes the envelope first
// (so agents always see valid JSON on stdout), then returns ErrEmptyResult
// if the data is empty.
func (f *Formatter) OutputList(data any) error {
	format := GetFormat(f.ctx)
	if format != "json" {
		return fmt.Errorf("use table methods for text output")
	}

	if IsRaw(f.ctx) {
		return f.outputJSON(data) // Raw mode outputs JSON Lines without wrapper
	}

	meta := AutoMeta(data)

	// Add limit/offset from context if present
	if limit, ok := GetLimit(f.ctx); ok {
		meta["limit"] = limit
	}
	if offset, ok := GetOffset(f.ctx); ok {
		meta["offset"] = offset
	}

	if err := f.OutputWithMeta(data, meta); err != nil {
		return err
	}

	// After writing the envelope, check if we should fail on empty
	if GetFailEmpty(f.ctx) && isEmptySlice(data) {
		return ErrEmptyResult
	}

	return nil
}

// isEmptySlice reports whether data represents an empty (or nil) slice/array.
func isEmptySlice(data any) bool {
	if data == nil {
		return true
	}
	rv := reflect.ValueOf(data)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return true
		}
		rv = rv.Elem()
	}
	if (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) && rv.Len() == 0 {
		return true
	}
	return false
}

// AutoMeta generates metadata for slices automatically.
func AutoMeta(data any) map[string]any {
	meta := map[string]any{}

	if data == nil {
		meta["count"] = 0
		return meta
	}

	rv := reflect.ValueOf(data)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			meta["count"] = 0
			return meta
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		meta["count"] = rv.Len()
	}

	return meta
}

func (f *Formatter) outputJSON(data any) error {
	query := GetQuery(f.ctx)
	if query != "" {
		return f.outputJQFiltered(data, query)
	}

	if IsRaw(f.ctx) {
		return outputJSONLines(f.out, data)
	}

	return encodeJSON(f.out, data, true)
}

func (f *Formatter) outputJQFiltered(data any, queryStr string) error {
	query, err := gojq.Parse(queryStr)
	if err != nil {
		return fmt.Errorf("invalid jq query: %w", err)
	}

	// gojq requires JSON-compatible types (maps, slices of interface{}, etc.)
	// not Go struct types. Serialize to JSON and back to get compatible types.
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to serialize data for jq: %w", err)
	}

	var jsonData any
	if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
		return fmt.Errorf("failed to prepare data for jq: %w", err)
	}

	iter := query.Run(jsonData)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return err
		}
		if err := encodeJSON(f.out, v, !IsRaw(f.ctx)); err != nil {
			return err
		}
	}
	return nil
}

// coerceNilSlice returns an empty []any{} when data is a nil slice,
// so JSON serialization produces [] instead of null.
// This prevents jq filters like .items[] from failing with
// "cannot iterate over null" when a list endpoint returns no results.
func coerceNilSlice(data any) any {
	if data == nil {
		return []any{}
	}
	rv := reflect.ValueOf(data)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return []any{}
		}
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Slice && rv.IsNil() {
		return []any{}
	}
	return data
}

func encodeJSON(w io.Writer, v any, pretty bool) error {
	enc := json.NewEncoder(w)
	if pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(v)
}

func outputJSONLines(w io.Writer, data any) error {
	if data == nil {
		return encodeJSON(w, []any{}, false)
	}

	val := reflect.ValueOf(data)
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return encodeJSON(w, []any{}, false)
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		if val.IsNil() {
			return encodeJSON(w, []any{}, false)
		}
		for i := 0; i < val.Len(); i++ {
			if err := encodeJSON(w, val.Index(i).Interface(), false); err != nil {
				return err
			}
		}
		return nil
	default:
		return encodeJSON(w, data, false)
	}
}

func (f *Formatter) StartTable(headers []string) {
	f.tabWriter = tabwriter.NewWriter(f.out, 0, 0, 2, ' ', 0)
	for i, h := range headers {
		if i > 0 {
			_, _ = fmt.Fprint(f.tabWriter, "\t")
		}
		_, _ = fmt.Fprint(f.tabWriter, h)
	}
	_, _ = fmt.Fprintln(f.tabWriter)
}

func (f *Formatter) Row(values ...string) {
	for i, v := range values {
		if i > 0 {
			_, _ = fmt.Fprint(f.tabWriter, "\t")
		}
		_, _ = fmt.Fprint(f.tabWriter, v)
	}
	_, _ = fmt.Fprintln(f.tabWriter)
}

func (f *Formatter) EndTable() {
	if f.tabWriter != nil {
		_ = f.tabWriter.Flush()
	}
}
