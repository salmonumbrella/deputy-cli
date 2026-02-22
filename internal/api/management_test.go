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

func TestManagementService_CreateMemo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPut, r.Method)

		// Verify path (includes /api/v1 prefix from BaseURL)
		assert.Equal(t, "/api/v1/supervise/memo", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CreateMemoInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, "Team meeting tomorrow at 9am", input.Content)
		assert.Equal(t, 100, input.Company)
		assert.Equal(t, int64(1704067200), input.ShowFrom)
		assert.Equal(t, int64(1704153600), input.ShowUntil)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Memo{
			Id:        1,
			Content:   "Team meeting tomorrow at 9am",
			Company:   100,
			Creator:   5,
			Created:   1704067200,
			ShowFrom:  1704067200,
			ShowUntil: 1704153600,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateMemoInput{
		Content:   "Team meeting tomorrow at 9am",
		Company:   100,
		ShowFrom:  1704067200,
		ShowUntil: 1704153600,
	}

	memo, err := client.Management().CreateMemo(context.Background(), input)
	require.NoError(t, err)

	assert.Equal(t, 1, memo.Id)
	assert.Equal(t, "Team meeting tomorrow at 9am", memo.Content)
	assert.Equal(t, 100, memo.Company)
	assert.Equal(t, 5, memo.Creator)
	assert.Equal(t, int64(1704067200), memo.ShowFrom)
	assert.Equal(t, int64(1704153600), memo.ShowUntil)
}

func TestManagementService_CreateMemo_MinimalInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/api/v1/supervise/memo", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CreateMemoInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		// Verify required fields only
		assert.Equal(t, "Quick note", input.Content)
		assert.Equal(t, 50, input.Company)
		// Optional fields should be zero
		assert.Zero(t, input.ShowFrom)
		assert.Zero(t, input.ShowUntil)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Memo{
			Id:      2,
			Content: "Quick note",
			Company: 50,
			Creator: 1,
			Created: 1704067200,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateMemoInput{
		Content: "Quick note",
		Company: 50,
	}

	memo, err := client.Management().CreateMemo(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 2, memo.Id)
	assert.Equal(t, "Quick note", memo.Content)
}

