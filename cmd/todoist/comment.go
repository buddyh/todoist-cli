package main

import (
	"os"
	"strings"

	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

func newCommentCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "comment <task-id> [message]",
		Aliases: []string{"note"},
		Short:   "Add or view comments on a task",
		Long: `Add a comment to a task, or view existing comments.

Examples:
  todoist comment 123456                    # View comments
  todoist comment 123456 "This is a note"   # Add comment`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)
			taskID := args[0]

			client, err := getClient()
			if err != nil {
				return err
			}

			if len(args) > 1 {
				// Add comment
				content := strings.Join(args[1:], " ")
				comment, err := client.AddComment(content, taskID, "")
				if err != nil {
					return err
				}

				if flags.asJSON {
					return out.JSON(comment)
				}
				out.WriteSuccess("Comment added")
				return nil
			}

			// View comments
			comments, err := client.GetComments(taskID, "")
			if err != nil {
				return err
			}

			return out.WriteComments(comments)
		},
	}

	return cmd
}
