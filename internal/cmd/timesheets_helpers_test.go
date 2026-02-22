package cmd

import (
	"testing"
	"time"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDateFlag(t *testing.T) {
	t.Run("empty returns false", func(t *testing.T) {
		parsed, ok, err := parseDateFlag("", "--from")
		require.NoError(t, err)
		assert.False(t, ok)
		assert.True(t, parsed.IsZero())
	})

	t.Run("invalid returns error", func(t *testing.T) {
		_, ok, err := parseDateFlag("2024-13-01", "--from")
		require.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("valid returns parsed date", func(t *testing.T) {
		parsed, ok, err := parseDateFlag("2024-01-15", "--from")
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), parsed)
	})
}

func TestFilterTimesheetsByDate(t *testing.T) {
	t.Run("filters by range and skips empty dates", func(t *testing.T) {
		timesheets := []api.Timesheet{
			{Id: 1, Date: "2024-01-01"},
			{Id: 2, Date: "2024-01-05"},
			{Id: 3, Date: ""},
		}
		from := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

		filtered, err := filterTimesheetsByDate(timesheets, from, to, true, true)
		require.NoError(t, err)
		require.Len(t, filtered, 1)
		assert.Equal(t, 2, filtered[0].Id)
	})

	t.Run("returns error for invalid timesheet date", func(t *testing.T) {
		timesheets := []api.Timesheet{
			{Id: 10, Date: "not-a-date"},
		}
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		_, err := filterTimesheetsByDate(timesheets, from, time.Time{}, true, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Date")
	})
}
