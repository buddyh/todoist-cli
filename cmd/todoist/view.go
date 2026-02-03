package main

import (
	"fmt"
	"os"

	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

func newViewCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "view <task-id>",
		Aliases: []string{"show", "get"},
		Short:   "View a single task in detail",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)
			taskID := args[0]

			client, err := getClient()
			if err != nil {
				return err
			}

			task, err := client.GetTask(taskID)
			if err != nil {
				return err
			}

			if flags.asJSON {
				return out.JSON(task)
			}

			// Detailed human output
			fmt.Printf("ID:       %s\n", task.ID)
			fmt.Printf("Content:  %s\n", task.Content)
			if task.Description != "" {
				fmt.Printf("Notes:    %s\n", task.Description)
			}
			if task.Due != nil {
				dueStr := task.Due.String
				if dueStr == "" {
					dueStr = task.Due.Date
				}
				fmt.Printf("Due:      %s\n", dueStr)
			}
			if task.Priority > 1 {
				fmt.Printf("Priority: p%d\n", 5-task.Priority)
			}
			if len(task.Labels) > 0 {
				fmt.Printf("Labels:   @%s\n", joinLabels(task.Labels))
			}
			fmt.Printf("URL:      %s\n", task.URL)

			// Show comments if any
			comments, err := client.GetComments(taskID, "")
			if err == nil && len(comments) > 0 {
				fmt.Printf("\nComments (%d):\n", len(comments))
				for _, c := range comments {
					date := c.PostedAt
					if len(date) >= 10 {
						date = date[:10]
					}
					fmt.Printf("  [%s] %s\n", date, c.Content)
				}
			}

			return nil
		},
	}

	return cmd
}

func joinLabels(labels []string) string {
	if len(labels) == 0 {
		return ""
	}
	result := labels[0]
	for _, l := range labels[1:] {
		result += " @" + l
	}
	return result
}
