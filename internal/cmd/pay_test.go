package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPayAgreementsUpdate_ConfigArrayError(t *testing.T) {
	t.Run("rejects array config", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"pay", "agreements", "update", "194", "--config", `[{"area":108}]`})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a JSON object")
	})

	t.Run("rejects primitive config", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"pay", "agreements", "update", "194", "--config", `"just a string"`})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a JSON object")
	})

	t.Run("rejects number config", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"pay", "agreements", "update", "194", "--config", `123`})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a JSON object")
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"pay", "agreements", "update", "194", "--config", `{invalid`})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be valid JSON")
	})

	t.Run("rejects boolean config", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"pay", "agreements", "update", "194", "--config", `true`})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a JSON object")
	})

	t.Run("rejects null config", func(t *testing.T) {
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"pay", "agreements", "update", "194", "--config", `null`})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a JSON object")
	})
}

func TestParseOverridePayRule(t *testing.T) {
	t.Run("colon separator", func(t *testing.T) {
		result, err := parseOverridePayRule("123:23")
		assert.NoError(t, err)
		assert.Equal(t, "123", result.Id)
		assert.Equal(t, 23.0, result.HourlyRate)
	})

	t.Run("equals separator", func(t *testing.T) {
		result, err := parseOverridePayRule("ruleA=19.5")
		assert.NoError(t, err)
		assert.Equal(t, "ruleA", result.Id)
		assert.Equal(t, 19.5, result.HourlyRate)
	})

	t.Run("missing separator", func(t *testing.T) {
		_, err := parseOverridePayRule("bad")
		assert.Error(t, err)
	})

	t.Run("invalid rate", func(t *testing.T) {
		_, err := parseOverridePayRule("123:abc")
		assert.Error(t, err)
	})

	t.Run("zero rate", func(t *testing.T) {
		_, err := parseOverridePayRule("123:0")
		assert.Error(t, err)
	})

	t.Run("negative rate", func(t *testing.T) {
		_, err := parseOverridePayRule("123:-1")
		assert.Error(t, err)
	})

	t.Run("empty id with colon", func(t *testing.T) {
		_, err := parseOverridePayRule(":23")
		assert.Error(t, err)
	})

	t.Run("empty rate with colon", func(t *testing.T) {
		_, err := parseOverridePayRule("id:")
		assert.Error(t, err)
	})

	t.Run("only colon", func(t *testing.T) {
		_, err := parseOverridePayRule(":")
		assert.Error(t, err)
	})
}
