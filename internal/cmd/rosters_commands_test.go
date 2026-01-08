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

func TestRostersGetCommand_WithMockClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/resource/Roster/123", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.Roster{
			Id:        123,
			Date:      "2024-01-15",
			StartTime: 1705312800,
			EndTime:   1705341600,
			Employee:  10,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newRostersGetCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "ID:         123")
	assert.Contains(t, buf.String(), "Date:       2024-01-15")
}

func TestRostersSwapCommand_WithMockClient(t *testing.T) {
	t.Run("outputs table", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/v1/supervise/roster/123/swap", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.SwapRoster{
				{Id: 1, Date: "2024-01-15", StartTime: 1705312800, EndTime: 1705341600, Employee: 10},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newRostersSwapCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "EMPLOYEE")
		assert.Contains(t, buf.String(), "10")
	})

	t.Run("outputs json when requested", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.SwapRoster{{Id: 1, Employee: 10}})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = outfmt.WithFormat(ctx, "json")
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newRostersSwapCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Employee": 10`)
	})
}

func TestRostersCreateCopyPublishDiscardCommands_WithMockClient(t *testing.T) {
	t.Run("create roster", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/v1/supervise/roster", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Roster{Id: 10})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newRostersCreateCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--employee", "123", "--opunit", "5", "--start-time", "1705312800", "--end-time", "1705341600"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Created roster 10")
	})

	t.Run("copy roster", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/v1/supervise/roster/copy", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newRostersCopyCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--from-date", "2024-01-01", "--to-date", "2024-01-08", "--location", "2"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Roster copied from 2024-01-01 to 2024-01-08")
	})

	t.Run("publish roster", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/v1/supervise/roster/publish", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newRostersPublishCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--from-date", "2024-01-01", "--to-date", "2024-01-08", "--location", "2"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Rosters published from 2024-01-01 to 2024-01-08")
	})

	t.Run("discard roster", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/v1/supervise/roster/discard", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newRostersDiscardCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--from-date", "2024-01-01", "--to-date", "2024-01-08", "--location", "2"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Roster changes discarded from 2024-01-01 to 2024-01-08")
	})
}
