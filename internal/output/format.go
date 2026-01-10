package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/buddyh/todoist-cli/internal/api"
)

// Envelope wraps all JSON responses for consistent parsing
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *string     `json:"error,omitempty"`
}

// Formatter handles output formatting
type Formatter struct {
	w      io.Writer
	asJSON bool
}

// NewFormatter creates a new output formatter
func NewFormatter(w io.Writer, asJSON bool) *Formatter {
	return &Formatter{w: w, asJSON: asJSON}
}

// JSON outputs data wrapped in envelope
func (f *Formatter) JSON(v interface{}) error {
	env := Envelope{Success: true, Data: v}
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(f.w, string(b))
	return err
}

// WriteError outputs an error
func (f *Formatter) WriteError(err error) {
	if f.asJSON {
		msg := err.Error()
		env := Envelope{Success: false, Error: &msg}
		b, _ := json.Marshal(env)
		fmt.Fprintln(f.w, string(b))
	} else {
		fmt.Fprintf(f.w, "Error: %v\n", err)
	}
}

// WriteSuccess outputs a success message
func (f *Formatter) WriteSuccess(msg string) {
	if f.asJSON {
		f.JSON(map[string]string{"message": msg})
	} else {
		fmt.Fprintln(f.w, msg)
	}
}

// priorityString converts Todoist priority to human-readable
func priorityString(p int) string {
	switch p {
	case 4:
		return "p1"
	case 3:
		return "p2"
	case 2:
		return "p3"
	default:
		return ""
	}
}

// priorityColor returns ANSI color for priority
func priorityColor(p int) string {
	switch p {
	case 4:
		return "\033[31m" // red
	case 3:
		return "\033[33m" // yellow
	case 2:
		return "\033[34m" // blue
	default:
		return ""
	}
}

const colorReset = "\033[0m"

// FormatTask formats a single task for human output
func FormatTask(t *api.Task) string {
	var parts []string

	// Priority indicator
	pColor := priorityColor(t.Priority)
	pStr := priorityString(t.Priority)
	if pStr != "" {
		parts = append(parts, fmt.Sprintf("%s[%s]%s", pColor, pStr, colorReset))
	}

	// Task content
	parts = append(parts, t.Content)

	// Due date
	if t.Due != nil {
		dueStr := t.Due.String
		if dueStr == "" {
			dueStr = t.Due.Date
		}
		parts = append(parts, fmt.Sprintf("\033[90m(%s)\033[0m", dueStr))
	}

	// Labels
	if len(t.Labels) > 0 {
		parts = append(parts, fmt.Sprintf("\033[36m@%s\033[0m", strings.Join(t.Labels, " @")))
	}

	return strings.Join(parts, " ")
}

// FormatTaskLine formats a task as a single line with ID
func FormatTaskLine(t *api.Task) string {
	return fmt.Sprintf("\033[90m%s\033[0m  %s", t.ID, FormatTask(t))
}

// WriteTasks outputs a list of tasks
func (f *Formatter) WriteTasks(tasks []api.Task) error {
	if f.asJSON {
		return f.JSON(tasks)
	}

	if len(tasks) == 0 {
		fmt.Fprintln(f.w, "No tasks found.")
		return nil
	}

	taskMap := make(map[string]*api.Task)
	childrenMap := make(map[string][]*api.Task)

	for i := range tasks {
		t := &tasks[i]
		taskMap[t.ID] = t
		if t.ParentID != "" {
			childrenMap[t.ParentID] = append(childrenMap[t.ParentID], t)
		}
	}

	var roots []*api.Task
	for i := range tasks {
		t := &tasks[i]
		if t.ParentID == "" {
			roots = append(roots, t)
		} else if _, ok := taskMap[t.ParentID]; !ok {
			roots = append(roots, t)
		}
	}

	sortTasks(roots)
	for _, children := range childrenMap {
		sortTasks(children)
	}

	for _, root := range roots {
		f.printTaskRecursive(root, 0, childrenMap)
	}

	return nil
}

