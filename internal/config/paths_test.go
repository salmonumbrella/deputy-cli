package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigDir_DefaultPath(t *testing.T) {
	// Clear env var to test default behavior
	_ = os.Unsetenv("DEPUTY_CONFIG_DIR")

	dir := ConfigDir()

	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, ".config", "deputy")
	assert.Equal(t, expected, dir)
}

func TestConfigDir_EnvOverride(t *testing.T) {
	customPath := "/tmp/deputy-test-config"
	_ = os.Setenv("DEPUTY_CONFIG_DIR", customPath)
	defer func() { _ = os.Unsetenv("DEPUTY_CONFIG_DIR") }()

	dir := ConfigDir()
	assert.Equal(t, customPath, dir)
}

func TestKeychainService(t *testing.T) {
	assert.Equal(t, "deputy-cli", KeychainService)
}

func TestEnsureConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	customDir := filepath.Join(tmpDir, "deputy-config")
	t.Setenv("DEPUTY_CONFIG_DIR", customDir)

	err := EnsureConfigDir()

	assert.NoError(t, err)
	info, statErr := os.Stat(customDir)
	assert.NoError(t, statErr)
	assert.True(t, info.IsDir())
}
