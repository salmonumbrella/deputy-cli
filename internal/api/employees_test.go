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

func TestEmployeesService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodGet, r.Method)

		// Verify path (includes /api/v1 prefix from BaseURL)
		assert.Equal(t, "/api/v1/supervise/employee", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Employee{
			{
				Id:          1,
				FirstName:   "John",
				LastName:    "Doe",
				DisplayName: "John Doe",
				Email:       "john@example.com",
				Mobile:      "+1234567890",
				Active:      true,
				Company:     100,
				Role:        1,
			},
			{
				Id:          2,
				FirstName:   "Jane",
				LastName:    "Smith",
				DisplayName: "Jane Smith",
				Email:       "jane@example.com",
				Mobile:      "+0987654321",
				Active:      true,
				Company:     100,
				Role:        2,
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	employees, err := client.Employees().List(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, employees, 2)

	assert.Equal(t, 1, employees[0].Id)
	assert.Equal(t, "John", employees[0].FirstName)
	assert.Equal(t, "Doe", employees[0].LastName)
	assert.Equal(t, "john@example.com", employees[0].Email)
	assert.True(t, employees[0].Active)

	assert.Equal(t, 2, employees[1].Id)
	assert.Equal(t, "Jane", employees[1].FirstName)
	assert.Equal(t, "Smith", employees[1].LastName)
}

func TestEmployeesService_List_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/supervise/employee", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Employee{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	employees, err := client.Employees().List(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, employees)
}

func TestEmployeesService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Employees().List(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestEmployeesService_List_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// When pagination is specified, use QUERY endpoint with POST
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/Employee/QUERY", r.URL.Path)

		// Verify pagination in request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input map[string]interface{}
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)
		assert.Equal(t, float64(10), input["max"])
		assert.Equal(t, float64(5), input["start"])

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Employee{
			{Id: 6, FirstName: "User", LastName: "Six", Active: true},
			{Id: 7, FirstName: "User", LastName: "Seven", Active: true},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	opts := &ListOptions{Limit: 10, Offset: 5}
	employees, err := client.Employees().List(context.Background(), opts)
	require.NoError(t, err)
	require.Len(t, employees, 2)
	assert.Equal(t, 6, employees[0].Id)
	assert.Equal(t, 7, employees[1].Id)
}

func TestEmployeesService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodGet, r.Method)

		// Verify path includes employee ID (with /api/v1 prefix)
		assert.Equal(t, "/api/v1/supervise/employee/42", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Employee{
			Id:          42,
			FirstName:   "Alice",
			LastName:    "Johnson",
			DisplayName: "Alice Johnson",
			Email:       "alice@example.com",
			Mobile:      "+1111111111",
			Active:      true,
			Company:     200,
			Role:        3,
			StartDate:   "2023-01-15",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	employee, err := client.Employees().Get(context.Background(), 42)
	require.NoError(t, err)

	assert.Equal(t, 42, employee.Id)
	assert.Equal(t, "Alice", employee.FirstName)
	assert.Equal(t, "Johnson", employee.LastName)
	assert.Equal(t, "alice@example.com", employee.Email)
	assert.Equal(t, "2023-01-15", employee.StartDate)
	assert.True(t, employee.Active)
}

func TestEmployeesService_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/supervise/employee/999", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Employee not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Employees().Get(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestEmployeesService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path (includes /api/v1 prefix from BaseURL)
		assert.Equal(t, "/api/v1/supervise/employee", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CreateEmployeeInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, "Bob", input.FirstName)
		assert.Equal(t, "Wilson", input.LastName)
		assert.Equal(t, "bob@example.com", input.Email)
		assert.Equal(t, "+2222222222", input.Mobile)
		assert.Equal(t, 100, input.Company)
		assert.Equal(t, 1, input.Role)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Employee{
			Id:          99,
			FirstName:   "Bob",
			LastName:    "Wilson",
			DisplayName: "Bob Wilson",
			Email:       "bob@example.com",
			Mobile:      "+2222222222",
			Active:      true,
			Company:     100,
			Role:        1,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateEmployeeInput{
		FirstName: "Bob",
		LastName:  "Wilson",
		Email:     "bob@example.com",
		Mobile:    "+2222222222",
		Company:   100,
		Role:      1,
	}

	employee, err := client.Employees().Create(context.Background(), input)
	require.NoError(t, err)

	assert.Equal(t, 99, employee.Id)
	assert.Equal(t, "Bob", employee.FirstName)
	assert.Equal(t, "Wilson", employee.LastName)
	assert.Equal(t, "bob@example.com", employee.Email)
	assert.True(t, employee.Active)
}

func TestEmployeesService_Create_MinimalInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/supervise/employee", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CreateEmployeeInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		// Verify required fields only
		assert.Equal(t, "First", input.FirstName)
		assert.Equal(t, "Last", input.LastName)
		assert.Equal(t, 50, input.Company)
		// Optional fields should be empty/zero
		assert.Empty(t, input.Email)
		assert.Empty(t, input.Mobile)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Employee{
			Id:        101,
			FirstName: "First",
			LastName:  "Last",
			Company:   50,
			Active:    true,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateEmployeeInput{
		FirstName: "First",
		LastName:  "Last",
		Company:   50,
	}

	employee, err := client.Employees().Create(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 101, employee.Id)
}

func TestEmployeesService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Invalid input"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateEmployeeInput{
		FirstName: "Test",
		LastName:  "User",
		Company:   100,
	}

	_, err := client.Employees().Create(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}

