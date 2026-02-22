package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadDotenv_Direct exercises LoadDotenv() in-process so coverage is captured.
// Because sync.Once fires only once per binary, we reset it before the call.
func TestLoadDotenv_Direct(t *testing.T) {
	// Create a temp .env file.
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "test.env")
	require.NoError(t, os.WriteFile(envFile, []byte("DEPUTY_DOTENV_DIRECT_VAR=direct_value\n"), 0o600))

	// Point DEPUTY_ENV_FILE at it.
	t.Setenv("DEPUTY_ENV_FILE", envFile)

	// Reset the package-level sync.Once so LoadDotenv actually runs.
	loadDotenvOnce = sync.Once{}

	LoadDotenv()

	assert.Equal(t, "direct_value", os.Getenv("DEPUTY_DOTENV_DIRECT_VAR"))
	// Clean up the env var so it doesn't leak to other tests.
	t.Cleanup(func() { _ = os.Unsetenv("DEPUTY_DOTENV_DIRECT_VAR") })
}

// TestLoadDotenv_Direct_DefaultCwd exercises the cwd fallback path.
func TestLoadDotenv_Direct_DefaultCwd(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	require.NoError(t, os.WriteFile(envFile, []byte("DEPUTY_CWD_DIRECT_VAR=cwd_value\n"), 0o600))

	// Unset DEPUTY_ENV_FILE so it uses the cwd fallback.
	t.Setenv("DEPUTY_ENV_FILE", "")

	// Change to tmpDir so the .env file is found.
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	// Reset sync.Once.
	loadDotenvOnce = sync.Once{}

	LoadDotenv()

	assert.Equal(t, "cwd_value", os.Getenv("DEPUTY_CWD_DIRECT_VAR"))
	t.Cleanup(func() { _ = os.Unsetenv("DEPUTY_CWD_DIRECT_VAR") })
}

// TestLoadDotenv_Direct_NoFile exercises the path where no .env file exists.
func TestLoadDotenv_Direct_NoFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Unset DEPUTY_ENV_FILE.
	t.Setenv("DEPUTY_ENV_FILE", "")

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	// Reset sync.Once.
	loadDotenvOnce = sync.Once{}

	// Should not panic or error.
	LoadDotenv()
}

// TestLoadDotenv_WithEnvFile uses a subprocess for behavioral correctness
// (fresh sync.Once per process invocation).
func TestLoadDotenv_WithEnvFile(t *testing.T) {
	if os.Getenv("DOTENV_SUBPROCESS") == "1" {
		LoadDotenv()
		val := os.Getenv("DEPUTY_TEST_VAR")
		_, _ = os.Stdout.WriteString("DEPUTY_TEST_VAR=" + val)
		return
	}

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "test.env")
	require.NoError(t, os.WriteFile(envFile, []byte("DEPUTY_TEST_VAR=hello_from_env\n"), 0o600))

	cmd := exec.Command(os.Args[0], "-test.run=^TestLoadDotenv_WithEnvFile$", "-test.v")
	cmd.Env = append(os.Environ(),
		"DOTENV_SUBPROCESS=1",
		"DEPUTY_ENV_FILE="+envFile,
	)
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "subprocess failed: %s", string(out))
	assert.Contains(t, string(out), "DEPUTY_TEST_VAR=hello_from_env")
}

// TestLoadDotenv_DefaultDotEnv tests loading from ./.env in the cwd via subprocess.
func TestLoadDotenv_DefaultDotEnv(t *testing.T) {
	if os.Getenv("DOTENV_SUBPROCESS_CWD") == "1" {
		LoadDotenv()
		val := os.Getenv("DEPUTY_CWD_VAR")
		_, _ = os.Stdout.WriteString("DEPUTY_CWD_VAR=" + val)
		return
	}

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	require.NoError(t, os.WriteFile(envFile, []byte("DEPUTY_CWD_VAR=from_cwd_env\n"), 0o600))

	cmd := exec.Command(os.Args[0], "-test.run=^TestLoadDotenv_DefaultDotEnv$", "-test.v")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "DOTENV_SUBPROCESS_CWD=1")
	filtered := make([]string, 0, len(cmd.Env))
	for _, e := range cmd.Env {
		if len(e) > 16 && e[:16] == "DEPUTY_ENV_FILE=" {
			continue
		}
		filtered = append(filtered, e)
	}
	cmd.Env = filtered

	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "subprocess failed: %s", string(out))
	assert.Contains(t, string(out), "DEPUTY_CWD_VAR=from_cwd_env")
}

func TestLoadDotenv_Direct_OpenClawFallback(t *testing.T) {
	tmpHome := t.TempDir()
	openClawDir := filepath.Join(tmpHome, ".openclaw")
	require.NoError(t, os.MkdirAll(openClawDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(openClawDir, ".env"), []byte("DEPUTY_OPENCLAW_VAR=openclaw_value\n"), 0o600))

	t.Setenv("HOME", tmpHome)
	t.Setenv("DEPUTY_ENV_FILE", "")

	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	loadDotenvOnce = sync.Once{}
	LoadDotenv()

	assert.Equal(t, "openclaw_value", os.Getenv("DEPUTY_OPENCLAW_VAR"))
	t.Cleanup(func() { _ = os.Unsetenv("DEPUTY_OPENCLAW_VAR") })
}

func TestLoadDotenv_Direct_CwdPrecedenceOverOpenClaw(t *testing.T) {
	tmpHome := t.TempDir()
	openClawDir := filepath.Join(tmpHome, ".openclaw")
	require.NoError(t, os.MkdirAll(openClawDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(openClawDir, ".env"), []byte("DEPUTY_PRECEDENCE_VAR=from_openclaw\n"), 0o600))

	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("DEPUTY_PRECEDENCE_VAR=from_cwd\n"), 0o600))

	t.Setenv("HOME", tmpHome)
	t.Setenv("DEPUTY_ENV_FILE", "")

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	loadDotenvOnce = sync.Once{}
	LoadDotenv()

	assert.Equal(t, "from_cwd", os.Getenv("DEPUTY_PRECEDENCE_VAR"))
	t.Cleanup(func() { _ = os.Unsetenv("DEPUTY_PRECEDENCE_VAR") })
}

func TestDefaultDotenvPaths(t *testing.T) {
	tmpHome := t.TempDir()
	tmpDir := t.TempDir()

	t.Setenv("HOME", tmpHome)

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	paths := defaultDotenvPaths()
	require.Len(t, paths, 2)
	normalize := func(p string) string {
		return strings.TrimPrefix(filepath.Clean(p), "/private")
	}
	assert.Equal(t, normalize(filepath.Join(tmpDir, ".env")), normalize(paths[0]))
	assert.Equal(t, filepath.Join(tmpHome, ".openclaw", ".env"), paths[1])
}
