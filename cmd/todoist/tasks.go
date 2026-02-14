package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/buddyh/todoist-cli/internal/api"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func newTasksCmd(flags *rootFlags) *cobra.Command {
	var (
		today   bool
		filter  string
		project string
		overdue bool
		all     bool
		details bool
		sortBy  string
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
  todoist tasks --overdue    # Shortcut for overdue filter
  todoist tasks --sort priority     # Sort by priority`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasks(cmd, flags, today, filter, project, details, sortBy)
		},
	}

	cmd.Flags().BoolVarP(&today, "today", "t", true, "show today's tasks (including overdue)")
	cmd.Flags().StringVarP(&filter, "filter", "f", "", "Todoist filter string")
	cmd.Flags().StringVarP(&project, "project", "p", "", "filter by project name")
	cmd.Flags().BoolVar(&overdue, "overdue", false, "show only overdue tasks")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "show all active tasks")
	cmd.Flags().BoolVar(&details, "details", false, "show task descriptions and comments")
	cmd.Flags().StringVar(&sortBy, "sort", "", "sort tasks: priority, due, name, created")

	return cmd
}

func runTasks(cmd *cobra.Command, flags *rootFlags, today bool, filter, project string, details bool, sortBy string) error {
	out := newFormatter(flags)

	client, err := getClientWithFlags(flags)
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

	// Apply client-side sort if requested
	if sortBy != "" {
		sortTasksBy(tasks, sortBy)
	}

	if !flags.asJSON && details {
		if len(tasks) == 0 {
			fmt.Fprintln(os.Stdout, "No tasks found.")
			return nil
		}

		// Fetch comments concurrently (bounded to 5)
		type taskComments struct {
			comments []api.Comment
		}
		commentsMap := make(map[string]taskComments)
		var mu sync.Mutex

		g, ctx := errgroup.WithContext(context.Background())
		g.SetLimit(5)

		for _, t := range tasks {
			t := t
			g.Go(func() error {
				comments, err := client.GetCommentsCtx(ctx, t.ID, "")
				if err != nil {
					return err
				}
				mu.Lock()
				commentsMap[t.ID] = taskComments{comments: comments}
				mu.Unlock()
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return err
		}

		for i, t := range tasks {
			fmt.Fprintln(os.Stdout, out.FormatTaskLine(&t))
			if t.Description != "" {
				fmt.Fprintf(os.Stdout, "    %s\n", out.Color().Wrap("\033[90m", t.Description))
			}

			if tc, ok := commentsMap[t.ID]; ok && len(tc.comments) > 0 {
				fmt.Fprintf(os.Stdout, "    Comments (%d):\n", len(tc.comments))
				for _, c := range tc.comments {
					date := c.PostedAt
					if len(date) >= 10 {
						date = date[:10]
					}
					fmt.Fprintf(os.Stdout, "      [%s] %s\n", date, c.Content)
				}
			}

			if i < len(tasks)-1 {
				fmt.Fprintln(os.Stdout)
			}
		}

		return nil
	}

	return out.WriteTasks(tasks)
}

// sortTasksBy sorts tasks by the given field.
func sortTasksBy(tasks []api.Task, field string) {
	sort.Slice(tasks, func(i, j int) bool {
		switch field {
		case "priority":
			return tasks[i].Priority > tasks[j].Priority // higher priority first
		case "due":
			di := taskDueDate(tasks[i])
			dj := taskDueDate(tasks[j])
			if di == dj {
				return tasks[i].ChildOrder < tasks[j].ChildOrder
			}
			if di == "" {
				return false // no due date sorts last
			}
			if dj == "" {
				return true
			}
			return di < dj
		case "name":
			return strings.ToLower(tasks[i].Content) < strings.ToLower(tasks[j].Content)
		case "created":
			return tasks[i].CreatedAt < tasks[j].CreatedAt
		default:
			return tasks[i].ChildOrder < tasks[j].ChildOrder
		}
	})
}

func taskDueDate(t api.Task) string {
	if t.Due == nil {
		return ""
	}
	if t.Due.Datetime != "" {
		return t.Due.Datetime
	}
	return t.Due.Date
}

// Helper to check if a string contains another (case-insensitive)
func containsCI(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
