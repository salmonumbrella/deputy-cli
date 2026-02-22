package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocationsService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/supervise/location/simplified"),
			"Expected path to end with /supervise/location/simplified, got %s", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Location{
			{
				Id:          1,
				CompanyName: "Main Office",
				Code:        "MO",
				Address:     json.RawMessage(`"123 Main St"`),
				Active:      true,
				Timezone:    "Australia/Sydney",
			},
			{
				Id:          2,
				CompanyName: "Branch Office",
				Code:        "BO",
				Address:     json.RawMessage(`"456 Branch Ave"`),
				Active:      true,
				Timezone:    "Australia/Melbourne",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	locations, err := client.Locations().List(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, locations, 2)
	assert.Equal(t, 1, locations[0].Id)
	assert.Equal(t, "Main Office", locations[0].CompanyName)
	assert.Equal(t, "MO", locations[0].Code)
	assert.Equal(t, "123 Main St", locations[0].AddressString())
	assert.True(t, locations[0].Active)
	assert.Equal(t, "Australia/Sydney", locations[0].Timezone)
	assert.Equal(t, 2, locations[1].Id)
	assert.Equal(t, "Branch Office", locations[1].CompanyName)
}

func TestLocationsService_List_FallbackCompany(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/supervise/location/simplified"):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error": "not found"}`))
		case strings.HasSuffix(r.URL.Path, "/resource/Company"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]Location{
				{
					Id:          10,
					CompanyName: "Fallback Location",
					Code:        "FB",
					Address:     json.RawMessage(`"1 Fallback Rd"`),
					Active:      true,
					Timezone:    "Australia/Sydney",
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	locations, err := client.Locations().List(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, locations, 1)
	assert.Equal(t, 10, locations[0].Id)
	assert.Equal(t, "Fallback Location", locations[0].CompanyName)
	assert.Equal(t, "FB", locations[0].Code)
}

func TestLocationsService_List_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Location{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	locations, err := client.Locations().List(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, locations)
}

func TestLocationsService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Locations().List(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestLocationsService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/resource/Company/42"),
			"Expected path to end with /resource/Company/42, got %s", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Location{
			Id:          42,
			CompanyName: "Test Location",
			Code:        "TL",
			Address:     json.RawMessage(`"789 Test Blvd"`),
			Active:      true,
			Timezone:    "Australia/Perth",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	location, err := client.Locations().Get(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, 42, location.Id)
	assert.Equal(t, "Test Location", location.CompanyName)
	assert.Equal(t, "TL", location.Code)
	assert.Equal(t, "789 Test Blvd", location.AddressString())
	assert.True(t, location.Active)
	assert.Equal(t, "Australia/Perth", location.Timezone)
}

func TestLocationsService_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Location not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Locations().Get(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestLocationsService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/supervise/location"),
			"Expected path to end with /supervise/location, got %s", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CreateLocationInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)
		assert.Equal(t, "New Location", input.CompanyName)
		assert.Equal(t, "NL", input.Code)
		assert.Equal(t, "100 New St", input.Address)
		assert.Equal(t, "Australia/Brisbane", input.Timezone)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Location{
			Id:          100,
			CompanyName: "New Location",
			Code:        "NL",
			Address:     json.RawMessage(`"100 New St"`),
			Active:      true,
			Timezone:    "Australia/Brisbane",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateLocationInput{
		CompanyName: "New Location",
		Code:        "NL",
		Address:     "100 New St",
		Timezone:    "Australia/Brisbane",
	}

	location, err := client.Locations().Create(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 100, location.Id)
	assert.Equal(t, "New Location", location.CompanyName)
	assert.Equal(t, "NL", location.Code)
	assert.Equal(t, "100 New St", location.AddressString())
	assert.True(t, location.Active)
	assert.Equal(t, "Australia/Brisbane", location.Timezone)
}

func TestLocationsService_Create_MinimalInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CreateLocationInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)
		assert.Equal(t, "Minimal Location", input.CompanyName)
		assert.Empty(t, input.Code)
		assert.Empty(t, input.Address)
		assert.Empty(t, input.Timezone)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Location{
			Id:          101,
			CompanyName: "Minimal Location",
			Active:      true,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateLocationInput{
		CompanyName: "Minimal Location",
	}

	location, err := client.Locations().Create(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 101, location.Id)
	assert.Equal(t, "Minimal Location", location.CompanyName)
}

func TestLocationsService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Company name is required"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateLocationInput{}

	_, err := client.Locations().Create(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}

func TestLocationsService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/supervise/location/42"),
			"Expected path to end with /supervise/location/42, got %s", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input UpdateLocationInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)
		assert.Equal(t, "Updated Location", input.CompanyName)
		assert.Equal(t, "UL", input.Code)
		assert.Equal(t, "200 Updated St", input.Address)
		assert.Equal(t, "Australia/Darwin", input.Timezone)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Location{
			Id:          42,
			CompanyName: "Updated Location",
			Code:        "UL",
			Address:     json.RawMessage(`"200 Updated St"`),
			Active:      true,
			Timezone:    "Australia/Darwin",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &UpdateLocationInput{
		CompanyName: "Updated Location",
		Code:        "UL",
		Address:     "200 Updated St",
		Timezone:    "Australia/Darwin",
	}

	location, err := client.Locations().Update(context.Background(), 42, input)
	require.NoError(t, err)
	assert.Equal(t, 42, location.Id)
	assert.Equal(t, "Updated Location", location.CompanyName)
	assert.Equal(t, "UL", location.Code)
	assert.Equal(t, "200 Updated St", location.AddressString())
	assert.True(t, location.Active)
	assert.Equal(t, "Australia/Darwin", location.Timezone)
}

func TestLocationsService_Update_PartialUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input UpdateLocationInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)
		assert.Equal(t, "Only Name Updated", input.CompanyName)
		assert.Empty(t, input.Code)
		assert.Empty(t, input.Address)
		assert.Empty(t, input.Timezone)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Location{
			Id:          42,
			CompanyName: "Only Name Updated",
			Code:        "OLD",
			Address:     json.RawMessage(`"Old Address"`),
			Active:      true,
			Timezone:    "Australia/Sydney",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &UpdateLocationInput{
		CompanyName: "Only Name Updated",
	}

	location, err := client.Locations().Update(context.Background(), 42, input)
	require.NoError(t, err)
	assert.Equal(t, "Only Name Updated", location.CompanyName)
	assert.Equal(t, "OLD", location.Code)
}

func TestLocationsService_Update_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Location not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &UpdateLocationInput{
		CompanyName: "Test",
	}

	_, err := client.Locations().Update(context.Background(), 999, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestLocationsService_Archive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/supervise/location/42/archive"),
			"Expected path to end with /supervise/location/42/archive, got %s", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Locations().Archive(context.Background(), 42)
	require.NoError(t, err)
}

func TestLocationsService_Archive_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Location not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Locations().Archive(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestLocationsService_Archive_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Cannot archive default location"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Locations().Archive(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestLocationsService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/supervise/location/42"),
			"Expected path to end with /supervise/location/42, got %s", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Locations().Delete(context.Background(), 42)
	require.NoError(t, err)
}

func TestLocationsService_Delete_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Location not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Locations().Delete(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestLocationsService_Delete_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Cannot delete default location"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	err := client.Locations().Delete(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestLocationsService_GetSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/supervise/location/42/settings"),
			"Expected path to end with /supervise/location/42/settings, got %s", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LocationSettings{
			Id: 42,
			Settings: map[string]interface{}{
				"DefaultBreakLength": float64(30),
				"AutoClockOut":       true,
				"TimesheetRounding":  float64(15),
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	settings, err := client.Locations().GetSettings(context.Background(), 42)
	require.NoError(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, 42, settings.Id)
	assert.Equal(t, float64(30), settings.Settings["DefaultBreakLength"])
	assert.Equal(t, true, settings.Settings["AutoClockOut"])
	assert.Equal(t, float64(15), settings.Settings["TimesheetRounding"])
}

func TestLocationsService_GetSettings_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Location not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Locations().GetSettings(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestLocationsService_UpdateSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/supervise/location/42/settings"),
			"Expected path to end with /supervise/location/42/settings, got %s", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input UpdateSettingsInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)
		assert.Equal(t, float64(45), input.Settings["DefaultBreakLength"])
		assert.Equal(t, false, input.Settings["AutoClockOut"])

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	settings := map[string]interface{}{
		"DefaultBreakLength": float64(45),
		"AutoClockOut":       false,
	}
	err := client.Locations().UpdateSettings(context.Background(), 42, settings)
	require.NoError(t, err)
}

func TestLocationsService_UpdateSettings_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Location not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	settings := map[string]interface{}{"DefaultBreakLength": float64(30)}
	err := client.Locations().UpdateSettings(context.Background(), 999, settings)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestLocationsService_UpdateSettings_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Insufficient permissions"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	settings := map[string]interface{}{"DefaultBreakLength": float64(30)}
	err := client.Locations().UpdateSettings(context.Background(), 1, settings)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestLocationsService_List_FallbackIntegerAddress(t *testing.T) {
	// Test that Address can be integer (foreign key) from /resource/Company fallback
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/supervise/location/simplified"):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"code":404,"message":"No method"}}`))
		case strings.HasSuffix(r.URL.Path, "/resource/Company"):
			w.Header().Set("Content-Type", "application/json")
			// Address is an integer (foreign key to Address table) in Company resource
			_, _ = w.Write([]byte(`[{"Id":1,"CompanyName":"Test Location","Address":123,"Active":true}]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	locations, err := client.Locations().List(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, locations, 1)
	assert.Equal(t, 1, locations[0].Id)
	assert.Equal(t, "Test Location", locations[0].CompanyName)
	// Address should be stored and accessible via AddressString()
	assert.Equal(t, "(ref:123)", locations[0].AddressString())
}

func TestLocation_AddressString(t *testing.T) {
	tests := []struct {
		name     string
		address  json.RawMessage
		expected string
	}{
		{
			name:     "string address",
			address:  json.RawMessage(`"123 Main St"`),
			expected: "123 Main St",
		},
		{
			name:     "integer address (as float64 from JSON)",
			address:  json.RawMessage(`456`),
			expected: "(ref:456)",
		},
		{
			name:     "nil address",
			address:  nil,
			expected: "",
		},
		{
			name:     "empty string address",
			address:  json.RawMessage(`""`),
			expected: "",
		},
		{
			name:     "null address",
			address:  json.RawMessage(`null`),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := Location{Address: tt.address}
			assert.Equal(t, tt.expected, loc.AddressString())
		})
	}
}
