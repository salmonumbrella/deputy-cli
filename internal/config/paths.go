package config

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	AppName         = "deputy-cli"
	KeychainService = "deputy-cli"
	configDirEnv    = "DEPUTY_CONFIG_DIR"
	credsDirEnv     = "DEPUTY_CREDENTIALS_DIR"
)

func ConfigDir() string {
	if dir := strings.TrimSpace(os.Getenv(configDirEnv)); dir != "" {
		return dir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".deputy"
	}

	return filepath.Join(home, ".config", "deputy")
}

func EnsureConfigDir() error {
	return os.MkdirAll(ConfigDir(), 0o700)
}

func CredentialsDir() string {
	if dir := strings.TrimSpace(os.Getenv(credsDirEnv)); dir != "" {
		return dir
	}
	return filepath.Join(ConfigDir(), "credentials")
}

func EnsureCredentialsDir() error {
	return os.MkdirAll(CredentialsDir(), 0o700)
}
