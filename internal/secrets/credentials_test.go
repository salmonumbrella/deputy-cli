package secrets

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCredentials_BaseURL(t *testing.T) {
	tests := []struct {
		name     string
		creds    Credentials
		expected string
	}{
		{
			name: "Australian region",
			creds: Credentials{
				Token:   "test-token",
				Install: "mycompany",
				Geo:     "au",
			},
			expected: "https://mycompany.au.deputy.com/api/v1",
		},
		{
			name: "UK region",
			creds: Credentials{
				Token:   "test-token",
				Install: "acme-corp",
				Geo:     "uk",
			},
			expected: "https://acme-corp.uk.deputy.com/api/v1",
		},
		{
			name: "North America region",
			creds: Credentials{
				Token:   "test-token",
				Install: "enterprise",
				Geo:     "na",
			},
			expected: "https://enterprise.na.deputy.com/api/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.creds.BaseURL())
		})
	}
}

func TestCredentials_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	creds := Credentials{
		Token:     "secret-token-123",
		Install:   "testcompany",
		Geo:       "au",
		CreatedAt: now,
	}

	data, err := json.Marshal(creds)
	assert.NoError(t, err)

	var decoded Credentials
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, creds.Token, decoded.Token)
	assert.Equal(t, creds.Install, decoded.Install)
	assert.Equal(t, creds.Geo, decoded.Geo)
	assert.Equal(t, creds.CreatedAt.Unix(), decoded.CreatedAt.Unix())
}

func TestCredentials_BaseURLV2(t *testing.T) {
	tests := []struct {
		name     string
		creds    Credentials
		expected string
	}{
		{
			name:     "Australian region v2",
			creds:    Credentials{Token: "test", Install: "mycompany", Geo: "au"},
			expected: "https://mycompany.au.deputy.com/api/v2",
		},
		{
			name:     "UK region v2",
			creds:    Credentials{Token: "test", Install: "acme-corp", Geo: "uk"},
			expected: "https://acme-corp.uk.deputy.com/api/v2",
		},
		{
			name:     "NA region v2",
			creds:    Credentials{Token: "test", Install: "enterprise", Geo: "na"},
			expected: "https://enterprise.na.deputy.com/api/v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.creds.BaseURLV2())
		})
	}
}

func TestCredentials_Marshal(t *testing.T) {
	creds := Credentials{
		Token:     "secret-token",
		Install:   "testco",
		Geo:       "na",
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	data, err := creds.Marshal()
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"token":"secret-token"`)
}

func TestUnmarshalCredentials(t *testing.T) {
	data := []byte(`{"token":"abc123","install":"myco","geo":"uk","created_at":"2024-01-15T10:30:00Z"}`)
	creds, err := UnmarshalCredentials(data)
	assert.NoError(t, err)
	assert.Equal(t, "abc123", creds.Token)
}

func TestUnmarshalCredentials_InvalidJSON(t *testing.T) {
	_, err := UnmarshalCredentials([]byte(`{invalid}`))
	assert.Error(t, err)
}
