package retry

import (
	"time"
)

// ExponentialBackoff implements a retry strategy where delays double after each attempt.
// For example, with InitialDelay=1s and Multiplier=2: 1s, 2s, 4s, 8s...
type ExponentialBackoff struct {
	InitialDelay time.Duration // Starting delay for first retry
	Multiplier   float64       // Factor to multiply delay by after each attempt
	MaxDelay     time.Duration // Maximum delay between attempts
}

// NextDelay calculates the delay for the given attempt using exponential growth.
func (e *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	delay := e.InitialDelay
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * e.Multiplier)
		if e.MaxDelay > 0 && delay > e.MaxDelay {
			return e.MaxDelay
		}
	}
	return delay
}

// ShouldRetry returns true unless the error is permanent.
func (e *ExponentialBackoff) ShouldRetry(attempt int, err error) bool {
	return !IsPermanentError(err)
}

// LinearBackoff implements a retry strategy with linearly increasing delays.
// For example, with InitialDelay=1s and Increment=2s: 1s, 3s, 5s, 7s...
type LinearBackoff struct {
	InitialDelay time.Duration // Starting delay for first retry
	Increment    time.Duration // Amount to add to delay after each attempt
}

// NextDelay calculates the delay by adding Increment for each attempt.
func (l *LinearBackoff) NextDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}
	return l.InitialDelay + time.Duration(attempt-1)*l.Increment
}

// ShouldRetry returns true unless the error is permanent.
func (l *LinearBackoff) ShouldRetry(attempt int, err error) bool {
	return !IsPermanentError(err)
}

// ConstantDelay implements a retry strategy with fixed delay between attempts.
type ConstantDelay struct {
	Delay time.Duration // Fixed delay between all retry attempts
}

// NextDelay returns the same delay for all attempts.
func (c *ConstantDelay) NextDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}
	return c.Delay
}

// ShouldRetry returns true unless the error is permanent.
func (c *ConstantDelay) ShouldRetry(attempt int, err error) bool {
	return !IsPermanentError(err)
}

// CustomStrategy allows users to define custom retry behavior with user-provided functions.
type CustomStrategy struct {
	DelayFunc       func(attempt int) time.Duration   // Custom delay calculation
	ShouldRetryFunc func(attempt int, err error) bool // Custom retry condition
}

// NextDelay delegates to the user-defined delay function.
func (c *CustomStrategy) NextDelay(attempt int) time.Duration {
	if c.DelayFunc == nil {
		return 0
	}
	return c.DelayFunc(attempt)
}

// ShouldRetry delegates to the user-defined retry function.
func (c *CustomStrategy) ShouldRetry(attempt int, err error) bool {
	if c.ShouldRetryFunc == nil {
		return true
	}
	return c.ShouldRetryFunc(attempt, err)
}

// NoDelay implements immediate retry without any delay between attempts.
type NoDelay struct{}

// NextDelay always returns zero delay.
func (n *NoDelay) NextDelay(attempt int) time.Duration {
	return 0
}

// ShouldRetry returns true unless the error is permanent.
func (n *NoDelay) ShouldRetry(attempt int, err error) bool {
	return !IsPermanentError(err)
}
