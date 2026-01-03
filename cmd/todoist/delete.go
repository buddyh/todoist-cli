package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

func newDeleteCmd(flags *rootFlags) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "delete <task-id>",
		Aliases: []string{"rm", "remove"},
		Short:   "Delete a task permanently",
		Long: `Delete a task permanently by its ID.

This action cannot be undone. Use 'todoist complete' to mark as done instead.

Examples:
  todoist delete 1234567890
  todoist delete 1234567890 --force  # Skip confirmation`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)
			taskID := args[0]

			client, err := getClient()
			if err != nil {
				return err
			}

			// Get task first to show what will be deleted
			task, err := client.GetTask(taskID)
			if err != nil {
				return err
			}

			// Confirm unless force flag
			if !force && !flags.asJSON {
				fmt.Printf("Delete task: %s\nThis cannot be undone. Continue? [y/N] ", task.Content)
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				if strings.ToLower(strings.TrimSpace(input)) != "y" {
					out.WriteSuccess("Cancelled")
					return nil
				}
			}

			if err := client.DeleteTask(taskID); err != nil {
				return err
			}

			out.WriteSuccess(fmt.Sprintf("Deleted: %s", task.Content))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation")

	return cmd
}
