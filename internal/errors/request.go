// Package errors classifies request failures while preserving their causes.
package errors

import "fmt"

type requestError struct {
	message string
	cause   error
	code    int
}

// Error returns context with the original failure message.
func (e *requestError) Error() string {
	if e.cause == nil {
		return e.message
	}
	return fmt.Sprintf("%s: %v", e.message, e.cause)
}

// Unwrap preserves errors.Is and errors.As traversal to the cause.
func (e *requestError) Unwrap() error { return e.cause }
// ExitCode identifies authentication or network failures for CLI callers.
func (e *requestError) ExitCode() int { return e.code }

// WrapAuthError adds authentication context and exit code 3 to an error.
func WrapAuthError(message string, cause error) error {
	return &requestError{message: message, cause: cause, code: 3}
}

// WrapNetworkError adds network context and exit code 4 to an error.
func WrapNetworkError(message string, cause error) error {
	return &requestError{message: message, cause: cause, code: 4}
}
