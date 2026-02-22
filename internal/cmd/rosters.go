package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

func newRostersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rosters",
		Aliases: []string{"roster", "shifts", "shift", "r"},
		Short:   "Manage rosters/shifts",
	}

	cmd.AddCommand(newRostersListCmd())
	cmd.AddCommand(newRostersGetCmd())
	cmd.AddCommand(newRostersCreateCmd())
	cmd.AddCommand(newRostersCopyCmd())
	cmd.AddCommand(newRostersPublishCmd())
	cmd.AddCommand(newRostersDiscardCmd())
	cmd.AddCommand(newRostersSwapCmd())

	return cmd
}

func newRostersListCmd() *cobra.Command {
	var limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List rosters (last 12h + next 36h)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			opts := &api.ListOptions{Limit: limit, Offset: offset}
			rosters, err := client.Rosters().List(cmd.Context(), opts)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				ctx := outfmt.WithLimit(cmd.Context(), limit)
				ctx = outfmt.WithOffset(ctx, offset)
				ctx = outfmt.WithFailEmpty(ctx, failEmpty)
				f := outfmt.New(ctx)
				return f.OutputList(rosters)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"ID", "DATE", "START", "END", "EMPLOYEE", "PUBLISHED"})
			for _, r := range rosters {
				start := time.Unix(r.StartTime, 0).Format("15:04")
				end := time.Unix(r.EndTime, 0).Format("15:04")
				published := "No"
				if r.Published {
					published = "Yes"
				}
				f.Row(
					strconv.Itoa(r.Id),
					r.Date,
					start,
					end,
					strconv.Itoa(r.Employee),
					published,
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

func newRostersGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get roster details",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid roster ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			roster, err := client.Rosters().Get(cmd.Context(), id)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(roster)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "ID:         %d\n", roster.Id)
			_, _ = fmt.Fprintf(io.Out, "Date:       %s\n", roster.Date)
			_, _ = fmt.Fprintf(io.Out, "Start:      %s\n", time.Unix(roster.StartTime, 0).Format("15:04"))
			_, _ = fmt.Fprintf(io.Out, "End:        %s\n", time.Unix(roster.EndTime, 0).Format("15:04"))
			_, _ = fmt.Fprintf(io.Out, "Employee:   %d\n", roster.Employee)
			_, _ = fmt.Fprintf(io.Out, "OpUnit:     %d\n", roster.OperationalUnit)
			_, _ = fmt.Fprintf(io.Out, "Published:  %t\n", roster.Published)
			_, _ = fmt.Fprintf(io.Out, "Open:       %t\n", roster.Open)
			return nil
		},
	}
}

