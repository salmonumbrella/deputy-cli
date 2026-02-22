package secrets

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/99designs/keyring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockStore_SetAndGet(t *testing.T) {
	store := NewMockStore()

	creds := &Credentials{
		Token:     "test-token",
		Install:   "testcompany",
		Geo:       "au",
		CreatedAt: time.Now(),
	}

	err := store.Set(creds)
	require.NoError(t, err)

	retrieved, err := store.Get()
	require.NoError(t, err)
	assert.Equal(t, creds.Token, retrieved.Token)
	assert.Equal(t, creds.Install, retrieved.Install)
	assert.Equal(t, creds.Geo, retrieved.Geo)
}

func TestMockStore_GetEmpty(t *testing.T) {
	store := NewMockStore()

	_, err := store.Get()
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMockStore_Delete(t *testing.T) {
	store := NewMockStore()

	creds := &Credentials{
		Token:   "test-token",
		Install: "testcompany",
		Geo:     "au",
	}

	err := store.Set(creds)
	require.NoError(t, err)

	err = store.Delete()
	require.NoError(t, err)

	_, err = store.Get()
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMockStore_DeleteWhenEmpty(t *testing.T) {
	store := NewMockStore()
	err := store.Delete()
	// Delete on empty store should return ErrNotFound (matching real keychain behavior)
	assert.ErrorIs(t, err, ErrNotFound)
	// Get should also return ErrNotFound
	_, err = store.Get()
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMockStore_SetOverwrites(t *testing.T) {
	store := NewMockStore()
	_ = store.Set(&Credentials{Token: "first"})
	_ = store.Set(&Credentials{Token: "second"})
	creds, _ := store.Get()
	assert.Equal(t, "second", creds.Token)
}

// testKeychainStore creates a KeychainStore with a file-based backend for testing
func testKeychainStore(t *testing.T) *KeychainStore {
	t.Helper()
	dir := t.TempDir()
	ring, err := keyring.Open(keyring.Config{
		ServiceName:                    "deputy-cli-test",
		AllowedBackends:                []keyring.BackendType{keyring.FileBackend},
		FileDir:                        dir,
		FilePasswordFunc:               func(_ string) (string, error) { return "testpass", nil },
		LibSecretCollectionName:        "deputy-cli-test",
		KWalletAppID:                   "deputy-cli-test",
		KWalletFolder:                  "deputy-cli-test",
		KeychainTrustApplication:       true,
		KeychainSynchronizable:         false,
		KeychainAccessibleWhenUnlocked: true,
	})
	require.NoError(t, err)
	return &KeychainStore{ring: ring}
}

func TestKeychainStore_SetAndGet(t *testing.T) {
	store := testKeychainStore(t)

	creds := &Credentials{
		Token:     "kc-test-token",
		Install:   "testcompany",
		Geo:       "au",
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	err := store.Set(creds)
	require.NoError(t, err)

	retrieved, err := store.Get()
	require.NoError(t, err)
	assert.Equal(t, creds.Token, retrieved.Token)
	assert.Equal(t, creds.Install, retrieved.Install)
	assert.Equal(t, creds.Geo, retrieved.Geo)
}

func TestKeychainStore_GetNotFound(t *testing.T) {
	store := testKeychainStore(t)

	_, err := store.Get()
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestKeychainStore_Delete(t *testing.T) {
	store := testKeychainStore(t)

	creds := &Credentials{
		Token:   "delete-test-token",
		Install: "testco",
		Geo:     "na",
	}

	err := store.Set(creds)
	require.NoError(t, err)

	err = store.Delete()
	require.NoError(t, err)

	_, err = store.Get()
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestKeychainStore_SetOverwrites(t *testing.T) {
	store := testKeychainStore(t)

	_ = store.Set(&Credentials{Token: "first-token", Install: "a", Geo: "au"})
	_ = store.Set(&Credentials{Token: "second-token", Install: "b", Geo: "uk"})

	creds, err := store.Get()
	require.NoError(t, err)
	assert.Equal(t, "second-token", creds.Token)
	assert.Equal(t, "b", creds.Install)
}

func TestKeychainStore_DeleteNonExistent(t *testing.T) {
	store := testKeychainStore(t)

	// Delete on empty store should return an error (key not found)
	err := store.Delete()
	assert.Error(t, err)
}

func TestNewKeychainStore_CanInitOrError(t *testing.T) {
	store, err := NewKeychainStore()
	if err != nil {
		t.Logf("NewKeychainStore returned error (acceptable in test env): %v", err)
		return
	}
	if store == nil {
		t.Fatal("expected store, got nil")
	}
}

func TestShouldForceFileBackend(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		backend  string
		dbusAddr string
		want     bool
	}{
		{name: "linux auto no dbus", goos: "linux", backend: keyringBackendAuto, dbusAddr: "", want: true},
		{name: "linux auto whitespace dbus", goos: "linux", backend: keyringBackendAuto, dbusAddr: "   ", want: true},
		{name: "linux auto with dbus", goos: "linux", backend: keyringBackendAuto, dbusAddr: "unix:path=/tmp/dbus", want: false},
		{name: "linux explicit file", goos: "linux", backend: keyringBackendFile, dbusAddr: "", want: false},
		{name: "darwin auto no dbus", goos: "darwin", backend: keyringBackendAuto, dbusAddr: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shouldForceFileBackend(tt.goos, tt.backend, tt.dbusAddr))
		})
	}
}

func TestParseKeyringBackend(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "empty defaults to auto", input: "", want: keyringBackendAuto},
		{name: "auto upper", input: "AUTO", want: keyringBackendAuto},
		{name: "secretservice alias", input: "secretservice", want: string(keyring.SecretServiceBackend)},
		{name: "secret-service canonical", input: "secret-service", want: string(keyring.SecretServiceBackend)},
		{name: "file", input: "file", want: keyringBackendFile},
		{name: "invalid", input: "nope", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseKeyringBackend(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), keyringBackendEnv)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilePasswordPrompt(t *testing.T) {
	t.Run("uses fixed prompt when env password is set", func(t *testing.T) {
		prompt, err := filePasswordPrompt("secret-password", false)
		require.NoError(t, err)
		require.NotNil(t, prompt)

		got, err := prompt("ignored")
		require.NoError(t, err)
		assert.Equal(t, "secret-password", got)
	})

	t.Run("uses terminal prompt when tty is available", func(t *testing.T) {
		prompt, err := filePasswordPrompt("", true)
		require.NoError(t, err)
		require.NotNil(t, prompt)
	})

	t.Run("fails in non-interactive mode without password", func(t *testing.T) {
		prompt, err := filePasswordPrompt("", false)
		require.Error(t, err)
		assert.Nil(t, prompt)
		assert.Contains(t, err.Error(), keyringPasswordEnv)
	})
}

func TestOpenKeyringWithTimeout(t *testing.T) {
	orig := openKeyring
	t.Cleanup(func() { openKeyring = orig })

	t.Run("returns keyring when opener completes", func(t *testing.T) {
		openKeyring = func(cfg keyring.Config) (keyring.Keyring, error) {
			return keyring.NewArrayKeyring(nil), nil
		}

		ring, err := openKeyringWithTimeout(keyring.Config{}, 50*time.Millisecond)
		require.NoError(t, err)
		assert.NotNil(t, ring)
	})

	t.Run("returns timeout error when opener hangs", func(t *testing.T) {
		block := make(chan struct{})
		done := make(chan struct{})
		openKeyring = func(cfg keyring.Config) (keyring.Keyring, error) {
			defer close(done)
			<-block
			return keyring.NewArrayKeyring(nil), nil
		}

		_, err := openKeyringWithTimeout(keyring.Config{}, 10*time.Millisecond)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "timed out opening keyring")

		close(block)
		<-done
	})
}

func TestOpenConfiguredKeyring(t *testing.T) {
	orig := openKeyring
	t.Cleanup(func() { openKeyring = orig })

	t.Run("forces file backend on headless linux auto mode", func(t *testing.T) {
		var captured keyring.Config
		openKeyring = func(cfg keyring.Config) (keyring.Keyring, error) {
			captured = cfg
			return keyring.NewArrayKeyring(nil), nil
		}

		ring, err := openConfiguredKeyring(keyringOptions{
			goos:           "linux",
			backend:        keyringBackendAuto,
			dbusAddr:       "",
			credentialsDir: "/tmp/deputy-test-credentials",
			password:       "secret",
			stdinIsTTY:     false,
		})
		require.NoError(t, err)
		assert.NotNil(t, ring)
		assert.Equal(t, []keyring.BackendType{keyring.FileBackend}, captured.AllowedBackends)
		assert.Equal(t, "/tmp/deputy-test-credentials", captured.FileDir)
		require.NotNil(t, captured.FilePasswordFunc)
		pwd, err := captured.FilePasswordFunc("ignored")
		require.NoError(t, err)
		assert.Equal(t, "secret", pwd)
	})

	t.Run("falls back to file backend when linux auto keyring open fails", func(t *testing.T) {
		var calls int
		openKeyring = func(cfg keyring.Config) (keyring.Keyring, error) {
			calls++
			if calls == 1 {
				assert.NotEqual(t, []keyring.BackendType{keyring.FileBackend}, cfg.AllowedBackends)
				return nil, errors.New("dbus unavailable")
			}
			assert.Equal(t, []keyring.BackendType{keyring.FileBackend}, cfg.AllowedBackends)
			return keyring.NewArrayKeyring(nil), nil
		}

		ring, err := openConfiguredKeyring(keyringOptions{
			goos:           "linux",
			backend:        keyringBackendAuto,
			dbusAddr:       "unix:path=/tmp/dbus",
			credentialsDir: "/tmp/deputy-test-credentials",
			password:       "secret",
			stdinIsTTY:     false,
		})
		require.NoError(t, err)
		assert.NotNil(t, ring)
		assert.Equal(t, 2, calls)
	})

	t.Run("returns explicit error when non-interactive file backend has no password", func(t *testing.T) {
		_, err := openConfiguredKeyring(keyringOptions{
			goos:           "linux",
			backend:        keyringBackendFile,
			dbusAddr:       "",
			credentialsDir: "/tmp/deputy-test-credentials",
			password:       "",
			stdinIsTTY:     false,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), keyringPasswordEnv)
	})
}

func TestKeyringOptionsFromEnv(t *testing.T) {
	t.Run("reads env configuration", func(t *testing.T) {
		t.Setenv("DEPUTY_KEYRING_BACKEND", "file")
		t.Setenv("DEPUTY_KEYRING_PASSWORD", "pass123")
		t.Setenv("DEPUTY_CREDENTIALS_DIR", "/tmp/deputy-creds")
		t.Setenv("DBUS_SESSION_BUS_ADDRESS", "")

		opts, err := keyringOptionsFromEnv()
		require.NoError(t, err)
		assert.Equal(t, keyringBackendFile, opts.backend)
		assert.Equal(t, "pass123", opts.password)
		assert.Equal(t, "/tmp/deputy-creds", opts.credentialsDir)
	})

	t.Run("invalid backend returns error", func(t *testing.T) {
		t.Setenv("DEPUTY_KEYRING_BACKEND", "definitely-invalid")
		_, err := keyringOptionsFromEnv()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DEPUTY_KEYRING_BACKEND")
	})

	t.Run("unset backend defaults to auto", func(t *testing.T) {
		_ = os.Unsetenv("DEPUTY_KEYRING_BACKEND")
		opts, err := keyringOptionsFromEnv()
		require.NoError(t, err)
		assert.Equal(t, keyringBackendAuto, opts.backend)
	})
}
