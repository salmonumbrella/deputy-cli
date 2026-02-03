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
		"data":  data,
		"_meta": meta,
	}
	return f.Output(wrapped)
}

// AutoMeta generates metadata for slices automatically.
func AutoMeta(data any) map[string]any {
	meta := map[string]any{}

	if data == nil {
		return meta
	}

	rv := reflect.ValueOf(data)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
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

	iter := query.Run(data)
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

func encodeJSON(w io.Writer, v any, pretty bool) error {
	enc := json.NewEncoder(w)
	if pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(v)
}

func outputJSONLines(w io.Writer, data any) error {
	if data == nil {
		return encodeJSON(w, nil, false)
	}

	val := reflect.ValueOf(data)
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return encodeJSON(w, nil, false)
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Slice, reflect.Array:
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
