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

func newWebhooksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "webhooks",
		Aliases: []string{"webhook", "wh"},
		Short:   "Manage webhooks",
	}

	cmd.AddCommand(newWebhooksListCmd())
	cmd.AddCommand(newWebhooksGetCmd())
	cmd.AddCommand(newWebhooksAddCmd())
	cmd.AddCommand(newWebhooksDeleteCmd())

	return cmd
}

func newWebhooksListCmd() *cobra.Command {
	var limit, offset int
	var failEmpty bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all webhooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			opts := &api.ListOptions{Limit: limit, Offset: offset}
			webhooks, err := client.Webhooks().List(cmd.Context(), opts)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				ctx := outfmt.WithLimit(cmd.Context(), limit)
				ctx = outfmt.WithOffset(ctx, offset)
				ctx = outfmt.WithFailEmpty(ctx, failEmpty)
				f := outfmt.New(ctx)
				return f.OutputList(webhooks)
			}

			f := outfmt.New(cmd.Context())
			f.StartTable([]string{"ID", "TOPIC", "URL", "ENABLED"})
			for _, w := range webhooks {
				enabled := "Yes"
				if !w.Enabled {
					enabled = "No"
				}
				f.Row(
					strconv.Itoa(w.Id),
					w.Topic,
					w.Url,
					enabled,
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

func newWebhooksGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get webhook details",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid webhook ID: %s", args[0])
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			webhook, err := client.Webhooks().Get(cmd.Context(), id)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(webhook)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "ID:      %d\n", webhook.Id)
			_, _ = fmt.Fprintf(io.Out, "Topic:   %s\n", webhook.Topic)
			_, _ = fmt.Fprintf(io.Out, "URL:     %s\n", webhook.Url)
			_, _ = fmt.Fprintf(io.Out, "Type:    %s\n", webhook.Type)
			_, _ = fmt.Fprintf(io.Out, "Enabled: %t\n", webhook.Enabled)
			return nil
		},
	}
}

func newWebhooksAddCmd() *cobra.Command {
	var topic, url, webhookType string
	var enabled bool

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new webhook",
		Long: `Add a new webhook subscription.

Topics use the format {Resource}.{Action}. Valid actions are:
  Insert  - Triggered when a new record is added
  Update  - Triggered only when a record is changed
  Save    - Triggered every time a record is saved (even without changes)
  Delete  - Triggered when a record is deleted

Common topic examples:
  Employee.Insert, Employee.Update, Employee.Save
  Timesheet.Insert, Timesheet.Update, Timesheet.Save
  Roster.Insert, Roster.Update, Roster.Publish
  Leave.Insert, Leave.Update

Additional resources that support webhooks:
  Comment, Company, EmployeeAvailability, Memo, OperationalUnit,
  RosterOpen, RosterSwap, Task, TimesheetPayReturn, TrainingRecord

Special topics:
  User.Login              - Triggered when a user logs in
  TimesheetExport.Begin   - Triggered when timesheet export begins
  TimesheetExport.End     - Triggered when timesheet export completes
  Device.Registration     - Triggered when a new device is registered

Webhook types:
  URL   - HTTP/HTTPS webhook (default)
  SQS   - AWS SQS queue
  SLACK - Slack notification`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if topic == "" {
				return errors.New("--topic is required")
			}
			if url == "" {
				return errors.New("--url is required")
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			input := &api.CreateWebhookInput{
				Topic:   topic,
				Url:     url,
				Type:    webhookType,
				Enabled: enabled,
			}

			webhook, err := client.Webhooks().Create(cmd.Context(), input)
			if err != nil {
				return err
			}

			format := outfmt.GetFormat(cmd.Context())
			if format == "json" {
				f := outfmt.New(cmd.Context())
				return f.Output(webhook)
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Created webhook %d for topic %s\n", webhook.Id, webhook.Topic)
			return nil
		},
	}

	cmd.Flags().StringVar(&topic, "topic", "", "Webhook topic (required)")
	cmd.Flags().StringVar(&url, "url", "", "Webhook URL (required)")
	cmd.Flags().StringVar(&webhookType, "type", "", "Webhook type")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable webhook")

	return cmd
}

func newWebhooksDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a webhook",
		Args:  RequireArg("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid webhook ID: %s", args[0])
			}

			if err := confirmDestructive(cmd.Context(), yes, fmt.Sprintf("Are you sure you want to delete webhook %d?", id)); err != nil {
				return err
			}

			client, err := getClientFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.Webhooks().Delete(cmd.Context(), id); err != nil {
				return err
			}

			io := iocontext.FromContext(cmd.Context())
			_, _ = fmt.Fprintf(io.Out, "Deleted webhook %d\n", id)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
