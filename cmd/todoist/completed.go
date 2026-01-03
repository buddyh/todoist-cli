package main

import (
	"os"

	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

func newCompletedCmd(flags *rootFlags) *cobra.Command {
	var (
		project string
		since   string
		until   string
		limit   int
	)

	cmd := &cobra.Command{
		Use:     "completed",
		Aliases: []string{"history"},
		Short:   "Show completed tasks",
		Long: `Show recently completed tasks.

Examples:
  todoist completed
  todoist completed --limit 20
  todoist completed --since 2024-01-01
  todoist completed -p Work`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)

			client, err := getClient()
			if err != nil {
				return err
			}

			var projectID string
			if project != "" {
				p, err := client.FindProject(project)
				if err != nil {
					return err
				}
				projectID = p.ID
			}

			resp, err := client.GetCompletedTasks(projectID, since, until, limit)
			if err != nil {
				return err
			}

			return out.WriteCompletedTasks(resp)
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "filter by project")
	cmd.Flags().StringVar(&since, "since", "", "start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&until, "until", "", "end date (YYYY-MM-DD)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 30, "max results")

	return cmd
}
