package secrets

import (
	"errors"

	"github.com/99designs/keyring"
	"github.com/salmonumbrella/deputy-cli/internal/config"
)

var ErrNotFound = errors.New("credentials not found")

type Store interface {
	Get() (*Credentials, error)
	Set(creds *Credentials) error
	Delete() error
}

type KeychainStore struct {
	ring keyring.Keyring
}

func NewKeychainStore() (*KeychainStore, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: config.KeychainService,
		// macOS: avoid repeated permission prompts by trusting the application in the ACL.
		// These fields are ignored by non-macOS backends.
		KeychainTrustApplication:       true,
		KeychainSynchronizable:         false,
		KeychainAccessibleWhenUnlocked: true,
	})
	if err != nil {
		return nil, err
	}
	return &KeychainStore{ring: ring}, nil
}

func (s *KeychainStore) Get() (*Credentials, error) {
	item, err := s.ring.Get("default")
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return UnmarshalCredentials(item.Data)
}

func (s *KeychainStore) Set(creds *Credentials) error {
	data, err := creds.Marshal()
	if err != nil {
		return err
	}
	return s.ring.Set(keyring.Item{
		Key:  "default",
		Data: data,
	})
}

func (s *KeychainStore) Delete() error {
	return s.ring.Remove("default")
}

// MockStore for testing
type MockStore struct {
	creds *Credentials
}

func NewMockStore() *MockStore {
	return &MockStore{}
}

func (s *MockStore) Get() (*Credentials, error) {
	if s.creds == nil {
		return nil, ErrNotFound
	}
	return s.creds, nil
}

func (s *MockStore) Set(creds *Credentials) error {
	s.creds = creds
	return nil
}

func (s *MockStore) Delete() error {
	if s.creds == nil {
		return ErrNotFound
	}
	s.creds = nil
	return nil
}
