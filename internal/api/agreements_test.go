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

func TestAgreementsService_ListByEmployee(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/EmployeeAgreement/QUERY", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var payload map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &payload))
		search := payload["search"].(map[string]interface{})
		cond := search["s1"].(map[string]interface{})
		assert.Equal(t, "Employee", cond["field"])
		assert.Equal(t, "eq", cond["type"])
		assert.Equal(t, float64(42), cond["data"])

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]EmployeeAgreement{{Id: 7, Employee: 42, Active: true}})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	agreements, err := client.Agreements().ListByEmployee(context.Background(), 42, false)
	require.NoError(t, err)
	require.Len(t, agreements, 1)
	assert.Equal(t, 7, agreements[0].Id)
	assert.Equal(t, 42, agreements[0].Employee)
}

func TestAgreementsService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/resource/EmployeeAgreement/9", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(EmployeeAgreement{Id: 9, Employee: 42, Active: true})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	agreement, err := client.Agreements().Get(context.Background(), 9)
	require.NoError(t, err)
	assert.Equal(t, 9, agreement.Id)
}

func TestAgreementsService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/EmployeeAgreement/9", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var payload map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &payload))
		assert.Equal(t, 23.0, payload["BaseRate"])
		_, ok := payload["Config"].([]interface{})
		assert.True(t, ok)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(EmployeeAgreement{Id: 9, Employee: 42, Active: true})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	config := json.RawMessage("[]")
	baseRate := 23.0

	updated, err := client.Agreements().Update(context.Background(), 9, &UpdateAgreementInput{BaseRate: &baseRate, Config: &config})
	require.NoError(t, err)
	assert.Equal(t, 9, updated.Id)
}

func TestAgreementsService_ListByEmployee_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")

	_, err := client.Agreements().ListByEmployee(context.Background(), 42, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestAgreementsService_Get_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")

	_, err := client.Agreements().Get(context.Background(), 9)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestAgreementsService_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"Not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Agreements().Get(context.Background(), 9)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestAgreementsService_Update_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	baseRate := 23.0

	_, err := client.Agreements().Update(context.Background(), 9, &UpdateAgreementInput{BaseRate: &baseRate})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestAgreementsService_Update_Unprocessable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"Invalid update"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	baseRate := 23.0

	_, err := client.Agreements().Update(context.Background(), 9, &UpdateAgreementInput{BaseRate: &baseRate})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 422")
}
