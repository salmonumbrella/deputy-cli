package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPayAwardsListCommand(t *testing.T) {
	t.Run("outputs table", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/v1/payroll/listAwardsLibrary", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.AwardLibraryEntry{
				{"AwardCode": "fastfood", "Name": "Fast Food", "CountryCode": "au"},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newPayAwardsListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "CODE")
		assert.Contains(t, buf.String(), "fastfood")
	})

	t.Run("outputs json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.AwardLibraryEntry{
				{"AwardCode": "fastfood"},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = outfmt.WithFormat(ctx, "json")
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newPayAwardsListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "AwardCode")
	})
}

func TestPayAwardsGetCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/payroll/listAwardsLibrary/fastfood", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.AwardLibraryEntry{"AwardCode": "fastfood", "Name": "Fast Food"})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newPayAwardsGetCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"fastfood"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "AwardCode")
	assert.Contains(t, buf.String(), "fastfood")
}

func TestPayAwardsSetCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/supervise/employee/123/setAwardFromLibrary", r.URL.Path)

		var payload map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "fastfood", payload["strAwardCode"])
		assert.Equal(t, "au", payload["strCountryCode"])
		overrides := payload["arrOverridePayRules"].([]interface{})
		require.Len(t, overrides, 1)
		override := overrides[0].(map[string]interface{})
		assert.Equal(t, "323208", override["Id"])
		assert.Equal(t, 23.0, override["HourlyRate"])

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"Status": "ok"})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newPayAwardsSetCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"123", "--award", "fastfood", "--country", "au", "--override", "323208:23"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Assigned award fastfood")
}

func TestPayAwardsSetCommand_Errors(t *testing.T) {
	t.Run("missing award", func(t *testing.T) {
		cmd := newPayAwardsSetCmd()
		cmd.SetArgs([]string{"123", "--country", "au"})
		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--award is required")
	})

	t.Run("missing country", func(t *testing.T) {
		cmd := newPayAwardsSetCmd()
		cmd.SetArgs([]string{"123", "--award", "fastfood"})
		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--country is required")
	})

	t.Run("invalid override", func(t *testing.T) {
		cmd := newPayAwardsSetCmd()
		cmd.SetArgs([]string{"123", "--award", "fastfood", "--country", "au", "--override", "bad"})
		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid override")
	})

	t.Run("zero override rate", func(t *testing.T) {
		cmd := newPayAwardsSetCmd()
		cmd.SetArgs([]string{"123", "--award", "fastfood", "--country", "au", "--override", "323208:0"})
		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be greater than 0")
	})

	t.Run("negative override rate", func(t *testing.T) {
		cmd := newPayAwardsSetCmd()
		cmd.SetArgs([]string{"123", "--award", "fastfood", "--country", "au", "--override", "323208:-1"})
		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be greater than 0")
	})
}

func TestPayAgreementsListCommand(t *testing.T) {
	t.Run("basic list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/v1/resource/EmployeeAgreement/QUERY", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.EmployeeAgreement{{Id: 7, Employee: 42, Active: true}})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newPayAgreementsListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--employee", "42"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "EMPLOYEE")
		assert.Contains(t, buf.String(), "42")
	})

	t.Run("with active-only flag", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/v1/resource/EmployeeAgreement/QUERY", r.URL.Path)

			var payload map[string]interface{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))

			search, ok := payload["search"].(map[string]interface{})
			require.True(t, ok, "expected search object in payload")

			// Verify s2 filter for Active field
			s2, ok := search["s2"].(map[string]interface{})
			require.True(t, ok, "expected s2 filter when --active-only is set")
			assert.Equal(t, "Active", s2["field"])
			assert.Equal(t, "eq", s2["type"])
			assert.Equal(t, true, s2["data"])

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]api.EmployeeAgreement{{Id: 7, Employee: 42, Active: true}})
		}))
		defer server.Close()

		client := newTestClient(server.URL, "test-token")
		mockFactory := &MockClientFactory{client: client}

		buf := &bytes.Buffer{}
		ctx := WithClientFactory(context.Background(), mockFactory)
		ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

		cmd := newPayAgreementsListCmd()
		cmd.SetContext(ctx)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--employee", "42", "--active-only"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "EMPLOYEE")
		assert.Contains(t, buf.String(), "42")
	})
}

func TestPayAgreementsListCommand_Errors(t *testing.T) {
	cmd := newPayAgreementsListCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--employee is required")
}

func TestPayAgreementsUpdateCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/EmployeeAgreement/9", r.URL.Path)

		var payload map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, 23.0, payload["BaseRate"])
		_, ok := payload["Config"].(map[string]interface{})
		assert.True(t, ok)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.EmployeeAgreement{Id: 9, Employee: 42, Active: true})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newPayAgreementsUpdateCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"9", "--base-rate", "23", "--config", `{"DepartmentalPay": []}`})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Updated agreement 9")
}

func TestPayAgreementsUpdateCommand_ConfigFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/resource/EmployeeAgreement/9", r.URL.Path)

		var payload map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		configMap, ok := payload["Config"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(108), configMap["area"])

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.EmployeeAgreement{Id: 9, Employee: 42, Active: true})
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")
	mockFactory := &MockClientFactory{client: client}

	dir := t.TempDir()
	configPath := dir + "/config.json"
	require.NoError(t, os.WriteFile(configPath, []byte(`{"area":108}`), 0o600))

	buf := &bytes.Buffer{}
	ctx := WithClientFactory(context.Background(), mockFactory)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: buf, ErrOut: buf})

	cmd := newPayAgreementsUpdateCmd()
	cmd.SetContext(ctx)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"9", "--config-file", configPath})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Updated agreement 9")
}

func TestPayAgreementsUpdateCommand_Errors(t *testing.T) {
	t.Run("missing flags", func(t *testing.T) {
		cmd := newPayAgreementsUpdateCmd()
		cmd.SetArgs([]string{"9"})
		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one")
	})

	t.Run("base rate must be greater than zero", func(t *testing.T) {
		cmd := newPayAgreementsUpdateCmd()
		cmd.SetArgs([]string{"9", "--base-rate", "0"})
		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--base-rate must be greater than 0")
	})

	t.Run("config and config-file together", func(t *testing.T) {
		cmd := newPayAgreementsUpdateCmd()
		cmd.SetArgs([]string{"9", "--config", "[]", "--config-file", "file.json"})
		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "use either --config or --config-file")
	})

	t.Run("invalid config json", func(t *testing.T) {
		cmd := newPayAgreementsUpdateCmd()
		cmd.SetArgs([]string{"9", "--config", "{bad"})
		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config must be valid JSON")
	})

	t.Run("missing config file", func(t *testing.T) {
		cmd := newPayAgreementsUpdateCmd()
		cmd.SetArgs([]string{"9", "--config-file", "does-not-exist.json"})
		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "read config file")
	})

	t.Run("invalid config file json", func(t *testing.T) {
		dir := t.TempDir()
		configPath := dir + "/bad.json"
		require.NoError(t, os.WriteFile(configPath, []byte("{bad"), 0o600))

		cmd := newPayAgreementsUpdateCmd()
		cmd.SetArgs([]string{"9", "--config-file", configPath})
		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config must be valid JSON")
	})
}
