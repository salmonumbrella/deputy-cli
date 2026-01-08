package outfmt

import (
	"bytes"
	"context"
	"testing"

	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to create a test context with custom output buffer
func testContext(out *bytes.Buffer) context.Context {
	ctx := context.Background()
	io := &iocontext.IO{
		Out:    out,
		ErrOut: out,
	}
	return iocontext.WithIO(ctx, io)
}

func TestFormatter_JSON(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "json")
	f := New(ctx)

	data := map[string]string{"name": "test", "value": "123"}
	err := f.Output(data)
	require.NoError(t, err)

	assert.Contains(t, out.String(), `"name": "test"`)
	assert.Contains(t, out.String(), `"value": "123"`)
}

func TestFormatter_JSONWithQuery(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "json")
	ctx = WithQuery(ctx, ".name")
	f := New(ctx)

	// gojq requires map[string]any for proper query execution
	data := map[string]any{"name": "test", "value": "123"}
	err := f.Output(data)
	require.NoError(t, err)

	assert.Equal(t, "\"test\"\n", out.String())
}

func TestFormatter_JSONWithInvalidQuery(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "json")
	ctx = WithQuery(ctx, "[invalid")
	f := New(ctx)

	data := map[string]string{"name": "test"}
	err := f.Output(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid jq query")
}

func TestFormatter_Text(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "text")
	f := New(ctx)

	// Text formatter uses StartTable/Row/EndTable for structured output
	f.StartTable([]string{"ID", "Name"})
	f.Row("1", "Alice")
	f.Row("2", "Bob")
	f.EndTable()

	output := out.String()
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "Alice")
	assert.Contains(t, output, "Bob")
}

func TestFormatter_TextOutputReturnsError(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "text")
	f := New(ctx)

	// Output() with text format returns an error directing to use table methods
	data := map[string]string{"name": "test"}
	err := f.Output(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "table methods")
}

func TestFormatter_TableAlignment(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := testContext(out)
	f := New(ctx)

	f.StartTable([]string{"Short", "LongerHeader"})
	f.Row("a", "value1")
	f.Row("longer", "v2")
	f.EndTable()

	output := out.String()
	// Verify all content is present
	assert.Contains(t, output, "Short")
	assert.Contains(t, output, "LongerHeader")
	assert.Contains(t, output, "a")
	assert.Contains(t, output, "value1")
	assert.Contains(t, output, "longer")
	assert.Contains(t, output, "v2")
}

func TestFormatter_JSONComplexData(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "json")
	f := New(ctx)

	data := map[string]any{
		"id":     123,
		"active": true,
		"nested": map[string]string{
			"key": "value",
		},
		"items": []string{"a", "b", "c"},
	}
	err := f.Output(data)
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, `"id": 123`)
	assert.Contains(t, output, `"active": true`)
	assert.Contains(t, output, `"key": "value"`)
}

func TestFormatter_JSONWithArrayQuery(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "json")
	ctx = WithQuery(ctx, ".items[]")
	f := New(ctx)

	// gojq requires []any for array iteration queries
	data := map[string]any{
		"items": []any{"a", "b", "c"},
	}
	err := f.Output(data)
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, `"a"`)
	assert.Contains(t, output, `"b"`)
	assert.Contains(t, output, `"c"`)
}

func TestFormatter_JSONLinesRaw(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "json")
	ctx = WithRaw(ctx, true)
	f := New(ctx)

	data := []map[string]any{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
	}

	err := f.Output(data)
	require.NoError(t, err)

	output := out.String()
	lines := bytes.Split(bytes.TrimSpace([]byte(output)), []byte("\n"))
	require.Len(t, lines, 2)
	assert.Contains(t, string(lines[0]), "\"id\":1")
	assert.Contains(t, string(lines[0]), "\"name\":\"Alice\"")
	assert.Contains(t, string(lines[1]), "\"id\":2")
	assert.Contains(t, string(lines[1]), "\"name\":\"Bob\"")
}

func TestOutputJSONLines_NilData(t *testing.T) {
	out := &bytes.Buffer{}

	err := outputJSONLines(out, nil)
	require.NoError(t, err)

	assert.Equal(t, "null\n", out.String())
}

func TestOutputJSONLines_PointerToSlice(t *testing.T) {
	out := &bytes.Buffer{}
	items := []map[string]any{
		{"id": 1},
		{"id": 2},
	}

	err := outputJSONLines(out, &items)
	require.NoError(t, err)

	lines := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))
	require.Len(t, lines, 2)
	assert.Contains(t, string(lines[0]), `"id":1`)
	assert.Contains(t, string(lines[1]), `"id":2`)
}

func TestOutputJSONLines_NilPointer(t *testing.T) {
	out := &bytes.Buffer{}
	var items *[]map[string]any

	err := outputJSONLines(out, items)
	require.NoError(t, err)
	assert.Equal(t, "null\n", out.String())
}

func TestFormatter_EndTableWithoutStart(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := testContext(out)
	f := New(ctx)

	// Should not panic when EndTable is called without StartTable
	f.EndTable()
	assert.Empty(t, out.String())
}
