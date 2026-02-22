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

func TestTimesheetsService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodGet, r.Method)

		// Verify path (includes /api/v1 prefix from BaseURL)
		assert.Equal(t, "/api/v1/my/timesheets", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Timesheet{
			{
				Id:              1,
				Employee:        100,
				Date:            "2024-01-15",
				StartTime:       1705300800,
				EndTime:         1705329600,
				Mealbreak:       "00:30",
				TotalTime:       7.5,
				TotalTimeStr:    "7h 30m",
				OperationalUnit: 10,
				IsInProgress:    false,
				IsLeave:         false,
				Comment:         "Regular shift",
			},
			{
				Id:              2,
				Employee:        100,
				Date:            "2024-01-16",
				StartTime:       1705387200,
				EndTime:         0,
				Mealbreak:       "",
				TotalTime:       0,
				TotalTimeStr:    "",
				OperationalUnit: 10,
				IsInProgress:    true,
				IsLeave:         false,
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	timesheets, err := client.Timesheets().List(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, timesheets, 2)

	assert.Equal(t, 1, timesheets[0].Id)
	assert.Equal(t, 100, timesheets[0].Employee)
	assert.Equal(t, "2024-01-15", timesheets[0].Date)
	assert.Equal(t, int64(1705300800), timesheets[0].StartTime)
	assert.Equal(t, int64(1705329600), timesheets[0].EndTime)
	assert.Equal(t, "00:30", timesheets[0].Mealbreak)
	assert.Equal(t, 7.5, timesheets[0].TotalTime)
	assert.Equal(t, "7h 30m", timesheets[0].TotalTimeStr)
	assert.Equal(t, 10, timesheets[0].OperationalUnit)
	assert.False(t, timesheets[0].IsInProgress)
	assert.False(t, timesheets[0].IsLeave)
	assert.Equal(t, "Regular shift", timesheets[0].Comment)

	assert.Equal(t, 2, timesheets[1].Id)
	assert.True(t, timesheets[1].IsInProgress)
}

func TestTimesheetsService_List_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/my/timesheets", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Timesheet{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	timesheets, err := client.Timesheets().List(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, timesheets)
}

func TestTimesheetsService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Timesheets().List(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestTimesheetsService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodGet, r.Method)

		// Verify path includes timesheet ID (with /api/v1 prefix)
		assert.Equal(t, "/api/v1/supervise/timesheet/42", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Timesheet{
			Id:              42,
			Employee:        100,
			Date:            "2024-01-15",
			StartTime:       1705300800,
			EndTime:         1705329600,
			Mealbreak:       "01:00",
			TotalTime:       7.0,
			TotalTimeStr:    "7h 00m",
			OperationalUnit: 20,
			IsInProgress:    false,
			IsLeave:         false,
			Comment:         "Morning shift",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	timesheet, err := client.Timesheets().Get(context.Background(), 42)
	require.NoError(t, err)

	assert.Equal(t, 42, timesheet.Id)
	assert.Equal(t, 100, timesheet.Employee)
	assert.Equal(t, "2024-01-15", timesheet.Date)
	assert.Equal(t, int64(1705300800), timesheet.StartTime)
	assert.Equal(t, int64(1705329600), timesheet.EndTime)
	assert.Equal(t, "01:00", timesheet.Mealbreak)
	assert.Equal(t, 7.0, timesheet.TotalTime)
	assert.Equal(t, "7h 00m", timesheet.TotalTimeStr)
	assert.Equal(t, 20, timesheet.OperationalUnit)
	assert.False(t, timesheet.IsInProgress)
	assert.Equal(t, "Morning shift", timesheet.Comment)
}

func TestTimesheetsService_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/supervise/timesheet/999", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Timesheet not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Timesheets().Get(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestTimesheetsService_ClockIn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path (with /api/v1 prefix)
		assert.Equal(t, "/api/v1/supervise/timesheet/start", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input ClockInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, 100, input.Employee)
		assert.Equal(t, 20, input.OperationalUnit)
		assert.Equal(t, "Starting work", input.Comment)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ClockResponse{
			Id:       123,
			Employee: 100,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &ClockInput{
		Employee:        100,
		OperationalUnit: 20,
		Comment:         "Starting work",
	}

	resp, err := client.Timesheets().ClockIn(context.Background(), input)
	require.NoError(t, err)

	assert.Equal(t, 123, resp.Id)
	assert.Equal(t, 100, resp.Employee)
}

func TestTimesheetsService_ClockIn_MinimalInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/supervise/timesheet/start", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input ClockInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		// Verify only required field
		assert.Equal(t, 100, input.Employee)
		// Optional fields should be empty/zero
		assert.Equal(t, 0, input.OperationalUnit)
		assert.Empty(t, input.Comment)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ClockResponse{
			Id:       124,
			Employee: 100,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &ClockInput{
		Employee: 100,
	}

	resp, err := client.Timesheets().ClockIn(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 124, resp.Id)
}

func TestTimesheetsService_ClockIn_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Employee already clocked in"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &ClockInput{
		Employee: 100,
	}

	_, err := client.Timesheets().ClockIn(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}

func TestTimesheetsService_ClockOut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path (with /api/v1 prefix)
		assert.Equal(t, "/api/v1/supervise/timesheet/stop", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input ClockInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, 100, input.Employee)
		assert.Equal(t, "Ending work", input.Comment)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ClockResponse{
			Id:       123,
			Employee: 100,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &ClockInput{
		Employee: 100,
		Comment:  "Ending work",
	}

	resp, err := client.Timesheets().ClockOut(context.Background(), input)
	require.NoError(t, err)

	assert.Equal(t, 123, resp.Id)
	assert.Equal(t, 100, resp.Employee)
}

func TestTimesheetsService_ClockOut_NotClockedIn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Employee not clocked in"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &ClockInput{
		Employee: 100,
	}

	_, err := client.Timesheets().ClockOut(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}

