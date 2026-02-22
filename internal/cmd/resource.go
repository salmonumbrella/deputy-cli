package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
)

func newResourceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resource",
		Aliases: []string{"res"},
		Short:   "Query any Deputy resource",
		Long:    "Generic commands for querying any Deputy resource type using the QUERY API.",
	}

	cmd.AddCommand(newResourceListCmd())
	cmd.AddCommand(newResourceInfoCmd())
	cmd.AddCommand(newResourceQueryCmd())
	cmd.AddCommand(newResourceGetCmd())

	return cmd
}

func newResourceListCmd() *cobra.Command {
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List known resource types",
		RunE: func(cmd *cobra.Command, args []string) error {
			resources := api.KnownResources()

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				ctx := outfmt.WithFailEmpty(cmd.Context(), failEmpty)
				f := outfmt.New(ctx)
				return f.OutputList(resources)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"RESOURCE"})
			for _, r := range resources {
				f.Row(r)
			}
			f.EndTable()
			return nil
		},
	}

	cmd.Flags().BoolVar(&failEmpty, "fail-empty", false, "Exit 4 when results are empty (JSON mode)")

	return cmd
}

func newResourceInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <ResourceName>",
		Short: "Get schema information for a resource",
		Args:  RequireArg("ResourceName"),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceName := args[0]

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			info, err := client.Resource(resourceName).Info(cmd.Context())
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(info)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Resource: %s\n\n", info.Name)

			_, _ = fmt.Fprintf(io.Out, "Fields:\n")
			fieldNames := make([]string, 0, len(info.Fields))
			for name := range info.Fields {
				fieldNames = append(fieldNames, name)
			}
			sort.Strings(fieldNames)
			for _, name := range fieldNames {
				_, _ = fmt.Fprintf(io.Out, "  %s: %v\n", name, info.Fields[name])
			}

			if info.HasAssocs() {
				_, _ = fmt.Fprintf(io.Out, "\nAssociations:\n")
				if assocMap := info.AssocsAsMap(); assocMap != nil {
					assocNames := make([]string, 0, len(assocMap))
					for name := range assocMap {
						assocNames = append(assocNames, name)
					}
					sort.Strings(assocNames)
					for _, name := range assocNames {
						_, _ = fmt.Fprintf(io.Out, "  %s: %v\n", name, assocMap[name])
					}
				} else if assocArr := info.AssocsAsArray(); assocArr != nil {
					for _, assocName := range assocArr {
						_, _ = fmt.Fprintf(io.Out, "  %s\n", assocName)
					}
				}
			}

			return nil
		},
	}
}

func newResourceQueryCmd() *cobra.Command {
	var filters []string
	var joins []string
	var sortField string
	var limit int
	var start int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "query <ResourceName>",
		Short: "Query a resource with filters",
		Long: `Query any Deputy resource with filters, joins, and sorting.

Filter syntax:
  field=value     Exact match
  field>value     Greater than
  field<value     Less than
  field>=value    Greater than or equal
  field<=value    Less than or equal

Examples:
  deputy resource query Employee --filter "Active=1"
  deputy resource query Timesheet --filter "Employee=123" --filter "Date>=2024-01-01"
  deputy resource query Roster --filter "StartTime>2024-01-01" --join Employee --sort StartTime --limit 100
  deputy resource query Leave --filter "Status=1" --join Employee
  one_month_ago=$(date -v-1m +%Y-%m-%d); deputy resource query Timesheet --filter "Date>=$one_month_ago" --raw`,
		Args: RequireArg("ResourceName"),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceName := args[0]

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			search, err := parseFilters(filters)
			if err != nil {
				return err
			}

			input := &api.QueryInput{
				Search: search,
				Join:   joins,
				Max:    limit,
				Start:  start,
			}

			if sortField != "" {
				input.Sort = map[string]string{sortField: "asc"}
			}

			results, err := client.Resource(resourceName).Query(cmd.Context(), input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				ctx := outfmt.WithLimit(cmd.Context(), limit)
				ctx = outfmt.WithOffset(ctx, start)
				ctx = outfmt.WithFailEmpty(ctx, failEmpty)
				f := outfmt.New(ctx)
				return f.OutputList(results)
			}

			io := iocontext.FromContext(cmd.Context())
			if len(results) == 0 {
				_, _ = fmt.Fprintf(io.Out, "No results found\n")
				return nil
			}

			// For text output, show a summary and first few fields
			_, _ = fmt.Fprintf(io.Out, "Found %d result(s)\n\n", len(results))
			for i, result := range results {
				_, _ = fmt.Fprintf(io.Out, "--- Result %d ---\n", i+1)
				keys := make([]string, 0, len(result))
				for k := range result {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					_, _ = fmt.Fprintf(io.Out, "  %s: %v\n", k, result[k])
				}
				_, _ = fmt.Fprintf(io.Out, "\n")
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&filters, "filter", "f", nil, "Filter expression (can be repeated)")
	cmd.Flags().StringArrayVarP(&joins, "join", "j", nil, "Join related resource (can be repeated)")
	cmd.Flags().StringVar(&sortField, "sort", "", "Sort by field")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum results to return (0 = unlimited)")
	cmd.Flags().IntVar(&start, "start", 0, "Starting offset for pagination")
	cmd.Flags().BoolVar(&failEmpty, "fail-empty", false, "Exit 4 when results are empty (JSON mode)")

	return cmd
}

func newResourceGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <ResourceName> <id>",
		Short: "Get a specific resource by ID",
		Args:  RequireArgs("ResourceName", "id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceName := args[0]
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid ID: %s", args[1])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.Resource(resourceName).Get(cmd.Context(), id)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(result)
			}

			io := iocontext.FromContext(cmd.Context())
			keys := make([]string, 0, len(result))
			for k := range result {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				_, _ = fmt.Fprintf(io.Out, "%s: %v\n", k, result[k])
			}

			return nil
		},
	}
}

// parseFilters converts filter expressions like "field=value", "field>value" into Deputy's search format.
// Deputy expects: { "f1": { "field": "FieldName", "type": "eq", "data": "value" }, ... }
func parseFilters(filters []string) (map[string]interface{}, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	search := make(map[string]interface{})

	// Operators in order of precedence (check longer operators first)
	operators := []struct {
		op       string
		deputyOp string
	}{
		{">=", "ge"},
		{"<=", "le"},
		{">", "gt"},
		{"<", "lt"},
		{"=", "eq"},
	}

	for i, filter := range filters {
		var field, value, deputyOp string
		found := false

		for _, op := range operators {
			if idx := strings.Index(filter, op.op); idx > 0 {
				field = filter[:idx]
				value = filter[idx+len(op.op):]
				deputyOp = op.deputyOp
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("invalid filter syntax: %s (expected field=value, field>value, etc.)", filter)
		}

		filterKey := fmt.Sprintf("f%d", i+1)
		search[filterKey] = map[string]interface{}{
			"field": field,
			"type":  deputyOp,
			"data":  value,
		}
	}

	return search, nil
}
