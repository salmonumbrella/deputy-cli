package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
	"github.com/salmonumbrella/deputy-cli/internal/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
Store injection is implemented via WithStore(ctx, store) in auth.go.
All auth commands use getStore(cmd.Context()) which checks context first,
then falls back to KeychainStore. Tests inject MockStore via context.
*/

// TestAuthStatusCommand_ViaRootCmd verifies the command is properly registered
func TestAuthStatusCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"auth", "status", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Show current authentication status")
}

// TestAuthStatusCommand_NotAuthenticated tests status when no credentials stored.
func TestAuthStatusCommand_NotAuthenticated(t *testing.T) {
	mockStore := secrets.NewMockStore()

	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = WithStore(ctx, mockStore)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	root := NewRootCmd()
	root.SetContext(ctx)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"auth", "status"})

	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Not authenticated")
}

// TestAuthStatusCommand_Authenticated tests status with credentials.
func TestAuthStatusCommand_Authenticated(t *testing.T) {
	mockStore := secrets.NewMockStore()
	err := mockStore.Set(&secrets.Credentials{
		Token:     "test-token-12345678",
		Install:   "testcompany",
		Geo:       "au",
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = WithStore(ctx, mockStore)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	root := NewRootCmd()
	root.SetContext(ctx)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"auth", "status"})

	err = root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "testcompany")
	assert.Contains(t, output, "AU")
	assert.Contains(t, output, "test...5678") // masked token
}

// TestAuthStatusCommand_JSONOutput tests JSON output for auth status.
func TestAuthStatusCommand_JSONOutput(t *testing.T) {
	createdAt := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	mockStore := secrets.NewMockStore()
	err := mockStore.Set(&secrets.Credentials{
		Token:     "abcd1234efgh5678",
		Install:   "testcompany",
		Geo:       "na",
		CreatedAt: createdAt,
	})
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = WithStore(ctx, mockStore)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})
	ctx = outfmt.WithFormat(ctx, "json")

	root := NewRootCmd()
	root.SetContext(ctx)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"auth", "status", "-o", "json"})

	err = root.Execute()

	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "output should be valid JSON")

	assert.Equal(t, "testcompany", result["install"])
	assert.Equal(t, "NA", result["region"])
	assert.Equal(t, "https://testcompany.na.deputy.com/api/v1", result["base_url"])
	assert.Equal(t, "abcd...5678", result["token_masked"])
	assert.Equal(t, "2025-01-15T10:30:00Z", result["added"])
}

// TestAuthLogoutCommand tests logout removes credentials.
func TestAuthLogoutCommand(t *testing.T) {
	mockStore := secrets.NewMockStore()
	err := mockStore.Set(&secrets.Credentials{
		Token:     "test-token-12345678",
		Install:   "testcompany",
		Geo:       "au",
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = WithStore(ctx, mockStore)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	root := NewRootCmd()
	root.SetContext(ctx)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"auth", "logout"})

	err = root.Execute()

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Credentials removed")

	// Verify store is now empty
	_, getErr := mockStore.Get()
	assert.ErrorIs(t, getErr, secrets.ErrNotFound)
}

// TestAuthLogoutCommand_ViaRootCmd verifies the logout command is properly registered
func TestAuthLogoutCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"auth", "logout", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Remove stored credentials")
}

// TestAuthLogoutCommand_NoCredentials tests logout when no credentials exist
func TestAuthLogoutCommand_NoCredentials(t *testing.T) {
	mockStore := secrets.NewMockStore() // Empty store

	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = WithStore(ctx, mockStore)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	root := NewRootCmd()
	root.SetContext(ctx)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"auth", "logout"})

	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No credentials to remove")
}

// TestAuthAddCommand_ViaRootCmd verifies the add command is properly registered
func TestAuthAddCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"auth", "add", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Add Deputy API credentials directly via command-line flags")
	assert.Contains(t, output, "--token")
	assert.Contains(t, output, "--install")
	assert.Contains(t, output, "--geo")
}