func newRostersCreateCmd() *cobra.Command {
	var employeeID, opunitID int
	var startTime, endTime int64
	var mealbreak, comment string
	var open, publish bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new roster/shift",
		RunE: func(cmd *cobra.Command, args []string) error {
			if employeeID == 0 {
				return errors.New("--employee is required")
			}
			if opunitID == 0 {
				return errors.New("--opunit is required")
			}
			if startTime == 0 || endTime == 0 {
				return errors.New("--start-time and --end-time are required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.CreateRosterInput{
				Employee:        employeeID,
				OperationalUnit: opunitID,
				StartTime:       startTime,
				EndTime:         endTime,
				Mealbreak:       mealbreak,
				Comment:         comment,
				Open:            open,
				Publish:         publish,
			}

			roster, err := client.Rosters().Create(cmd.Context(), input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(roster)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Created roster %d\n", roster.Id)
			return nil
		},
	}

	cmd.Flags().IntVar(&employeeID, "employee", 0, "Employee ID (required)")
	cmd.Flags().IntVar(&opunitID, "opunit", 0, "Operational unit ID (required)")
	cmd.Flags().Int64Var(&startTime, "start-time", 0, "Start time (Unix timestamp, required)")
	cmd.Flags().Int64Var(&endTime, "end-time", 0, "End time (Unix timestamp, required)")
	cmd.Flags().StringVar(&mealbreak, "mealbreak", "", "Mealbreak duration")
	cmd.Flags().StringVar(&comment, "comment", "", "Comment")
	cmd.Flags().BoolVar(&open, "open", false, "Create as open shift")
	cmd.Flags().BoolVar(&publish, "publish", false, "Publish immediately")

	return cmd
}

func newRostersCopyCmd() *cobra.Command {
	var fromDate, toDate string
	var locationID int

	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy roster from one week to another",
		RunE: func(cmd *cobra.Command, args []string) error {
			if fromDate == "" || toDate == "" {
				return errors.New("--from-date and --to-date are required")
			}
			if locationID == 0 {
				return errors.New("--location is required")
			}
			if err := validateDateFormat(fromDate); err != nil {
				return err
			}
			if err := validateDateFormat(toDate); err != nil {
				return err
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.CopyRosterInput{
				FromDate: fromDate,
				ToDate:   toDate,
				Location: locationID,
			}

			if err := client.Rosters().Copy(cmd.Context(), input); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Roster copied from %s to %s\n", fromDate, toDate)
			return nil
		},
	}

	cmd.Flags().StringVar(&fromDate, "from-date", "", "Source date (YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&toDate, "to-date", "", "Target date (YYYY-MM-DD, required)")
	cmd.Flags().IntVar(&locationID, "location", 0, "Location ID (required)")

	return cmd
}

func newRostersPublishCmd() *cobra.Command {
	var fromDate, toDate string
	var locationID int

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish rosters for a date range",
		RunE: func(cmd *cobra.Command, args []string) error {
			if fromDate == "" || toDate == "" {
				return errors.New("--from-date and --to-date are required")
			}
			if locationID == 0 {
				return errors.New("--location is required")
			}
			if err := validateDateFormat(fromDate); err != nil {
				return err
			}
			if err := validateDateFormat(toDate); err != nil {
				return err
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.PublishRosterInput{
				FromDate: fromDate,
				ToDate:   toDate,
				Location: locationID,
			}

			if err := client.Rosters().Publish(cmd.Context(), input); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Rosters published from %s to %s\n", fromDate, toDate)
			return nil
		},
	}

	cmd.Flags().StringVar(&fromDate, "from-date", "", "Start date (YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&toDate, "to-date", "", "End date (YYYY-MM-DD, required)")
	cmd.Flags().IntVar(&locationID, "location", 0, "Location ID (required)")

	return cmd
}

func newRostersDiscardCmd() *cobra.Command {
	var fromDate, toDate string
	var locationID int

	cmd := &cobra.Command{
		Use:   "discard",
		Short: "Discard unpublished roster changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if fromDate == "" || toDate == "" {
				return errors.New("--from-date and --to-date are required")
			}
			if locationID == 0 {
				return errors.New("--location is required")
			}
			if err := validateDateFormat(fromDate); err != nil {
				return err
			}
			if err := validateDateFormat(toDate); err != nil {
				return err
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.PublishRosterInput{
				FromDate: fromDate,
				ToDate:   toDate,
				Location: locationID,
			}

			if err := client.Rosters().Discard(cmd.Context(), input); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Roster changes discarded from %s to %s\n", fromDate, toDate)
			return nil
		},
	}

	cmd.Flags().StringVar(&fromDate, "from-date", "", "Start date (YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&toDate, "to-date", "", "End date (YYYY-MM-DD, required)")
	cmd.Flags().IntVar(&locationID, "location", 0, "Location ID (required)")

	return cmd
}

func newRostersSwapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "swap <roster-id>",
		Short: "List rosters available for swap",
		Long: `List rosters that can be swapped with the specified roster.

This command queries the Deputy API for shift swap candidates. It does not
perform the actual swap - that must be done through the Deputy web interface
or by using the resource API to update roster assignments.

Example:
  deputy rosters swap 12345`,
		Args: RequireArg("roster-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid roster ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			rosters, err := client.Rosters().GetSwappable(cmd.Context(), id)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.OutputList(rosters)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"ID", "DATE", "START", "END", "EMPLOYEE"})
			for _, r := range rosters {
				start := time.Unix(r.StartTime, 0).Format("15:04")
				end := time.Unix(r.EndTime, 0).Format("15:04")
				f.Row(
					strconv.Itoa(r.Id),
					r.Date,
					start,
					end,
					strconv.Itoa(r.Employee),
				)
			}
			f.EndTable()
			return nil
		},
	}
}
