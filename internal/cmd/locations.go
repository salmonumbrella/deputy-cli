package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/iocontext"
	"github.com/salmonumbrella/deputy-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

func newLocationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "locations",
		Aliases: []string{"location", "loc"},
		Short:   "Manage locations",
	}

	cmd.AddCommand(newLocationsListCmd())
	cmd.AddCommand(newLocationsGetCmd())
	cmd.AddCommand(newLocationsAddCmd())
	cmd.AddCommand(newLocationsUpdateCmd())
	cmd.AddCommand(newLocationsArchiveCmd())
	cmd.AddCommand(newLocationsDeleteCmd())
	cmd.AddCommand(newLocationsSettingsCmd())
	cmd.AddCommand(newLocationsSettingsUpdateCmd())

	return cmd
}

func newLocationsListCmd() *cobra.Command {
	var limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all locations",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			opts := &api.ListOptions{Limit: limit, Offset: offset}
			locations, err := client.Locations().List(cmd.Context(), opts)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				ctx := outfmt.WithLimit(cmd.Context(), limit)
				ctx = outfmt.WithOffset(ctx, offset)
				ctx = outfmt.WithFailEmpty(ctx, failEmpty)
				f := outfmt.New(ctx)
				return f.OutputList(locations)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"ID", "NAME", "CODE", "ACTIVE"})
			for _, l := range locations {
				code := l.Code
				if code == "" {
					code = l.CompanyCode
				}
				active := "Yes"
				if !l.Active {
					active = "No"
				}
				f.Row(
					strconv.Itoa(l.Id),
					l.CompanyName,
					code,
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

func newLocationsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get location details",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid location ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			location, err := client.Locations().Get(cmd.Context(), id)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(location)
			}

			code := location.Code
			if code == "" {
				code = location.CompanyCode
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "ID:       %d\n", location.Id)
			_, _ = fmt.Fprintf(io.Out, "Name:     %s\n", location.CompanyName)
			_, _ = fmt.Fprintf(io.Out, "Code:     %s\n", code)
			_, _ = fmt.Fprintf(io.Out, "Address:  %s\n", location.AddressString())
			_, _ = fmt.Fprintf(io.Out, "Timezone: %s\n", location.Timezone)
			_, _ = fmt.Fprintf(io.Out, "Active:   %t\n", location.Active)
			return nil
		},
	}
}

func newLocationsAddCmd() *cobra.Command {
	var name, code, address, timezone string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new location",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return errors.New("--name is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.CreateLocationInput{
				CompanyName: name,
				Code:        code,
				Address:     address,
				Timezone:    timezone,
			}

			location, err := client.Locations().Create(cmd.Context(), input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(location)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Created location %d: %s\n", location.Id, location.CompanyName)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Location name (required)")
	cmd.Flags().StringVar(&code, "code", "", "Location code")
	cmd.Flags().StringVar(&address, "address", "", "Address")
	cmd.Flags().StringVar(&timezone, "timezone", "", "Timezone")

	return cmd
}

func newLocationsUpdateCmd() *cobra.Command {
	var name, code, address, timezone string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a location",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid location ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.UpdateLocationInput{
				CompanyName: name,
				Code:        code,
				Address:     address,
				Timezone:    timezone,
			}

			location, err := client.Locations().Update(cmd.Context(), id, input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(location)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Updated location %d: %s\n", location.Id, location.CompanyName)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Location name")
	cmd.Flags().StringVar(&code, "code", "", "Location code")
	cmd.Flags().StringVar(&address, "address", "", "Address")
	cmd.Flags().StringVar(&timezone, "timezone", "", "Timezone")

	return cmd
}

func newLocationsArchiveCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a location",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid location ID: %s", args[0])
			}

			if err := confirmDestructive(cmd.Context(), yes, fmt.Sprintf("Are you sure you want to archive location %d?", id)); err != nil {
				return err
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.Locations().Archive(cmd.Context(), id); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Location %d archived\n", id)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func newLocationsDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a location",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid location ID: %s", args[0])
			}

			if err := confirmDestructive(cmd.Context(), yes, fmt.Sprintf("Are you sure you want to delete location %d?", id)); err != nil {
				return err
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.Locations().Delete(cmd.Context(), id); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Location %d deleted\n", id)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func newLocationsSettingsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "settings <id>",
		Short: "Get location settings",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid location ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			settings, err := client.Locations().GetSettings(cmd.Context(), id)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(settings)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Location %d Settings:\n", id)
			keys := make([]string, 0, len(settings.Settings))
			for k := range settings.Settings {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				_, _ = fmt.Fprintf(io.Out, "  %s: %v\n", k, settings.Settings[k])
			}
			return nil
		},
	}
}

func newLocationsSettingsUpdateCmd() *cobra.Command {
	var settingsJSON string

	cmd := &cobra.Command{
		Use:   "settings-update <id>",
		Short: "Update location settings",
		Long: `Update location settings with a JSON object.

Example:
  deputy locations settings-update 123 --settings '{"key": "value"}'`,
		Args: RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid location ID: %s", args[0])
			}
			if settingsJSON == "" {
				return errors.New("--settings is required")
			}

			var settings map[string]interface{}
			if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.Locations().UpdateSettings(cmd.Context(), id, settings); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Updated settings for location %d\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&settingsJSON, "settings", "", "Settings as JSON object (required)")

	return cmd
}
