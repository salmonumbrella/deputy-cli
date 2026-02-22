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

func TestSalesService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/resource/SalesData", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]SalesData{
			{
				Id:        1,
				Company:   10,
				Area:      5,
				Timestamp: 1705312800,
				Value:     1500.50,
				Type:      "revenue",
			},
			{
				Id:        2,
				Company:   10,
				Area:      6,
				Timestamp: 1705316400,
				Value:     2300.75,
				Type:      "revenue",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Sales().List(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].Id)
	assert.Equal(t, 10, result[0].Company)
	assert.Equal(t, 5, result[0].Area)
	assert.Equal(t, 1500.50, result[0].Value)
	assert.Equal(t, "revenue", result[0].Type)
}

func TestSalesService_List_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Sales().List(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestSalesService_List_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]SalesData{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Sales().List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result)
}

// newTestClientV2 creates a client configured to handle v2 API calls
// The test transport preserves query strings
type testServerTransportWithQuery struct {
	testServerURL string
	underlying    http.RoundTripper
}

func (t *testServerTransportWithQuery) RoundTrip(req *http.Request) (*http.Response, error) {
	testURL := t.testServerURL + req.URL.Path
	if req.URL.RawQuery != "" {
		testURL += "?" + req.URL.RawQuery
	}
	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, testURL, req.Body)
	if err != nil {
		return nil, err
	}
	newReq.Header = req.Header
	return t.underlying.RoundTrip(newReq)
}

func TestSalesService_Add(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		// Note: Add uses v2 API
		assert.Equal(t, "/api/v2/metrics", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var input CreateSalesInput
		_ = json.Unmarshal(body, &input)
		assert.Equal(t, 10, input.Company)
		assert.Equal(t, int64(1705312800), input.Timestamp)
		assert.Equal(t, 500.25, input.Value)
		assert.Equal(t, "daily_revenue", input.Type)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SalesData{
			Id:        100,
			Company:   10,
			Timestamp: 1705312800,
			Value:     500.25,
			Type:      "daily_revenue",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &CreateSalesInput{
		Company:   10,
		Timestamp: 1705312800,
		Value:     500.25,
		Type:      "daily_revenue",
	}
	result, err := client.Sales().Add(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 100, result.Id)
	assert.Equal(t, 10, result.Company)
	assert.Equal(t, 500.25, result.Value)
}

func TestSalesService_Add_WithArea(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var input CreateSalesInput
		_ = json.Unmarshal(body, &input)
		assert.Equal(t, 5, input.Area)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SalesData{
			Id:    101,
			Area:  5,
			Value: 100.0,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &CreateSalesInput{
		Company:   10,
		Area:      5,
		Timestamp: 1705312800,
		Value:     100.0,
	}
	result, err := client.Sales().Add(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 5, result.Area)
}

func TestSalesService_Add_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Sales().Add(context.Background(), &CreateSalesInput{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestSalesService_Query(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/resource/SalesData")
		assert.Equal(t, "10", r.URL.Query().Get("company"))
		assert.Equal(t, "1705000000", r.URL.Query().Get("start"))
		assert.Equal(t, "1705999999", r.URL.Query().Get("end"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]SalesData{
			{
				Id:        1,
				Company:   10,
				Timestamp: 1705312800,
				Value:     1500.50,
			},
		})
	}))
	defer server.Close()

	// Use the transport that preserves query strings
	client := newTestClient(server.URL, "test-token")
	client.httpClient.Transport = &testServerTransportWithQuery{
		testServerURL: server.URL,
		underlying:    http.DefaultTransport,
	}

	input := &SalesQueryInput{
		Company:   10,
		StartTime: 1705000000,
		EndTime:   1705999999,
	}
	result, err := client.Sales().Query(context.Background(), input)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 10, result[0].Company)
}

func TestSalesService_Query_CompanyOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "10", r.URL.Query().Get("company"))
		assert.Empty(t, r.URL.Query().Get("start"))
		assert.Empty(t, r.URL.Query().Get("end"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]SalesData{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	client.httpClient.Transport = &testServerTransportWithQuery{
		testServerURL: server.URL,
		underlying:    http.DefaultTransport,
	}

	input := &SalesQueryInput{
		Company: 10,
	}
	_, err := client.Sales().Query(context.Background(), input)
	require.NoError(t, err)
}

func TestSalesService_Query_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	client.httpClient.Transport = &testServerTransportWithQuery{
		testServerURL: server.URL,
		underlying:    http.DefaultTransport,
	}

	_, err := client.Sales().Query(context.Background(), &SalesQueryInput{Company: 10})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestSalesService_Query_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]SalesData{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	client.httpClient.Transport = &testServerTransportWithQuery{
		testServerURL: server.URL,
		underlying:    http.DefaultTransport,
	}

	result, err := client.Sales().Query(context.Background(), &SalesQueryInput{Company: 10})
	require.NoError(t, err)
	assert.Empty(t, result)
}
