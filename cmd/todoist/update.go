package main

import (
	"os"

	"github.com/buddyh/todoist-cli/internal/api"
	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

func newUpdateCmd(flags *rootFlags) *cobra.Command {
	var (
		content     string
		description string
		due         string
		priority    int
		labels      []string
	)

	cmd := &cobra.Command{
		Use:     "update <task-id>",
		Aliases: []string{"edit", "modify"},
		Short:   "Update a task",
		Long: `Update an existing task by its ID.

Examples:
  todoist update 123 --content "New title"
  todoist update 123 --due "tomorrow"
  todoist update 123 -P 1
  todoist update 123 --labels "urgent,important"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)
			taskID := args[0]

			client, err := getClient()
			if err != nil {
				return err
			}

			params := api.UpdateTaskParams{
				Content:     content,
				Description: description,
				DueString:   due,
			}

			// Convert priority
			if priority > 0 {
				params.Priority = 5 - priority
			}

			// Parse labels
			if len(labels) > 0 {
				params.Labels = labels
			}

			task, err := client.UpdateTask(taskID, params)
			if err != nil {
				return err
			}

			return out.WriteTask(task)
		},
	}

	cmd.Flags().StringVar(&content, "content", "", "new task content")
	cmd.Flags().StringVar(&description, "description", "", "new description")
	cmd.Flags().StringVarP(&due, "due", "d", "", "new due date")
	cmd.Flags().IntVarP(&priority, "priority", "P", 0, "new priority 1-4")
	cmd.Flags().StringSliceVarP(&labels, "labels", "l", nil, "replace labels (comma-separated)")

	return cmd
}

func newReopenCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reopen <task-id>",
		Short: "Reopen a completed task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)
			taskID := args[0]

			client, err := getClient()
			if err != nil {
				return err
			}

			if err := client.ReopenTask(taskID); err != nil {
				return err
			}

			out.WriteSuccess("Task reopened")
			return nil
		},
	}

	return cmd
}