func TestManagementService_CreateMemo_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateMemoInput{
		Content: "Test",
		Company: 100,
	}

	_, err := client.Management().CreateMemo(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestManagementService_CreateMemo_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Access denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateMemoInput{
		Content: "Test",
		Company: 100,
	}

	_, err := client.Management().CreateMemo(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestManagementService_ListMemos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodGet, r.Method)

		// Verify path and query parameter
		assert.Equal(t, "/api/v1/supervise/memo", r.URL.Path)
		assert.Equal(t, "100", r.URL.Query().Get("company"))

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Memo{
			{
				Id:        1,
				Content:   "First memo",
				Company:   100,
				Creator:   5,
				Created:   1704067200,
				ShowFrom:  1704067200,
				ShowUntil: 1704153600,
			},
			{
				Id:      2,
				Content: "Second memo",
				Company: 100,
				Creator: 3,
				Created: 1704153600,
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	memos, err := client.Management().ListMemos(context.Background(), 100)
	require.NoError(t, err)
	require.Len(t, memos, 2)

	assert.Equal(t, 1, memos[0].Id)
	assert.Equal(t, "First memo", memos[0].Content)
	assert.Equal(t, 100, memos[0].Company)
	assert.Equal(t, 5, memos[0].Creator)
	assert.Equal(t, int64(1704067200), memos[0].ShowFrom)

	assert.Equal(t, 2, memos[1].Id)
	assert.Equal(t, "Second memo", memos[1].Content)
	assert.Equal(t, 3, memos[1].Creator)
}

func TestManagementService_ListMemos_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/supervise/memo", r.URL.Path)
		assert.Equal(t, "999", r.URL.Query().Get("company"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Memo{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	memos, err := client.Management().ListMemos(context.Background(), 999)
	require.NoError(t, err)
	assert.Empty(t, memos)
}

func TestManagementService_ListMemos_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Management().ListMemos(context.Background(), 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestManagementService_ListMemos_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Access denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Management().ListMemos(context.Background(), 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestManagementService_PostJournal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify path (includes /api/v1 prefix from BaseURL)
		assert.Equal(t, "/api/v1/supervise/journal", r.URL.Path)

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CreateJournalInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		assert.Equal(t, 42, input.Employee)
		assert.Equal(t, 100, input.Company)
		assert.Equal(t, "Performance review completed", input.Comment)
		assert.Equal(t, 3, input.Category)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Journal{
			Id:       1,
			Employee: 42,
			Company:  100,
			Comment:  "Performance review completed",
			Created:  1704067200,
			Category: 3,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateJournalInput{
		Employee: 42,
		Company:  100,
		Comment:  "Performance review completed",
		Category: 3,
	}

	journal, err := client.Management().PostJournal(context.Background(), input)
	require.NoError(t, err)

	assert.Equal(t, 1, journal.Id)
	assert.Equal(t, 42, journal.Employee)
	assert.Equal(t, 100, journal.Company)
	assert.Equal(t, "Performance review completed", journal.Comment)
	assert.Equal(t, 3, journal.Category)
}

func TestManagementService_PostJournal_MinimalInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/supervise/journal", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var input CreateJournalInput
		err = json.Unmarshal(body, &input)
		require.NoError(t, err)

		// Verify required fields only
		assert.Equal(t, 10, input.Employee)
		assert.Equal(t, 50, input.Company)
		assert.Equal(t, "Quick note", input.Comment)
		// Optional field should be zero
		assert.Zero(t, input.Category)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Journal{
			Id:       2,
			Employee: 10,
			Company:  50,
			Comment:  "Quick note",
			Created:  1704067200,
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateJournalInput{
		Employee: 10,
		Company:  50,
		Comment:  "Quick note",
	}

	journal, err := client.Management().PostJournal(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 2, journal.Id)
	assert.Equal(t, "Quick note", journal.Comment)
}

func TestManagementService_PostJournal_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateJournalInput{
		Employee: 42,
		Company:  100,
		Comment:  "Test",
	}

	_, err := client.Management().PostJournal(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestManagementService_PostJournal_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Access denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	input := &CreateJournalInput{
		Employee: 42,
		Company:  100,
		Comment:  "Test",
	}

	_, err := client.Management().PostJournal(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}

func TestManagementService_ListJournals(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify HTTP method
		assert.Equal(t, http.MethodGet, r.Method)

		// Verify path and query parameter
		assert.Equal(t, "/api/v1/supervise/journal", r.URL.Path)
		assert.Equal(t, "42", r.URL.Query().Get("employee"))

		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Journal{
			{
				Id:       1,
				Employee: 42,
				Company:  100,
				Comment:  "First journal entry",
				Created:  1704067200,
				Category: 1,
			},
			{
				Id:       2,
				Employee: 42,
				Company:  100,
				Comment:  "Second journal entry",
				Created:  1704153600,
				Category: 2,
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	journals, err := client.Management().ListJournals(context.Background(), 42)
	require.NoError(t, err)
	require.Len(t, journals, 2)

	assert.Equal(t, 1, journals[0].Id)
	assert.Equal(t, 42, journals[0].Employee)
	assert.Equal(t, 100, journals[0].Company)
	assert.Equal(t, "First journal entry", journals[0].Comment)
	assert.Equal(t, 1, journals[0].Category)

	assert.Equal(t, 2, journals[1].Id)
	assert.Equal(t, "Second journal entry", journals[1].Comment)
	assert.Equal(t, 2, journals[1].Category)
}

func TestManagementService_ListJournals_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/supervise/journal", r.URL.Path)
		assert.Equal(t, "999", r.URL.Query().Get("employee"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Journal{})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	journals, err := client.Management().ListJournals(context.Background(), 999)
	require.NoError(t, err)
	assert.Empty(t, journals)
}

func TestManagementService_ListJournals_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Management().ListJournals(context.Background(), 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 401")
}

func TestManagementService_ListJournals_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Access denied"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	_, err := client.Management().ListJournals(context.Background(), 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 403")
}
