package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/buddyh/todoist-cli/internal/api"
)

func TestWriteTasks_Hierarchy(t *testing.T) {
	// Setup tasks with hierarchy
	// Parent (1)
	//   Child 1 (2)
	//     Grandchild (4)
	//   Child 2 (3)
	tasks := []api.Task{
		{ID: "1", Content: "Parent", Order: 1},
		{ID: "2", Content: "Child 1", ParentID: "1", Order: 1},
		{ID: "3", Content: "Child 2", ParentID: "1", Order: 2},
		{ID: "4", Content: "Grandchild", ParentID: "2", Order: 1},
	}

	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	err := f.WriteTasks(tasks)
	if err != nil {
		t.Fatalf("WriteTasks failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	if len(lines) != 4 {
		t.Fatalf("Expected 4 lines, got %d. Output:\n%s", len(lines), output)
	}

	// Verify hierarchy (indentation)
	// Note: FormatTaskLine output starts with color code if not indented.
	// If indented, it starts with spaces.
	
	expectedOrder := []string{"1", "2", "4", "3"}
	
	for i, id := range expectedOrder {
		line := lines[i]
		if !strings.Contains(line, id) {
			t.Errorf("Line %d expected to contain ID %s, got: %s", i, id, line)
		}
	}

	// Check indentation levels
	// Line 0: ID 1 (Root) -> 0 spaces
	if strings.HasPrefix(lines[0], " ") {
		t.Errorf("Line 0 (Root) should not be indented. Got: %q", lines[0])
	}

	// Line 1: ID 2 (Child of 1) -> 2 spaces
	if !strings.HasPrefix(lines[1], "  \033") {
		t.Errorf("Line 1 (Child) should be indented by 2 spaces. Got: %q", lines[1])
	}

	// Line 2: ID 4 (Child of 2) -> 4 spaces
	if !strings.HasPrefix(lines[2], "    \033") {
		t.Errorf("Line 2 (Grandchild) should be indented by 4 spaces. Got: %q", lines[2])
	}

	// Line 3: ID 3 (Child of 1) -> 2 spaces
	if !strings.HasPrefix(lines[3], "  \033") {
		t.Errorf("Line 3 (Child) should be indented by 2 spaces. Got: %q", lines[3])
	}
}

func TestWriteTasks_JSON(t *testing.T) {
	tasks := []api.Task{
		{ID: "1", Content: "Parent"},
		{ID: "2", Content: "Child", ParentID: "1"},
	}

	var buf bytes.Buffer
	f := NewFormatter(&buf, true)

	err := f.WriteTasks(tasks)
	if err != nil {
		t.Fatalf("WriteTasks failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"content":"Parent"`) {
		t.Error("JSON output should contain Parent task")
	}
	if !strings.Contains(output, `"content":"Child"`) {
		t.Error("JSON output should contain Child task")
	}
}

func TestWriteTasks_Empty(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	err := f.WriteTasks([]api.Task{})
	if err != nil {
		t.Fatalf("WriteTasks failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No tasks found") {
		t.Errorf("Expected 'No tasks found' message, got: %q", output)
	}
}

func TestWriteTasks_EmptyJSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, true)

	err := f.WriteTasks([]api.Task{})
	if err != nil {
		t.Fatalf("WriteTasks failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"success":true`) {
		t.Error("JSON output should contain success:true")
	}
	if !strings.Contains(output, `"data":[]`) {
		t.Error("JSON output should contain empty data array")
	}
}

func TestFormatTask_Priority(t *testing.T) {
	tests := []struct {
		priority int
		want     string
	}{
		{4, "[p1]"},
		{3, "[p2]"},
		{2, "[p3]"},
		{1, ""},
	}

	for _, tt := range tests {
		task := &api.Task{Content: "Test", Priority: tt.priority}
		output := FormatTask(task)

		if tt.want == "" {
			if strings.Contains(output, "[p") {
				t.Errorf("Priority %d should not show indicator, got: %s", tt.priority, output)
			}
		} else {
			if !strings.Contains(output, tt.want) {
				t.Errorf("Priority %d should show %s, got: %s", tt.priority, tt.want, output)
			}
		}
	}
}

func TestFormatTask_DueDate(t *testing.T) {
	task := &api.Task{
		Content: "Test",
		Due: &api.Due{
			String: "tomorrow",
			Date:   "2024-01-15",
		},
	}

	output := FormatTask(task)
	if !strings.Contains(output, "(tomorrow)") {
		t.Errorf("Should contain due string 'tomorrow', got: %s", output)
	}
}

func TestFormatTask_DueDateFallback(t *testing.T) {
	task := &api.Task{
		Content: "Test",
		Due: &api.Due{
			String: "",
			Date:   "2024-01-15",
		},
	}

	output := FormatTask(task)
	if !strings.Contains(output, "(2024-01-15)") {
		t.Errorf("Should fall back to date when string is empty, got: %s", output)
	}
}

func TestFormatTask_Labels(t *testing.T) {
	task := &api.Task{
		Content: "Test",
		Labels:  []string{"urgent", "work"},
	}

	output := FormatTask(task)
	if !strings.Contains(output, "@urgent") {
		t.Errorf("Should contain @urgent label, got: %s", output)
	}
	if !strings.Contains(output, "@work") {
		t.Errorf("Should contain @work label, got: %s", output)
	}
}

func TestWriteError(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	testErr := &testError{msg: "test error message"}
	f.WriteError(testErr)

	output := buf.String()
	if !strings.Contains(output, "Error: test error message") {
		t.Errorf("Expected error message, got: %q", output)
	}
}

func TestWriteError_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, true)

	testErr := &testError{msg: "test error message"}
	f.WriteError(testErr)

	output := buf.String()
	if !strings.Contains(output, `"success":false`) {
		t.Error("JSON error should have success:false")
	}
	if !strings.Contains(output, `"error":"test error message"`) {
		t.Errorf("JSON error should contain error message, got: %s", output)
	}
}

