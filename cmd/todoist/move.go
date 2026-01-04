package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/buddyh/todoist-cli/internal/api"
	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

func newMoveCmd(flags *rootFlags) *cobra.Command {
	var (
		section string
		project string
	)

	cmd := &cobra.Command{
		Use:   "move <task-id>",
		Short: "Move a task to a different section or project",
		Long: `Move a task to a different section or project.

This is useful for Kanban-style workflows where tasks move between sections.

Examples:
  todoist move 123 --section "In Progress"
  todoist move 123 -s "Done"
  todoist move 123 --project "Work"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.NewFormatter(os.Stdout, flags.asJSON)
			taskID := args[0]

			if section == "" && project == "" {
				return fmt.Errorf("must specify either --section or --project")
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			var sectionID, projectID string

			if section != "" {
				// Need to find section by name - first get the task to know its project
				task, err := client.GetTask(taskID)
				if err != nil {
					return fmt.Errorf("failed to get task: %w", err)
				}

				sections, err := client.GetSections(task.ProjectID)
				if err != nil {
					return fmt.Errorf("failed to get sections: %w", err)
				}

				sectionID = findSectionID(sections, section)
				if sectionID == "" {
					return fmt.Errorf("section not found: %s", section)
				}
			}

			if project != "" {
				p, err := client.FindProject(project)
				if err != nil {
					return err
				}
				projectID = p.ID
			}

			if err := client.MoveTask(taskID, sectionID, projectID); err != nil {
				return err
			}

			if section != "" {
				out.WriteSuccess(fmt.Sprintf("Moved task to section: %s", section))
			} else {
				out.WriteSuccess(fmt.Sprintf("Moved task to project: %s", project))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&section, "section", "s", "", "target section name")
	cmd.Flags().StringVarP(&project, "project", "p", "", "target project name")

	return cmd
}

// findSectionID finds a section by name (case-insensitive partial match)
func findSectionID(sections []api.Section, name string) string {
	nameLower := strings.ToLower(name)
	for _, s := range sections {
		if strings.Contains(strings.ToLower(s.Name), nameLower) {
			return s.ID
		}
	}
	return ""
}
