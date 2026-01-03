package main

import (
	"fmt"
	"os"

	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

func newCompleteCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete <task-id>",
		Short: "Mark a task as complete",
		Long: `Mark a task as complete by its ID.

Examples:
  todoist complete 1234567890
  todoist done 1234567890`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runComplete(flags, args[0])
		},
	}

	return cmd
}

func newDoneCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done <task-id>",
		Short: "Mark a task as complete (alias for complete)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runComplete(flags, args[0])
		},
	}

	return cmd
}

func runComplete(flags *rootFlags, taskID string) error {
	out := output.NewFormatter(os.Stdout, flags.asJSON)

	client, err := getClient()
	if err != nil {
		return err
	}

	// Get task first to show what was completed
	task, err := client.GetTask(taskID)
	if err != nil {
		return err
	}

	if err := client.CompleteTask(taskID); err != nil {
		return err
	}

	out.WriteSuccess(fmt.Sprintf("Completed: %s", task.Content))
	return nil
}
