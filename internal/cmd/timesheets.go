package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
)

func newTimesheetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "timesheets",
		Aliases: []string{"timesheet", "ts", "t"},
		Short:   "Manage timesheets",
		Long: `Manage timesheets including viewing, clocking in/out, and pay rules.

Pay Rule Commands:
  list-pay-rules    List available pay rules (filter by --hourly-rate)
  select-pay-rule   Assign a pay rule to an approved timesheet

Example workflow for setting pay rates:
  # Find pay rules with $190/hr rate
  deputy timesheets list-pay-rules --hourly-rate 190

  # Assign pay rule 304 to timesheet 19379
  deputy timesheets select-pay-rule 19379 --pay-rule 304`,
	}

	cmd.AddCommand(newTimesheetsListCmd())
	cmd.AddCommand(newTimesheetsGetCmd())
	cmd.AddCommand(newTimesheetsUpdateCmd())
	cmd.AddCommand(newTimesheetsListPayRulesCmd())
	cmd.AddCommand(newTimesheetsSetPayRuleCmd())
	cmd.AddCommand(newTimesheetsClockInCmd())
	cmd.AddCommand(newTimesheetsClockOutCmd())
	cmd.AddCommand(newTimesheetsStartBreakCmd())
	cmd.AddCommand(newTimesheetsEndBreakCmd())

	return cmd
}

func newTimesheetsListCmd() *cobra.Command {
	var limit, offset int
	var failEmpty bool
	var fromDate string
	var toDate string
	var employeeID int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List my timesheets",
		Long: `List timesheets for the authenticated user.

By default, this returns only your own timesheets via the /my/timesheets endpoint.
Use --employee to query a specific employee's timesheets (uses a different API endpoint
and requires permission). Use --from/--to to filter by date (YYYY-MM-DD).`,
		Example: `  deputy timesheets list --from 2024-01-01 --to 2024-01-31
  deputy timesheets list --employee 123 --from 2024-01-01 --to 2024-01-31 -o json -q '.items[].Id'
  deputy timesheets list --employee 123 --from 2024-01-01 --to 2024-01-31 --raw`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			from, hasFrom, err := parseDateFlag(fromDate, "--from")
			if err != nil {
				return err
			}
			to, hasTo, err := parseDateFlag(toDate, "--to")
			if err != nil {
				return err
			}
			if hasFrom && hasTo && from.After(to) {
				return fmt.Errorf("--from must be on or before --to")
			}

			if employeeID != 0 {
				filters := []string{fmt.Sprintf("Employee=%d", employeeID)}
				if hasFrom {
					filters = append(filters, "Date>="+fromDate)
				}
				if hasTo {
					filters = append(filters, "Date<="+toDate)
				}

				search, err := parseFilters(filters)
				if err != nil {
					return err
				}

				input := &api.QueryInput{
					Search: search,
					Max:    limit,
					Start:  offset,
				}

				timesheets, err := client.Timesheets().Query(cmd.Context(), input)
				if err != nil {
					return err
				}

				return outputTimesheets(cmd, timesheets, limit, offset, failEmpty)
			}

			opts := &api.ListOptions{Limit: limit, Offset: offset}
			timesheets, err := client.Timesheets().List(cmd.Context(), opts)
			if err != nil {
				return err
			}

			if hasFrom || hasTo {
				timesheets, err = filterTimesheetsByDate(timesheets, from, to, hasFrom, hasTo)
				if err != nil {
					return err
				}
			}

			return outputTimesheets(cmd, timesheets, limit, offset, failEmpty)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (0 = unlimited)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&failEmpty, "fail-empty", false, "Exit 4 when results are empty (JSON mode)")
	cmd.Flags().StringVar(&fromDate, "from", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&toDate, "to", "", "End date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&employeeID, "employee", 0, "Filter by employee ID (uses resource query)")

	return cmd
}

func newTimesheetsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get timesheet details",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid timesheet ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			timesheet, err := client.Timesheets().Get(cmd.Context(), id)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(timesheet)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "ID:         %d\n", timesheet.Id)
			_, _ = fmt.Fprintf(io.Out, "Employee:   %d\n", timesheet.Employee)
			_, _ = fmt.Fprintf(io.Out, "Date:       %s\n", timesheet.Date)
			_, _ = fmt.Fprintf(io.Out, "Start:      %s\n", time.Unix(timesheet.StartTime, 0).Format("15:04"))
			if timesheet.EndTime > 0 {
				_, _ = fmt.Fprintf(io.Out, "End:        %s\n", time.Unix(timesheet.EndTime, 0).Format("15:04"))
			}
			_, _ = fmt.Fprintf(io.Out, "Total:      %s\n", timesheet.TotalTimeStr)
			_, _ = fmt.Fprintf(io.Out, "Mealbreak:  %s\n", timesheet.Mealbreak)
			_, _ = fmt.Fprintf(io.Out, "In Progress: %t\n", timesheet.IsInProgress)
			return nil
		},
	}
}

