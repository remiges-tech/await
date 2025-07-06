package retry

import (
	"errors"
)

// RetryIf creates a condition function that retries only on specific errors.
// Example: RetryIf(io.EOF, context.DeadlineExceeded) retries only these errors.
func RetryIf(errs ...error) func(error) bool {
	return func(err error) bool {
		for _, e := range errs {
			if errors.Is(err, e) {
				return true
			}
		}
		return false
	}
}
