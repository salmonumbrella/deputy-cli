package secrets

import (
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
