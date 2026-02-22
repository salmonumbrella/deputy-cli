package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeService_Info(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/me", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		// Simulate Deputy API response with PascalCase field names
		_, _ = w.Write([]byte(`{
			"UserId": 123,
			"EmployeeId": 456,
			"Login": "jdoe",
			"Name": "John Doe",
			"FirstName": "John",
			"LastName": "Doe",
			"PrimaryEmail": "john@example.com",
			"PrimaryPhone": "+1234567890",
			"Photo": "https://example.com/photo.jpg",
			"Company": 1,
			"Portfolio": "Test Company",
			"Role": 5
		}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Me().Info(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 123, result.UserId)
	assert.Equal(t, 456, result.EmployeeId)
	assert.Equal(t, "jdoe", result.Login)
	assert.Equal(t, "John Doe", result.Name)
	assert.Equal(t, "John", result.FirstName)
	assert.Equal(t, "Doe", result.LastName)
	assert.Equal(t, "john@example.com", result.PrimaryEmail)
	assert.Equal(t, "+1234567890", result.PrimaryPhone)
	assert.Equal(t, "https://example.com/photo.jpg", result.Photo)
	assert.Equal(t, 1, result.Company)
	assert.Equal(t, "Test Company", result.Portfolio)
	assert.Equal(t, 5, result.Role)
}

func TestMeInfo_JSONMarshal_SnakeCase(t *testing.T) {
	// Test that MeInfo marshals to snake_case JSON with id field
	info := MeInfo{
		UserId:       123,
		EmployeeId:   456,
		Login:        "jdoe",
		Name:         "John Doe",
		FirstName:    "John",
		LastName:     "Doe",
		PrimaryEmail: "john@example.com",
		PrimaryPhone: "+1234567890",
		Photo:        "https://example.com/photo.jpg",
		Company:      1,
		Portfolio:    "Test Company",
		Role:         5,
	}

	data, err := json.Marshal(info)
	require.NoError(t, err)

	// Verify snake_case field names
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &result))

	// Check id field exists and equals user_id
	assert.Equal(t, float64(123), result["id"])
	assert.Equal(t, float64(123), result["user_id"])
	assert.Equal(t, float64(456), result["employee_id"])
	assert.Equal(t, "jdoe", result["login"])
	assert.Equal(t, "John Doe", result["name"])
	assert.Equal(t, "John", result["first_name"])
	assert.Equal(t, "Doe", result["last_name"])
	assert.Equal(t, "john@example.com", result["primary_email"])
	assert.Equal(t, "+1234567890", result["primary_phone"])
	assert.Equal(t, "https://example.com/photo.jpg", result["photo"])
	assert.Equal(t, float64(1), result["company"])
	assert.Equal(t, "Test Company", result["portfolio"])
	assert.Equal(t, float64(5), result["role"])

	// Verify PascalCase fields do NOT exist
	assert.Nil(t, result["UserId"])
	assert.Nil(t, result["EmployeeId"])
	assert.Nil(t, result["PrimaryEmail"])
}

func TestMeService_Info_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Me().Info(context.Background())
	assert.Error(t, err)
	// In non-debug mode, error should be sanitized (no response body content)
	assert.Equal(t, "API error 401: unauthorized", err.Error())
}

func TestMeService_Info_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Access denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	_, err := client.Me().Info(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestMeService_Timesheets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/my/timesheets", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Timesheet{
			{
				Id:              1,
				Employee:        123,
				Date:            "2024-01-15",
				StartTime:       1705312800,
				EndTime:         1705341600,
				TotalTime:       8.0,
				TotalTimeStr:    "8:00",
				OperationalUnit: 10,
				IsInProgress:    false,
				IsLeave:         false,
			},
			{
				Id:              2,
				Employee:        123,
				Date:            "2024-01-16",
				StartTime:       1705399200,
				EndTime:         0,
				TotalTime:       0,
				OperationalUnit: 10,
				IsInProgress:    true,
				IsLeave:         false,
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Me().Timesheets(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].Id)
	assert.Equal(t, "2024-01-15", result[0].Date)
	assert.False(t, result[0].IsInProgress)
	assert.Equal(t, 2, result[1].Id)
	assert.True(t, result[1].IsInProgress)
}

func TestMeService_Timesheets_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Me().Timesheets(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestMeService_Timesheets_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Access denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	_, err := client.Me().Timesheets(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestMeService_Timesheets_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Timesheet{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Me().Timesheets(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestMeService_Rosters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/my/rosters", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Roster{
			{
				Id:              100,
				Date:            "2024-01-20",
				StartTime:       1705744800,
				EndTime:         1705773600,
				Mealbreak:       "30",
				Employee:        123,
				OperationalUnit: 10,
				Open:            false,
				Published:       true,
				Comment:         "Morning shift",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Me().Rosters(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 100, result[0].Id)
	assert.Equal(t, "2024-01-20", result[0].Date)
	assert.Equal(t, 123, result[0].Employee)
	assert.True(t, result[0].Published)
	assert.Equal(t, "Morning shift", result[0].Comment)
}

func TestMeService_Rosters_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Me().Rosters(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestMeService_Rosters_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Access denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	_, err := client.Me().Rosters(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestMeService_Rosters_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Roster{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Me().Rosters(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestMeService_Leave(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/my/leave", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Leave{
			{
				Id:        200,
				Employee:  123,
				Company:   1,
				DateStart: "2024-02-01",
				DateEnd:   "2024-02-05",
				Status:    1, // approved
				Hours:     32,
				Days:      4,
				Comment:   "Annual leave",
				LeaveRule: 5,
			},
			{
				Id:        201,
				Employee:  123,
				Company:   1,
				DateStart: "2024-03-10",
				DateEnd:   "2024-03-10",
				Status:    0, // awaiting
				Hours:     8,
				Days:      1,
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Me().Leave(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 200, result[0].Id)
	assert.Equal(t, "2024-02-01", result[0].DateStart)
	assert.Equal(t, 1, result[0].Status)
	assert.Equal(t, "Annual leave", result[0].Comment)
	assert.Equal(t, 201, result[1].Id)
	assert.Equal(t, 0, result[1].Status)
}

func TestMeService_Leave_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Me().Leave(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestMeService_Leave_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Access denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	_, err := client.Me().Leave(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestMeService_Leave_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Leave{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Me().Leave(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result)
}
