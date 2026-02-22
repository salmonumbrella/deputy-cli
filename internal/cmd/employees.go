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

func newEmployeesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "employees",
		Aliases: []string{"employee", "emp", "e"},
		Short:   "Manage employees",
	}

	cmd.AddCommand(newEmployeesListCmd())
	cmd.AddCommand(newEmployeesGetCmd())
	cmd.AddCommand(newEmployeesAddCmd())
	cmd.AddCommand(newEmployeesUpdateCmd())
	cmd.AddCommand(newEmployeesTerminateCmd())
	cmd.AddCommand(newEmployeesInviteCmd())
	cmd.AddCommand(newEmployeesAssignLocationCmd())
	cmd.AddCommand(newEmployeesRemoveLocationCmd())
	cmd.AddCommand(newEmployeesReactivateCmd())
	cmd.AddCommand(newEmployeesDeleteCmd())
	cmd.AddCommand(newEmployeesAddUnavailabilityCmd())

	return cmd
}

func newEmployeesListCmd() *cobra.Command {
	var limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all employees",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			opts := &api.ListOptions{Limit: limit, Offset: offset}
			employees, err := client.Employees().List(cmd.Context(), opts)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				ctx := outfmt.WithLimit(cmd.Context(), limit)
				ctx = outfmt.WithOffset(ctx, offset)
				ctx = outfmt.WithFailEmpty(ctx, failEmpty)
				f := outfmt.New(ctx)
				return f.OutputList(employees)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"ID", "NAME", "EMAIL", "ACTIVE"})
			for _, e := range employees {
				active := "Yes"
				if !e.Active {
					active = "No"
				}
				f.Row(
					strconv.Itoa(e.Id),
					e.DisplayName,
					e.Email,
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

func newEmployeesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get employee details",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid employee ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			employee, err := client.Employees().Get(cmd.Context(), id)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(employee)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "ID:         %d\n", employee.Id)
			_, _ = fmt.Fprintf(io.Out, "Name:       %s\n", employee.DisplayName)
			_, _ = fmt.Fprintf(io.Out, "First Name: %s\n", employee.FirstName)
			_, _ = fmt.Fprintf(io.Out, "Last Name:  %s\n", employee.LastName)
			_, _ = fmt.Fprintf(io.Out, "Email:      %s\n", employee.Email)
			_, _ = fmt.Fprintf(io.Out, "Mobile:     %s\n", employee.Mobile)
			_, _ = fmt.Fprintf(io.Out, "Active:     %t\n", employee.Active)
			_, _ = fmt.Fprintf(io.Out, "Company:    %d\n", employee.Company)
			return nil
		},
	}
}

func newEmployeesAddCmd() *cobra.Command {
	var firstName, lastName, email, mobile, startDate string
	var company, role int

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new employee",
		RunE: func(cmd *cobra.Command, args []string) error {
			if firstName == "" || lastName == "" {
				return errors.New("--first-name and --last-name are required")
			}
			if company == 0 {
				return errors.New("--company is required")
			}
			if startDate != "" {
				if err := validateDateFormat(startDate); err != nil {
					return err
				}
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.CreateEmployeeInput{
				FirstName: firstName,
				LastName:  lastName,
				Email:     email,
				Mobile:    mobile,
				StartDate: startDate,
				Company:   company,
				Role:      role,
			}

			employee, err := client.Employees().Create(cmd.Context(), input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(employee)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Created employee %d: %s\n", employee.Id, employee.DisplayName)
			return nil
		},
	}

	cmd.Flags().StringVar(&firstName, "first-name", "", "First name (required)")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name (required)")
	cmd.Flags().StringVar(&email, "email", "", "Email address")
	cmd.Flags().StringVar(&mobile, "mobile", "", "Mobile phone")
	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&company, "company", 0, "Company/location ID (required)")
	cmd.Flags().IntVar(&role, "role", 0, "Role ID")

	return cmd
}

func newEmployeesUpdateCmd() *cobra.Command {
	var firstName, lastName, email, mobile string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an employee",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid employee ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.UpdateEmployeeInput{
				FirstName: firstName,
				LastName:  lastName,
				Email:     email,
				Mobile:    mobile,
			}

			employee, err := client.Employees().Update(cmd.Context(), id, input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(employee)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Updated employee %d: %s\n", employee.Id, employee.DisplayName)
			return nil
		},
	}

	cmd.Flags().StringVar(&firstName, "first-name", "", "First name")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name")
	cmd.Flags().StringVar(&email, "email", "", "Email address")
	cmd.Flags().StringVar(&mobile, "mobile", "", "Mobile phone")

	return cmd
}

