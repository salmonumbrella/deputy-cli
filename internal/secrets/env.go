package secrets

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// FromEnv returns credentials from environment variables (or .env after it's been loaded).
//
// Variables:
// - DEPUTY_TOKEN (required)
// - DEPUTY_INSTALL (optional if DEPUTY_BASE_URL is set)
// - DEPUTY_GEO (optional; if omitted, base URL defaults to install.deputy.com)
// - DEPUTY_BASE_URL (optional; overrides computed base URL, can be host or /api/v1 URL)
// - DEPUTY_AUTH_SCHEME (optional; defaults to "Bearer", can be "OAuth")
func FromEnv() (*Credentials, bool, error) {
	token := strings.TrimSpace(os.Getenv("DEPUTY_TOKEN"))
	if token == "" {
		return nil, false, nil
	}

	install := strings.TrimSpace(os.Getenv("DEPUTY_INSTALL"))
	geo := strings.TrimSpace(os.Getenv("DEPUTY_GEO"))
	base := strings.TrimSpace(os.Getenv("DEPUTY_BASE_URL"))
	scheme := strings.TrimSpace(os.Getenv("DEPUTY_AUTH_SCHEME"))

	if base == "" && install == "" {
		return nil, true, fmt.Errorf("DEPUTY_TOKEN is set, but neither DEPUTY_BASE_URL nor DEPUTY_INSTALL is set")
	}

	creds := &Credentials{
		Token:      token,
		Install:    strings.ToLower(install),
		Geo:        strings.ToLower(geo),
		AuthScheme: scheme,
		CreatedAt:  time.Now(),
	}
	if base != "" {
		creds.BaseURLOverride = base
	}
	return creds, true, nil
}
