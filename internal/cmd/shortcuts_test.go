package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListShortcut(t *testing.T) {
	t.Run("list command is registered", func(t *testing.T) {
		root := NewRootCmd()
		listCmd, _, err := root.Find([]string{"list"})
		require.NoError(t, err)
		require.NotNil(t, listCmd)
		assert.Equal(t, "list <resource>", listCmd.Use)
	})

	t.Run("list command has correct short description", func(t *testing.T) {
		root := NewRootCmd()
		listCmd, _, err := root.Find([]string{"list"})
		require.NoError(t, err)
		assert.Equal(t, "List resources (shortcut)", listCmd.Short)
	})

	t.Run("list command has limit flag", func(t *testing.T) {
		root := NewRootCmd()
		listCmd, _, err := root.Find([]string{"list"})
		require.NoError(t, err)

		limitFlag := listCmd.Flags().Lookup("limit")
		require.NotNil(t, limitFlag)
		assert.Equal(t, "0", limitFlag.DefValue)
	})

	t.Run("list command has offset flag", func(t *testing.T) {
		root := NewRootCmd()
		listCmd, _, err := root.Find([]string{"list"})
		require.NoError(t, err)

		offsetFlag := listCmd.Flags().Lookup("offset")
		require.NotNil(t, offsetFlag)
		assert.Equal(t, "0", offsetFlag.DefValue)
	})

	t.Run("list command requires exactly one argument", func(t *testing.T) {
		root := NewRootCmd()
		listCmd, _, err := root.Find([]string{"list"})
		require.NoError(t, err)

		// RequireArg("resource") should be set
		assert.NotNil(t, listCmd.Args)
	})

	t.Run("list command long description contains examples", func(t *testing.T) {
		root := NewRootCmd()
		listCmd, _, err := root.Find([]string{"list"})
		require.NoError(t, err)

		assert.Contains(t, listCmd.Long, "deputy list employees")
		assert.Contains(t, listCmd.Long, "deputy list locations")
	})
}

func TestGetShortcut(t *testing.T) {
	t.Run("get command is registered", func(t *testing.T) {
		root := NewRootCmd()
		getCmd, _, err := root.Find([]string{"get"})
		require.NoError(t, err)
		require.NotNil(t, getCmd)
		assert.Equal(t, "get <resource> <id>", getCmd.Use)
	})

	t.Run("get command has correct short description", func(t *testing.T) {
		root := NewRootCmd()
		getCmd, _, err := root.Find([]string{"get"})
		require.NoError(t, err)
		assert.Equal(t, "Get a resource by ID (shortcut)", getCmd.Short)
	})

	t.Run("get command requires exactly two arguments", func(t *testing.T) {
		root := NewRootCmd()
		getCmd, _, err := root.Find([]string{"get"})
		require.NoError(t, err)

		// RequireArgs("resource", "id") should be set
		assert.NotNil(t, getCmd.Args)
	})

	t.Run("get command long description contains examples", func(t *testing.T) {
		root := NewRootCmd()
		getCmd, _, err := root.Find([]string{"get"})
		require.NoError(t, err)

		assert.Contains(t, getCmd.Long, "deputy get employee 123")
		assert.Contains(t, getCmd.Long, "deputy get location 1")
	})
}

func TestResourceMap(t *testing.T) {
	tests := []struct {
		alias    string
		expected string
	}{
		// Employees
		{"employees", "employees"},
		{"employee", "employees"},
		{"emp", "employees"},
		{"e", "employees"},
		// Locations
		{"locations", "locations"},
		{"location", "locations"},
		{"loc", "locations"},
		// Timesheets
		{"timesheets", "timesheets"},
		{"timesheet", "timesheets"},
		{"ts", "timesheets"},
		{"t", "timesheets"},
		// Rosters
		{"rosters", "rosters"},
		{"roster", "rosters"},
		{"shifts", "rosters"},
		{"shift", "rosters"},
		{"r", "rosters"},
		// Departments
		{"departments", "departments"},
		{"department", "departments"},
		{"dept", "departments"},
		{"area", "departments"},
		{"areas", "departments"},
		{"d", "departments"},
		// Leave
		{"leave", "leave"},
		// Webhooks
		{"webhooks", "webhooks"},
		{"webhook", "webhooks"},
		{"wh", "webhooks"},
		// Sales
		{"sales", "sales"},
		{"sale", "sales"},
	}

	for _, tt := range tests {
		t.Run(tt.alias+" -> "+tt.expected, func(t *testing.T) {
			resolved, ok := resourceMap[tt.alias]
			require.True(t, ok, "alias %q should exist in resourceMap", tt.alias)
			assert.Equal(t, tt.expected, resolved)
		})
	}
}

func TestRootCmd_HasShortcutCommands(t *testing.T) {
	cmd := NewRootCmd()
	subCmds := cmd.Commands()
	names := make([]string, len(subCmds))
	for i, c := range subCmds {
		names[i] = c.Name()
	}

	assert.Contains(t, names, "list", "root should have list shortcut command")
	assert.Contains(t, names, "get", "root should have get shortcut command")
}

func TestResolveKnownResourceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"exact canonical casing", "Employee", "Employee"},
		{"all lowercase", "employee", "Employee"},
		{"all uppercase", "EMPLOYEE", "Employee"},
		{"mixed case multi-word", "employeeagreement", "EmployeeAgreement"},
		{"unknown resource passes through unchanged", "FooBar", "FooBar"},
		{"empty string returns empty", "", ""},
		{"whitespace-padded input gets trimmed and resolved", "  Employee  ", "Employee"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveKnownResourceName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceMapKeysAreLowercase(t *testing.T) {
	for key := range resourceMap {
		t.Run(key, func(t *testing.T) {
			assert.Equal(t, strings.ToLower(key), key,
				"resourceMap key %q must be lowercase", key)
		})
	}
}
