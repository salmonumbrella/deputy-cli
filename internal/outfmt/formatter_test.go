package outfmt

import (
	"bytes"
	"context"
	"encoding/json"
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

	assert.Equal(t, "[]\n", out.String())
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
	assert.Equal(t, "[]\n", out.String())
}

func TestOutputJSONLines_NilSlice(t *testing.T) {
	out := &bytes.Buffer{}
	var items []map[string]any // nil slice

	err := outputJSONLines(out, items)
	require.NoError(t, err)
	assert.Equal(t, "[]\n", out.String())
}

func TestFormatter_EndTableWithoutStart(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := testContext(out)
	f := New(ctx)

	// Should not panic when EndTable is called without StartTable
	f.EndTable()
	assert.Empty(t, out.String())
}

func TestFormatter_OutputWithMeta(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "json")
	f := New(ctx)

	data := []map[string]any{
		{"Id": 1, "Name": "Test"},
		{"Id": 2, "Name": "Test2"},
	}

	err := f.OutputWithMeta(data, map[string]any{
		"count":  2,
		"limit":  10,
		"offset": 0,
	})
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(out.Bytes(), &result)
	require.NoError(t, err)

	assert.Contains(t, result, "items")
	assert.Contains(t, result, "meta")
	meta := result["meta"].(map[string]any)
	assert.Equal(t, float64(2), meta["count"])
	assert.Equal(t, float64(10), meta["limit"])
	assert.Equal(t, float64(0), meta["offset"])
}

func TestFormatter_OutputWithMeta_RawMode(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "json")
	ctx = WithRaw(ctx, true)
	f := New(ctx)

	data := []map[string]any{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
	}

	err := f.OutputWithMeta(data, map[string]any{
		"count":  2,
		"limit":  10,
		"offset": 0,
	})
	require.NoError(t, err)

	// In raw mode, should output JSON lines without meta wrapper
	lines := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))
	require.Len(t, lines, 2)
	assert.Contains(t, string(lines[0]), "\"id\":1")
	assert.Contains(t, string(lines[1]), "\"id\":2")
}

func TestAutoMeta(t *testing.T) {
	data := []int{1, 2, 3}
	meta := AutoMeta(data)
	assert.Equal(t, 3, meta["count"])
}

func TestAutoMeta_PointerToSlice(t *testing.T) {
	data := []string{"a", "b"}
	meta := AutoMeta(&data)
	assert.Equal(t, 2, meta["count"])
}

func TestAutoMeta_EmptySlice(t *testing.T) {
	data := []int{}
	meta := AutoMeta(data)
	assert.Equal(t, 0, meta["count"])
}

func TestAutoMeta_Nil(t *testing.T) {
	meta := AutoMeta(nil)
	assert.Equal(t, 0, meta["count"])
}

func TestAutoMeta_NilPointer(t *testing.T) {
	var data *[]int
	meta := AutoMeta(data)
	assert.Equal(t, 0, meta["count"])
}

func TestAutoMeta_NonSlice(t *testing.T) {
	data := map[string]string{"key": "value"}
	meta := AutoMeta(data)
	assert.Empty(t, meta)
}

func TestFormatter_OutputList(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "json")
	ctx = WithLimit(ctx, 10)
	ctx = WithOffset(ctx, 5)
	f := New(ctx)

	data := []map[string]any{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
	}

	err := f.OutputList(data)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(out.Bytes(), &result)
	require.NoError(t, err)

	assert.Contains(t, result, "items")
	assert.Contains(t, result, "meta")
	meta := result["meta"].(map[string]any)
	assert.Equal(t, float64(2), meta["count"])
	assert.Equal(t, float64(10), meta["limit"])
	assert.Equal(t, float64(5), meta["offset"])
}

func TestFormatter_OutputList_RawMode(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "json")
	ctx = WithRaw(ctx, true)
	ctx = WithLimit(ctx, 10)
	f := New(ctx)

	data := []map[string]any{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
	}

	err := f.OutputList(data)
	require.NoError(t, err)

	// In raw mode, should output JSON lines without wrapper
	lines := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))
	require.Len(t, lines, 2)
	assert.Contains(t, string(lines[0]), "\"id\":1")
	assert.Contains(t, string(lines[1]), "\"id\":2")
}

func TestFormatter_OutputList_TextReturnsError(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "text")
	f := New(ctx)

	data := []string{"a", "b"}
	err := f.OutputList(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "table methods")
}

func TestLimitOffsetContext(t *testing.T) {
	ctx := context.Background()

	// Test limit
	limit, ok := GetLimit(ctx)
	assert.False(t, ok)
	assert.Equal(t, 0, limit)

	ctx = WithLimit(ctx, 50)
	limit, ok = GetLimit(ctx)
	assert.True(t, ok)
	assert.Equal(t, 50, limit)

	// Test offset
	offset, ok := GetOffset(ctx)
	assert.False(t, ok)
	assert.Equal(t, 0, offset)

	ctx = WithOffset(ctx, 25)
	offset, ok = GetOffset(ctx)
	assert.True(t, ok)
	assert.Equal(t, 25, offset)
}

