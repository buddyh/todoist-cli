package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/buddyh/todoist-cli/internal/api"
)

func TestWriteTasks_Hierarchy(t *testing.T) {
	tasks := []api.Task{
		{ID: "1", Content: "Parent", ChildOrder: 1},
		{ID: "2", Content: "Child 1", ParentID: "1", ChildOrder: 1},
		{ID: "3", Content: "Child 2", ParentID: "1", ChildOrder: 2},
		{ID: "4", Content: "Grandchild", ParentID: "2", ChildOrder: 1},
	}

	var buf bytes.Buffer
	f := NewFormatterWithColor(&buf, false, ColorAlways)

	err := f.WriteTasks(tasks)
	if err != nil {
		t.Fatalf("WriteTasks failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 4 {
		t.Fatalf("Expected 4 lines, got %d. Output:\n%s", len(lines), output)
	}

	expectedOrder := []string{"1", "2", "4", "3"}
	for i, id := range expectedOrder {
		if !strings.Contains(lines[i], id) {
			t.Errorf("Line %d expected to contain ID %s, got: %s", i, id, lines[i])
		}
	}

	// Check indentation levels
	if strings.HasPrefix(lines[0], " ") {
		t.Errorf("Root should not be indented. Got: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "  ") {
		t.Errorf("Child should be indented by 2. Got: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "    ") {
		t.Errorf("Grandchild should be indented by 4. Got: %q", lines[2])
	}
	if !strings.HasPrefix(lines[3], "  ") {
		t.Errorf("Child 2 should be indented by 2. Got: %q", lines[3])
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

func TestWriteTasks_NoColor(t *testing.T) {
	tasks := []api.Task{
		{ID: "1", Content: "Test Task", Priority: 4},
	}

	var buf bytes.Buffer
	f := NewFormatterWithColor(&buf, false, ColorNever)

	err := f.WriteTasks(tasks)
	if err != nil {
		t.Fatalf("WriteTasks failed: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "\033[") {
		t.Errorf("ColorNever should produce no ANSI codes. Got: %q", output)
	}
	if !strings.Contains(output, "[p1]") {
		t.Error("Should still contain priority marker")
	}
	if !strings.Contains(output, "Test Task") {
		t.Error("Should contain task content")
	}
}

func TestFormatTask_WithDue(t *testing.T) {
	f := NewFormatterWithColor(nil, false, ColorNever)
	task := &api.Task{
		Content: "Buy milk",
		Due:     &api.Due{String: "tomorrow", Date: "2024-01-15"},
	}

	got := f.FormatTask(task)
	if !strings.Contains(got, "Buy milk") {
		t.Error("Should contain content")
	}
	if !strings.Contains(got, "(tomorrow)") {
		t.Error("Should contain due string")
	}
}

func TestFormatTask_WithLabels(t *testing.T) {
	f := NewFormatterWithColor(nil, false, ColorNever)
	task := &api.Task{
		Content: "Review PR",
		Labels:  []string{"urgent", "work"},
	}

	got := f.FormatTask(task)
	if !strings.Contains(got, "@urgent @work") {
		t.Errorf("Should contain labels, got: %q", got)
	}
}
