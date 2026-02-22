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

func newSalesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sales",
		Aliases: []string{"metrics"},
		Short:   "Manage sales data",
	}

	cmd.AddCommand(newSalesListCmd())
	cmd.AddCommand(newSalesAddCmd())

	return cmd
}

func newSalesListCmd() *cobra.Command {
	var companyID int
	var limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sales data",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			var sales []api.SalesData
			if companyID > 0 {
				input := &api.SalesQueryInput{
					Company: companyID,
				}
				sales, err = client.Sales().Query(cmd.Context(), input)
			} else {
				sales, err = client.Sales().List(cmd.Context())
			}
			if err != nil {
				return err
			}

			// Apply client-side pagination
			sales = applyPagination(sales, offset, limit)

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				ctx := outfmt.WithLimit(cmd.Context(), limit)
				ctx = outfmt.WithOffset(ctx, offset)
				ctx = outfmt.WithFailEmpty(ctx, failEmpty)
				f := outfmt.New(ctx)
				return f.OutputList(sales)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"ID", "COMPANY", "TIMESTAMP", "VALUE", "TYPE"})
			for _, s := range sales {
				timestamp := time.Unix(s.Timestamp, 0).Format(time.RFC3339)
				f.Row(
					strconv.Itoa(s.Id),
					strconv.Itoa(s.Company),
					timestamp,
					fmt.Sprintf("%.2f", s.Value),
					s.Type,
				)
			}
			f.EndTable()
			return nil
		},
	}

	cmd.Flags().IntVar(&companyID, "company", 0, "Filter by company ID")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (0 = unlimited)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&failEmpty, "fail-empty", false, "Exit 4 when results are empty (JSON mode)")

	return cmd
}

func newSalesAddCmd() *cobra.Command {
	var companyID, areaID int
	var timestamp int64
	var value float64
	var salesType string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add sales data",
		RunE: func(cmd *cobra.Command, args []string) error {
			if companyID == 0 {
				return errors.New("--company is required")
			}
			if timestamp == 0 {
				return errors.New("--timestamp is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.CreateSalesInput{
				Company:   companyID,
				Area:      areaID,
				Timestamp: timestamp,
				Value:     value,
				Type:      salesType,
			}

			sale, err := client.Sales().Add(cmd.Context(), input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(sale)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Created sales data %d for company %d\n", sale.Id, sale.Company)
			return nil
		},
	}

	cmd.Flags().IntVar(&companyID, "company", 0, "Company ID (required)")
	cmd.Flags().IntVar(&areaID, "area", 0, "Area ID")
	cmd.Flags().Int64Var(&timestamp, "timestamp", 0, "Unix timestamp (required)")
	cmd.Flags().Float64Var(&value, "value", 0, "Sales value")
	cmd.Flags().StringVar(&salesType, "type", "", "Sales type")

	return cmd
}
