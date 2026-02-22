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

func TestEmployeesGetCommand_WithMockClient(t *testing.T) {
	t.Run("outputs text details", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/v1/supervise/employee/123", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Employee{
				Id:          123,
				DisplayName: "Jane Doe",
				FirstName:   "Jane",
				LastName:    "Doe",
				Email:       "jane@example.com",
				Mobile:      "1234",
				Active:      true,
				Company:     1,
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesGetCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})

		err := cmd.Execute()
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Name:       Jane Doe")
		assert.Contains(t, output, "Email:      jane@example.com")
	})

	t.Run("outputs json when requested", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Employee{Id: 123, DisplayName: "Jane Doe"})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = outfmt.WithFormat(ctx, "json")
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesGetCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Id": 123`)
	})
}

func TestEmployeesInviteCommand_WithMockClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/supervise/employee/123/invite", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newEmployeesInviteCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Invitation sent to employee 123")
}

func TestEmployeesReactivateCommand_WithMockClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/Employee/123", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.Employee{Id: 123, Active: true})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newEmployeesReactivateCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"123"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Employee 123 reactivated")
}

func TestEmployeesInviteReactivate_InvalidID(t *testing.T) {
	t.Run("invite invalid id", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesInviteCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"bad"})

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid employee ID")
	})

	t.Run("reactivate invalid id", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newEmployeesReactivateCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"bad"})

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid employee ID")
	})
}

func TestEmployeesDeleteCommand_InvalidID(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newEmployeesDeleteCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"bad"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid employee ID")
}

func TestEmployeesTerminateCommand_Cancelled(t *testing.T) {
	inBuf := bytes.NewBufferString("n\n")
	outBuf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = iocontext.WithIO(ctx, &iocontext.IO{In: inBuf, Out: outBuf, ErrOut: outBuf})

	cmd := newEmployeesTerminateCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(outBuf)
	cmd.SetArgs([]string{"123", "--date", "2024-01-01"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}
