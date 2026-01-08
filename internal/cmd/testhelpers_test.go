package cmd

import (
	"net/http"
	"net/url"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/secrets"
)

// testServerTransport redirects API requests to a test server.
type testServerTransport struct {
	testServerURL string
	underlying    http.RoundTripper
}

func (t *testServerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	testURL, _ := url.Parse(t.testServerURL)
	req.URL.Scheme = testURL.Scheme
	req.URL.Host = testURL.Host
	return t.underlying.RoundTrip(req)
}

// newTestClient creates a client configured to use a test server.
// Use this in all command tests that need to mock the Deputy API.
func newTestClient(serverURL, token string) *api.Client {
	creds := &secrets.Credentials{
		Token:   token,
		Install: "test",
		Geo:     "au",
	}
	client := api.NewClient(creds)
	client.SetHTTPClient(&http.Client{
		Transport: &testServerTransport{
			testServerURL: serverURL,
			underlying:    http.DefaultTransport,
		},
	})
	return client
}
