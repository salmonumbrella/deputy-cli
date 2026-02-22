package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/salmonumbrella/deputy-cli/internal/secrets"
	"github.com/stretchr/testify/assert"
)

func TestClient_buildURL(t *testing.T) {
	creds := &secrets.Credentials{
		Token:   "test-token",
		Install: "test",
		Geo:     "au",
	}
	client := NewClient(creds)
	base := creds.BaseURL()

	assert.Equal(t, base+"/resource", client.buildURL("/resource", nil))
	assert.Equal(t, base+"/resource?max=10", client.buildURL("/resource", &ListOptions{Limit: 10}))
	assert.Equal(t, base+"/resource?start=5", client.buildURL("/resource", &ListOptions{Offset: 5}))
	assert.Equal(t, base+"/resource?max=10&start=5", client.buildURL("/resource", &ListOptions{Limit: 10, Offset: 5}))
}

func TestClient_SetHTTPClient(t *testing.T) {
	creds := &secrets.Credentials{
		Token:   "test-token",
		Install: "test",
		Geo:     "au",
	}
	client := NewClient(creds)
	original := client.httpClient

	client.SetHTTPClient(nil)
	assert.Equal(t, original, client.httpClient)

	custom := &http.Client{}
	client.SetHTTPClient(custom)
	assert.Equal(t, custom, client.httpClient)
}

func TestClient_doV2_NoResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v2/ping", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	err := client.doV2(context.Background(), http.MethodGet, "/ping", nil, nil)
	assert.NoError(t, err)
}
