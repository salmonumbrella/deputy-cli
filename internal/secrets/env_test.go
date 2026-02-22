package secrets

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromEnv_NoToken(t *testing.T) {
	t.Setenv("DEPUTY_TOKEN", "")
	t.Setenv("DEPUTY_INSTALL", "")
	t.Setenv("DEPUTY_GEO", "")
	t.Setenv("DEPUTY_BASE_URL", "")
	t.Setenv("DEPUTY_AUTH_SCHEME", "")

	creds, found, err := FromEnv()
	assert.Nil(t, creds)
	assert.False(t, found)
	assert.NoError(t, err)
}

func TestFromEnv_TokenOnly_NoInstallOrBaseURL(t *testing.T) {
	t.Setenv("DEPUTY_TOKEN", "test-token")
	t.Setenv("DEPUTY_INSTALL", "")
	t.Setenv("DEPUTY_GEO", "")
	t.Setenv("DEPUTY_BASE_URL", "")
	t.Setenv("DEPUTY_AUTH_SCHEME", "")

	creds, found, err := FromEnv()
	assert.Nil(t, creds)
	assert.True(t, found)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DEPUTY_TOKEN is set")
	assert.Contains(t, err.Error(), "DEPUTY_BASE_URL")
	assert.Contains(t, err.Error(), "DEPUTY_INSTALL")
}

func TestFromEnv_TokenAndInstall(t *testing.T) {
	t.Setenv("DEPUTY_TOKEN", "test-token")
	t.Setenv("DEPUTY_INSTALL", "MyCompany")
	t.Setenv("DEPUTY_GEO", "")
	t.Setenv("DEPUTY_BASE_URL", "")
	t.Setenv("DEPUTY_AUTH_SCHEME", "")

	creds, found, err := FromEnv()
	require.NoError(t, err)
	assert.True(t, found)
	require.NotNil(t, creds)

	assert.Equal(t, "test-token", creds.Token)
	assert.Equal(t, "mycompany", creds.Install) // lowercased
	assert.Equal(t, "", creds.Geo)
	assert.Equal(t, "", creds.BaseURLOverride)
}

func TestFromEnv_TokenAndBaseURL(t *testing.T) {
	t.Setenv("DEPUTY_TOKEN", "test-token")
	t.Setenv("DEPUTY_INSTALL", "")
	t.Setenv("DEPUTY_GEO", "")
	t.Setenv("DEPUTY_BASE_URL", "https://custom.example.com/api/v1")
	t.Setenv("DEPUTY_AUTH_SCHEME", "")

	creds, found, err := FromEnv()
	require.NoError(t, err)
	assert.True(t, found)
	require.NotNil(t, creds)

	assert.Equal(t, "test-token", creds.Token)
	assert.Equal(t, "https://custom.example.com/api/v1", creds.BaseURLOverride)
}

func TestFromEnv_TokenInstallAndGeo(t *testing.T) {
	t.Setenv("DEPUTY_TOKEN", "test-token")
	t.Setenv("DEPUTY_INSTALL", "AcmeCorp")
	t.Setenv("DEPUTY_GEO", "AU")
	t.Setenv("DEPUTY_BASE_URL", "")
	t.Setenv("DEPUTY_AUTH_SCHEME", "")

	creds, found, err := FromEnv()
	require.NoError(t, err)
	assert.True(t, found)
	require.NotNil(t, creds)

	assert.Equal(t, "test-token", creds.Token)
	assert.Equal(t, "acmecorp", creds.Install) // lowercased
	assert.Equal(t, "au", creds.Geo)           // lowercased
	assert.Equal(t, "", creds.BaseURLOverride)
}

func TestFromEnv_WithAuthScheme(t *testing.T) {
	t.Setenv("DEPUTY_TOKEN", "oauth-token")
	t.Setenv("DEPUTY_INSTALL", "myco")
	t.Setenv("DEPUTY_GEO", "na")
	t.Setenv("DEPUTY_BASE_URL", "")
	t.Setenv("DEPUTY_AUTH_SCHEME", "OAuth")

	creds, found, err := FromEnv()
	require.NoError(t, err)
	assert.True(t, found)
	require.NotNil(t, creds)

	assert.Equal(t, "OAuth", creds.AuthScheme)
}

func TestFromEnv_WhitespaceToken(t *testing.T) {
	t.Setenv("DEPUTY_TOKEN", "   ")
	t.Setenv("DEPUTY_INSTALL", "myco")
	t.Setenv("DEPUTY_GEO", "")
	t.Setenv("DEPUTY_BASE_URL", "")
	t.Setenv("DEPUTY_AUTH_SCHEME", "")

	creds, found, err := FromEnv()
	assert.Nil(t, creds)
	assert.False(t, found)
	assert.NoError(t, err)
}

func TestFromEnv_BaseURLTakesPrecedence(t *testing.T) {
	t.Setenv("DEPUTY_TOKEN", "test-token")
	t.Setenv("DEPUTY_INSTALL", "myco")
	t.Setenv("DEPUTY_GEO", "au")
	t.Setenv("DEPUTY_BASE_URL", "https://override.example.com/api/v1")
	t.Setenv("DEPUTY_AUTH_SCHEME", "")

	creds, found, err := FromEnv()
	require.NoError(t, err)
	assert.True(t, found)
	require.NotNil(t, creds)

	// Both install and base URL override should be populated
	assert.Equal(t, "myco", creds.Install)
	assert.Equal(t, "https://override.example.com/api/v1", creds.BaseURLOverride)
	// BaseURL method should use the override
	assert.Equal(t, "https://override.example.com/api/v1", creds.BaseURL())
}

func TestFromEnv_CreatedAtIsSet(t *testing.T) {
	t.Setenv("DEPUTY_TOKEN", "test-token")
	t.Setenv("DEPUTY_INSTALL", "myco")
	t.Setenv("DEPUTY_GEO", "")
	t.Setenv("DEPUTY_BASE_URL", "")
	t.Setenv("DEPUTY_AUTH_SCHEME", "")

	creds, found, err := FromEnv()
	require.NoError(t, err)
	assert.True(t, found)
	require.NotNil(t, creds)

	assert.False(t, creds.CreatedAt.IsZero(), "CreatedAt should be set to a non-zero time")
}
