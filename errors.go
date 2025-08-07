package await

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrNoTasks is returned when an empty task slice is provided to All, Any, or Race.
	ErrNoTasks = errors.New("no tasks provided")
)

// AggregateError contains multiple errors from concurrent operations.
// Returned by Any when all tasks fail.
type AggregateError struct {
	Errors []error // All errors that occurred during execution
}

// Error returns a formatted message listing all contained errors.
func (e *AggregateError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}

	var messages []string
	for _, err := range e.Errors {
		if err != nil {
			messages = append(messages, err.Error())
		}
	}

	return fmt.Sprintf("multiple errors occurred: [%s]", strings.Join(messages, "; "))
}

// Unwrap returns all contained errors for use with errors.Is and errors.As.
// Allows checking if an AggregateError contains a specific error type.
func (e *AggregateError) Unwrap() []error {
	return e.Errors
}
