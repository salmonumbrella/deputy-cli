package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var loadDotenvOnce sync.Once

const deputyEnvFileEnv = "DEPUTY_ENV_FILE"

// LoadDotenv loads environment variables from a .env file (if present).
//
// Precedence:
// - If DEPUTY_ENV_FILE is set, only that file is loaded.
// - Otherwise, it first attempts "./.env" from the current working directory.
// - Then it attempts "~/.openclaw/.env" for OpenClaw/systemd deployments.
//
// Existing environment variables are NOT overridden.
func LoadDotenv() {
	loadDotenvOnce.Do(func() {
		if p := strings.TrimSpace(os.Getenv(deputyEnvFileEnv)); p != "" {
			_ = godotenv.Load(p)
			return
		}

		for _, path := range defaultDotenvPaths() {
			_ = godotenv.Load(path)
		}
	})
}

func defaultDotenvPaths() []string {
	paths := make([]string, 0, 2)

	if wd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(wd, ".env"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".openclaw", ".env"))
	}

	return paths
}
