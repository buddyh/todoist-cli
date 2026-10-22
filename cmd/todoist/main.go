package main

import (
	"errors"
	"os"
)

func main() {
	if err := execute(os.Args[1:]); err != nil {
		os.Exit(exitCode(err))
	}
}

// exitCode preserves request classifications through wrapped command errors.
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var classified interface{ ExitCode() int }
	if errors.As(err, &classified) {
		return classified.ExitCode()
	}
	return 1
}
