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

func TestResourceService_Info(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/resource/Employee/INFO", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ResourceInfo{
			Name: "Employee",
			Fields: map[string]interface{}{
				"Id":          map[string]interface{}{"type": "Integer", "primary": true},
				"DisplayName": map[string]interface{}{"type": "String"},
				"FirstName":   map[string]interface{}{"type": "String"},
				"LastName":    map[string]interface{}{"type": "String"},
			},
			Assocs: map[string]interface{}{
				"Contact":    "Contact",
				"Timesheets": "Timesheet",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Resource("Employee").Info(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Employee", result.Name)
	assert.NotNil(t, result.Fields)
	assert.NotNil(t, result.Fields["Id"])
	assert.NotNil(t, result.Fields["DisplayName"])
	assert.NotNil(t, result.Assocs)
	assocMap := result.AssocsAsMap()
	assert.NotNil(t, assocMap)
	assert.NotNil(t, assocMap["Contact"])
}

func TestResourceService_Info_ArrayAssocs(t *testing.T) {
	// Test that associations returned as an array (like Roster) are handled correctly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/resource/Roster/INFO", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		// Return assocs as an array (as Roster does in the real API)
		_, _ = w.Write([]byte(`{
			"name": "Roster",
			"fields": {"Id": {"type": "Integer", "primary": true}},
			"assocs": ["Employee", "OperationalUnit", "Schedule"]
		}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Resource("Roster").Info(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Roster", result.Name)
	assert.NotNil(t, result.Assocs)

	// AssocsAsMap should return nil for array-type assocs
	assert.Nil(t, result.AssocsAsMap())

	// AssocsAsArray should return the associations
	assocArr := result.AssocsAsArray()
	assert.NotNil(t, assocArr)
	assert.Len(t, assocArr, 3)
	assert.Contains(t, assocArr, "Employee")
	assert.Contains(t, assocArr, "OperationalUnit")
	assert.Contains(t, assocArr, "Schedule")

	// HasAssocs should return true
	assert.True(t, result.HasAssocs())
}

func TestResourceService_Info_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Resource("Employee").Info(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestResourceInfo_HasAssocs_NilOrInvalid(t *testing.T) {
	info := &ResourceInfo{}
	assert.False(t, info.HasAssocs())

	info.Assocs = "not-a-map-or-array"
	assert.False(t, info.HasAssocs())
}

func TestResourceService_Info_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Resource not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	_, err := client.Resource("NonExistent").Info(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestResourceService_Query(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/Employee/QUERY", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var input QueryInput
		_ = json.Unmarshal(body, &input)
		assert.Equal(t, 10, input.Max)
		assert.NotNil(t, input.Search)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{"Id": float64(1), "DisplayName": "John Doe", "Active": true},
			{"Id": float64(2), "DisplayName": "Jane Smith", "Active": true},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &QueryInput{
		Search: map[string]interface{}{"Active": true},
		Max:    10,
	}
	result, err := client.Resource("Employee").Query(context.Background(), input)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, float64(1), result[0]["Id"])
	assert.Equal(t, "John Doe", result[0]["DisplayName"])
}

func TestResourceService_Query_WithJoinAndSort(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var input QueryInput
		_ = json.Unmarshal(body, &input)
		assert.Contains(t, input.Join, "Contact")
		assert.Equal(t, "desc", input.Sort["Id"])

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &QueryInput{
		Join: []string{"Contact"},
		Sort: map[string]string{"Id": "desc"},
	}
	_, err := client.Resource("Employee").Query(context.Background(), input)
	require.NoError(t, err)
}

func TestResourceService_Query_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Resource("Employee").Query(context.Background(), &QueryInput{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestResourceService_Query_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Resource("Employee").Query(context.Background(), &QueryInput{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestResourceService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/resource/Employee/123", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"Id":          float64(123),
			"DisplayName": "John Doe",
			"FirstName":   "John",
			"LastName":    "Doe",
			"Email":       "john@example.com",
			"Active":      true,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Resource("Employee").Get(context.Background(), 123)
	require.NoError(t, err)
	assert.Equal(t, float64(123), result["Id"])
	assert.Equal(t, "John Doe", result["DisplayName"])
	assert.Equal(t, "john@example.com", result["Email"])
	assert.Equal(t, true, result["Active"])
}

func TestResourceService_Get_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Resource("Employee").Get(context.Background(), 123)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestResourceService_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Resource not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	_, err := client.Resource("Employee").Get(context.Background(), 99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestResourceService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/resource/Timesheet", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{"Id": float64(1), "Employee": float64(100), "Date": "2024-01-15"},
			{"Id": float64(2), "Employee": float64(101), "Date": "2024-01-15"},
			{"Id": float64(3), "Employee": float64(100), "Date": "2024-01-16"},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Resource("Timesheet").List(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, float64(1), result[0]["Id"])
	assert.Equal(t, float64(100), result[0]["Employee"])
}

func TestResourceService_List_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Resource("Timesheet").List(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestResourceService_List_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Resource("Leave").List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestResourceService_DifferentResources(t *testing.T) {
	tests := []struct {
		name         string
		resourceName string
		expectedPath string
	}{
		{"Employee", "Employee", "/api/v1/resource/Employee"},
		{"Timesheet", "Timesheet", "/api/v1/resource/Timesheet"},
		{"Leave", "Leave", "/api/v1/resource/Leave"},
		{"Roster", "Roster", "/api/v1/resource/Roster"},
		{"OperationalUnit", "OperationalUnit", "/api/v1/resource/OperationalUnit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedPath, r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode([]map[string]interface{}{})
			}))
			defer server.Close()

			client := newTestClient(server.URL, "test-token")
			_, err := client.Resource(tt.resourceName).List(context.Background())
			require.NoError(t, err)
		})
	}
}

func TestKnownResources(t *testing.T) {
	resources := KnownResources()

	// Verify we have a good number of known resources
	assert.True(t, len(resources) > 10, "Expected at least 10 known resources")

	// Verify some key resources are present
	expectedResources := []string{
		"Employee",
		"Timesheet",
		"Roster",
		"Leave",
		"Company",
		"OperationalUnit",
		"Webhook",
		"SalesData",
	}

	for _, expected := range expectedResources {
		found := false
		for _, r := range resources {
			if r == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected to find resource %s in KnownResources()", expected)
	}
}

func TestKnownResources_NoDuplicates(t *testing.T) {
	resources := KnownResources()
	seen := make(map[string]bool)

	for _, r := range resources {
		assert.False(t, seen[r], "Duplicate resource found: %s", r)
		seen[r] = true
	}
}