// TestAuthAddCommand_MissingRequiredFlags tests validation for missing flags.
// These tests work because validation happens before store access.
func TestAuthAddCommand_MissingRequiredFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "missing token",
			args:        []string{"auth", "add", "--install", "test", "--geo", "au"},
			expectedErr: "--token is required",
		},
		{
			name:        "missing install",
			args:        []string{"auth", "add", "--token", "abc123", "--geo", "au"},
			expectedErr: "--install is required",
		},
		{
			name:        "missing geo",
			args:        []string{"auth", "add", "--token", "abc123", "--install", "test"},
			expectedErr: "--geo is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := NewRootCmd()
			buf := &bytes.Buffer{}
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs(tt.args)

			err := root.Execute()

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// TestAuthAddCommand_InvalidGeo tests geo validation.
// This works because geo validation happens before store access.
func TestAuthAddCommand_InvalidGeo(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"auth", "add", "--token", "abc123", "--install", "test", "--geo", "invalid"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid geo")
	assert.Contains(t, err.Error(), "must be one of")
}

// TestAuthAddCommand_ValidGeos tests that valid geo values pass validation.
// NOTE: This test actually writes to the system keychain on macOS.
// It validates that valid geos pass the validation check.
func TestAuthAddCommand_ValidGeos(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that writes to system keychain in short mode")
	}

	// These will pass geo validation and attempt keychain write
	// On macOS, this succeeds and writes actual credentials
	validGeos := []string{"au", "uk", "na", "AU", "UK", "NA"}

	for _, geo := range validGeos {
		t.Run(geo, func(t *testing.T) {
			root := NewRootCmd()
			buf := &bytes.Buffer{}
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs([]string{"auth", "add", "--token", "test-token-for-geo-validation", "--install", "test", "--geo", geo})

			err := root.Execute()
			// Should pass geo validation (either succeed fully or fail at keychain, but NOT at geo)
			if err != nil {
				assert.NotContains(t, err.Error(), "invalid geo")
			}
		})
	}
}

// TestAuthAddCommand tests add with valid flags.
func TestAuthAddCommand(t *testing.T) {
	mockStore := secrets.NewMockStore()

	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = WithStore(ctx, mockStore)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	root := NewRootCmd()
	root.SetContext(ctx)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"auth", "add", "--token", "secret", "--install", "myco", "--geo", "au"})

	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Credentials saved")

	creds, getErr := mockStore.Get()
	require.NoError(t, getErr)
	assert.Equal(t, "secret", creds.Token)
	assert.Equal(t, "myco", creds.Install)
	assert.Equal(t, "au", creds.Geo)
}

// TestAuthTestCommand_ViaRootCmd verifies the test command is properly registered
func TestAuthTestCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"auth", "test", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Test authentication by calling /me endpoint")
}

// TestAuthLoginCommand_ViaRootCmd verifies the login command is properly registered
func TestAuthLoginCommand_ViaRootCmd(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"auth", "login", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Opens a browser window to authenticate")
}

// TestGeoValidation tests geo validation through the command.
// Note: isValidGeo is unexported, so we test it indirectly through the command.
func TestGeoValidation(t *testing.T) {
	invalidGeos := []struct {
		geo string
	}{
		{"eu"},
		{"australia"},
		{"invalid"},
		{""},
	}

	for _, tt := range invalidGeos {
		name := tt.geo
		if name == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			root := NewRootCmd()
			buf := &bytes.Buffer{}
			root.SetOut(buf)
			root.SetErr(buf)

			args := []string{"auth", "add", "--token", "abc", "--install", "test"}
			if tt.geo != "" {
				args = append(args, "--geo", tt.geo)
			}
			root.SetArgs(args)

			err := root.Execute()
			require.Error(t, err)

			if tt.geo == "" {
				assert.Contains(t, err.Error(), "--geo is required")
			} else {
				assert.Contains(t, err.Error(), "invalid geo")
			}
		})
	}
}

// TestAuthCommand_HasAllSubcommands verifies all auth subcommands are registered
func TestAuthCommand_HasAllSubcommands(t *testing.T) {
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"auth", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "login")
	assert.Contains(t, output, "add")
	assert.Contains(t, output, "status")
	assert.Contains(t, output, "logout")
	assert.Contains(t, output, "test")
}