func TestEmployeesService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path includes employee ID (with /api/v1 prefix)
		assert.Equal(t, "/api/v1/resource/Employee/42", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input UpdateEmployeeInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, "UpdatedFirst", input.FirstName)
		assert.Equal(t, "UpdatedLast", input.LastName)
		assert.Equal(t, "updated@example.com", input.Email)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Employee{
			Id:          42,
			FirstName:   "UpdatedFirst",
			LastName:    "UpdatedLast",
			DisplayName: "UpdatedFirst UpdatedLast",
			Email:       "updated@example.com",
			Active:      true,
			Company:     100,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &UpdateEmployeeInput{
		FirstName: "UpdatedFirst",
		LastName:  "UpdatedLast",
		Email:     "updated@example.com",
	}

	employee, err := client.Employees().Update(context.Background(), 42, input)
	require.NoError(t, err)

	assert.Equal(t, 42, employee.Id)
	assert.Equal(t, "UpdatedFirst", employee.FirstName)
	assert.Equal(t, "UpdatedLast", employee.LastName)
	assert.Equal(t, "updated@example.com", employee.Email)
}

func TestEmployeesService_Update_ActiveStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/Employee/10", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		// Verify the Active field is properly set in request
		var rawInput map[string]interface{}
		err = json.Unmarshal(body, &rawInput)
		require.NoError(t, err)
		assert.Equal(t, false, rawInput["blnActive"])

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Employee{
			Id:        10,
			FirstName: "Test",
			LastName:  "User",
			Active:    false,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	active := false
	input := &UpdateEmployeeInput{
		Active: &active,
	}

	employee, err := client.Employees().Update(context.Background(), 10, input)
	require.NoError(t, err)
	assert.False(t, employee.Active)
}

func TestEmployeesService_Update_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Employee not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &UpdateEmployeeInput{
		FirstName: "Test",
	}

	_, err := client.Employees().Update(context.Background(), 999, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestEmployeesService_Terminate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path includes employee ID and terminate endpoint (with /api/v1 prefix)
		assert.Equal(t, "/api/v1/supervise/employee/42/terminate", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input TerminateInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, "2024-12-31", input.TerminationDate)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Employees().Terminate(context.Background(), 42, "2024-12-31")
	require.NoError(t, err)
}

func TestEmployeesService_Terminate_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/supervise/employee/999/terminate", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Employee not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Employees().Terminate(context.Background(), 999, "2024-12-31")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestEmployeesService_Terminate_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Invalid date format"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Employees().Terminate(context.Background(), 42, "invalid-date")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}

