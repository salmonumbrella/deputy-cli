package secrets

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

type Credentials struct {
	Token           string    `json:"token"`
	Install         string    `json:"install,omitempty"`
	Geo             string    `json:"geo,omitempty"`
	BaseURLOverride string    `json:"base_url_override,omitempty"` // If set, should point at /api/v1
	AuthScheme      string    `json:"auth_scheme,omitempty"`       // e.g. "Bearer" (default), "OAuth"
	CreatedAt       time.Time `json:"created_at"`
}

func (c *Credentials) BaseURL() string {
	if c.BaseURLOverride != "" {
		return normalizeBaseURLToVersion(c.BaseURLOverride, "v1")
	}
	if c.Install == "" {
		return ""
	}
	// Backwards-compatible default (install.geo.deputy.com).
	if c.Geo != "" {
		return "https://" + c.Install + "." + c.Geo + ".deputy.com/api/v1"
	}
	// Some tenants use install.deputy.com without a geo subdomain.
	return "https://" + c.Install + ".deputy.com/api/v1"
}

func (c *Credentials) BaseURLV2() string {
	if c.BaseURLOverride != "" {
		return normalizeBaseURLToVersion(c.BaseURLOverride, "v2")
	}
	if c.Install == "" {
		return ""
	}
	if c.Geo != "" {
		return "https://" + c.Install + "." + c.Geo + ".deputy.com/api/v2"
	}
	return "https://" + c.Install + ".deputy.com/api/v2"
}

func (c *Credentials) AuthorizationHeaderValue() string {
	scheme := strings.TrimSpace(c.AuthScheme)
	if scheme == "" {
		scheme = "Bearer"
	}
	return scheme + " " + c.Token
}

var apiVersionRe = regexp.MustCompile(`/api/v\d+`)

func normalizeBaseURLToVersion(baseURL, version string) string {
	u := strings.TrimSpace(baseURL)
	u = strings.TrimRight(u, "/")
	if u == "" {
		return ""
	}

	// Accept a host without scheme for convenience in .env.
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		u = "https://" + u
	}

	// Replace any existing /api/vN with the target version.
	if apiVersionRe.MatchString(u) {
		return apiVersionRe.ReplaceAllString(u, "/api/"+version)
	}

	// No /api/vN present — assume base host and append.
	return u + "/api/" + version
}

func (c *Credentials) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

func UnmarshalCredentials(data []byte) (*Credentials, error) {
	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
