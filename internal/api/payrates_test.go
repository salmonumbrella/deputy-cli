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

func TestPayRatesService_ListAwardsLibrary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/payroll/listAwardsLibrary", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]AwardLibraryEntry{
			{"AwardCode": "fastfood", "Name": "Fast Food", "CountryCode": "au"},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	awards, err := client.PayRates().ListAwardsLibrary(context.Background())
	require.NoError(t, err)
	require.Len(t, awards, 1)
	assert.Equal(t, "fastfood", awards[0]["AwardCode"])
}

func TestPayRatesService_GetAwardFromLibrary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/payroll/listAwardsLibrary/fastfood", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(AwardLibraryEntry{"AwardCode": "fastfood", "Name": "Fast Food"})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	award, err := client.PayRates().GetAwardFromLibrary(context.Background(), "fastfood")
	require.NoError(t, err)
	assert.Equal(t, "fastfood", award["AwardCode"])
}

func TestPayRatesService_SetAwardFromLibrary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/supervise/employee/123/setAwardFromLibrary", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var payload map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &payload))
		assert.Equal(t, "au", payload["strCountryCode"])
		assert.Equal(t, "fastfood", payload["strAwardCode"])

		overrides := payload["arrOverridePayRules"].([]interface{})
		require.Len(t, overrides, 1)
		override := overrides[0].(map[string]interface{})
		assert.Equal(t, "323208", override["Id"])
		assert.Equal(t, 23.0, override["HourlyRate"])

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"Status": "ok"})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	result, err := client.PayRates().SetAwardFromLibrary(context.Background(), 123, &SetAwardFromLibraryInput{
		CountryCode:     "au",
		AwardCode:       "fastfood",
		OverridePayRule: []OverridePayRule{{Id: "323208", HourlyRate: 23.0}},
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", result["Status"])
}

func TestPayRatesService_ListAwardsLibrary_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")

	_, err := client.PayRates().ListAwardsLibrary(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestPayRatesService_GetAwardFromLibrary_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")

	_, err := client.PayRates().GetAwardFromLibrary(context.Background(), "fastfood")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestPayRatesService_GetAwardFromLibrary_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"Not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.PayRates().GetAwardFromLibrary(context.Background(), "missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestPayRatesService_SetAwardFromLibrary_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")

	_, err := client.PayRates().SetAwardFromLibrary(context.Background(), 123, &SetAwardFromLibraryInput{
		CountryCode: "au",
		AwardCode:   "fastfood",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestPayRatesService_SetAwardFromLibrary_Unprocessable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"Invalid award"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.PayRates().SetAwardFromLibrary(context.Background(), 123, &SetAwardFromLibraryInput{
		CountryCode: "au",
		AwardCode:   "fastfood",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 422")
}
