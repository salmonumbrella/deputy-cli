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

func TestWebhooksService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/resource/Webhook", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Webhook{
			{
				Id:       1,
				Topic:    "Timesheet.Created",
				Url:      "https://example.com/webhooks/timesheet",
				Type:     "REST",
				Enabled:  true,
				Created:  "2024-01-15T10:00:00-08:00",
				Modified: "2024-01-15T11:00:00-08:00",
			},
			{
				Id:       2,
				Topic:    "Employee.Updated",
				Url:      "https://example.com/webhooks/employee",
				Type:     "REST",
				Enabled:  false,
				Created:  "2024-01-16T10:00:00-08:00",
				Modified: "2024-01-16T10:00:00-08:00",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Webhooks().List(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].Id)
	assert.Equal(t, "Timesheet.Created", result[0].Topic)
	assert.Equal(t, "https://example.com/webhooks/timesheet", result[0].Url)
	assert.True(t, result[0].Enabled)
	assert.Equal(t, 2, result[1].Id)
	assert.False(t, result[1].Enabled)
}

func TestWebhooksService_List_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Webhooks().List(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestWebhooksService_List_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Webhook{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Webhooks().List(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestWebhooksService_List_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Access denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	_, err := client.Webhooks().List(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestWebhooksService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/resource/Webhook/123", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Webhook{
			Id:       123,
			Topic:    "Roster.Published",
			Url:      "https://myapp.com/hooks/roster",
			Type:     "REST",
			Enabled:  true,
			Created:  "2024-01-15T10:00:00-08:00",
			Modified: "2024-01-15T12:00:00-08:00",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	result, err := client.Webhooks().Get(context.Background(), 123)
	require.NoError(t, err)
	assert.Equal(t, 123, result.Id)
	assert.Equal(t, "Roster.Published", result.Topic)
	assert.Equal(t, "https://myapp.com/hooks/roster", result.Url)
	assert.Equal(t, "REST", result.Type)
	assert.True(t, result.Enabled)
}

func TestWebhooksService_Get_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Webhooks().Get(context.Background(), 123)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestWebhooksService_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Webhook not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	_, err := client.Webhooks().Get(context.Background(), 99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestWebhooksService_Get_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	_, err := client.Webhooks().Get(context.Background(), 123)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestWebhooksService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/Webhook", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var input CreateWebhookInput
		_ = json.Unmarshal(body, &input)
		assert.Equal(t, "Leave.Created", input.Topic)
		assert.Equal(t, "https://myapp.com/webhooks/leave", input.Url)
		assert.Equal(t, "REST", input.Type)
		assert.True(t, input.Enabled)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Webhook{
			Id:       500,
			Topic:    "Leave.Created",
			Url:      "https://myapp.com/webhooks/leave",
			Type:     "REST",
			Enabled:  true,
			Created:  "2024-01-17T12:00:00-08:00",
			Modified: "2024-01-17T12:00:00-08:00",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &CreateWebhookInput{
		Topic:   "Leave.Created",
		Url:     "https://myapp.com/webhooks/leave",
		Type:    "REST",
		Enabled: true,
	}
	result, err := client.Webhooks().Create(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 500, result.Id)
	assert.Equal(t, "Leave.Created", result.Topic)
	assert.Equal(t, "https://myapp.com/webhooks/leave", result.Url)
	assert.True(t, result.Enabled)
}

func TestWebhooksService_Create_Disabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var input CreateWebhookInput
		_ = json.Unmarshal(body, &input)
		assert.False(t, input.Enabled)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Webhook{
			Id:      501,
			Enabled: false,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	input := &CreateWebhookInput{
		Topic:   "Timesheet.Updated",
		Url:     "https://example.com/hook",
		Enabled: false,
	}
	result, err := client.Webhooks().Create(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, result.Enabled)
}

func TestWebhooksService_Create_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	_, err := client.Webhooks().Create(context.Background(), &CreateWebhookInput{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestWebhooksService_Create_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Invalid topic"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	_, err := client.Webhooks().Create(context.Background(), &CreateWebhookInput{
		Topic: "Invalid.Topic",
		Url:   "https://example.com",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
}

func TestWebhooksService_Create_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	_, err := client.Webhooks().Create(context.Background(), &CreateWebhookInput{
		Topic: "Employee.Insert",
		Url:   "https://example.com/webhook",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestWebhooksService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/v1/resource/Webhook/123", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Webhooks().Delete(context.Background(), 123)
	require.NoError(t, err)
}

func TestWebhooksService_Delete_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "bad-token")
	err := client.Webhooks().Delete(context.Background(), 123)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestWebhooksService_Delete_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Webhook not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Webhooks().Delete(context.Background(), 99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 404")
}

func TestWebhooksService_Delete_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Not authorized to delete this webhook"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.Webhooks().Delete(context.Background(), 123)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}
