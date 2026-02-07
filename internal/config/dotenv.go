package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
)

var loadDotenvOnce sync.Once

// LoadDotenv loads environment variables from a .env file (if present).
//
// Precedence:
// - If DEPUTY_ENV_FILE is set, only that file is loaded.
// - Otherwise, it attempts to load "./.env" from the current working directory.
//
// Existing environment variables are NOT overridden.
func LoadDotenv() {
	loadDotenvOnce.Do(func() {
		if p := os.Getenv("DEPUTY_ENV_FILE"); p != "" {
			_ = godotenv.Load(p)
			return
		}

		wd, err := os.Getwd()
		if err != nil {
			return
		}
		_ = godotenv.Load(filepath.Join(wd, ".env"))
	})
}
