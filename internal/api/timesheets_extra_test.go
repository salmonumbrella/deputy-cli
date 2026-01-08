package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimesheetsService_Query(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/Timesheet/QUERY", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"Id":           1,
				"Employee":     123,
				"Date":         "2024-01-01",
				"TotalTime":    8.0,
				"TotalTimeStr": "08:00",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &QueryInput{
		Search: map[string]interface{}{
			"s1": map[string]interface{}{"field": "Employee", "type": "eq", "data": 123},
		},
	}

	results, err := client.Timesheets().Query(context.Background(), input)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 1, results[0].Id)
	assert.Equal(t, 123, results[0].Employee)
	assert.Equal(t, "2024-01-01", results[0].Date)
}

func TestTimesheetsService_Query_InvalidTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"Id":"not-a-number"}]`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &QueryInput{}

	_, err := client.Timesheets().Query(context.Background(), input)
	require.Error(t, err)
}

func TestTimesheetsService_Query_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	input := &QueryInput{}

	_, err := client.Timesheets().Query(context.Background(), input)
	require.Error(t, err)
}

func TestTimesheetsService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/Timesheet/123", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"Cost":123.45`)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Timesheet{
			Id:   123,
			Cost: 123.45,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	cost := 123.45
	updated, err := client.Timesheets().Update(context.Background(), 123, &UpdateTimesheetInput{Cost: &cost})
	require.NoError(t, err)
	assert.Equal(t, 123, updated.Id)
	assert.Equal(t, 123.45, updated.Cost)
}

func TestTimesheetsService_ListPayRules(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/PayRules/QUERY", r.URL.Path)

		var payload map[string]any
		err := json.NewDecoder(r.Body).Decode(&payload)
		require.NoError(t, err)
		search := payload["search"].(map[string]any)
		cond := search["s1"].(map[string]any)
		assert.Equal(t, "HourlyRate", cond["field"])
		assert.Equal(t, "eq", cond["type"])
		assert.Equal(t, 190.0, cond["data"])

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]PayRule{
			{Id: 7, PayTitle: "Standard", HourlyRate: 190},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	rate := 190.0
	rules, err := client.Timesheets().ListPayRules(context.Background(), &rate)
	require.NoError(t, err)
	require.Len(t, rules, 1)
	assert.Equal(t, 7, rules[0].Id)
	assert.Equal(t, 190.0, rules[0].HourlyRate)
}

func TestTimesheetsService_GetPayReturn(t *testing.T) {
	t.Run("returns first pay return", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/v1/resource/TimesheetPayReturn/QUERY", r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]TimesheetPayReturn{
				{Id: 55, Timesheet: 123, PayRule: 7, Cost: 200},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		result, err := client.Timesheets().GetPayReturn(context.Background(), 123)
		require.NoError(t, err)
		assert.Equal(t, 55, result.Id)
	})

	t.Run("returns error when no results", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]TimesheetPayReturn{})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		_, err := client.Timesheets().GetPayReturn(context.Background(), 999)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no pay return found")
	})

	t.Run("returns error on api failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		_, err := client.Timesheets().GetPayReturn(context.Background(), 123)
		require.Error(t, err)
	})
}

func TestTimesheetsService_SetPayRule(t *testing.T) {
	t.Run("sets pay rule and updates cost", func(t *testing.T) {
		var payReturnCost float64
		var timesheetCost float64

		mux := http.NewServeMux()
		mux.HandleFunc("/api/v1/resource/TimesheetPayReturn/QUERY", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]TimesheetPayReturn{
				{Id: 99, Timesheet: 123, PayRule: 1, Cost: 0},
			})
		})
		mux.HandleFunc("/api/v1/supervise/timesheet/123", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Timesheet{
				Id:        123,
				TotalTime: 8.0,
			})
		})
		mux.HandleFunc("/api/v1/resource/PayRules/QUERY", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]PayRule{
				{Id: 7, HourlyRate: 25.0},
			})
		})
		mux.HandleFunc("/api/v1/resource/TimesheetPayReturn/99", func(w http.ResponseWriter, r *http.Request) {
			var input SetPayRuleInput
			err := json.NewDecoder(r.Body).Decode(&input)
			require.NoError(t, err)
			payReturnCost = input.Cost
			assert.Equal(t, 7, input.PayRule)
			assert.True(t, input.Overridden)

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(TimesheetPayReturn{
				Id:        99,
				Timesheet: 123,
				PayRule:   7,
				Cost:      input.Cost,
			})
		})
		mux.HandleFunc("/api/v1/resource/Timesheet/123", func(w http.ResponseWriter, r *http.Request) {
			var input UpdateTimesheetInput
			err := json.NewDecoder(r.Body).Decode(&input)
			require.NoError(t, err)
			if input.Cost != nil {
				timesheetCost = *input.Cost
			}
			w.WriteHeader(http.StatusOK)
		})

		server := httptest.NewServer(mux)
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		result, err := client.Timesheets().SetPayRule(context.Background(), 123, 7)
		require.NoError(t, err)
		assert.Equal(t, 7, result.PayRule)
		assert.Equal(t, 200.0, result.Cost)
		assert.Equal(t, 200.0, payReturnCost)
		assert.Equal(t, 200.0, timesheetCost)
	})

	t.Run("returns error when pay rule not found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v1/resource/TimesheetPayReturn/QUERY", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]TimesheetPayReturn{
				{Id: 99, Timesheet: 123, PayRule: 1, Cost: 0},
			})
		})
		mux.HandleFunc("/api/v1/supervise/timesheet/123", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Timesheet{
				Id:        123,
				TotalTime: 8.0,
			})
		})
		mux.HandleFunc("/api/v1/resource/PayRules/QUERY", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]PayRule{})
		})

		server := httptest.NewServer(mux)
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		_, err := client.Timesheets().SetPayRule(context.Background(), 123, 7)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pay rule 7 not found")
	})

	t.Run("returns error when timesheet has no hours", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v1/resource/TimesheetPayReturn/QUERY", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]TimesheetPayReturn{
				{Id: 99, Timesheet: 123, PayRule: 1, Cost: 0},
			})
		})
		mux.HandleFunc("/api/v1/supervise/timesheet/123", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Timesheet{
				Id:        123,
				TotalTime: 0,
			})
		})
		mux.HandleFunc("/api/v1/resource/PayRules/QUERY", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]PayRule{
				{Id: 7, HourlyRate: 25.0},
			})
		})

		server := httptest.NewServer(mux)
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		_, err := client.Timesheets().SetPayRule(context.Background(), 123, 7)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has no hours recorded")
	})
}