func newTimesheetsUpdateCmd() *cobra.Command {
	var cost float64

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a timesheet",
		Long: `Update a timesheet's properties.

Currently supports updating the cost/pay amount for a timesheet.`,
		Args: RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid timesheet ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.UpdateTimesheetInput{}
			if cmd.Flags().Changed("cost") {
				input.Cost = &cost
			}

			timesheet, err := client.Timesheets().Update(cmd.Context(), id, input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(timesheet)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Updated timesheet %d (cost: %.2f)\n", timesheet.Id, timesheet.Cost)
			return nil
		},
	}

	cmd.Flags().Float64Var(&cost, "cost", 0, "Total cost/pay amount for the timesheet")

	return cmd
}

func newTimesheetsListPayRulesCmd() *cobra.Command {
	var hourlyRate float64

	cmd := &cobra.Command{
		Use:   "list-pay-rules",
		Short: "List available pay rules",
		Long: `List pay rules available in Deputy.

Use --hourly-rate to filter by a specific rate.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			var ratePtr *float64
			if cmd.Flags().Changed("hourly-rate") {
				ratePtr = &hourlyRate
			}

			rules, err := client.Timesheets().ListPayRules(cmd.Context(), ratePtr)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.OutputList(rules)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"ID", "TITLE", "HOURLY RATE"})
			for _, r := range rules {
				f.Row(
					strconv.Itoa(r.Id),
					r.PayTitle,
					fmt.Sprintf("%.2f", r.HourlyRate),
				)
			}
			f.EndTable()
			return nil
		},
	}

	cmd.Flags().Float64Var(&hourlyRate, "hourly-rate", 0, "Filter by hourly rate")

	return cmd
}

func newTimesheetsSetPayRuleCmd() *cobra.Command {
	var payRuleID int

	cmd := &cobra.Command{
		Use:     "select-pay-rule <timesheet-id>",
		Aliases: []string{"set-pay-rule"},
		Short:   "Select a pay rule for a timesheet",
		Long: `Select a pay rule for an approved timesheet.

Use list-pay-rules to find available pay rules, then assign one to a timesheet.
The cost is automatically calculated as: hourly_rate × hours.

Example:
  # Find 190/hr pay rules
  deputy timesheets list-pay-rules --hourly-rate 190

  # Assign pay rule 304 to timesheet 19379
  deputy timesheets select-pay-rule 19379 --pay-rule 304`,
		Args: RequireArg("timesheet-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			timesheetID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid timesheet ID: %s", args[0])
			}

			if payRuleID == 0 {
				return errors.New("--pay-rule is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.Timesheets().SetPayRule(cmd.Context(), timesheetID, payRuleID)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(result)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Assigned pay rule %d to timesheet %d (total: $%.2f)\n", result.PayRule, result.Timesheet, result.Cost)
			return nil
		},
	}

	cmd.Flags().IntVar(&payRuleID, "pay-rule", 0, "Pay rule ID (required)")

	return cmd
}

func parseDateFlag(value, flagName string) (time.Time, bool, error) {
	if value == "" {
		return time.Time{}, false, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("invalid %s date %q (expected YYYY-MM-DD)", flagName, value)
	}
	return parsed, true, nil
}

func filterTimesheetsByDate(timesheets []api.Timesheet, from, to time.Time, hasFrom, hasTo bool) ([]api.Timesheet, error) {
	if !hasFrom && !hasTo {
		return timesheets, nil
	}

	filtered := make([]api.Timesheet, 0, len(timesheets))
	for _, t := range timesheets {
		if t.Date == "" {
			continue
		}
		parsed, err := time.Parse("2006-01-02", t.Date)
		if err != nil {
			return nil, fmt.Errorf("timesheet %d has invalid Date %q", t.Id, t.Date)
		}
		if hasFrom && parsed.Before(from) {
			continue
		}
		if hasTo && parsed.After(to) {
			continue
		}
		filtered = append(filtered, t)
	}

	return filtered, nil
}

func outputTimesheets(cmd *cobra.Command, timesheets []api.Timesheet, limit, offset int, failEmpty bool) error {
	format := outfmt.GetFormat(cmd.Context())
	if format == "json" {
		ctx := outfmt.WithLimit(cmd.Context(), limit)
		ctx = outfmt.WithOffset(ctx, offset)
		ctx = outfmt.WithFailEmpty(ctx, failEmpty)
		f := outfmt.New(ctx)
		return f.OutputList(timesheets)
	}

	f := outfmt.New(cmd.Context())
	f.StartTable([]string{"ID", "DATE", "START", "END", "TOTAL", "STATUS"})
	for _, t := range timesheets {
		start := time.Unix(t.StartTime, 0).Format("15:04")
		end := "-"
		status := "In Progress"
		if t.EndTime > 0 {
			end = time.Unix(t.EndTime, 0).Format("15:04")
			status = "Complete"
		}
		f.Row(
			strconv.Itoa(t.Id),
			t.Date,
			start,
			end,
			t.TotalTimeStr,
			status,
		)
	}
	f.EndTable()
	return nil
}

func newTimesheetsClockInCmd() *cobra.Command {
	var employeeID, opunitID int
	var comment string

	cmd := &cobra.Command{
		Use:   "clock-in",
		Short: "Clock in an employee",
		RunE: func(cmd *cobra.Command, args []string) error {
			if employeeID == 0 {
				return errors.New("--employee is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.ClockInput{
				Employee:        employeeID,
				OperationalUnit: opunitID,
				Comment:         comment,
			}

			resp, err := client.Timesheets().ClockIn(cmd.Context(), input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(resp)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Clocked in employee %d (timesheet %d)\n", resp.Employee, resp.Id)
			return nil
		},
	}

	cmd.Flags().IntVar(&employeeID, "employee", 0, "Employee ID (required)")
	cmd.Flags().IntVar(&opunitID, "opunit", 0, "Operational unit ID")
	cmd.Flags().StringVar(&comment, "comment", "", "Comment")

	return cmd
}

func newTimesheetsClockOutCmd() *cobra.Command {
	var timesheetID, employeeID int
	var comment string

	cmd := &cobra.Command{
		Use:   "clock-out",
		Short: "Clock out an employee",
		Long: `Clock out an employee by ending their active timesheet.

Use --timesheet to specify the timesheet ID directly (preferred method).
Use --employee as a fallback if the API supports stopping by employee ID.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if timesheetID == 0 && employeeID == 0 {
				return errors.New("--timesheet or --employee is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.ClockInput{
				Timesheet: timesheetID,
				Employee:  employeeID,
				Comment:   comment,
			}

			resp, err := client.Timesheets().ClockOut(cmd.Context(), input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(resp)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Clocked out employee %d (timesheet %d)\n", resp.Employee, resp.Id)
			return nil
		},
	}

	cmd.Flags().IntVarP(&timesheetID, "timesheet", "t", 0, "Timesheet ID (preferred)")
	cmd.Flags().IntVarP(&employeeID, "employee", "e", 0, "Employee ID (fallback)")
	cmd.Flags().StringVar(&comment, "comment", "", "Comment")

	return cmd
}

func newTimesheetsStartBreakCmd() *cobra.Command {
	var timesheetID, employeeID int

	cmd := &cobra.Command{
		Use:   "start-break",
		Short: "Start break for employee",
		Long: `Start a break on an active timesheet.

Use --timesheet to specify the timesheet ID directly (preferred method).
Use --employee as a fallback if the API supports pausing by employee ID.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if timesheetID == 0 && employeeID == 0 {
				return errors.New("--timesheet or --employee is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.ClockInput{
				Timesheet: timesheetID,
				Employee:  employeeID,
			}

			if err := client.Timesheets().StartBreak(cmd.Context(), input); err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				result := map[string]any{"status": "break_started"}
				if timesheetID != 0 {
					result["timesheet"] = timesheetID
				}
				if employeeID != 0 {
					result["employee"] = employeeID
				}
				return f.Output(result)
			}

			io := iocontext.FromContext(cmd.Context())
			if timesheetID != 0 {
				_, _ = fmt.Fprintf(io.Out, "Break started on timesheet %d\n", timesheetID)
			} else {
				_, _ = fmt.Fprintf(io.Out, "Break started for employee %d\n", employeeID)
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&timesheetID, "timesheet", "t", 0, "Timesheet ID (preferred)")
	cmd.Flags().IntVarP(&employeeID, "employee", "e", 0, "Employee ID (fallback)")

	return cmd
}

func newTimesheetsEndBreakCmd() *cobra.Command {
	var timesheetID, employeeID int

	cmd := &cobra.Command{
		Use:   "end-break",
		Short: "End break for employee",
		Long: `End a break on an active timesheet.

Use --timesheet to specify the timesheet ID directly (preferred method).
Use --employee as a fallback if the API supports resuming by employee ID.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if timesheetID == 0 && employeeID == 0 {
				return errors.New("--timesheet or --employee is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.ClockInput{
				Timesheet: timesheetID,
				Employee:  employeeID,
			}

			if err := client.Timesheets().EndBreak(cmd.Context(), input); err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				result := map[string]any{"status": "break_ended"}
				if timesheetID != 0 {
					result["timesheet"] = timesheetID
				}
				if employeeID != 0 {
					result["employee"] = employeeID
				}
				return f.Output(result)
			}

			io := iocontext.FromContext(cmd.Context())
			if timesheetID != 0 {
				_, _ = fmt.Fprintf(io.Out, "Break ended on timesheet %d\n", timesheetID)
			} else {
				_, _ = fmt.Fprintf(io.Out, "Break ended for employee %d\n", employeeID)
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&timesheetID, "timesheet", "t", 0, "Timesheet ID (preferred)")
	cmd.Flags().IntVarP(&employeeID, "employee", "e", 0, "Employee ID (fallback)")

	return cmd
}
