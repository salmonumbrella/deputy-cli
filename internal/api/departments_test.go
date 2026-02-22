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

func TestDepartmentsService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodGet, r.Method)

		// Verify path (includes /api/v1 prefix from BaseURL)
		assert.Equal(t, "/api/v1/resource/OperationalUnit", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Department{
			{
				Id:          1,
				Company:     100,
				ParentId:    0,
				CompanyName: "Engineering",
				CompanyCode: "ENG",
				Active:      true,
				SortOrder:   1,
			},
			{
				Id:          2,
				Company:     100,
				ParentId:    1,
				CompanyName: "Frontend Team",
				CompanyCode: "ENG-FE",
				Active:      true,
				SortOrder:   2,
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	departments, err := client.Departments().List(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, departments, 2)

	assert.Equal(t, 1, departments[0].Id)
	assert.Equal(t, 100, departments[0].Company)
	assert.Equal(t, "Engineering", departments[0].CompanyName)
	assert.Equal(t, "ENG", departments[0].CompanyCode)
	assert.True(t, departments[0].Active)

	assert.Equal(t, 2, departments[1].Id)
	assert.Equal(t, 1, departments[1].ParentId)
	assert.Equal(t, "Frontend Team", departments[1].CompanyName)
}

func TestDepartmentsService_List_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/resource/OperationalUnit", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Department{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	departments, err := client.Departments().List(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, departments)
}

func TestDepartmentsService_List_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/OperationalUnit/QUERY", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"max":5`)
		assert.Contains(t, string(body), `"start":10`)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Department{
			{Id: 1, CompanyName: "Engineering"},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	departments, err := client.Departments().List(context.Background(), &ListOptions{Limit: 5, Offset: 10})
	require.NoError(t, err)
	require.Len(t, departments, 1)
	assert.Equal(t, 1, departments[0].Id)
}

func TestDepartmentsService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Departments().List(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestDepartmentsService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodGet, r.Method)

		// Verify path includes department ID (with /api/v1 prefix)
		assert.Equal(t, "/api/v1/resource/OperationalUnit/42", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Department{
			Id:          42,
			Company:     100,
			ParentId:    10,
			CompanyName: "Backend Team",
			CompanyCode: "ENG-BE",
			Active:      true,
			SortOrder:   3,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	department, err := client.Departments().Get(context.Background(), 42)
	require.NoError(t, err)

	assert.Equal(t, 42, department.Id)
	assert.Equal(t, 100, department.Company)
	assert.Equal(t, 10, department.ParentId)
	assert.Equal(t, "Backend Team", department.CompanyName)
	assert.Equal(t, "ENG-BE", department.CompanyCode)
	assert.True(t, department.Active)
	assert.Equal(t, 3, department.SortOrder)
}

func TestDepartmentsService_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/resource/OperationalUnit/999", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Department not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Departments().Get(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestDepartmentsService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path (includes /api/v1 prefix from BaseURL)
		assert.Equal(t, "/api/v1/resource/OperationalUnit", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CreateDepartmentInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, 100, input.Company)
		assert.Equal(t, 10, input.ParentId)
		assert.Equal(t, "DevOps Team", input.CompanyName)
		assert.Equal(t, "ENG-DO", input.CompanyCode)
		assert.Equal(t, 5, input.SortOrder)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Department{
			Id:          99,
			Company:     100,
			ParentId:    10,
			CompanyName: "DevOps Team",
			CompanyCode: "ENG-DO",
			Active:      true,
			SortOrder:   5,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateDepartmentInput{
		Company:     100,
		ParentId:    10,
		CompanyName: "DevOps Team",
		CompanyCode: "ENG-DO",
		SortOrder:   5,
	}

	department, err := client.Departments().Create(context.Background(), input)
	require.NoError(t, err)

	assert.Equal(t, 99, department.Id)
	assert.Equal(t, 100, department.Company)
	assert.Equal(t, 10, department.ParentId)
	assert.Equal(t, "DevOps Team", department.CompanyName)
	assert.Equal(t, "ENG-DO", department.CompanyCode)
	assert.True(t, department.Active)
}

func TestDepartmentsService_Create_MinimalInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/OperationalUnit", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CreateDepartmentInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		// Verify required fields only
		assert.Equal(t, 50, input.Company)
		assert.Equal(t, "New Department", input.CompanyName)
		// Optional fields should be empty/zero
		assert.Empty(t, input.CompanyCode)
		assert.Zero(t, input.ParentId)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Department{
			Id:          101,
			Company:     50,
			CompanyName: "New Department",
			Active:      true,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateDepartmentInput{
		Company:     50,
		CompanyName: "New Department",
	}

	department, err := client.Departments().Create(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 101, department.Id)
}

func TestDepartmentsService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Invalid input"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateDepartmentInput{
		Company:     100,
		CompanyName: "Test",
	}

	_, err := client.Departments().Create(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}

func TestDepartmentsService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path includes department ID (with /api/v1 prefix)
		assert.Equal(t, "/api/v1/resource/OperationalUnit/42", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input UpdateDepartmentInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, "Updated Department", input.CompanyName)
		assert.Equal(t, "UPD", input.CompanyCode)
		assert.Equal(t, 10, input.SortOrder)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Department{
			Id:          42,
			Company:     100,
			CompanyName: "Updated Department",
			CompanyCode: "UPD",
			Active:      true,
			SortOrder:   10,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &UpdateDepartmentInput{
		CompanyName: "Updated Department",
		CompanyCode: "UPD",
		SortOrder:   10,
	}

	department, err := client.Departments().Update(context.Background(), 42, input)
	require.NoError(t, err)

	assert.Equal(t, 42, department.Id)
	assert.Equal(t, "Updated Department", department.CompanyName)
	assert.Equal(t, "UPD", department.CompanyCode)
	assert.Equal(t, 10, department.SortOrder)
}

func TestDepartmentsService_Update_ActiveStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/OperationalUnit/10", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		// Verify the Active field is properly set in request
		var rawInput map[string]interface{}
		err = json.Unmarshal(body, &rawInput)
		require.NoError(t, err)
		assert.Equal(t, false, rawInput["blnActive"])

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Department{
			Id:          10,
			Company:     100,
			CompanyName: "Test Department",
			Active:      false,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	active := false
	input := &UpdateDepartmentInput{
		Active: &active,
	}

	department, err := client.Departments().Update(context.Background(), 10, input)
	require.NoError(t, err)
	assert.False(t, department.Active)
}

func TestDepartmentsService_Update_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Department not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &UpdateDepartmentInput{
		CompanyName: "Test",
	}

	_, err := client.Departments().Update(context.Background(), 999, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestDepartmentsService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodDelete, r.Method)

		// Verify path includes department ID (with /api/v1 prefix)
		assert.Equal(t, "/api/v1/resource/OperationalUnit/42", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Departments().Delete(context.Background(), 42)
	require.NoError(t, err)
}

func TestDepartmentsService_Delete_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/resource/OperationalUnit/999", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Department not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Departments().Delete(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestDepartmentsService_Delete_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Departments().Delete(context.Background(), 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestDepartmentsService_List_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Access denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Departments().List(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}
