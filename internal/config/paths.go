package config

import (
	"os"
	"path/filepath"
)

const (
	AppName         = "deputy-cli"
	KeychainService = "deputy-cli"
)

func ConfigDir() string {
	if dir := os.Getenv("DEPUTY_CONFIG_DIR"); dir != "" {
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
