package main

import (
	"os"
	"strings"

	"github.com/buddyh/todoist-cli/internal/output"
	"github.com/spf13/cobra"
)

func newTasksCmd(flags *rootFlags) *cobra.Command {
	var (
		today   bool
		filter  string
		project string
		overdue bool
		all     bool
	)

	cmd := &cobra.Command{
		Use:     "tasks",
		Aliases: []string{"list", "ls"},
		Short:   "List tasks",
		Long: `List tasks with optional filters.

Examples:
  todoist tasks              # Today's tasks (default)
  todoist tasks --all        # All active tasks
  todoist tasks --filter "p1"       # High priority
  todoist tasks --filter "overdue"  # Overdue tasks
  todoist tasks -p Work      # Tasks in Work project
  todoist tasks --overdue    # Shortcut for overdue filter`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasks(cmd, flags, today, filter, project)
		},
	}

	cmd.Flags().BoolVarP(&today, "today", "t", true, "show today's tasks (including overdue)")
	cmd.Flags().StringVarP(&filter, "filter", "f", "", "Todoist filter string")
	cmd.Flags().StringVarP(&project, "project", "p", "", "filter by project name")
	cmd.Flags().BoolVar(&overdue, "overdue", false, "show only overdue tasks")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "show all active tasks")

	return cmd
}

func runTasks(cmd *cobra.Command, flags *rootFlags, today bool, filter, project string) error {
	out := output.NewFormatter(os.Stdout, flags.asJSON)

	client, err := getClient()
	if err != nil {
		return err
	}

	// Determine project ID if project name given
	var projectID string
	if project != "" {
		p, err := client.FindProject(project)
		if err != nil {
			return err
		}
		projectID = p.ID
	}

	// Build filter
	if filter == "" {
		// If project is specified and no explicit filter flags were set,
		// default to all active tasks in that project.
		if project != "" && cmd.Flag("today") != nil && !cmd.Flag("today").Changed &&
			cmd.Flag("overdue") != nil && !cmd.Flag("overdue").Changed &&
			cmd.Flag("all") != nil && !cmd.Flag("all").Changed {
			today = false
		}

		// Check flags
		if cmd.Flag("overdue") != nil && cmd.Flag("overdue").Changed {
			filter = "overdue"
		} else if cmd.Flag("all") != nil && cmd.Flag("all").Changed {
			// No filter - get all tasks
			filter = ""
		} else if today {
			filter = "today | overdue"
		}
	}

	tasks, err := client.GetTasks(projectID, filter)
	if err != nil {
		return err
	}

	return out.WriteTasks(tasks)
}

// Helper to check if a string contains another (case-insensitive)
func containsCI(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
