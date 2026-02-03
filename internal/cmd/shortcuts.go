package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// resourceMap provides aliases for common resource names.
// This allows users to use singular, plural, or abbreviated forms.
var resourceMap = map[string]string{
	"employees":   "employees",
	"employee":    "employees",
	"emp":         "employees",
	"locations":   "locations",
	"location":    "locations",
	"loc":         "locations",
	"timesheets":  "timesheets",
	"timesheet":   "timesheets",
	"ts":          "timesheets",
	"rosters":     "rosters",
	"roster":      "rosters",
	"shifts":      "rosters",
	"departments": "departments",
	"department":  "departments",
	"dept":        "departments",
	"areas":       "departments",
	"leave":       "leave",
	"webhooks":    "webhooks",
	"webhook":     "webhooks",
	"sales":       "sales",
	"sale":        "sales",
}

func newListCmd() *cobra.Command {
	var limit, offset int

	cmd := &cobra.Command{
		Use:   "list <resource>",
		Short: "List resources (shortcut)",
		Long: `List resources - shortcut for 'deputy <resource> list'.

Examples:
  deputy list employees
  deputy list locations
  deputy list timesheets
  deputy list emp           # alias for employees
  deputy list ts            # alias for timesheets`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resource := args[0]
			root := cmd.Root()

			resolved := resource
			if mapped, ok := resourceMap[resource]; ok {
				resolved = mapped
			}

			// Build args list
			cmdArgs := []string{resolved, "list"}
			if limit > 0 {
				cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", limit))
			}
			if offset > 0 {
				cmdArgs = append(cmdArgs, "--offset", fmt.Sprintf("%d", offset))
			}

			root.SetArgs(cmdArgs)
			return root.Execute()
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (0 = unlimited)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")

	return cmd
}

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <resource> <id>",
		Short: "Get a resource by ID (shortcut)",
		Long: `Get a resource by ID - shortcut for 'deputy <resource> get <id>'.

Examples:
  deputy get employee 123
  deputy get location 1
  deputy get emp 123        # alias for employees
  deputy get ts 456         # alias for timesheets`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resource := args[0]
			id := args[1]
			root := cmd.Root()

			resolved := resource
			if mapped, ok := resourceMap[resource]; ok {
				resolved = mapped
			}

			root.SetArgs([]string{resolved, "get", id})
			return root.Execute()
		},
	}

	return cmd
}
