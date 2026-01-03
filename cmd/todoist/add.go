package main

import (
	"os"
	"strings"

	"github.com/buddyh/todoist-cli/internal/api"
	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

func newAddCmd(flags *rootFlags) *cobra.Command {
	var (
		description string
		due         string
		priority    int
		project     string
		section     string
		labels      []string
	)

	cmd := &cobra.Command{
		Use:   "add <task content>",
		Short: "Create a new task",
		Long: `Create a new task with optional parameters.

Examples:
  todoist add "Buy groceries"
  todoist add "Call mom" -d tomorrow
  todoist add "Urgent task" -P 1 -d "today 5pm"
  todoist add "Work task" -p Work -l urgent -l followup
  todoist add "Meeting prep" --description "Prepare slides for Q4 review"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)
			content := strings.Join(args, " ")

			client, err := getClient()
			if err != nil {
				return err
			}

			params := api.AddTaskParams{
				Content:     content,
				Description: description,
				DueString:   due,
				Labels:      labels,
			}

			// Convert priority (user: 1=highest, API: 4=highest)
			if priority > 0 {
				params.Priority = 5 - priority
			}

			// Find project ID if name given
			if project != "" {
				p, err := client.FindProject(project)
				if err != nil {
					return err
				}
				params.ProjectID = p.ID
			}

			// Find section ID if name given
			if section != "" && params.ProjectID != "" {
				sections, err := client.GetSections(params.ProjectID)
				if err != nil {
					return err
				}
				for _, s := range sections {
					if containsCI(s.Name, section) {
						params.SectionID = s.ID
						break
					}
				}
			}

			task, err := client.AddTask(params)
			if err != nil {
				return err
			}

			return out.WriteTask(task)
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "task description/notes")
	cmd.Flags().StringVarP(&due, "due", "d", "", "due date (e.g., 'tomorrow', 'next monday 3pm')")
	cmd.Flags().IntVarP(&priority, "priority", "P", 0, "priority 1-4 (1=highest)")
	cmd.Flags().StringVarP(&project, "project", "p", "", "project name")
	cmd.Flags().StringVarP(&section, "section", "s", "", "section name (requires project)")
	cmd.Flags().StringArrayVarP(&labels, "label", "l", nil, "add label (can be repeated)")

	return cmd
}