func TestWriteSuccess(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	f.WriteSuccess("Operation completed")

	output := buf.String()
	if !strings.Contains(output, "Operation completed") {
		t.Errorf("Expected success message, got: %q", output)
	}
}

func TestWriteSuccess_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, true)

	f.WriteSuccess("Operation completed")

	output := buf.String()
	if !strings.Contains(output, `"success":true`) {
		t.Error("JSON success should have success:true")
	}
	if !strings.Contains(output, `"message":"Operation completed"`) {
		t.Errorf("JSON should contain message, got: %s", output)
	}
}

func TestWriteProjects(t *testing.T) {
	projects := []api.Project{
		{ID: "1", Name: "Work", IsFavorite: true},
		{ID: "2", Name: "Personal", IsInboxProject: true},
	}

	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	err := f.WriteProjects(projects)
	if err != nil {
		t.Fatalf("WriteProjects failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Work") {
		t.Error("Output should contain project name 'Work'")
	}
	if !strings.Contains(output, "Personal") {
		t.Error("Output should contain project name 'Personal'")
	}
}

func TestWriteProjects_Empty(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	err := f.WriteProjects([]api.Project{})
	if err != nil {
		t.Fatalf("WriteProjects failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No projects found") {
		t.Errorf("Expected 'No projects found' message, got: %q", output)
	}
}

func TestWriteLabels(t *testing.T) {
	labels := []api.Label{
		{ID: "1", Name: "urgent"},
		{ID: "2", Name: "work"},
	}

	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	err := f.WriteLabels(labels)
	if err != nil {
		t.Fatalf("WriteLabels failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "@urgent") {
		t.Error("Output should contain @urgent")
	}
	if !strings.Contains(output, "@work") {
		t.Error("Output should contain @work")
	}
}

func TestWriteLabels_Empty(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	err := f.WriteLabels([]api.Label{})
	if err != nil {
		t.Fatalf("WriteLabels failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No labels found") {
		t.Errorf("Expected 'No labels found' message, got: %q", output)
	}
}

func TestWriteSections(t *testing.T) {
	sections := []api.Section{
		{ID: "1", Name: "To Do"},
		{ID: "2", Name: "In Progress"},
	}

	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	err := f.WriteSections(sections)
	if err != nil {
		t.Fatalf("WriteSections failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "To Do") {
		t.Error("Output should contain 'To Do'")
	}
	if !strings.Contains(output, "In Progress") {
		t.Error("Output should contain 'In Progress'")
	}
}

func TestWriteSections_Empty(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	err := f.WriteSections([]api.Section{})
	if err != nil {
		t.Fatalf("WriteSections failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No sections found") {
		t.Errorf("Expected 'No sections found' message, got: %q", output)
	}
}

func TestWriteComments(t *testing.T) {
	comments := []api.Comment{
		{ID: "1", Content: "First comment", PostedAt: "2024-01-15T10:00:00Z"},
		{ID: "2", Content: "Second comment", PostedAt: "2024-01-16T11:00:00Z"},
	}

	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	err := f.WriteComments(comments)
	if err != nil {
		t.Fatalf("WriteComments failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "First comment") {
		t.Error("Output should contain 'First comment'")
	}
	if !strings.Contains(output, "Second comment") {
		t.Error("Output should contain 'Second comment'")
	}
}

func TestWriteComments_Empty(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	err := f.WriteComments([]api.Comment{})
	if err != nil {
		t.Fatalf("WriteComments failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No comments found") {
		t.Errorf("Expected 'No comments found' message, got: %q", output)
	}
}

func TestWriteCompletedTasks(t *testing.T) {
	resp := &api.CompletedTasksResponse{
		Items: []api.CompletedTask{
			{ID: "1", Content: "Completed task 1", CompletedAt: "2024-01-15T10:00:00Z"},
			{ID: "2", Content: "Completed task 2", CompletedAt: "2024-01-14T09:00:00Z"},
		},
	}

	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	err := f.WriteCompletedTasks(resp)
	if err != nil {
		t.Fatalf("WriteCompletedTasks failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Completed task 1") {
		t.Error("Output should contain 'Completed task 1'")
	}
	if !strings.Contains(output, "2024-01-15") {
		t.Error("Output should contain date '2024-01-15'")
	}
}

func TestWriteCompletedTasks_Empty(t *testing.T) {
	resp := &api.CompletedTasksResponse{Items: []api.CompletedTask{}}

	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	err := f.WriteCompletedTasks(resp)
	if err != nil {
		t.Fatalf("WriteCompletedTasks failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No completed tasks found") {
		t.Errorf("Expected 'No completed tasks found' message, got: %q", output)
	}
}

func TestFormatProject_Markers(t *testing.T) {
	tests := []struct {
		name     string
		project  api.Project
		contains []string
	}{
		{
			name:     "favorite project",
			project:  api.Project{Name: "Work", IsFavorite: true},
			contains: []string{"Work", "*"},
		},
		{
			name:     "inbox project",
			project:  api.Project{Name: "Inbox", IsInboxProject: true},
			contains: []string{"Inbox", "inbox"},
		},
		{
			name:     "regular project",
			project:  api.Project{Name: "Personal"},
			contains: []string{"Personal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatProject(&tt.project)
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("FormatProject() = %q, want to contain %q", output, want)
				}
			}
		})
	}
}

func TestWriteTask_WithDescription(t *testing.T) {
	task := &api.Task{
		ID:          "123",
		Content:     "Test task",
		Description: "This is a description",
	}

	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	err := f.WriteTask(task)
	if err != nil {
		t.Fatalf("WriteTask failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test task") {
		t.Error("Output should contain task content")
	}
	if !strings.Contains(output, "This is a description") {
		t.Error("Output should contain description")
	}
}

func TestJSON_Envelope(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, true)

	data := map[string]string{"key": "value"}
	err := f.JSON(data)
	if err != nil {
		t.Fatalf("JSON() failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"success":true`) {
		t.Error("JSON envelope should have success:true")
	}
	if !strings.Contains(output, `"data"`) {
		t.Error("JSON envelope should have data field")
	}
}

func TestOrphanedSubtasks(t *testing.T) {
	// Test tasks where parent is not in the list
	tasks := []api.Task{
		{ID: "1", Content: "Orphan", ParentID: "999", Order: 1},
		{ID: "2", Content: "Root", Order: 2},
	}

	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	err := f.WriteTasks(tasks)
	if err != nil {
		t.Fatalf("WriteTasks failed: %v", err)
	}

	output := buf.String()
	// Orphan should be treated as root level
	if !strings.Contains(output, "Orphan") {
		t.Error("Output should contain orphan task")
	}
	if !strings.Contains(output, "Root") {
		t.Error("Output should contain root task")
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}