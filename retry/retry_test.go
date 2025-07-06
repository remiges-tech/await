package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestCustomError is used for testing error type conditions.
type TestCustomError struct {
	msg string
}

func (e *TestCustomError) Error() string {
	return e.msg
}

func TestDo(t *testing.T) {
	t.Run("successful operation", func(t *testing.T) {
		attempts := 0
		fn := func(ctx context.Context) (int, error) {
			attempts++
			return 42, nil
		}

		result, err := Do(context.Background(), fn, WithMaxAttempts(3))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != 42 {
			t.Fatalf("expected 42, got %d", result)
		}
		if attempts != 1 {
			t.Fatalf("expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("retry until success", func(t *testing.T) {
		attempts := 0
		fn := func(ctx context.Context) (string, error) {
			attempts++
			if attempts < 3 {
				return "", errors.New("temporary error")
			}
			return "success", nil
		}

		opts := Options{
			Strategy:    &ConstantDelay{Delay: 10 * time.Millisecond},
			MaxAttempts: 5,
		}

		result, err := Do(context.Background(), fn, opts)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != "success" {
			t.Fatalf("expected 'success', got %s", result)
		}
		if attempts != 3 {
			t.Fatalf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("max attempts exceeded", func(t *testing.T) {
		attempts := 0
		fn := func(ctx context.Context) (int, error) {
			attempts++
			return 0, errors.New("always fails")
		}

		opts := Options{
			Strategy:    &NoDelay{},
			MaxAttempts: 3,
		}

		_, err := Do(context.Background(), fn, opts)
		if err == nil {
			t.Fatal("expected error")
		}

		var retryErr *RetryError
		if !errors.As(err, &retryErr) {
			t.Fatalf("expected RetryError, got %T", err)
		}
		if retryErr.Attempts != 3 {
			t.Fatalf("expected 3 attempts in error, got %d", retryErr.Attempts)
		}
		if attempts != 3 {
			t.Fatalf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		attempts := 0

		fn := func(ctx context.Context) (int, error) {
			attempts++
			if attempts == 2 {
				cancel()
			}
			return 0, errors.New("error")
		}

		opts := Options{
			Strategy:    &ConstantDelay{Delay: 100 * time.Millisecond},
			MaxAttempts: 5,
		}

		_, err := Do(ctx, fn, opts)
		if err != context.Canceled {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
		if attempts > 3 {
			t.Fatalf("expected at most 3 attempts, got %d", attempts)
		}
	})

	t.Run("permanent error", func(t *testing.T) {
		attempts := 0
		permanentErr := Permanent(errors.New("permanent failure"))

		fn := func(ctx context.Context) (int, error) {
			attempts++
			return 0, permanentErr
		}

		opts := Options{
			Strategy:    &NoDelay{},
			MaxAttempts: 5,
		}

		_, err := Do(context.Background(), fn, opts)
		if !errors.Is(err, permanentErr) {
			t.Fatalf("expected permanent error, got %v", err)
		}
		if attempts != 1 {
			t.Fatalf("expected 1 attempt for permanent error, got %d", attempts)
		}
	})

	t.Run("retry callback", func(t *testing.T) {
		var callbacks []int
		fn := func(ctx context.Context) (int, error) {
			return 0, errors.New("error")
		}

		opts := Options{
			Strategy:    &NoDelay{},
			MaxAttempts: 3,
			OnRetry: func(attempt int, err error) {
				callbacks = append(callbacks, attempt)
			},
		}

		_, err := Do(context.Background(), fn, opts)
		if err == nil {
			t.Fatal("expected error but got none")
		}

		expected := []int{1, 2}
		if len(callbacks) != len(expected) {
			t.Fatalf("expected %d callbacks, got %d", len(expected), len(callbacks))
		}
		for i, v := range expected {
			if callbacks[i] != v {
				t.Fatalf("callback %d: expected attempt %d, got %d", i, v, callbacks[i])
			}
		}
	})

	t.Run("retry if condition", func(t *testing.T) {
		specificErr := errors.New("retry this")
		otherErr := errors.New("don't retry this")
		attempts := 0

		fn := func(ctx context.Context) (int, error) {
			attempts++
			if attempts == 1 {
				return 0, specificErr
			}
			return 0, otherErr
		}

		opts := Options{
			Strategy:    &NoDelay{},
			MaxAttempts: 5,
			RetryIf:     RetryIf(specificErr),
		}

		_, err := Do(context.Background(), fn, opts)
		if err != otherErr {
			t.Fatalf("expected otherErr, got %v", err)
		}
		if attempts != 2 {
			t.Fatalf("expected 2 attempts, got %d", attempts)
		}
	})

	t.Run("invalid max attempts", func(t *testing.T) {
		fn := func(ctx context.Context) (int, error) {
			return 42, nil
		}

		opts := Options{
			Strategy:    &NoDelay{},
			MaxAttempts: 0,
		}

		_, err := Do(context.Background(), fn, opts)
		if err != ErrMaxAttemptsInvalid {
			t.Fatalf("expected ErrMaxAttemptsInvalid, got %v", err)
		}
	})
}

func TestStrategies(t *testing.T) {
	t.Run("ExponentialBackoff", func(t *testing.T) {
		strategy := &ExponentialBackoff{
			InitialDelay: 100 * time.Millisecond,
			Multiplier:   2,
			MaxDelay:     1 * time.Second,
		}

		delays := []time.Duration{
			strategy.NextDelay(1),
			strategy.NextDelay(2),
			strategy.NextDelay(3),
			strategy.NextDelay(4),
			strategy.NextDelay(5),
		}

		expected := []time.Duration{
			100 * time.Millisecond,
			200 * time.Millisecond,
			400 * time.Millisecond,
			800 * time.Millisecond,
			1 * time.Second, // capped at max
		}

		for i, delay := range delays {
			if delay != expected[i] {
				t.Errorf("attempt %d: expected %v, got %v", i+1, expected[i], delay)
			}
		}
	})

	t.Run("LinearBackoff", func(t *testing.T) {
		strategy := &LinearBackoff{
			InitialDelay: 100 * time.Millisecond,
			Increment:    50 * time.Millisecond,
		}

		delays := []time.Duration{
			strategy.NextDelay(1),
			strategy.NextDelay(2),
			strategy.NextDelay(3),
		}

		expected := []time.Duration{
			100 * time.Millisecond,
			150 * time.Millisecond,
			200 * time.Millisecond,
		}

		for i, delay := range delays {
			if delay != expected[i] {
				t.Errorf("attempt %d: expected %v, got %v", i+1, expected[i], delay)
			}
		}
	})

	t.Run("CustomStrategy", func(t *testing.T) {
		callCount := 0
		strategy := &CustomStrategy{
			DelayFunc: func(attempt int) time.Duration {
				return time.Duration(attempt) * 100 * time.Millisecond
			},
			ShouldRetryFunc: func(attempt int, err error) bool {
				callCount++
				return attempt < 3
			},
		}

		if delay := strategy.NextDelay(2); delay != 200*time.Millisecond {
			t.Errorf("expected 200ms, got %v", delay)
		}

		if strategy.ShouldRetry(2, errors.New("test")) != true {
			t.Error("expected ShouldRetry to return true for attempt 2")
		}

		if strategy.ShouldRetry(3, errors.New("test")) != false {
			t.Error("expected ShouldRetry to return false for attempt 3")
		}

		if callCount != 2 {
			t.Errorf("expected 2 calls to ShouldRetryFunc, got %d", callCount)
		}
	})
}

func TestConditions(t *testing.T) {
	err1 := errors.New("error1")
	err2 := errors.New("error2")
	err3 := errors.New("error3")

	t.Run("RetryIf", func(t *testing.T) {
		cond := RetryIf(err1, err2)

		if !cond(err1) {
			t.Error("expected true for err1")
		}
		if !cond(err2) {
			t.Error("expected true for err2")
		}
		if cond(err3) {
			t.Error("expected false for err3")
		}
	})
}

func TestDefaultOptions(t *testing.T) {
	attempts := 0
	fn := func(ctx context.Context) (string, error) {
		attempts++
		if attempts < 2 {
			return "", errors.New("temporary")
		}
		return "success", nil
	}

	result, err := Do(context.Background(), fn, DefaultOptions())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "success" {
		t.Fatalf("expected 'success', got %s", result)
	}
}

func TestPermanentError(t *testing.T) {
	err := errors.New("base error")
	permErr := Permanent(err)

	if permErr.Error() != err.Error() {
		t.Errorf("expected error message %q, got %q", err.Error(), permErr.Error())
	}

	if !errors.Is(permErr, ErrPermanent) {
		t.Error("expected permanent error to match ErrPermanent")
	}

	if !IsPermanentError(permErr) {
		t.Error("expected IsPermanentError to return true")
	}

	if IsPermanentError(err) {
		t.Error("expected IsPermanentError to return false for regular error")
	}

	if Permanent(nil) != nil {
		t.Error("expected Permanent(nil) to return nil")
	}
}

func TestWithOnRetry(t *testing.T) {
	var callbackCalls []struct {
		attempt int
		err     error
	}

	onRetry := func(attempt int, err error) {
		callbackCalls = append(callbackCalls, struct {
			attempt int
			err     error
		}{attempt, err})
	}

	attempts := 0
	fn := func(ctx context.Context) (string, error) {
		attempts++
		if attempts < 3 {
			return "", errors.New("retry error")
		}
		return "success", nil
	}

	result, err := Do(context.Background(), fn, WithOnRetry(onRetry))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "success" {
		t.Fatalf("expected 'success', got %s", result)
	}

	// Verify callback was called twice (before 2nd and 3rd attempts)
	if len(callbackCalls) != 2 {
		t.Fatalf("expected 2 callback calls, got %d", len(callbackCalls))
	}

	// Verify callback parameters
	if callbackCalls[0].attempt != 1 {
		t.Errorf("expected first callback attempt to be 1, got %d", callbackCalls[0].attempt)
	}
	if callbackCalls[0].err.Error() != "retry error" {
		t.Errorf("expected first callback error to be 'retry error', got %v", callbackCalls[0].err)
	}

	if callbackCalls[1].attempt != 2 {
		t.Errorf("expected second callback attempt to be 2, got %d", callbackCalls[1].attempt)
	}
	if callbackCalls[1].err.Error() != "retry error" {
		t.Errorf("expected second callback error to be 'retry error', got %v", callbackCalls[1].err)
	}
}
