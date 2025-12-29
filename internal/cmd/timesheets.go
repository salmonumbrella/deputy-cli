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
		Aliases: []string{"timesheet", "ts"},
		Short:   "Manage timesheets",
	}

	cmd.AddCommand(newTimesheetsListCmd())
	cmd.AddCommand(newTimesheetsGetCmd())
	cmd.AddCommand(newTimesheetsClockInCmd())
	cmd.AddCommand(newTimesheetsClockOutCmd())
	cmd.AddCommand(newTimesheetsStartBreakCmd())
	cmd.AddCommand(newTimesheetsEndBreakCmd())

	return cmd
}

func newTimesheetsListCmd() *cobra.Command {
	var limit, offset int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List my timesheets",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			opts := &api.ListOptions{Limit: limit, Offset: offset}
			timesheets, err := client.Timesheets().List(cmd.Context(), opts)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(timesheets)
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
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (0 = unlimited)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")

	return cmd
}

func newTimesheetsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get timesheet details",
		Args:  cobra.ExactArgs(1),
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
