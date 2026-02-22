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

func newManagementCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "management",
		Short: "Memos and journals",
	}

	cmd.AddCommand(newMemoCmd())
	cmd.AddCommand(newJournalCmd())

	return cmd
}

func newMemoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memo",
		Short: "Manage memos",
	}

	cmd.AddCommand(newMemoListCmd())
	cmd.AddCommand(newMemoAddCmd())

	return cmd
}

func newMemoListCmd() *cobra.Command {
	var companyID int
	var limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List memos",
		RunE: func(cmd *cobra.Command, args []string) error {
			if companyID == 0 {
				return errors.New("--company is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			memos, err := client.Management().ListMemos(cmd.Context(), companyID)
			if err != nil {
				return err
			}

			// Apply client-side pagination
			memos = applyPagination(memos, offset, limit)

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				ctx := outfmt.WithLimit(cmd.Context(), limit)
				ctx = outfmt.WithOffset(ctx, offset)
				ctx = outfmt.WithFailEmpty(ctx, failEmpty)
				f := outfmt.New(ctx)
				return f.OutputList(memos)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"ID", "CREATED", "CONTENT"})
			for _, m := range memos {
				created := time.Unix(m.Created, 0).Format("2006-01-02")
				content := m.Content
				if len(content) > 50 {
					content = content[:50] + "..."
				}
				f.Row(
					strconv.Itoa(m.Id),
					created,
					content,
				)
			}
			f.EndTable()
			return nil
		},
	}

	cmd.Flags().IntVar(&companyID, "company", 0, "Company ID (required)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (0 = unlimited)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&failEmpty, "fail-empty", false, "Exit 4 when results are empty (JSON mode)")

	return cmd
}

func newMemoAddCmd() *cobra.Command {
	var companyID int
	var content string
	var locations []int
	var employees []int

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Create a memo",
		RunE: func(cmd *cobra.Command, args []string) error {
			if companyID == 0 {
				return errors.New("--company is required")
			}
			if content == "" {
				return errors.New("--content is required")
			}
			if len(locations) == 0 && len(employees) == 0 {
				return errors.New("at least one --location or --employee is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.CreateMemoInput{
				Content:   content,
				Company:   companyID,
				Locations: locations,
				Employees: employees,
			}

			memo, err := client.Management().CreateMemo(cmd.Context(), input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(memo)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Created memo %d\n", memo.Id)
			return nil
		},
	}

	cmd.Flags().IntVar(&companyID, "company", 0, "Company ID (required)")
	cmd.Flags().StringVar(&content, "content", "", "Memo content (required)")
	cmd.Flags().IntSliceVar(&locations, "location", nil, "Location IDs to target (can be repeated)")
	cmd.Flags().IntSliceVar(&employees, "employee", nil, "Employee IDs to target (can be repeated)")

	return cmd
}

func newJournalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "journal",
		Short: "Manage journals",
	}

	cmd.AddCommand(newJournalListCmd())
	cmd.AddCommand(newJournalAddCmd())

	return cmd
}

func newJournalListCmd() *cobra.Command {
	var employeeID int
	var limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List journal entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			if employeeID == 0 {
				return errors.New("--employee is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			journals, err := client.Management().ListJournals(cmd.Context(), employeeID)
			if err != nil {
				return err
			}

			// Apply client-side pagination
			journals = applyPagination(journals, offset, limit)

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				ctx := outfmt.WithLimit(cmd.Context(), limit)
				ctx = outfmt.WithOffset(ctx, offset)
				ctx = outfmt.WithFailEmpty(ctx, failEmpty)
				f := outfmt.New(ctx)
				return f.OutputList(journals)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"ID", "CREATED", "COMMENT"})
			for _, j := range journals {
				created := time.Unix(j.Created, 0).Format("2006-01-02")
				comment := j.Comment
				if len(comment) > 50 {
					comment = comment[:50] + "..."
				}
				f.Row(
					strconv.Itoa(j.Id),
					created,
					comment,
				)
			}
			f.EndTable()
			return nil
		},
	}

	cmd.Flags().IntVar(&employeeID, "employee", 0, "Employee ID (required)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (0 = unlimited)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&failEmpty, "fail-empty", false, "Exit 4 when results are empty (JSON mode)")

	return cmd
}

func newJournalAddCmd() *cobra.Command {
	var employeeID, companyID, category int
	var comment string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Post a journal entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			if employeeID == 0 {
				return errors.New("--employee is required")
			}
			if companyID == 0 {
				return errors.New("--company is required")
			}
			if comment == "" {
				return errors.New("--comment is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.CreateJournalInput{
				Employee: employeeID,
				Company:  companyID,
				Comment:  comment,
				Category: category,
			}

			journal, err := client.Management().PostJournal(cmd.Context(), input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(journal)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Posted journal %d for employee %d\n", journal.Id, journal.Employee)
			return nil
		},
	}

	cmd.Flags().IntVar(&employeeID, "employee", 0, "Employee ID (required)")
	cmd.Flags().IntVar(&companyID, "company", 0, "Company ID (required)")
	cmd.Flags().StringVar(&comment, "comment", "", "Journal comment (required)")
	cmd.Flags().IntVar(&category, "category", 0, "Category ID")

	return cmd
}
