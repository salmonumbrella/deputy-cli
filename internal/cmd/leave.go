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

func newLeaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "leave",
		Aliases: []string{"leaves"},
		Short:   "Manage leave requests",
	}

	cmd.AddCommand(newLeaveListCmd())
	cmd.AddCommand(newLeaveGetCmd())
	cmd.AddCommand(newLeaveAddCmd())
	cmd.AddCommand(newLeaveApproveCmd())
	cmd.AddCommand(newLeaveDeclineCmd())

	return cmd
}

func newLeaveListCmd() *cobra.Command {
	var employeeID, limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List leave requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			var leaves []api.Leave
			if employeeID > 0 {
				input := &api.LeaveQueryInput{
					Search: map[string]interface{}{
						"Employee": employeeID,
					},
					Max:   limit,
					Start: offset,
				}
				leaves, err = client.Leave().Query(cmd.Context(), input)
			} else {
				opts := &api.ListOptions{Limit: limit, Offset: offset}
				leaves, err = client.Leave().List(cmd.Context(), opts)
			}
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				ctx := outfmt.WithLimit(cmd.Context(), limit)
				ctx = outfmt.WithOffset(ctx, offset)
				ctx = outfmt.WithFailEmpty(ctx, failEmpty)
				f := outfmt.New(ctx)
				return f.OutputList(leaves)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"ID", "EMPLOYEE", "START", "END", "DAYS", "STATUS"})
			for _, l := range leaves {
				f.Row(
					strconv.Itoa(l.Id),
					strconv.Itoa(l.Employee),
					l.DateStart,
					l.DateEnd,
					fmt.Sprintf("%.1f", l.Days),
					leaveStatusText(l.Status),
				)
			}
			f.EndTable()
			return nil
		},
	}

	cmd.Flags().IntVar(&employeeID, "employee", 0, "Filter by employee ID")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (0 = unlimited)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&failEmpty, "fail-empty", false, "Exit 4 when results are empty (JSON mode)")

	return cmd
}

func newLeaveGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get leave request details",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid leave ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			leave, err := client.Leave().Get(cmd.Context(), id)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(leave)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "ID:         %d\n", leave.Id)
			_, _ = fmt.Fprintf(io.Out, "Employee:   %d\n", leave.Employee)
			_, _ = fmt.Fprintf(io.Out, "Company:    %d\n", leave.Company)
			_, _ = fmt.Fprintf(io.Out, "Start:      %s\n", leave.DateStart)
			_, _ = fmt.Fprintf(io.Out, "End:        %s\n", leave.DateEnd)
			_, _ = fmt.Fprintf(io.Out, "Days:       %.1f\n", leave.Days)
			_, _ = fmt.Fprintf(io.Out, "Hours:      %.1f\n", leave.Hours)
			_, _ = fmt.Fprintf(io.Out, "Status:     %s\n", leaveStatusText(leave.Status))
			if leave.Comment != "" {
				_, _ = fmt.Fprintf(io.Out, "Comment:    %s\n", leave.Comment)
			}
			if leave.LeaveRule > 0 {
				_, _ = fmt.Fprintf(io.Out, "Leave Rule: %d\n", leave.LeaveRule)
			}
			return nil
		},
	}
}

func newLeaveAddCmd() *cobra.Command {
	var employeeID, leaveRule int
	var startDate, endDate, comment string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a leave request",
		RunE: func(cmd *cobra.Command, args []string) error {
			if employeeID == 0 {
				return errors.New("--employee is required")
			}
			if startDate == "" {
				return errors.New("--start-date is required")
			}
			if endDate == "" {
				return errors.New("--end-date is required")
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

			input := &api.CreateLeaveInput{
				Employee:  employeeID,
				DateStart: startDate,
				DateEnd:   endDate,
				LeaveRule: leaveRule,
				Comment:   comment,
			}

			leave, err := client.Leave().Create(cmd.Context(), input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(leave)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Created leave request %d for employee %d (%s to %s)\n",
				leave.Id, leave.Employee, leave.DateStart, leave.DateEnd)
			return nil
		},
	}

	cmd.Flags().IntVar(&employeeID, "employee", 0, "Employee ID (required)")
	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date (YYYY-MM-DD) (required)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date (YYYY-MM-DD) (required)")
	cmd.Flags().IntVar(&leaveRule, "leave-rule", 0, "Leave rule ID")
	cmd.Flags().StringVar(&comment, "comment", "", "Comment")

	return cmd
}

func newLeaveApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <id>",
		Short: "Approve a leave request",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid leave ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.Leave().Approve(cmd.Context(), id); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Leave request %d approved\n", id)
			return nil
		},
	}
}

func newLeaveDeclineCmd() *cobra.Command {
	var comment string
	var yes bool

	cmd := &cobra.Command{
		Use:   "decline <id>",
		Short: "Decline a leave request",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid leave ID: %s", args[0])
			}

			if err := confirmDestructive(cmd.Context(), yes, fmt.Sprintf("Are you sure you want to decline leave request %d?", id)); err != nil {
				return err
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.Leave().Decline(cmd.Context(), id, comment); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Leave request %d declined\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&comment, "comment", "", "Reason for declining")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func leaveStatusText(status int) string {
	switch status {
	case 0:
		return "Awaiting"
	case 1:
		return "Approved"
	case 2:
		return "Declined"
	case 3:
		return "Cancelled"
	case 4:
		return "Pay Pending"
	case 5:
		return "Pay Approved"
	default:
		return fmt.Sprintf("Unknown (%d)", status)
	}
}
