package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocationsGetCommand_WithMockClient(t *testing.T) {
	t.Run("outputs text details", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/v1/resource/Company/123", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Location{
				Id:          123,
				CompanyName: "HQ",
				Code:        "HQ",
				Address:     json.RawMessage(`"1 Main St"`),
				Timezone:    "UTC",
				Active:      true,
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newLocationsGetCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})

		err := cmd.Execute()
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Name:     HQ")
		assert.Contains(t, output, "Code:     HQ")
		assert.Contains(t, output, "Timezone: UTC")
	})

	t.Run("outputs json when requested", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Location{Id: 123, CompanyName: "HQ"})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = outfmt.WithFormat(ctx, "json")
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newLocationsGetCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Id": 123`)
	})
}

func TestLocationsAddCommand_WithMockClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/supervise/location", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.Location{Id: 9, CompanyName: "HQ"})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newLocationsAddCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--name", "HQ"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Created location 9")
}

func TestLocationsAddCommand_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.Location{Id: 11, CompanyName: "HQ"})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
	ctx = outfmt.WithFormat(ctx, "json")

	cmd := newLocationsAddCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--name", "HQ"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"Id": 11`)
}

func TestLocationsUpdateCommand_WithMockClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/api/v1/supervise/location/9", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.Location{Id: 9, CompanyName: "HQ"})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newLocationsUpdateCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"9", "--name", "HQ"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Updated location 9")
}

func TestLocationsListCommand_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]api.Location{
			{Id: 1, CompanyName: "HQ"},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
	ctx = outfmt.WithFormat(ctx, "json")

	cmd := newLocationsListCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	err := cmd.Execute()

	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"Id": 1`)
}

func TestLocationsSettingsCommand_WithMockClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/supervise/location/123/settings", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.LocationSettings{
			Id: 123,
			Settings: map[string]interface{}{
				"key": "value",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newLocationsSettingsCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Location 123 Settings:")
	assert.Contains(t, buf.String(), "key: value")
}

func TestLocationsSettingsCommand_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.LocationSettings{
			Id:       123,
			Settings: map[string]interface{}{"key": "value"},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
	ctx = outfmt.WithFormat(ctx, "json")

	cmd := newLocationsSettingsCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"Id": 123`)
}

func TestLocationsSettingsUpdateCommand_WithMockClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/supervise/location/123/settings", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newLocationsSettingsUpdateCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"123", "--settings", `{"key":"value"}`})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Updated settings for location 123")
}

func TestLocationsSettingsUpdateCommand_InvalidInput(t *testing.T) {
	t.Run("missing settings", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newLocationsSettingsUpdateCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--settings is required")
	})

	t.Run("invalid json", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newLocationsSettingsUpdateCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--settings", "{invalid"})

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON")
	})
}

func TestLocationsArchiveDeleteCommands_WithMockClient(t *testing.T) {
	t.Run("archive location", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/v1/supervise/location/123/archive", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newLocationsArchiveCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--yes"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Location 123 archived")
	})

	t.Run("delete location", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "/api/v1/supervise/location/123", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newLocationsDeleteCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--yes"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Location 123 deleted")
	})
}

func TestLocationsArchiveDeleteCommands_Cancelled(t *testing.T) {
	inBuf := bytes.NewBufferString("n\n")
	outBuf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = iocontext.WithIO(ctx, &iocontext.IO{In: inBuf, Out: outBuf, ErrOut: outBuf})

	cmd := newLocationsArchiveCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(outBuf)
	cmd.SetArgs([]string{"123"})
	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}

func TestLocationsDeleteCommand_Cancelled(t *testing.T) {
	inBuf := bytes.NewBufferString("n\n")
	outBuf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = iocontext.WithIO(ctx, &iocontext.IO{In: inBuf, Out: outBuf, ErrOut: outBuf})

	cmd := newLocationsDeleteCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(outBuf)
	cmd.SetArgs([]string{"123"})
	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}
