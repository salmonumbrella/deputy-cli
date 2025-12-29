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

func TestLeaveService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/resource/Leave", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Leave{
			{Id: 1, Employee: 10, Company: 100, DateStart: "2024-01-15", DateEnd: "2024-01-20", Status: 1, Hours: 40, Days: 5},
			{Id: 2, Employee: 11, Company: 100, DateStart: "2024-02-01", DateEnd: "2024-02-02", Status: 0, Hours: 16, Days: 2},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	leaves, err := client.Leave().List(context.Background(), nil)

	require.NoError(t, err)
	assert.Len(t, leaves, 2)
	assert.Equal(t, 1, leaves[0].Id)
	assert.Equal(t, 10, leaves[0].Employee)
	assert.Equal(t, "2024-01-15", leaves[0].DateStart)
	assert.Equal(t, "2024-01-20", leaves[0].DateEnd)
	assert.Equal(t, 1, leaves[0].Status) // approved
	assert.Equal(t, 40.0, leaves[0].Hours)
	assert.Equal(t, 5.0, leaves[0].Days)
}

func TestLeaveService_List_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Leave().List(context.Background(), nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestLeaveService_List_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Leave{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	leaves, err := client.Leave().List(context.Background(), nil)

	require.NoError(t, err)
	assert.Empty(t, leaves)
}

func TestLeaveService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/resource/Leave/42", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Leave{
			Id:        42,
			Employee:  10,
			Company:   100,
			DateStart: "2024-03-01",
			DateEnd:   "2024-03-05",
			Status:    1,
			Hours:     32,
			Days:      4,
			ApproveBy: 5,
			Comment:   "Annual leave",
			LeaveRule: 1,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	leave, err := client.Leave().Get(context.Background(), 42)

	require.NoError(t, err)
	assert.Equal(t, 42, leave.Id)
	assert.Equal(t, 10, leave.Employee)
	assert.Equal(t, "2024-03-01", leave.DateStart)
	assert.Equal(t, "2024-03-05", leave.DateEnd)
	assert.Equal(t, 1, leave.Status)
	assert.Equal(t, 5, leave.ApproveBy)
	assert.Equal(t, "Annual leave", leave.Comment)
	assert.Equal(t, 1, leave.LeaveRule)
}

