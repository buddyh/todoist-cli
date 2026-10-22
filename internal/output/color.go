package output

import (
	"os"

	"golang.org/x/term"
)

// ColorMode determines whether a formatter emits ANSI escape sequences.
type ColorMode int

const (
	// ColorAuto follows the output terminal and NO_COLOR environment setting.
	ColorAuto ColorMode = iota
	// ColorAlways emits ANSI colors even when output is redirected.
	ColorAlways
	// ColorNever emits plain text.
	ColorNever
)

// Color applies ANSI styles when the selected output mode allows them.
type Color struct{ enabled bool }

// NewColor resolves the requested mode against the current output environment.
func NewColor(mode ColorMode) *Color {
	enabled := false
	switch mode {
	case ColorAlways:
		enabled = true
	case ColorNever:
	default:
		_, disabled := os.LookupEnv("NO_COLOR")
		enabled = !disabled && os.Getenv("TERM") != "dumb" && term.IsTerminal(int(os.Stdout.Fd()))
	}
	return &Color{enabled: enabled}
}

// Wrap styles text when enabled, otherwise returning it unchanged.
func (c *Color) Wrap(code, text string) string {
	if !c.enabled || code == "" {
		return text
	}
	return code + text + "\033[0m"
}

// ANSI styles used by the committed task and project formatters.
const (
	ANSIRed    = "\033[31m"
	ANSIYellow = "\033[33m"
	ANSIBlue   = "\033[34m"
	ANSICyan   = "\033[36m"
	ANSIGray   = "\033[90m"
	ANSIStrike = "\033[9m"
)
