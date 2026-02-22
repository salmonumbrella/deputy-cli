package secrets

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"github.com/salmonumbrella/deputy-cli/internal/config"
	"golang.org/x/term"
)

var ErrNotFound = errors.New("credentials not found")

const (
	defaultProfileKey     = "default"
	keyringBackendEnv     = "DEPUTY_KEYRING_BACKEND"
	keyringPasswordEnv    = "DEPUTY_KEYRING_PASSWORD"
	keyringBackendAuto    = "auto"
	keyringBackendFile    = "file"
	defaultOpenTimeout    = 5 * time.Second
	dbusSessionAddressEnv = "DBUS_SESSION_BUS_ADDRESS"
)

var (
	openKeyring        = keyring.Open
	keyringOpenTimeout = defaultOpenTimeout
)

type Store interface {
	Get() (*Credentials, error)
	Set(creds *Credentials) error
	Delete() error
}

type KeychainStore struct {
	ring keyring.Keyring
}

func NewKeychainStore() (*KeychainStore, error) {
	opts, err := keyringOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	ring, err := openConfiguredKeyring(opts)
	if err != nil {
		return nil, err
	}
	return &KeychainStore{ring: ring}, nil
}

func (s *KeychainStore) Get() (*Credentials, error) {
	item, err := s.ring.Get(defaultProfileKey)
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
		Key:  defaultProfileKey,
		Data: data,
	})
}

func (s *KeychainStore) Delete() error {
	return s.ring.Remove(defaultProfileKey)
}

type keyringOptions struct {
	goos           string
	backend        string
	dbusAddr       string
	credentialsDir string
	password       string
	stdinIsTTY     bool
}

func keyringOptionsFromEnv() (keyringOptions, error) {
	backend, err := parseKeyringBackend(os.Getenv(keyringBackendEnv))
	if err != nil {
		return keyringOptions{}, err
	}

	stdinIsTTY := false
	if os.Stdin != nil {
		stdinIsTTY = term.IsTerminal(int(os.Stdin.Fd()))
	}

	return keyringOptions{
		goos:           runtime.GOOS,
		backend:        backend,
		dbusAddr:       strings.TrimSpace(os.Getenv(dbusSessionAddressEnv)),
		credentialsDir: config.CredentialsDir(),
		password:       os.Getenv(keyringPasswordEnv),
		stdinIsTTY:     stdinIsTTY,
	}, nil
}

func baseKeyringConfig() keyring.Config {
	return keyring.Config{
		ServiceName: config.KeychainService,
		// macOS: avoid repeated permission prompts by trusting the application in the ACL.
		// These fields are ignored by non-macOS backends.
		KeychainTrustApplication:       true,
		KeychainSynchronizable:         false,
		KeychainAccessibleWhenUnlocked: true,
		LibSecretCollectionName:        config.KeychainService,
		KWalletAppID:                   config.KeychainService,
		KWalletFolder:                  config.KeychainService,
	}
}

func openConfiguredKeyring(opts keyringOptions) (keyring.Keyring, error) {
	cfg := baseKeyringConfig()

	if shouldForceFileBackend(opts.goos, opts.backend, opts.dbusAddr) {
		return openFileBackend(cfg, opts)
	}

	if backendType, ok := keyringBackendType(opts.backend); ok {
		if backendType == keyring.FileBackend {
			return openFileBackend(cfg, opts)
		}
		cfg.AllowedBackends = []keyring.BackendType{backendType}
		return openWithOptionalTimeout(cfg, opts)
	}

	ring, err := openWithOptionalTimeout(cfg, opts)
	if err == nil {
		return ring, nil
	}

	// Linux auto-mode fallback: if desktop keyring lookup fails or hangs, fall back
	// to encrypted file storage so headless/server deployments still work.
	if shouldTryLinuxFileFallback(opts.goos, opts.backend) {
		fileRing, fileErr := openFileBackend(cfg, opts)
		if fileErr == nil {
			return fileRing, nil
		}
		return nil, fmt.Errorf("failed to open system keyring: %w; file backend fallback failed: %v", err, fileErr)
	}

	return nil, err
}

