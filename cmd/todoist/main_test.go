package main

import (
	"errors"
	"fmt"
	"testing"

	clerrors "github.com/buddyh/todoist-cli/internal/errors"
)

func TestExitCode(t *testing.T) {
	cause := errors.New("request failed")
	for _, test := range []struct {
		name string
		err  error
		want int
	}{
		{"success", nil, 0},
		{"unclassified", cause, 1},
		{"authentication", clerrors.WrapAuthError("authentication failed", cause), 3},
		{"network", clerrors.WrapNetworkError("network failed", cause), 4},
		{"wrapped authentication", fmt.Errorf("load tasks: %w", clerrors.WrapAuthError("authentication failed", cause)), 3},
		{"wrapped network", fmt.Errorf("load tasks: %w", clerrors.WrapNetworkError("network failed", cause)), 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := exitCode(test.err); got != test.want {
				t.Fatalf("exit code = %d, want %d", got, test.want)
			}
		})
	}
}
