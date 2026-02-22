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

func TestRostersService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodGet, r.Method)

		// Verify path (includes /api/v1 prefix from BaseURL)
		assert.Equal(t, "/api/v1/supervise/roster", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Roster{
			{
				Id:              1,
				Date:            "2024-01-15",
				StartTime:       1705312800,
				EndTime:         1705341600,
				Mealbreak:       "00:30",
				Employee:        100,
				OperationalUnit: 10,
				Open:            false,
				Published:       true,
				Comment:         "Morning shift",
			},
			{
				Id:              2,
				Date:            "2024-01-15",
				StartTime:       1705341600,
				EndTime:         1705370400,
				Mealbreak:       "01:00",
				Employee:        101,
				OperationalUnit: 10,
				Open:            false,
				Published:       true,
				Comment:         "Evening shift",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	rosters, err := client.Rosters().List(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, rosters, 2)

	assert.Equal(t, 1, rosters[0].Id)
	assert.Equal(t, "2024-01-15", rosters[0].Date)
	assert.Equal(t, int64(1705312800), rosters[0].StartTime)
	assert.Equal(t, int64(1705341600), rosters[0].EndTime)
	assert.Equal(t, "00:30", rosters[0].Mealbreak)
	assert.Equal(t, 100, rosters[0].Employee)
	assert.Equal(t, 10, rosters[0].OperationalUnit)
	assert.False(t, rosters[0].Open)
	assert.True(t, rosters[0].Published)
	assert.Equal(t, "Morning shift", rosters[0].Comment)

	assert.Equal(t, 2, rosters[1].Id)
	assert.Equal(t, 101, rosters[1].Employee)
	assert.Equal(t, "Evening shift", rosters[1].Comment)
}

func TestRostersService_List_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/supervise/roster", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Roster{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	rosters, err := client.Rosters().List(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, rosters)
}

func TestRostersService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Rosters().List(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestRostersService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodGet, r.Method)

		// Verify path includes roster ID (with /api/v1 prefix)
		// Note: Get uses /resource/Roster/{id} path
		assert.Equal(t, "/api/v1/resource/Roster/42", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Roster{
			Id:              42,
			Date:            "2024-01-20",
			StartTime:       1705744800,
			EndTime:         1705773600,
			Mealbreak:       "00:45",
			Employee:        200,
			OperationalUnit: 20,
			Open:            true,
			Published:       false,
			Comment:         "Open shift",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	roster, err := client.Rosters().Get(context.Background(), 42)
	require.NoError(t, err)

	assert.Equal(t, 42, roster.Id)
	assert.Equal(t, "2024-01-20", roster.Date)
	assert.Equal(t, int64(1705744800), roster.StartTime)
	assert.Equal(t, int64(1705773600), roster.EndTime)
	assert.Equal(t, "00:45", roster.Mealbreak)
	assert.Equal(t, 200, roster.Employee)
	assert.Equal(t, 20, roster.OperationalUnit)
	assert.True(t, roster.Open)
	assert.False(t, roster.Published)
	assert.Equal(t, "Open shift", roster.Comment)
}

func TestRostersService_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/resource/Roster/999", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Roster not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Rosters().Get(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestRostersService_GetSwappable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/supervise/roster/123/swap", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]SwapRoster{
			{Id: 1, Employee: 10, OperationalUnit: 5},
			{Id: 2, Employee: 11, OperationalUnit: 5},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	rosters, err := client.Rosters().GetSwappable(context.Background(), 123)
	require.NoError(t, err)
	require.Len(t, rosters, 2)
	assert.Equal(t, 1, rosters[0].Id)
	assert.Equal(t, 2, rosters[1].Id)
}

func TestRostersService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path (includes /api/v1 prefix from BaseURL)
		assert.Equal(t, "/api/v1/supervise/roster", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CreateRosterInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, 100, input.Employee)
		assert.Equal(t, 10, input.OperationalUnit)
		assert.Equal(t, int64(1705312800), input.StartTime)
		assert.Equal(t, int64(1705341600), input.EndTime)
		assert.Equal(t, "00:30", input.Mealbreak)
		assert.Equal(t, "Test shift", input.Comment)
		assert.False(t, input.Open)
		assert.True(t, input.Publish)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Roster{
			Id:              99,
			Date:            "2024-01-15",
			StartTime:       1705312800,
			EndTime:         1705341600,
			Mealbreak:       "00:30",
			Employee:        100,
			OperationalUnit: 10,
			Open:            false,
			Published:       true,
			Comment:         "Test shift",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateRosterInput{
		Employee:        100,
		OperationalUnit: 10,
		StartTime:       1705312800,
		EndTime:         1705341600,
		Mealbreak:       "00:30",
		Comment:         "Test shift",
		Open:            false,
		Publish:         true,
	}

	roster, err := client.Rosters().Create(context.Background(), input)
	require.NoError(t, err)

	assert.Equal(t, 99, roster.Id)
	assert.Equal(t, "2024-01-15", roster.Date)
	assert.Equal(t, int64(1705312800), roster.StartTime)
	assert.Equal(t, int64(1705341600), roster.EndTime)
	assert.Equal(t, 100, roster.Employee)
	assert.Equal(t, 10, roster.OperationalUnit)
	assert.True(t, roster.Published)
}

func TestRostersService_Create_OpenShift(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/supervise/roster", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CreateRosterInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		// Verify Open shift
		assert.True(t, input.Open)
		assert.Equal(t, 0, input.Employee) // No employee for open shifts

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Roster{
			Id:              100,
			Date:            "2024-01-15",
			StartTime:       1705312800,
			EndTime:         1705341600,
			OperationalUnit: 10,
			Open:            true,
			Published:       false,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateRosterInput{
		OperationalUnit: 10,
		StartTime:       1705312800,
		EndTime:         1705341600,
		Open:            true,
	}

	roster, err := client.Rosters().Create(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 100, roster.Id)
	assert.True(t, roster.Open)
}

func TestRostersService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Invalid input"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateRosterInput{
		OperationalUnit: 10,
		StartTime:       1705312800,
		EndTime:         1705341600,
	}

	_, err := client.Rosters().Create(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}

func TestRostersService_Copy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path (includes /api/v1 prefix from BaseURL)
		assert.Equal(t, "/api/v1/supervise/roster/copy", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CopyRosterInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, "2024-01-15", input.FromDate)
		assert.Equal(t, "2024-01-21", input.ToDate)
		assert.Equal(t, 5, input.Location)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CopyRosterInput{
		FromDate: "2024-01-15",
		ToDate:   "2024-01-21",
		Location: 5,
	}

	err := client.Rosters().Copy(context.Background(), input)
	require.NoError(t, err)
}