func openWithOptionalTimeout(cfg keyring.Config, opts keyringOptions) (keyring.Keyring, error) {
	if opts.goos == "linux" && opts.backend == keyringBackendAuto && keyringOpenTimeout > 0 {
		return openKeyringWithTimeout(cfg, keyringOpenTimeout)
	}
	return openKeyring(cfg)
}

func openKeyringWithTimeout(cfg keyring.Config, timeout time.Duration) (keyring.Keyring, error) {
	type result struct {
		ring keyring.Keyring
		err  error
	}

	done := make(chan result, 1)
	go func() {
		ring, err := openKeyring(cfg)
		done <- result{ring: ring, err: err}
	}()

	select {
	case res := <-done:
		return res.ring, res.err
	case <-time.After(timeout):
		return nil, fmt.Errorf("timed out opening keyring after %s", timeout)
	}
}

func openFileBackend(cfg keyring.Config, opts keyringOptions) (keyring.Keyring, error) {
	passwordPrompt, err := filePasswordPrompt(opts.password, opts.stdinIsTTY)
	if err != nil {
		return nil, err
	}

	cfg.AllowedBackends = []keyring.BackendType{keyring.FileBackend}
	cfg.FileDir = opts.credentialsDir
	cfg.FilePasswordFunc = passwordPrompt

	return openKeyring(cfg)
}

func filePasswordPrompt(password string, stdinIsTTY bool) (keyring.PromptFunc, error) {
	if strings.TrimSpace(password) != "" {
		return keyring.FixedStringPrompt(password), nil
	}
	if stdinIsTTY {
		return keyring.TerminalPrompt, nil
	}
	return nil, fmt.Errorf("%s is required when using file keyring in non-interactive mode", keyringPasswordEnv)
}

func parseKeyringBackend(raw string) (string, error) {
	backend := strings.ToLower(strings.TrimSpace(raw))
	if backend == "" {
		return keyringBackendAuto, nil
	}

	switch backend {
	case keyringBackendAuto,
		keyringBackendFile,
		string(keyring.KeychainBackend),
		string(keyring.SecretServiceBackend),
		"secretservice",
		string(keyring.KWalletBackend),
		string(keyring.KeyCtlBackend),
		string(keyring.PassBackend),
		string(keyring.WinCredBackend):
		if backend == "secretservice" {
			return string(keyring.SecretServiceBackend), nil
		}
		return backend, nil
	default:
		return "", fmt.Errorf("invalid %s %q (expected one of: auto, file, keychain, secret-service, kwallet, keyctl, pass, wincred)", keyringBackendEnv, raw)
	}
}

func keyringBackendType(backend string) (keyring.BackendType, bool) {
	switch backend {
	case keyringBackendAuto:
		return keyring.InvalidBackend, false
	case keyringBackendFile:
		return keyring.FileBackend, true
	case string(keyring.KeychainBackend):
		return keyring.KeychainBackend, true
	case string(keyring.SecretServiceBackend):
		return keyring.SecretServiceBackend, true
	case string(keyring.KWalletBackend):
		return keyring.KWalletBackend, true
	case string(keyring.KeyCtlBackend):
		return keyring.KeyCtlBackend, true
	case string(keyring.PassBackend):
		return keyring.PassBackend, true
	case string(keyring.WinCredBackend):
		return keyring.WinCredBackend, true
	default:
		return keyring.InvalidBackend, false
	}
}

func shouldForceFileBackend(goos string, backend string, dbusAddr string) bool {
	return goos == "linux" && backend == keyringBackendAuto && strings.TrimSpace(dbusAddr) == ""
}

func shouldTryLinuxFileFallback(goos string, backend string) bool {
	return goos == "linux" && backend == keyringBackendAuto
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
