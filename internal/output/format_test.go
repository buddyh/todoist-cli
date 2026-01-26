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