func TestRostersService_Copy_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Invalid date range"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CopyRosterInput{
		FromDate: "invalid-date",
		ToDate:   "2024-01-21",
		Location: 5,
	}

	err := client.Rosters().Copy(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}

func TestRostersService_Copy_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CopyRosterInput{
		FromDate: "2024-01-15",
		ToDate:   "2024-01-21",
		Location: 5,
	}

	err := client.Rosters().Copy(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestRostersService_Publish(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path (includes /api/v1 prefix from BaseURL)
		assert.Equal(t, "/api/v1/supervise/roster/publish", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input PublishRosterInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, "2024-01-15", input.FromDate)
		assert.Equal(t, "2024-01-21", input.ToDate)
		assert.Equal(t, 10, input.Location)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &PublishRosterInput{
		FromDate: "2024-01-15",
		ToDate:   "2024-01-21",
		Location: 10,
	}

	err := client.Rosters().Publish(context.Background(), input)
	require.NoError(t, err)
}

func TestRostersService_Publish_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Invalid date range"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &PublishRosterInput{
		FromDate: "invalid",
		ToDate:   "2024-01-21",
		Location: 10,
	}

	err := client.Rosters().Publish(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}

func TestRostersService_Publish_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Location not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &PublishRosterInput{
		FromDate: "2024-01-15",
		ToDate:   "2024-01-21",
		Location: 999,
	}

	err := client.Rosters().Publish(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestRostersService_Discard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path (includes /api/v1 prefix from BaseURL)
		assert.Equal(t, "/api/v1/supervise/roster/discard", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input PublishRosterInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, "2024-01-22", input.FromDate)
		assert.Equal(t, "2024-01-28", input.ToDate)
		assert.Equal(t, 15, input.Location)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &PublishRosterInput{
		FromDate: "2024-01-22",
		ToDate:   "2024-01-28",
		Location: 15,
	}

	err := client.Rosters().Discard(context.Background(), input)
	require.NoError(t, err)
}

func TestRostersService_Discard_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Invalid date range"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &PublishRosterInput{
		FromDate: "bad-date",
		ToDate:   "2024-01-28",
		Location: 15,
	}

	err := client.Rosters().Discard(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}

func TestRostersService_Discard_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &PublishRosterInput{
		FromDate: "2024-01-22",
		ToDate:   "2024-01-28",
		Location: 15,
	}

	err := client.Rosters().Discard(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestRostersService_Discard_NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/supervise/roster/discard", r.URL.Path)

		// Server returns 204 No Content for successful discard with no changes
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &PublishRosterInput{
		FromDate: "2024-01-22",
		ToDate:   "2024-01-28",
		Location: 15,
	}

	err := client.Rosters().Discard(context.Background(), input)
	require.NoError(t, err)
}
