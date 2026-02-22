package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
)

func newPayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pay",
		Aliases: []string{"payroll", "rates"},
		Short:   "Manage pay rates and agreements",
	}

	cmd.AddCommand(newPayAwardsCmd())
	cmd.AddCommand(newPayAgreementsCmd())

	return cmd
}

func newPayAwardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "awards",
		Short: "Manage award library pay rates",
	}

	cmd.AddCommand(newPayAwardsListCmd())
	cmd.AddCommand(newPayAwardsGetCmd())
	cmd.AddCommand(newPayAwardsSetCmd())

	return cmd
}

func newPayAwardsListCmd() *cobra.Command {
	var limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List awards from the pay rate library",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			awards, err := client.PayRates().ListAwardsLibrary(cmd.Context())
			if err != nil {
				return err
			}

			awards = applyPagination(awards, offset, limit)

			ctx := cmd.Context()
			if limit > 0 {
				ctx = outfmt.WithLimit(ctx, limit)
			}
			if offset > 0 {
				ctx = outfmt.WithOffset(ctx, offset)
			}
			ctx = outfmt.WithFailEmpty(ctx, failEmpty)

			format := outfmt.GetFormat(ctx)
			if format == "json" {
				f := outfmt.New(ctx)
				return f.OutputList(awards)
			}

			f := outfmt.New(ctx)
			f.StartTable([]string{"CODE", "NAME", "COUNTRY"})
			for _, award := range awards {
				code := stringFromMap(award, "AwardCode", "Code", "Id")
				name := stringFromMap(award, "Name", "AwardName", "Description")
				country := stringFromMap(award, "CountryCode", "Country")
				f.Row(code, name, country)
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

func newPayAwardsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <award-code>",
		Short: "Get details for an award from the library",
		Args:  RequireArg("award-code"),
		RunE: func(cmd *cobra.Command, args []string) error {
			awardCode := args[0]

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			award, err := client.PayRates().GetAwardFromLibrary(cmd.Context(), awardCode)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(award)
			}

			io := iocontext.FromContext(cmd.Context())
			keys := make([]string, 0, len(award))
			for k := range award {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				_, _ = fmt.Fprintf(io.Out, "%s: %v\n", k, award[k])
			}
			return nil
		},
	}
}

func newPayAwardsSetCmd() *cobra.Command {
	var awardCode string
	var countryCode string
	var overrides []string

	cmd := &cobra.Command{
		Use:   "set <employee-id>",
		Short: "Assign an award from the library to an employee",
		Long: `Assign an award from the pay rate library to an employee.

Overrides are provided as payRuleId:hourlyRate (repeatable).`,
		Args: RequireArg("employee-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			employeeID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid employee ID: %s", args[0])
			}
			if awardCode == "" {
				return errors.New("--award is required")
			}
			if countryCode == "" {
				return errors.New("--country is required")
			}

			overrideRules := make([]api.OverridePayRule, 0, len(overrides))
			for _, raw := range overrides {
				override, err := parseOverridePayRule(raw)
				if err != nil {
					return err
				}
				overrideRules = append(overrideRules, override)
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.SetAwardFromLibraryInput{
				CountryCode:     countryCode,
				AwardCode:       awardCode,
				OverridePayRule: overrideRules,
			}

			result, err := client.PayRates().SetAwardFromLibrary(cmd.Context(), employeeID, input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(result)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Assigned award %s to employee %d\n", awardCode, employeeID)
			return nil
		},
	}

	cmd.Flags().StringVar(&awardCode, "award", "", "Award code (required)")
	cmd.Flags().StringVar(&countryCode, "country", "", "Country code (required)")
	cmd.Flags().StringArrayVar(&overrides, "override", nil, "Override pay rule in format payRuleId:hourlyRate (repeatable)")

	return cmd
}

func newPayAgreementsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agreements",
		Short: "Manage employee agreements (base rate and area config)",
	}

	cmd.AddCommand(newPayAgreementsListCmd())
	cmd.AddCommand(newPayAgreementsGetCmd())
	cmd.AddCommand(newPayAgreementsUpdateCmd())

	return cmd
}

func newPayAgreementsListCmd() *cobra.Command {
	var employeeID int
	var activeOnly bool
	var limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List agreements for an employee",
		RunE: func(cmd *cobra.Command, args []string) error {
			if employeeID == 0 {
				return errors.New("--employee is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			agreements, err := client.Agreements().ListByEmployee(cmd.Context(), employeeID, activeOnly)
			if err != nil {
				return err
			}

			agreements = applyPagination(agreements, offset, limit)

			ctx := cmd.Context()
			if limit > 0 {
				ctx = outfmt.WithLimit(ctx, limit)
			}
			if offset > 0 {
				ctx = outfmt.WithOffset(ctx, offset)
			}
			ctx = outfmt.WithFailEmpty(ctx, failEmpty)

			format := outfmt.GetFormat(ctx)
			if format == "json" {
				f := outfmt.New(ctx)
				return f.OutputList(agreements)
			}

			f := outfmt.New(ctx)
			f.StartTable([]string{"ID", "EMPLOYEE", "ACTIVE", "BASE RATE"})
			for _, agreement := range agreements {
				baseRate := ""
				if agreement.BaseRate != nil {
					baseRate = fmt.Sprintf("%.2f", *agreement.BaseRate)
				}
				active := "No"
				if agreement.Active {
					active = "Yes"
				}
				f.Row(
					strconv.Itoa(agreement.Id),
					strconv.Itoa(agreement.Employee),
					active,
					baseRate,
				)
			}
			f.EndTable()
			return nil
		},
	}

	cmd.Flags().IntVar(&employeeID, "employee", 0, "Employee ID (required)")
	cmd.Flags().BoolVar(&activeOnly, "active-only", false, "Only show active agreements")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (0 = unlimited)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&failEmpty, "fail-empty", false, "Exit 4 when results are empty (JSON mode)")

	return cmd
}

func newPayAgreementsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <agreement-id>",
		Short: "Get agreement details",
		Args:  RequireArg("agreement-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			agreementID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid agreement ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			agreement, err := client.Agreements().Get(cmd.Context(), agreementID)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(agreement)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "ID:        %d\n", agreement.Id)
			_, _ = fmt.Fprintf(io.Out, "Employee:  %d\n", agreement.Employee)
			_, _ = fmt.Fprintf(io.Out, "Active:    %t\n", agreement.Active)
			if agreement.BaseRate != nil {
				_, _ = fmt.Fprintf(io.Out, "Base Rate: %.2f\n", *agreement.BaseRate)
			}
			if len(agreement.Config) > 0 {
				_, _ = fmt.Fprintf(io.Out, "Config:    %s\n", string(agreement.Config))
			}
			return nil
		},
	}
}

func newPayAgreementsUpdateCmd() *cobra.Command {
	var baseRate float64
	var config string
	var configFile string

	cmd := &cobra.Command{
		Use:   "update <agreement-id>",
		Short: "Update agreement base rate or config",
		Long: `Update an employee agreement.

Config should be a JSON string (or provided via --config-file).`,
		Args: RequireArg("agreement-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			agreementID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid agreement ID: %s", args[0])
			}

			baseRateSet := cmd.Flags().Changed("base-rate")
			configSet := config != ""
			configFileSet := configFile != ""

			if !baseRateSet && !configSet && !configFileSet {
				return errors.New("at least one of --base-rate, --config, or --config-file is required")
			}
			if baseRateSet && baseRate <= 0 {
				return errors.New("--base-rate must be greater than 0")
			}
			if configSet && configFileSet {
				return errors.New("use either --config or --config-file, not both")
			}

			var configValue string
			if configFileSet {
				data, err := os.ReadFile(configFile)
				if err != nil {
					return fmt.Errorf("read config file: %w", err)
				}
				configValue = string(data)
			} else if configSet {
				configValue = config
			}

			var configPayload json.RawMessage
			if configValue != "" {
				if !json.Valid([]byte(configValue)) {
					return errors.New("config must be valid JSON")
				}
				// Verify it's an object, not an array or primitive
				var parsed interface{}
				if err := json.Unmarshal([]byte(configValue), &parsed); err != nil {
					return fmt.Errorf("config parse error: %w", err)
				}
				if _, isMap := parsed.(map[string]interface{}); !isMap {
					return errors.New("config must be a JSON object (not an array or primitive)")
				}
				configPayload = json.RawMessage(configValue)
			}

			input := &api.UpdateAgreementInput{}
			if baseRateSet {
				input.BaseRate = &baseRate
			}
			if configValue != "" {
				input.Config = &configPayload
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			agreement, err := client.Agreements().Update(cmd.Context(), agreementID, input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(agreement)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Updated agreement %d\n", agreement.Id)
			return nil
		},
	}

	cmd.Flags().Float64Var(&baseRate, "base-rate", 0, "Base rate (hourly) to set")
	cmd.Flags().StringVar(&config, "config", "", "Config JSON string")
	cmd.Flags().StringVar(&configFile, "config-file", "", "Config JSON file path")

	return cmd
}

func stringFromMap(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := m[key]; ok && value != nil {
			return fmt.Sprint(value)
		}
	}
	return ""
}

func parseOverridePayRule(input string) (api.OverridePayRule, error) {
	separator := ""
	if strings.Contains(input, "=") {
		separator = "="
	} else if strings.Contains(input, ":") {
		separator = ":"
	}
	if separator == "" {
		return api.OverridePayRule{}, fmt.Errorf("invalid override %q (expected payRuleId:hourlyRate)", input)
	}

	parts := strings.SplitN(input, separator, 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return api.OverridePayRule{}, fmt.Errorf("invalid override %q (expected payRuleId:hourlyRate)", input)
	}

	rate, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return api.OverridePayRule{}, fmt.Errorf("invalid override rate %q", parts[1])
	}
	if rate <= 0 {
		return api.OverridePayRule{}, fmt.Errorf("invalid override rate %q (must be greater than 0)", parts[1])
	}

	return api.OverridePayRule{Id: parts[0], HourlyRate: rate}, nil
}
