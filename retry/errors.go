package retry

import (
	"errors"
	"fmt"
)

var (
	// ErrMaxAttemptsInvalid is returned when max attempts is <= 0.
	ErrMaxAttemptsInvalid = errors.New("max attempts must be greater than 0")

	// ErrPermanent is a sentinel error used to mark errors as non-retryable.
	// Wrap errors with Permanent() to prevent retry attempts.
	ErrPermanent = errors.New("permanent error")
)

// RetryError is returned when all retry attempts fail.
// It contains the last error and the number of attempts made.
type RetryError struct {
	LastError error // The error from the final attempt
	Attempts  int   // Total number of attempts made
}

// Error returns a formatted message with attempt count and last error.
func (e *RetryError) Error() string {
	return fmt.Sprintf("retry failed after %d attempts: %v", e.Attempts, e.LastError)
}

// Unwrap returns the underlying error.
func (e *RetryError) Unwrap() error {
	return e.LastError
}

// PermanentError wraps an error to mark it as non-retryable.
// Any error wrapped with PermanentError will cause retry logic to stop immediately.
type PermanentError struct {
	Err error // The wrapped error
}

// Error returns the wrapped error's message.
func (p *PermanentError) Error() string {
	return p.Err.Error()
}

// Unwrap returns the wrapped error.
func (p *PermanentError) Unwrap() error {
	return p.Err
}

// Is reports whether the target is ErrPermanent.
func (p *PermanentError) Is(target error) bool {
	return target == ErrPermanent
}

// Permanent wraps an error to mark it as non-retryable.
// Use this for errors that won't succeed on retry (e.g., invalid input, auth failures).
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return &PermanentError{Err: err}
}

// IsPermanentError checks if an error is marked as permanent using errors.Is.
func IsPermanentError(err error) bool {
	return errors.Is(err, ErrPermanent)
}
