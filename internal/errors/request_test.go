package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestRequestErrorClassification(t *testing.T) {
	for _, test := range []struct {
		name string
		wrap func(string, error) error
		code int
	}{
		{"auth", WrapAuthError, 3},
		{"network", WrapNetworkError, 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			cause := errors.New("original failure")
			err := test.wrap("request failed", cause)
			if !errors.Is(err, cause) {
				t.Fatal("original cause was lost")
			}
			if got := err.Error(); got != "request failed: original failure" {
				t.Fatalf("message = %q", got)
			}
			var classified interface{ ExitCode() int }
			if !errors.As(fmt.Errorf("outer context: %w", err), &classified) || classified.ExitCode() != test.code {
				t.Fatalf("exit classification was lost: %v", err)
			}
			if got := test.wrap("request failed", nil).Error(); got != "request failed" {
				t.Fatalf("nil-cause message = %q", got)
			}
		})
	}
}
