package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
)

func newDepartmentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "departments",
		Aliases: []string{"department", "dept", "opunit", "d", "areas", "area"},
		Short:   "Manage departments (operational units)",
	}

	cmd.AddCommand(newDepartmentsListCmd())
	cmd.AddCommand(newDepartmentsGetCmd())
	cmd.AddCommand(newDepartmentsAddCmd())
	cmd.AddCommand(newDepartmentsUpdateCmd())
	cmd.AddCommand(newDepartmentsDeleteCmd())

	return cmd
}

func newDepartmentsListCmd() *cobra.Command {
	var limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all departments",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			opts := &api.ListOptions{Limit: limit, Offset: offset}
			departments, err := client.Departments().List(cmd.Context(), opts)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				ctx := outfmt.WithLimit(cmd.Context(), limit)
				ctx = outfmt.WithOffset(ctx, offset)
				ctx = outfmt.WithFailEmpty(ctx, failEmpty)
				f := outfmt.New(ctx)
				return f.OutputList(departments)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"ID", "NAME", "CODE", "COMPANY", "ACTIVE"})
			for _, d := range departments {
				active := "Yes"
				if !d.Active {
					active = "No"
				}
				f.Row(
					strconv.Itoa(d.Id),
					d.CompanyName,
					d.CompanyCode,
					strconv.Itoa(d.Company),
					active,
				)
			}
			f.EndTable()
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (0 = unlimited)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&failEmpty, "fail-empty", false, "Exit 4 when results are empty (JSON mode)")

	return cmd
}

func newDepartmentsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get department details",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid department ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			department, err := client.Departments().Get(cmd.Context(), id)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(department)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "ID:         %d\n", department.Id)
			_, _ = fmt.Fprintf(io.Out, "Name:       %s\n", department.CompanyName)
			_, _ = fmt.Fprintf(io.Out, "Code:       %s\n", department.CompanyCode)
			_, _ = fmt.Fprintf(io.Out, "Company:    %d\n", department.Company)
			_, _ = fmt.Fprintf(io.Out, "Parent ID:  %d\n", department.ParentId)
			_, _ = fmt.Fprintf(io.Out, "Sort Order: %d\n", department.SortOrder)
			_, _ = fmt.Fprintf(io.Out, "Active:     %t\n", department.Active)
			return nil
		},
	}
}

func newDepartmentsAddCmd() *cobra.Command {
	var name, code string
	var company, parent, sortOrder int

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new department",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return errors.New("--name is required")
			}
			if company == 0 {
				return errors.New("--company is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.CreateDepartmentInput{
				Company:     company,
				ParentId:    parent,
				CompanyName: name,
				CompanyCode: code,
				SortOrder:   sortOrder,
			}

			department, err := client.Departments().Create(cmd.Context(), input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(department)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Created department %d: %s\n", department.Id, department.CompanyName)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Department name (required)")
	cmd.Flags().IntVar(&company, "company", 0, "Company/location ID (required)")
	cmd.Flags().StringVar(&code, "code", "", "Department code")
	cmd.Flags().IntVar(&parent, "parent", 0, "Parent department ID")
	cmd.Flags().IntVar(&sortOrder, "sort-order", 0, "Sort order")

	return cmd
}

func newDepartmentsUpdateCmd() *cobra.Command {
	var name, code string
	var sortOrder int
	var active, setActive bool

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a department",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid department ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.UpdateDepartmentInput{
				CompanyName: name,
				CompanyCode: code,
				SortOrder:   sortOrder,
			}

			// Only set Active if --set-active was explicitly provided
			if setActive {
				input.Active = &active
			}

			department, err := client.Departments().Update(cmd.Context(), id, input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(department)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Updated department %d: %s\n", department.Id, department.CompanyName)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Department name")
	cmd.Flags().StringVar(&code, "code", "", "Department code")
	cmd.Flags().IntVar(&sortOrder, "sort-order", 0, "Sort order")
	cmd.Flags().BoolVar(&active, "active", false, "Active status (use with --set-active)")
	cmd.Flags().BoolVar(&setActive, "set-active", false, "Update the active status")

	return cmd
}

func newDepartmentsDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a department",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid department ID: %s", args[0])
			}

			if err := confirmDestructive(cmd.Context(), yes, fmt.Sprintf("Are you sure you want to delete department %d?", id)); err != nil {
				return err
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.Departments().Delete(cmd.Context(), id); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Deleted department %d\n", id)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
