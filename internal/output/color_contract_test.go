package output

import "testing"

func TestColorModesForCommittedFormatter(t *testing.T) {
	if got := NewColor(ColorAlways).Wrap(ANSIRed, "task"); got != "\033[31mtask\033[0m" {
		t.Fatalf("forced color = %q", got)
	}
	if got := NewColor(ColorNever).Wrap(ANSIRed, "task"); got != "task" {
		t.Fatalf("disabled color = %q", got)
	}
	t.Setenv("NO_COLOR", "")
	if got := NewColor(ColorAuto).Wrap(ANSIRed, "task"); got != "task" {
		t.Fatalf("NO_COLOR did not suppress ANSI sequences: %q", got)
	}
	if got := NewColor(ColorAlways).Wrap("", "task"); got != "task" {
		t.Fatalf("empty style changed text: %q", got)
	}
}