func TestTimesheetsService_ClockOut_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &ClockInput{
		Employee: 100,
	}

	_, err := client.Timesheets().ClockOut(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestTimesheetsService_StartBreak(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path (with /api/v1 prefix)
		assert.Equal(t, "/api/v1/supervise/timesheet/pause", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input ClockInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, 100, input.Employee)
		assert.Equal(t, "Lunch break", input.Comment)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &ClockInput{
		Employee: 100,
		Comment:  "Lunch break",
	}

	err := client.Timesheets().StartBreak(context.Background(), input)
	require.NoError(t, err)
}

func TestTimesheetsService_StartBreak_MinimalInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/supervise/timesheet/pause", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input ClockInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		// Verify only required field
		assert.Equal(t, 100, input.Employee)
		assert.Empty(t, input.Comment)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &ClockInput{
		Employee: 100,
	}

	err := client.Timesheets().StartBreak(context.Background(), input)
	require.NoError(t, err)
}

func TestTimesheetsService_StartBreak_NotClockedIn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Employee not clocked in"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &ClockInput{
		Employee: 100,
	}

	err := client.Timesheets().StartBreak(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}

func TestTimesheetsService_StartBreak_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &ClockInput{
		Employee: 100,
	}

	err := client.Timesheets().StartBreak(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestTimesheetsService_EndBreak(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path (with /api/v1 prefix)
		assert.Equal(t, "/api/v1/supervise/timesheet/resume", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input ClockInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, 100, input.Employee)
		assert.Equal(t, "Back from break", input.Comment)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &ClockInput{
		Employee: 100,
		Comment:  "Back from break",
	}

	err := client.Timesheets().EndBreak(context.Background(), input)
	require.NoError(t, err)
}

func TestTimesheetsService_EndBreak_MinimalInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/supervise/timesheet/resume", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input ClockInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		// Verify only required field
		assert.Equal(t, 100, input.Employee)
		assert.Empty(t, input.Comment)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &ClockInput{
		Employee: 100,
	}

	err := client.Timesheets().EndBreak(context.Background(), input)
	require.NoError(t, err)
}

func TestTimesheetsService_EndBreak_NotOnBreak(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Employee not on break"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &ClockInput{
		Employee: 100,
	}

	err := client.Timesheets().EndBreak(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}

func TestTimesheetsService_EndBreak_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &ClockInput{
		Employee: 100,
	}

	err := client.Timesheets().EndBreak(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}
