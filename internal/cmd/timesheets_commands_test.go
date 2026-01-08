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

func TestTimesheetsGetCommand_WithMockClient(t *testing.T) {
	t.Run("outputs text details", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/v1/supervise/timesheet/123", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Timesheet{
				Id:           123,
				Employee:     10,
				Date:         "2024-01-15",
				StartTime:    1705309200,
				EndTime:      1705330800,
				TotalTimeStr: "06:00",
				Mealbreak:    "00:30",
				IsInProgress: false,
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newTimesheetsGetCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})

		err := cmd.Execute()
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "ID:         123")
		assert.Contains(t, output, "Employee:   10")
		assert.Contains(t, output, "Date:       2024-01-15")
		assert.Contains(t, output, "Total:      06:00")
	})

	t.Run("outputs json when requested", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Timesheet{
				Id:           123,
				Employee:     10,
				Date:         "2024-01-15",
				TotalTimeStr: "06:00",
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = outfmt.WithFormat(ctx, "json")
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newTimesheetsGetCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Id": 123`)
	})
}

func TestTimesheetsUpdateCommand_WithMockClient(t *testing.T) {
	t.Run("updates cost and outputs text", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/v1/resource/Timesheet/123", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Timesheet{
				Id:   123,
				Cost: 150.50,
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newTimesheetsUpdateCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--cost", "150.5"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Updated timesheet 123")
	})

	t.Run("outputs json when requested", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Timesheet{Id: 123, Cost: 150.5})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = outfmt.WithFormat(ctx, "json")
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newTimesheetsUpdateCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--cost", "150.5"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Cost": 150.5`)
	})
}

func TestTimesheetsListPayRulesCommand_WithMockClient(t *testing.T) {
	t.Run("outputs table", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.PayRule{
				{Id: 7, PayTitle: "Standard", HourlyRate: 25.0},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newTimesheetsListPayRulesCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "HOURLY RATE")
		assert.Contains(t, buf.String(), "Standard")
	})

	t.Run("outputs json when requested", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.PayRule{
				{Id: 7, PayTitle: "Standard", HourlyRate: 25.0},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = outfmt.WithFormat(ctx, "json")
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newTimesheetsListPayRulesCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--hourly-rate", "25"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"HourlyRate": 25`)
	})
}

func TestTimesheetsSetPayRuleCommand_WithMockClient(t *testing.T) {
	t.Run("sets pay rule and outputs text", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v1/resource/TimesheetPayReturn/QUERY", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.TimesheetPayReturn{
				{Id: 99, Timesheet: 123, PayRule: 1, Cost: 0},
			})
		})
		mux.HandleFunc("/api/v1/supervise/timesheet/123", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Timesheet{
				Id:        123,
				TotalTime: 8.0,
			})
		})
		mux.HandleFunc("/api/v1/resource/PayRules/QUERY", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.PayRule{
				{Id: 7, HourlyRate: 25.0},
			})
		})
		mux.HandleFunc("/api/v1/resource/TimesheetPayReturn/99", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.TimesheetPayReturn{
				Id:        99,
				Timesheet: 123,
				PayRule:   7,
				Cost:      200.0,
			})
		})
		mux.HandleFunc("/api/v1/resource/Timesheet/123", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		server := httptest.NewServer(mux)
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newTimesheetsSetPayRuleCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123", "--pay-rule", "7"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Assigned pay rule 7 to timesheet 123")
	})

	t.Run("requires pay-rule flag", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newTimesheetsSetPayRuleCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"123"})

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--pay-rule is required")
	})

	t.Run("invalid timesheet id", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := iocontext.WithIO(context.Background(), &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newTimesheetsSetPayRuleCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"bad", "--pay-rule", "7"})

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid timesheet ID")
	})
}

func TestTimesheetsListCommand_WithEmployeeFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/Timesheet/QUERY", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]api.Timesheet{
			{Id: 1, Employee: 123, Date: "2024-01-02", TotalTimeStr: "08:00"},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newTimesheetsListCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--employee", "123", "--from", "2024-01-01", "--to", "2024-01-31"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "2024-01-02")
}

func TestTimesheetsListCommand_WithEmployeeFilter_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]api.Timesheet{
			{Id: 2, Employee: 456, Date: "2024-02-01"},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
	ctx = outfmt.WithFormat(ctx, "json")

	cmd := newTimesheetsListCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--employee", "456"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"Id": 2`)
}

func TestTimesheetsClockInOutCommands_WithMockClient(t *testing.T) {
	t.Run("clock-in outputs success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/v1/supervise/timesheet/start", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.ClockResponse{Id: 99, Employee: 123})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newTimesheetsClockInCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--employee", "123"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Clocked in employee 123")
	})

	t.Run("clock-out outputs success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/v1/supervise/timesheet/stop", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.ClockResponse{Id: 77, Employee: 123})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newTimesheetsClockOutCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--timesheet", "77"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Clocked out employee 123")
	})

	t.Run("clock-in outputs json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.ClockResponse{Id: 101, Employee: 321})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newTimesheetsClockInCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--employee", "321"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Id": 101`)
	})

	t.Run("clock-out outputs json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.ClockResponse{Id: 202, Employee: 321})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
		ctx = outfmt.WithFormat(ctx, "json")

		cmd := newTimesheetsClockOutCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--timesheet", "202"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"Id": 202`)
	})
}

func TestTimesheetsListCommand_InvalidRanges(t *testing.T) {
	// Use mock client factory to avoid keychain access in CI
	// The validation happens after getClientFromContext, so we need a valid client
	client := newTestClient("http://localhost", "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newTimesheetsListCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--from", "2024-02-01", "--to", "2024-01-01"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--from must be on or before --to")
}

func TestTimesheetsListCommand_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]api.Timesheet{
			{Id: 1, Employee: 123, Date: "2024-01-15"},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
	ctx = outfmt.WithFormat(ctx, "json")

	cmd := newTimesheetsListCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	err := cmd.Execute()

	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"Id": 1`)
}

func TestTimesheetsListCommand_DateFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]api.Timesheet{
			{Id: 1, Date: "2024-01-01", StartTime: 1, EndTime: 2, TotalTimeStr: "01:00"},
			{Id: 2, Date: "2024-02-01", StartTime: 1, EndTime: 2, TotalTimeStr: "01:00"},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newTimesheetsListCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--from", "2024-02-01", "--to", "2024-02-28"})

	err := cmd.Execute()
	require.NoError(t, err)
	output := buf.String()
	assert.NotContains(t, output, "2024-01-01")
	assert.Contains(t, output, "2024-02-01")
}

func TestTimesheetsSetPayRuleCommand_JSONOutput(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/resource/TimesheetPayReturn/QUERY", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]api.TimesheetPayReturn{
			{Id: 99, Timesheet: 123, PayRule: 1, Cost: 0},
		})
	})
	mux.HandleFunc("/api/v1/supervise/timesheet/123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.Timesheet{
			Id:        123,
			TotalTime: 8.0,
		})
	})
	mux.HandleFunc("/api/v1/resource/PayRules/QUERY", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]api.PayRule{
			{Id: 7, HourlyRate: 25.0},
		})
	})
	mux.HandleFunc("/api/v1/resource/TimesheetPayReturn/99", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.TimesheetPayReturn{
			Id:        99,
			Timesheet: 123,
			PayRule:   7,
			Cost:      200.0,
		})
	})
	mux.HandleFunc("/api/v1/resource/Timesheet/123", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
	ctx = outfmt.WithFormat(ctx, "json")

	cmd := newTimesheetsSetPayRuleCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"123", "--pay-rule", "7"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"PayRule": 7`)
}
