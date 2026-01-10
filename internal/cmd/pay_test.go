package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
}