func TestFormatter_OutputList_NilSlice(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "json")
	f := New(ctx)

	type Employee struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}

	var employees []Employee // nil slice — simulates empty API response

	err := f.OutputList(employees)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(out.Bytes(), &result)
	require.NoError(t, err)

	// items must be [] not null — otherwise jq `.items[]` fails
	assert.NotNil(t, result["items"])
	items, ok := result["items"].([]any)
	require.True(t, ok, "items should be a JSON array, got %T", result["items"])
	assert.Empty(t, items)

	meta := result["meta"].(map[string]any)
	assert.Equal(t, float64(0), meta["count"])
}

func TestFormatter_OutputList_NilSlice_WithJQ(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "json")
	ctx = WithQuery(ctx, ".items[]")
	f := New(ctx)

	var employees []map[string]any // nil slice

	// Must not return error — this is the core bug fix
	err := f.OutputList(employees)
	require.NoError(t, err)

	// Empty output is correct — no items to iterate
	assert.Empty(t, out.String())
}

func TestFormatter_OutputWithMeta_NilSlice(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "json")
	f := New(ctx)

	var data []string // nil slice

	err := f.OutputWithMeta(data, map[string]any{"count": 0})
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(out.Bytes(), &result)
	require.NoError(t, err)

	// items must serialize as [] not null
	assert.NotNil(t, result["items"])
	items, ok := result["items"].([]any)
	require.True(t, ok, "items should be a JSON array, got %T", result["items"])
	assert.Empty(t, items)
}

// TestFormatter_OutputList_WithQuery tests that JQ queries work on OutputList
// which wraps data with {items: [...], meta: {...}} structure.
// This specifically tests that Go struct types are properly serialized before
// being passed to gojq (which requires JSON-compatible types).
func TestFormatter_OutputList_WithQuery(t *testing.T) {
	type Employee struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "items length",
			query:    ".items | length",
			expected: "2\n",
		},
		{
			name:     "meta count",
			query:    ".meta.count",
			expected: "2\n",
		},
		{
			name:     "first item id",
			query:    ".items[0].id",
			expected: "1\n",
		},
		{
			name:     "all names",
			query:    ".items[].name",
			expected: "\"Alice\"\n\"Bob\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			ctx := WithFormat(testContext(out), "json")
			ctx = WithQuery(ctx, tt.query)
			f := New(ctx)

			data := []Employee{
				{Id: 1, Name: "Alice"},
				{Id: 2, Name: "Bob"},
			}

			err := f.OutputList(data)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, out.String())
		})
	}
}

func TestOutputList_FailEmpty_ReturnsError(t *testing.T) {
	var buf bytes.Buffer
	ctx := context.Background()
	ctx = WithFormat(ctx, "json")
	ctx = WithFailEmpty(ctx, true)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{
		In:     bytes.NewReader(nil),
		Out:    &buf,
		ErrOut: &bytes.Buffer{},
	})

	f := New(ctx)
	err := f.OutputList([]any{})
	assert.ErrorIs(t, err, ErrEmptyResult)
	// Envelope should still have been written
	assert.Contains(t, buf.String(), `"items"`)
	assert.Contains(t, buf.String(), `"count"`)
}

func TestOutputList_FailEmpty_NilSlice(t *testing.T) {
	var buf bytes.Buffer
	ctx := context.Background()
	ctx = WithFormat(ctx, "json")
	ctx = WithFailEmpty(ctx, true)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{
		In:     bytes.NewReader(nil),
		Out:    &buf,
		ErrOut: &bytes.Buffer{},
	})

	f := New(ctx)
	var data []string // nil slice
	err := f.OutputList(data)
	assert.ErrorIs(t, err, ErrEmptyResult)
	// Envelope should still have been written
	assert.Contains(t, buf.String(), `"items"`)
}

func TestOutputList_FailEmpty_NoErrorWhenResults(t *testing.T) {
	var buf bytes.Buffer
	ctx := context.Background()
	ctx = WithFormat(ctx, "json")
	ctx = WithFailEmpty(ctx, true)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{
		In:     bytes.NewReader(nil),
		Out:    &buf,
		ErrOut: &bytes.Buffer{},
	})

	f := New(ctx)
	err := f.OutputList([]any{"item1"})
	assert.NoError(t, err)
}

func TestOutputList_NoFailEmpty_EmptyIsOK(t *testing.T) {
	var buf bytes.Buffer
	ctx := context.Background()
	ctx = WithFormat(ctx, "json")
	ctx = WithFailEmpty(ctx, false)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{
		In:     bytes.NewReader(nil),
		Out:    &buf,
		ErrOut: &bytes.Buffer{},
	})

	f := New(ctx)
	err := f.OutputList([]any{})
	assert.NoError(t, err)
}

// TestFormatter_JSONWithQuery_GoStruct tests that JQ queries work on Go structs
// (not just map[string]any)
func TestFormatter_JSONWithQuery_GoStruct(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	out := &bytes.Buffer{}
	ctx := WithFormat(testContext(out), "json")
	ctx = WithQuery(ctx, ".name")
	f := New(ctx)

	data := Person{Name: "Alice", Age: 30}
	err := f.Output(data)
	require.NoError(t, err)

	assert.Equal(t, "\"Alice\"\n", out.String())
}