func TestLeaveService_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Leave not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	_, err := client.Leave().Get(context.Background(), 99999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestLeaveService_Get_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Leave().Get(context.Background(), 42)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestLeaveService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/Leave", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var input CreateLeaveInput
		_ = json.Unmarshal(body, &input)
		assert.Equal(t, 10, input.Employee)
		assert.Equal(t, "2024-04-01", input.DateStart)
		assert.Equal(t, "2024-04-05", input.DateEnd)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Leave{
			Id:        99,
			Employee:  10,
			Company:   100,
			DateStart: "2024-04-01",
			DateEnd:   "2024-04-05",
			Status:    0, // awaiting approval
			Hours:     32,
			Days:      4,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &CreateLeaveInput{
		Employee:  10,
		DateStart: "2024-04-01",
		DateEnd:   "2024-04-05",
	}
	leave, err := client.Leave().Create(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 99, leave.Id)
	assert.Equal(t, 10, leave.Employee)
	assert.Equal(t, "2024-04-01", leave.DateStart)
	assert.Equal(t, 0, leave.Status)
}

func TestLeaveService_Create_WithOptionalFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var input CreateLeaveInput
		_ = json.Unmarshal(body, &input)
		assert.Equal(t, 2, input.LeaveRule)
		assert.Equal(t, "Family vacation", input.Comment)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Leave{
			Id:        99,
			Employee:  10,
			LeaveRule: 2,
			Comment:   "Family vacation",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &CreateLeaveInput{
		Employee:  10,
		DateStart: "2024-04-01",
		DateEnd:   "2024-04-05",
		LeaveRule: 2,
		Comment:   "Family vacation",
	}
	leave, err := client.Leave().Create(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 2, leave.LeaveRule)
	assert.Equal(t, "Family vacation", leave.Comment)
}

func TestLeaveService_Create_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	input := &CreateLeaveInput{
		Employee:  10,
		DateStart: "2024-04-01",
		DateEnd:   "2024-04-05",
	}
	_, err := client.Leave().Create(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestLeaveService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/Leave/42", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var input UpdateLeaveInput
		_ = json.Unmarshal(body, &input)
		assert.Equal(t, 1, input.Status) // approved
		assert.Equal(t, "Approved by manager", input.Comment)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Leave{
			Id:      42,
			Status:  1,
			Comment: "Approved by manager",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &UpdateLeaveInput{
		Status:  1,
		Comment: "Approved by manager",
	}
	leave, err := client.Leave().Update(context.Background(), 42, input)

	require.NoError(t, err)
	assert.Equal(t, 42, leave.Id)
	assert.Equal(t, 1, leave.Status)
	assert.Equal(t, "Approved by manager", leave.Comment)
}

func TestLeaveService_Update_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Leave not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &UpdateLeaveInput{Status: 1}
	_, err := client.Leave().Update(context.Background(), 99999, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestLeaveService_Update_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	input := &UpdateLeaveInput{Status: 1}
	_, err := client.Leave().Update(context.Background(), 42, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestLeaveService_Approve(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/Leave/42", r.URL.Path)

		// Verify status is set to 1 (approved)
		body, _ := io.ReadAll(r.Body)
		var input UpdateLeaveInput
		_ = json.Unmarshal(body, &input)
		assert.Equal(t, 1, input.Status)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Leave{
			Id:     42,
			Status: 1,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Leave().Approve(context.Background(), 42)

	require.NoError(t, err)
}

func TestLeaveService_Approve_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Leave not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Leave().Approve(context.Background(), 99999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestLeaveService_Approve_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	err := client.Leave().Approve(context.Background(), 42)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestLeaveService_Decline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/Leave/42", r.URL.Path)

		// Verify status is set to 2 (declined) and comment is included
		body, _ := io.ReadAll(r.Body)
		var input UpdateLeaveInput
		_ = json.Unmarshal(body, &input)
		assert.Equal(t, 2, input.Status)
		assert.Equal(t, "Insufficient staff coverage", input.Comment)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Leave{
			Id:      42,
			Status:  2,
			Comment: "Insufficient staff coverage",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Leave().Decline(context.Background(), 42, "Insufficient staff coverage")

	require.NoError(t, err)
}

func TestLeaveService_Decline_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Leave not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Leave().Decline(context.Background(), 99999, "reason")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestLeaveService_Decline_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	err := client.Leave().Decline(context.Background(), 42, "reason")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestLeaveService_Query(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/Leave/QUERY", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var input LeaveQueryInput
		_ = json.Unmarshal(body, &input)
		assert.NotNil(t, input.Search)
		// JSON unmarshals numbers as float64
		assert.Equal(t, float64(10), input.Search["Employee"])
		assert.Equal(t, 50, input.Max)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Leave{
			{Id: 1, Employee: 10, Status: 1},
			{Id: 2, Employee: 10, Status: 0},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &LeaveQueryInput{
		Search: map[string]interface{}{"Employee": 10},
		Max:    50,
	}
	leaves, err := client.Leave().Query(context.Background(), input)

	require.NoError(t, err)
	assert.Len(t, leaves, 2)
	assert.Equal(t, 10, leaves[0].Employee)
	assert.Equal(t, 10, leaves[1].Employee)
}

func TestLeaveService_Query_WithJoin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var input LeaveQueryInput
		_ = json.Unmarshal(body, &input)
		assert.Contains(t, input.Join, "Employee")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Leave{
			{Id: 1, Employee: 10},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &LeaveQueryInput{
		Search: map[string]interface{}{"Status": 0},
		Join:   []string{"Employee"},
	}
	leaves, err := client.Leave().Query(context.Background(), input)

	require.NoError(t, err)
	assert.Len(t, leaves, 1)
}

func TestLeaveService_Query_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var input LeaveQueryInput
		_ = json.Unmarshal(body, &input)
		assert.Equal(t, 25, input.Max)
		assert.Equal(t, 50, input.Start)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Leave{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &LeaveQueryInput{
		Max:   25,
		Start: 50,
	}
	leaves, err := client.Leave().Query(context.Background(), input)

	require.NoError(t, err)
	assert.Empty(t, leaves)
}

func TestLeaveService_Query_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Leave{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &LeaveQueryInput{
		Search: map[string]interface{}{"Employee": 99999},
	}
	leaves, err := client.Leave().Query(context.Background(), input)

	require.NoError(t, err)
	assert.Empty(t, leaves)
}

func TestLeaveService_Query_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	input := &LeaveQueryInput{
		Search: map[string]interface{}{"Employee": 10},
	}
	_, err := client.Leave().Query(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

// Forbidden (403) error tests

func TestLeaveService_List_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Access denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	_, err := client.Leave().List(context.Background(), nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestLeaveService_Get_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Access denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	_, err := client.Leave().Get(context.Background(), 42)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestLeaveService_Create_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &CreateLeaveInput{
		Employee:  10,
		DateStart: "2024-04-01",
		DateEnd:   "2024-04-05",
	}
	_, err := client.Leave().Create(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestLeaveService_Update_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &UpdateLeaveInput{Status: 1}
	_, err := client.Leave().Update(context.Background(), 42, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestLeaveService_Approve_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Leave().Approve(context.Background(), 42)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestLeaveService_Decline_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Leave().Decline(context.Background(), 42, "reason")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestLeaveService_Query_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Access denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &LeaveQueryInput{
		Search: map[string]interface{}{"Employee": 10},
	}
	_, err := client.Leave().Query(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}
