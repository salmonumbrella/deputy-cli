package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deputy-cli/internal/api"
)

// resourceMap provides aliases for common resource names.
// This allows users to use singular, plural, or abbreviated forms.
var resourceMap = map[string]string{
	"employees":   "employees",
	"employee":    "employees",
	"emp":         "employees",
	"e":           "employees",
	"locations":   "locations",
	"location":    "locations",
	"loc":         "locations",
	"timesheets":  "timesheets",
	"timesheet":   "timesheets",
	"ts":          "timesheets",
	"t":           "timesheets",
	"rosters":     "rosters",
	"roster":      "rosters",
	"shifts":      "rosters",
	"shift":       "rosters",
	"r":           "rosters",
	"departments": "departments",
	"department":  "departments",
	"dept":        "departments",
	"area":        "departments",
	"areas":       "departments",
	"d":           "departments",
	"leave":       "leave",
	"webhooks":    "webhooks",
	"webhook":     "webhooks",
	"wh":          "webhooks",
	"sales":       "sales",
	"sale":        "sales",
}

// persistentFlagArgs forwards persistent flags from the root command so that
// shortcut commands (list, get) behave identically to their full-form equivalents.
//
// Keep in sync with the persistent flags defined in root.go (NewRootCmd).
// When a new persistent flag is added to the root command, it must also be
// forwarded here.
func persistentFlagArgs(root *cobra.Command) []string {
	out, _ := root.PersistentFlags().GetString("output")
	query, _ := root.PersistentFlags().GetString("query")
	raw, _ := root.PersistentFlags().GetBool("raw")
	debug, _ := root.PersistentFlags().GetBool("debug")
	noColor, _ := root.PersistentFlags().GetBool("no-color")

	args := []string{"--output", out}
	if query != "" {
		args = append(args, "--query", query)
	}
	if raw {
		args = append(args, "--raw")
	}
	if debug {
		args = append(args, "--debug")
	}
	if noColor {
		args = append(args, "--no-color")
	}

	return args
}

// knownResourceLower maps lowercase resource names to their canonical PascalCase
// form as used by the Deputy API (e.g. "employeeagreement" -> "EmployeeAgreement").
var knownResourceLower = func() map[string]string {
	resources := api.KnownResources()
	m := make(map[string]string, len(resources))
	for _, r := range resources {
		m[strings.ToLower(r)] = r
	}
	return m
}()

// resolveKnownResourceName tries to map a user-provided resource identifier to a
// canonical Deputy API resource name (case-insensitive), falling back to the
// original input when no match is found.
func resolveKnownResourceName(input string) string {
	key := strings.TrimSpace(input)
	if key == "" {
		return input
	}

	if canonical, ok := knownResourceLower[strings.ToLower(key)]; ok {
		return canonical
	}
	return input
}

func newListCmd() *cobra.Command {
	var limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "list <resource>",
		Short: "List resources (shortcut)",
		Long: `List resources - shortcut for 'deputy <resource> list'.

Examples:
  deputy list employees
  deputy list locations
  deputy list timesheets
  deputy list emp           # alias for employees
  deputy list ts            # alias for timesheets
  deputy list EmployeeAgreement  # falls back to 'deputy resource query EmployeeAgreement'`,
		Args: RequireArg("resource"),
		RunE: func(cmd *cobra.Command, args []string) error {
			resource := args[0]
			root := cmd.Root()

			cmdArgs := persistentFlagArgs(root)

			// Normal shortcut path: map a common alias (case-insensitive) to the
			// corresponding first-class CLI noun.
			resourceKey := strings.ToLower(resource)
			if mapped, ok := resourceMap[resourceKey]; ok {
				cmdArgs = append(cmdArgs, mapped, "list")
				if limit > 0 {
					cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", limit))
				}
				if offset > 0 {
					// First-class commands use --offset for pagination
					// (the else branch uses --start for resource query).
					cmdArgs = append(cmdArgs, "--offset", fmt.Sprintf("%d", offset))
				}
				if failEmpty {
					cmdArgs = append(cmdArgs, "--fail-empty")
				}
			} else {
				// Agent desire path: if the user supplies an API resource name (e.g.
				// EmployeeAgreement), transparently route to the generic resource query.
				resourceName := resolveKnownResourceName(resource)
				cmdArgs = append(cmdArgs, "resource", "query", resourceName)
				if limit > 0 {
					cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", limit))
				}
				if offset > 0 {
					// resource query uses --start for pagination offsets
					cmdArgs = append(cmdArgs, "--start", fmt.Sprintf("%d", offset))
				}
				if failEmpty {
					cmdArgs = append(cmdArgs, "--fail-empty")
				}
			}

			root.SetArgs(cmdArgs)
			return root.ExecuteContext(cmd.Context())
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (0 = unlimited)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&failEmpty, "fail-empty", false, "Exit 4 when results are empty (JSON mode)")

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
  deputy get ts 456         # alias for timesheets
  deputy get EmployeeAgreement 194  # falls back to 'deputy resource get EmployeeAgreement 194'`,
		Args: RequireArgs("resource", "id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			resource := args[0]
			id := args[1]
			root := cmd.Root()

			cmdArgs := persistentFlagArgs(root)

			resourceKey := strings.ToLower(resource)
			if mapped, ok := resourceMap[resourceKey]; ok {
				cmdArgs = append(cmdArgs, mapped, "get", id)
			} else {
				resourceName := resolveKnownResourceName(resource)
				cmdArgs = append(cmdArgs, "resource", "get", resourceName, id)
			}

			root.SetArgs(cmdArgs)
			return root.ExecuteContext(cmd.Context())
		},
	}

	return cmd
}