func newEmployeesTerminateCmd() *cobra.Command {
	var date string
	var yes bool

	cmd := &cobra.Command{
		Use:   "terminate <id>",
		Short: "Terminate an employee",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid employee ID: %s", args[0])
			}

			if date == "" {
				return errors.New("--date is required")
			}
			if err := validateDateFormat(date); err != nil {
				return err
			}

			if err := confirmDestructive(cmd.Context(), yes, fmt.Sprintf("Are you sure you want to terminate employee %d?", id)); err != nil {
				return err
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.Employees().Terminate(cmd.Context(), id, date); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Employee %d terminated as of %s\n", id, date)
			return nil
		},
	}

	cmd.Flags().StringVar(&date, "date", "", "Termination date (YYYY-MM-DD) (required)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func newEmployeesInviteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "invite <id>",
		Short: "Send invitation to employee",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid employee ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.Employees().Invite(cmd.Context(), id); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Invitation sent to employee %d\n", id)
			return nil
		},
	}
}

func newEmployeesAssignLocationCmd() *cobra.Command {
	var locationID int

	cmd := &cobra.Command{
		Use:   "assign-location <employee-id>",
		Short: "Assign employee to a location",
		Args:  RequireArg("employee-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			employeeID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid employee ID: %s", args[0])
			}
			if locationID == 0 {
				return errors.New("--location is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.Employees().AssignLocation(cmd.Context(), employeeID, locationID); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Employee %d assigned to location %d\n", employeeID, locationID)
			return nil
		},
	}

	cmd.Flags().IntVar(&locationID, "location", 0, "Location ID (required)")

	return cmd
}

func newEmployeesRemoveLocationCmd() *cobra.Command {
	var locationID int

	cmd := &cobra.Command{
		Use:   "remove-location <employee-id>",
		Short: "Remove employee from a location",
		Args:  RequireArg("employee-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			employeeID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid employee ID: %s", args[0])
			}
			if locationID == 0 {
				return errors.New("--location is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.Employees().RemoveLocation(cmd.Context(), employeeID, locationID); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Employee %d removed from location %d\n", employeeID, locationID)
			return nil
		},
	}

	cmd.Flags().IntVar(&locationID, "location", 0, "Location ID (required)")

	return cmd
}

func newEmployeesReactivateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reactivate <id>",
		Short: "Reactivate a terminated employee",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid employee ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.Employees().Reactivate(cmd.Context(), id); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Employee %d reactivated\n", id)
			return nil
		},
	}
}

func newEmployeesDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an employee account",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid employee ID: %s", args[0])
			}

			if err := confirmDestructive(cmd.Context(), yes, fmt.Sprintf("Are you sure you want to delete employee %d?", id)); err != nil {
				return err
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.Employees().Delete(cmd.Context(), id); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Employee %d deleted\n", id)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func newEmployeesAddUnavailabilityCmd() *cobra.Command {
	var startDate, endDate, comment string

	cmd := &cobra.Command{
		Use:   "add-unavailability <employee-id>",
		Short: "Add unavailability for employee",
		Args:  RequireArg("employee-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			employeeID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid employee ID: %s", args[0])
			}
			if startDate == "" || endDate == "" {
				return errors.New("--start-date and --end-date are required")
			}
			if err := validateDateFormat(startDate); err != nil {
				return err
			}
			if err := validateDateFormat(endDate); err != nil {
				return err
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.CreateUnavailabilityInput{
				Employee:  employeeID,
				DateStart: startDate,
				DateEnd:   endDate,
				Comment:   comment,
			}

			unavail, err := client.Employees().AddUnavailability(cmd.Context(), input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(unavail)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Added unavailability %d for employee %d\n", unavail.Id, employeeID)
			return nil
		},
	}

	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date YYYY-MM-DD (required)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date YYYY-MM-DD (required)")
	cmd.Flags().StringVar(&comment, "comment", "", "Comment")

	return cmd
}
