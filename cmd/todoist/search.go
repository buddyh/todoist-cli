package main

import (
	"os"
	"strings"

	"github.com/buddyh/todoist-cli/internal/api"
	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

func newSearchCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search tasks by content",
		Long: `Search tasks by content (case-insensitive).

Examples:
  todoist search "meeting"
  todoist search "buy"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)
			query := strings.ToLower(args[0])

			client, err := getClient()
			if err != nil {
				return err
			}

			// Get all tasks and filter locally
			tasks, err := client.GetTasks("", "")
			if err != nil {
				return err
			}

			var matches []api.Task
			for _, t := range tasks {
				if strings.Contains(strings.ToLower(t.Content), query) ||
					strings.Contains(strings.ToLower(t.Description), query) {
					matches = append(matches, t)
				}
			}

			return out.WriteTasks(matches)
		},
	}

	return cmd
}
