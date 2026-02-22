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

func TestNormalizeBaseURLToVersion(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		version string
		want    string
	}{
		{
			name:    "empty string",
			baseURL: "",
			version: "v1",
			want:    "",
		},
		{
			name:    "whitespace only",
			baseURL: "   ",
			version: "v1",
			want:    "",
		},
		{
			name:    "host without scheme",
			baseURL: "mycompany.au.deputy.com/api/v1",
			version: "v1",
			want:    "https://mycompany.au.deputy.com/api/v1",
		},
		{
			name:    "host without scheme to v2",
			baseURL: "mycompany.au.deputy.com/api/v1",
			version: "v2",
			want:    "https://mycompany.au.deputy.com/api/v2",
		},
		{
			name:    "URL with /api/v1 to v2",
			baseURL: "https://mycompany.au.deputy.com/api/v1",
			version: "v2",
			want:    "https://mycompany.au.deputy.com/api/v2",
		},
		{
			name:    "URL with /api/v2 to v1",
			baseURL: "https://mycompany.au.deputy.com/api/v2",
			version: "v1",
			want:    "https://mycompany.au.deputy.com/api/v1",
		},
		{
			name:    "URL with trailing slash",
			baseURL: "https://mycompany.au.deputy.com/api/v1/",
			version: "v2",
			want:    "https://mycompany.au.deputy.com/api/v2",
		},
		{
			name:    "URL with no /api/vN",
			baseURL: "https://mycompany.au.deputy.com",
			version: "v1",
			want:    "https://mycompany.au.deputy.com/api/v1",
		},
		{
			name:    "URL with no /api/vN to v2",
			baseURL: "https://mycompany.au.deputy.com",
			version: "v2",
			want:    "https://mycompany.au.deputy.com/api/v2",
		},
		{
			name:    "bare host without scheme or api path",
			baseURL: "mycompany.au.deputy.com",
			version: "v1",
			want:    "https://mycompany.au.deputy.com/api/v1",
		},
		{
			name:    "http scheme preserved",
			baseURL: "http://localhost:8080/api/v1",
			version: "v2",
			want:    "http://localhost:8080/api/v2",
		},
		{
			name:    "higher version number replaced",
			baseURL: "https://example.com/api/v99",
			version: "v1",
			want:    "https://example.com/api/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeBaseURLToVersion(tt.baseURL, tt.version)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCredentials_BaseURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		creds    Credentials
		expected string
	}{
		{
			name:     "install only without geo",
			creds:    Credentials{Token: "t", Install: "mycompany"},
			expected: "https://mycompany.deputy.com/api/v1",
		},
		{
			name:     "empty install returns empty",
			creds:    Credentials{Token: "t"},
			expected: "",
		},
		{
			name:     "base URL override takes precedence",
			creds:    Credentials{Token: "t", Install: "mycompany", Geo: "au", BaseURLOverride: "https://custom.example.com/api/v2"},
			expected: "https://custom.example.com/api/v1",
		},
		{
			name:     "base URL override bare host",
			creds:    Credentials{Token: "t", BaseURLOverride: "custom.example.com"},
			expected: "https://custom.example.com/api/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.creds.BaseURL())
		})
	}
}

func TestCredentials_BaseURLV2_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		creds    Credentials
		expected string
	}{
		{
			name:     "install only without geo",
			creds:    Credentials{Token: "t", Install: "mycompany"},
			expected: "https://mycompany.deputy.com/api/v2",
		},
		{
			name:     "empty install returns empty",
			creds:    Credentials{Token: "t"},
			expected: "",
		},
		{
			name:     "base URL override takes precedence",
			creds:    Credentials{Token: "t", Install: "mycompany", Geo: "au", BaseURLOverride: "https://custom.example.com/api/v1"},
			expected: "https://custom.example.com/api/v2",
		},
		{
			name:     "base URL override bare host",
			creds:    Credentials{Token: "t", BaseURLOverride: "custom.example.com"},
			expected: "https://custom.example.com/api/v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.creds.BaseURLV2())
		})
	}
}

func TestCredentials_AuthorizationHeaderValue(t *testing.T) {
	tests := []struct {
		name     string
		creds    Credentials
		expected string
	}{
		{
			name:     "default empty scheme uses Bearer",
			creds:    Credentials{Token: "mytoken"},
			expected: "Bearer mytoken",
		},
		{
			name:     "explicit Bearer scheme",
			creds:    Credentials{Token: "mytoken", AuthScheme: "Bearer"},
			expected: "Bearer mytoken",
		},
		{
			name:     "OAuth scheme",
			creds:    Credentials{Token: "oauthtoken", AuthScheme: "OAuth"},
			expected: "OAuth oauthtoken",
		},
		{
			name:     "whitespace-only scheme uses Bearer",
			creds:    Credentials{Token: "mytoken", AuthScheme: "   "},
			expected: "Bearer mytoken",
		},
		{
			name:     "scheme with leading/trailing whitespace is trimmed",
			creds:    Credentials{Token: "mytoken", AuthScheme: "  Bearer  "},
			expected: "Bearer mytoken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.creds.AuthorizationHeaderValue())
		})
	}
}

func TestCredentials_MarshalRoundTrip(t *testing.T) {
	original := Credentials{
		Token:           "round-trip-token",
		Install:         "testco",
		Geo:             "na",
		BaseURLOverride: "https://custom.example.com/api/v1",
		AuthScheme:      "OAuth",
		CreatedAt:       time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC),
	}

	data, err := original.Marshal()
	assert.NoError(t, err)

	restored, err := UnmarshalCredentials(data)
	assert.NoError(t, err)

	assert.Equal(t, original.Token, restored.Token)
	assert.Equal(t, original.Install, restored.Install)
	assert.Equal(t, original.Geo, restored.Geo)
	assert.Equal(t, original.BaseURLOverride, restored.BaseURLOverride)
	assert.Equal(t, original.AuthScheme, restored.AuthScheme)
	assert.Equal(t, original.CreatedAt.Unix(), restored.CreatedAt.Unix())
}