func sortTasks(tasks []*api.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Order < tasks[j].Order
	})
}

func (f *Formatter) printTaskRecursive(t *api.Task, level int, childrenMap map[string][]*api.Task) {
	indent := strings.Repeat("  ", level)
	fmt.Fprintf(f.w, "%s%s\n", indent, FormatTaskLine(t))

	if children, ok := childrenMap[t.ID]; ok {
		for _, child := range children {
			f.printTaskRecursive(child, level+1, childrenMap)
		}
	}
}

// WriteTask outputs a single task
func (f *Formatter) WriteTask(t *api.Task) error {
	if f.asJSON {
		return f.JSON(t)
	}

	fmt.Fprintln(f.w, FormatTaskLine(t))
	if t.Description != "" {
		fmt.Fprintf(f.w, "    \033[90m%s\033[0m\n", t.Description)
	}

	return nil
}

// FormatProject formats a project for human output
func FormatProject(p *api.Project) string {
	var markers []string
	if p.IsFavorite {
		markers = append(markers, "*")
	}
	if p.IsInboxProject {
		markers = append(markers, "inbox")
	}

	result := p.Name
	if len(markers) > 0 {
		result = fmt.Sprintf("%s \033[90m[%s]\033[0m", result, strings.Join(markers, ", "))
	}

	return result
}

// WriteProjects outputs a list of projects
func (f *Formatter) WriteProjects(projects []api.Project) error {
	if f.asJSON {
		return f.JSON(projects)
	}

	if len(projects) == 0 {
		fmt.Fprintln(f.w, "No projects found.")
		return nil
	}

	for _, p := range projects {
		fmt.Fprintf(f.w, "\033[90m%s\033[0m  %s\n", p.ID, FormatProject(&p))
	}

	return nil
}

// WriteProject outputs a single project
func (f *Formatter) WriteProject(p *api.Project) error {
	if f.asJSON {
		return f.JSON(p)
	}

	fmt.Fprintf(f.w, "\033[90m%s\033[0m  %s\n", p.ID, FormatProject(p))
	return nil
}

// WriteLabels outputs a list of labels
func (f *Formatter) WriteLabels(labels []api.Label) error {
	if f.asJSON {
		return f.JSON(labels)
	}

	if len(labels) == 0 {
		fmt.Fprintln(f.w, "No labels found.")
		return nil
	}

	for _, l := range labels {
		fmt.Fprintf(f.w, "\033[90m%s\033[0m  \033[36m@%s\033[0m\n", l.ID, l.Name)
	}

	return nil
}

// WriteSections outputs a list of sections
func (f *Formatter) WriteSections(sections []api.Section) error {
	if f.asJSON {
		return f.JSON(sections)
	}

	if len(sections) == 0 {
		fmt.Fprintln(f.w, "No sections found.")
		return nil
	}

	for _, s := range sections {
		fmt.Fprintf(f.w, "\033[90m%s\033[0m  %s\n", s.ID, s.Name)
	}

	return nil
}

// WriteComments outputs a list of comments
func (f *Formatter) WriteComments(comments []api.Comment) error {
	if f.asJSON {
		return f.JSON(comments)
	}

	if len(comments) == 0 {
		fmt.Fprintln(f.w, "No comments found.")
		return nil
	}

	for _, c := range comments {
		fmt.Fprintf(f.w, "\033[90m%s\033[0m  %s\n", c.PostedAt, c.Content)
	}

	return nil
}

// WriteCompletedTasks outputs completed tasks
func (f *Formatter) WriteCompletedTasks(resp *api.CompletedTasksResponse) error {
	if f.asJSON {
		return f.JSON(resp)
	}

	if len(resp.Items) == 0 {
		fmt.Fprintln(f.w, "No completed tasks found.")
		return nil
	}

	for _, t := range resp.Items {
		fmt.Fprintf(f.w, "\033[90m%s\033[0m  \033[9m%s\033[0m\n", t.CompletedAt[:10], t.Content)
	}

	return nil
}