func TestEmployeesService_Invite(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path includes employee ID and invite endpoint (with /api/v1 prefix)
		assert.Equal(t, "/api/v1/supervise/employee/42/invite", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify no body (Invite sends nil body)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Empty(t, body)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Employees().Invite(context.Background(), 42)
	require.NoError(t, err)
}

func TestEmployeesService_Invite_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/supervise/employee/999/invite", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Employee not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Employees().Invite(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestEmployeesService_Invite_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Employees().Invite(context.Background(), 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestEmployeesService_AssignLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/supervise/employee/42/location", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input AssignLocationInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)
		assert.Equal(t, 42, input.Employee)
		assert.Equal(t, 10, input.Location)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Employees().AssignLocation(context.Background(), 42, 10)
	require.NoError(t, err)
}

func TestEmployeesService_AssignLocation_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Invalid location"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Employees().AssignLocation(context.Background(), 42, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}

func TestEmployeesService_RemoveLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/v1/supervise/employee/42/location/10", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Employees().RemoveLocation(context.Background(), 42, 10)
	require.NoError(t, err)
}

func TestEmployeesService_RemoveLocation_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/supervise/employee/42/location/999", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Location not assigned"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Employees().RemoveLocation(context.Background(), 42, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestEmployeesService_Reactivate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/Employee/42", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify body contains Active=true (uses Update internally)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"blnActive":true}`, string(body))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"Id":42,"FirstName":"John","LastName":"Doe","Active":true}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Employees().Reactivate(context.Background(), 42)
	require.NoError(t, err)
}

func TestEmployeesService_Reactivate_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/resource/Employee/999", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Employee not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Employees().Reactivate(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestEmployeesService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/v1/supervise/employee/42", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Employees().Delete(context.Background(), 42)
	require.NoError(t, err)
}

func TestEmployeesService_Delete_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/supervise/employee/999", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Employee not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Employees().Delete(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestEmployeesService_Delete_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Employees().Delete(context.Background(), 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestEmployeesService_AddUnavailability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/EmployeeAvailability", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CreateUnavailabilityInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)
		assert.Equal(t, 42, input.Employee)
		assert.Equal(t, "2024-03-01", input.DateStart)
		assert.Equal(t, "2024-03-05", input.DateEnd)
		assert.Equal(t, "Medical leave", input.Comment)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Unavailability{
			Id:        10,
			Employee:  42,
			DateStart: "2024-03-01",
			DateEnd:   "2024-03-05",
			Comment:   "Medical leave",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &CreateUnavailabilityInput{
		Employee:  42,
		DateStart: "2024-03-01",
		DateEnd:   "2024-03-05",
		Comment:   "Medical leave",
	}

	unavail, err := client.Employees().AddUnavailability(context.Background(), input)
	require.NoError(t, err)

	assert.Equal(t, 10, unavail.Id)
	assert.Equal(t, 42, unavail.Employee)
	assert.Equal(t, "2024-03-01", unavail.DateStart)
	assert.Equal(t, "2024-03-05", unavail.DateEnd)
	assert.Equal(t, "Medical leave", unavail.Comment)
}

func TestEmployeesService_AddUnavailability_NoComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/EmployeeAvailability", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CreateUnavailabilityInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)
		assert.Empty(t, input.Comment)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Unavailability{
			Id:        11,
			Employee:  42,
			DateStart: "2024-04-01",
			DateEnd:   "2024-04-02",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &CreateUnavailabilityInput{
		Employee:  42,
		DateStart: "2024-04-01",
		DateEnd:   "2024-04-02",
	}

	unavail, err := client.Employees().AddUnavailability(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 11, unavail.Id)
}

func TestEmployeesService_AddUnavailability_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Invalid date range"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &CreateUnavailabilityInput{
		Employee:  42,
		DateStart: "2024-01-10",
		DateEnd:   "2024-01-05", // End before start
	}

	_, err := client.Employees().AddUnavailability(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}
