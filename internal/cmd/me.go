package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
)

func newMeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Commands for the current user",
	}

	cmd.AddCommand(newMeInfoCmd())
	cmd.AddCommand(newMeTimesheetsCmd())
	cmd.AddCommand(newMeRostersCmd())
	cmd.AddCommand(newMeLeaveCmd())

	return cmd
}

func newMeInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show current user info",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			info, err := client.Me().Info(cmd.Context())
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(info)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "User ID:      %d\n", info.UserId)
			_, _ = fmt.Fprintf(io.Out, "Employee ID:  %d\n", info.EmployeeId)
			_, _ = fmt.Fprintf(io.Out, "Login:        %s\n", info.Login)
			_, _ = fmt.Fprintf(io.Out, "Name:         %s\n", info.Name)
			_, _ = fmt.Fprintf(io.Out, "First Name:   %s\n", info.FirstName)
			_, _ = fmt.Fprintf(io.Out, "Last Name:    %s\n", info.LastName)
			_, _ = fmt.Fprintf(io.Out, "Email:        %s\n", info.PrimaryEmail)
			_, _ = fmt.Fprintf(io.Out, "Phone:        %s\n", info.PrimaryPhone)
			_, _ = fmt.Fprintf(io.Out, "Company:      %d\n", info.Company)
			_, _ = fmt.Fprintf(io.Out, "Portfolio:    %s\n", info.Portfolio)
			return nil
		},
	}
}

func newMeTimesheetsCmd() *cobra.Command {
	var limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "timesheets",
		Short: "List my timesheets",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			timesheets, err := client.Me().Timesheets(cmd.Context())
			if err != nil {
				return err
			}

			// Apply client-side pagination
			timesheets = applyPagination(timesheets, offset, limit)

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				ctx := outfmt.WithLimit(cmd.Context(), limit)
				ctx = outfmt.WithOffset(ctx, offset)
				ctx = outfmt.WithFailEmpty(ctx, failEmpty)
				f := outfmt.New(ctx)
				return f.OutputList(timesheets)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"ID", "DATE", "START", "END", "TOTAL"})
			for _, t := range timesheets {
				start := time.Unix(t.StartTime, 0).Format("15:04")
				end := "-"
				if t.EndTime > 0 {
					end = time.Unix(t.EndTime, 0).Format("15:04")
				}
				f.Row(
					strconv.Itoa(t.Id),
					t.Date,
					start,
					end,
					t.TotalTimeStr,
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

func newMeRostersCmd() *cobra.Command {
	var limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "rosters",
		Short: "List my rosters",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			rosters, err := client.Me().Rosters(cmd.Context())
			if err != nil {
				return err
			}

			// Apply client-side pagination
			rosters = applyPagination(rosters, offset, limit)

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				ctx := outfmt.WithLimit(cmd.Context(), limit)
				ctx = outfmt.WithOffset(ctx, offset)
				ctx = outfmt.WithFailEmpty(ctx, failEmpty)
				f := outfmt.New(ctx)
				return f.OutputList(rosters)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"ID", "DATE", "START", "END"})
			for _, r := range rosters {
				start := time.Unix(r.StartTime, 0).Format("15:04")
				end := time.Unix(r.EndTime, 0).Format("15:04")
				f.Row(
					strconv.Itoa(r.Id),
					r.Date,
					start,
					end,
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

func newMeLeaveCmd() *cobra.Command {
	var limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "leave",
		Short: "List my leave requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			leaves, err := client.Me().Leave(cmd.Context())
			if err != nil {
				return err
			}

			// Apply client-side pagination
			leaves = applyPagination(leaves, offset, limit)

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				ctx := outfmt.WithLimit(cmd.Context(), limit)
				ctx = outfmt.WithOffset(ctx, offset)
				ctx = outfmt.WithFailEmpty(ctx, failEmpty)
				f := outfmt.New(ctx)
				return f.OutputList(leaves)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"ID", "START", "END", "STATUS", "HOURS"})
			for _, l := range leaves {
				f.Row(
					strconv.Itoa(l.Id),
					l.DateStart,
					l.DateEnd,
					leaveStatusText(l.Status),
					fmt.Sprintf("%.1f", l.Hours),
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
