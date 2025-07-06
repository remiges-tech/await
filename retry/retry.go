package retry

import (
	"context"
	"time"
)

// Strategy defines the retry behavior including delays and retry conditions.
type Strategy interface {
	// NextDelay returns the delay before the next retry attempt.
	NextDelay(attempt int) time.Duration
	// ShouldRetry determines if a retry should be attempted.
	ShouldRetry(attempt int, err error) bool
}

// Options configures retry behavior including strategy, conditions, and callbacks.
type Options struct {
	Strategy    Strategy                     // Determines delay between attempts
	MaxAttempts int                          // Maximum number of attempts (must be > 0)
	OnRetry     func(attempt int, err error) // Called before each retry
	RetryIf     func(error) bool             // Optional condition to check if error is retryable
}

// DefaultOptions returns default options with exponential backoff and 3 attempts.
func DefaultOptions() Options {
	return Options{
		Strategy: &ExponentialBackoff{
			InitialDelay: 100 * time.Millisecond,
			Multiplier:   2,
			MaxDelay:     30 * time.Second,
		},
		MaxAttempts: 3,
	}
}

// Do executes the function with retry logic, attempting up to MaxAttempts times.
// It stops retrying when the function succeeds, a permanent error occurs,
// or the context is cancelled. Returns the last error wrapped in RetryError
// if all attempts fail.
func Do[T any](ctx context.Context, fn func(context.Context) (T, error), opts Options) (T, error) {
	var zero T
	if opts.MaxAttempts <= 0 {
		return zero, ErrMaxAttemptsInvalid
	}

	var lastErr error
	for attempt := 1; attempt <= opts.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return zero, err
		}

		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err

		if !shouldRetryError(opts, err) {
			return zero, err
		}

		if !opts.Strategy.ShouldRetry(attempt, err) {
			return zero, err
		}

		if isLastAttempt(attempt, opts.MaxAttempts) {
			break
		}

		if opts.OnRetry != nil {
			opts.OnRetry(attempt, err)
		}

		delay := calculateDelay(opts, attempt)

		if err := waitForRetry(ctx, delay); err != nil {
			return zero, err
		}
	}

	return zero, &RetryError{
		LastError: lastErr,
		Attempts:  opts.MaxAttempts,
	}
}

// WithMaxAttempts creates options with specified max attempts and default strategy.
func WithMaxAttempts(attempts int) Options {
	opts := DefaultOptions()
	opts.MaxAttempts = attempts
	return opts
}

// WithStrategy creates options with specified strategy and default max attempts.
func WithStrategy(strategy Strategy) Options {
	opts := DefaultOptions()
	opts.Strategy = strategy
	return opts
}

// WithOnRetry creates options with specified callback and default values.
func WithOnRetry(onRetry func(attempt int, err error)) Options {
	opts := DefaultOptions()
	opts.OnRetry = onRetry
	return opts
}

func shouldRetryError(opts Options, err error) bool {
	if opts.RetryIf == nil {
		return true
	}
	return opts.RetryIf(err)
}

func isLastAttempt(attempt, maxAttempts int) bool {
	return attempt >= maxAttempts
}

func calculateDelay(opts Options, attempt int) time.Duration {
	return opts.Strategy.NextDelay(attempt)
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}
