package cmd

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/auth"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSetupServer struct {
	started bool
	result  *auth.SetupResult
	err     error
}

func (f *fakeSetupServer) Start(ctx context.Context) (*auth.SetupResult, error) {
	f.started = true
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

func TestAuthLoginCommand_UsesSetupServerFactory(t *testing.T) {
	server := &fakeSetupServer{
		result: &auth.SetupResult{Install: "acme", Geo: "au"},
	}
	factory := func(store secrets.Store) (setupServer, error) {
		return server, nil
	}

	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = WithStore(ctx, secrets.NewMockStore())
	ctx = WithSetupServerFactory(ctx, factory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newAuthLoginCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, server.started)
	assert.Contains(t, buf.String(), "Authenticated successfully")
	assert.Contains(t, buf.String(), "acme.au.deputy.com")
}

func TestAuthLoginCommand_FactoryError(t *testing.T) {
	expectedErr := errors.New("setup failed")
	factory := func(store secrets.Store) (setupServer, error) {
		return nil, expectedErr
	}

	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = WithStore(ctx, secrets.NewMockStore())
	ctx = WithSetupServerFactory(ctx, factory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newAuthLoginCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start auth server")
}

func TestAuthLoginCommand_StartError(t *testing.T) {
	expectedErr := errors.New("failed to start server")
	server := &fakeSetupServer{
		err: expectedErr,
	}
	factory := func(store secrets.Store) (setupServer, error) {
		return server, nil
	}

	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = WithStore(ctx, secrets.NewMockStore())
	ctx = WithSetupServerFactory(ctx, factory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newAuthLoginCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.Error(t, err)
	assert.True(t, server.started)
	assert.Contains(t, err.Error(), "failed to start server")
}

func TestAuthTestCommand_UsesAuthClientFactory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/me", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		// Simulate Deputy API response with PascalCase field names
		_, _ = w.Write([]byte(`{
			"Name": "Ada Lovelace",
			"PrimaryEmail": "ada@example.com",
			"EmployeeId": 42
		}`))
	}))
	defer server.Close()

	store := secrets.NewMockStore()
	require.NoError(t, store.Set(&secrets.Credentials{
		Token:     "test-token",
		Install:   "acme",
		Geo:       "au",
		CreatedAt: time.Now(),
	}))

	factory := func(creds *secrets.Credentials) (*api.Client, error) {
		return newTestClient(server.URL, creds.Token), nil
	}

	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = WithStore(ctx, store)
	ctx = WithAuthClientFactory(ctx, factory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newAuthTestCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Authentication successful")
	assert.Contains(t, buf.String(), "Ada Lovelace")
	assert.Contains(t, buf.String(), "ada@example.com")
}

func TestAuthTestCommand_NotAuthenticated(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = WithStore(ctx, secrets.NewMockStore())
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newAuthTestCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not authenticated")
}

func TestAuthFactories_Defaults(t *testing.T) {
	t.Run("setup server factory default returns server", func(t *testing.T) {
		ctx := context.Background()
		factory := setupServerFactoryFromContext(ctx)
		if factory == nil {
			t.Fatal("expected default factory")
		}
		server, err := factory(secrets.NewMockStore())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if server == nil {
			t.Fatal("expected server")
		}
	})

	t.Run("auth client factory default returns client", func(t *testing.T) {
		ctx := context.Background()
		factory := authClientFactoryFromContext(ctx)
		if factory == nil {
			t.Fatal("expected default factory")
		}
		client, err := factory(&secrets.Credentials{
			Token:   "test-token",
			Install: "test",
			Geo:     "au",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client == nil {
			t.Fatal("expected client")
		}
	})
}

type errStore struct {
	err error
}

func (s errStore) Get() (*secrets.Credentials, error) { return nil, s.err }
func (s errStore) Set(*secrets.Credentials) error     { return s.err }
func (s errStore) Delete() error                      { return s.err }

func TestAuthTestCommand_StoreError(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = WithStore(ctx, errStore{err: errors.New("store failed")})
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newAuthTestCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "store failed")
}
