package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func (f *Formatter) outputJSON(data any) error {
	query := GetQuery(f.ctx)
	if query != "" {
		return f.outputJQFiltered(data, query)
	}

	enc := json.NewEncoder(f.out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
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
		enc := json.NewEncoder(f.out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(v); err != nil {
			return err
		}
	}
	return nil
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
