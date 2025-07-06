package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/remiges-tech/await/retry"
)

func main() {
	ctx := context.Background()

	// Example 1: Retry with exponential backoff
	fmt.Println("Example 1: Exponential backoff retry")
	attempts := 0
	result, err := retry.Do(ctx, func(ctx context.Context) (string, error) {
		attempts++
		log.Printf("Attempt %d", attempts)
		if attempts < 3 {
			return "", errors.New("temporary failure")
		}
		return "Success!", nil
	}, retry.Options{
		Strategy: &retry.ExponentialBackoff{
			InitialDelay: 100 * time.Millisecond,
			Multiplier:   2,
			MaxDelay:     1 * time.Second,
		},
		MaxAttempts: 5,
	})

	if err != nil {
		log.Printf("Failed: %v", err)
	} else {
		log.Printf("Result: %s", result)
	}

	fmt.Println("\nExample 2: Linear backoff with retry callback")
	attempts = 0
	result, err = retry.Do(ctx, func(ctx context.Context) (string, error) {
		attempts++
		if attempts < 4 {
			return "", fmt.Errorf("attempt %d failed", attempts)
		}
		return "Eventually succeeded", nil
	}, retry.Options{
		Strategy: &retry.LinearBackoff{
			InitialDelay: 50 * time.Millisecond,
			Increment:    50 * time.Millisecond,
		},
		MaxAttempts: 5,
		OnRetry: func(attempt int, err error) {
			log.Printf("Retry %d after error: %v", attempt, err)
		},
	})

	if err != nil {
		log.Printf("Failed: %v", err)
	} else {
		log.Printf("Result: %s", result)
	}

	fmt.Println("\nExample 3: Custom retry strategy")
	customStrategy := &retry.CustomStrategy{
		DelayFunc: func(attempt int) time.Duration {
			// Custom delay pattern: 100ms, 200ms, 500ms, 1s
			delays := []time.Duration{100, 200, 500, 1000}
			if attempt > len(delays) {
				return delays[len(delays)-1] * time.Millisecond
			}
			return delays[attempt-1] * time.Millisecond
		},
		ShouldRetryFunc: func(attempt int, err error) bool {
			// Only retry if error message contains "retry"
			return attempt < 5 && errors.Is(err, errRetryable)
		},
	}

	attempts = 0
	result, err = retry.Do(ctx, func(ctx context.Context) (string, error) {
		attempts++
		if attempts < 3 {
			return "", errRetryable
		}
		return "Custom strategy worked", nil
	}, retry.Options{
		Strategy:    customStrategy,
		MaxAttempts: 10,
	})

	if err != nil {
		log.Printf("Failed: %v", err)
	} else {
		log.Printf("Result: %s", result)
	}

	fmt.Println("\nExample 4: Permanent error (no retry)")
	_, err = retry.Do(ctx, func(ctx context.Context) (string, error) {
		log.Printf("This will only be called once")
		return "", retry.Permanent(errors.New("permanent failure"))
	}, retry.WithMaxAttempts(5))

	if err != nil {
		log.Printf("Failed as expected: %v", err)
	}

	fmt.Println("\nExample 5: Retry with specific error conditions")
	attempts = 0
	result, err = retry.Do(ctx, func(ctx context.Context) (string, error) {
		attempts++
		switch attempts {
		case 1:
			return "", errRetryable
		case 2:
			return "", errNotRetryable
		default:
			return "Should not reach here", nil
		}
	}, retry.Options{
		Strategy:    &retry.ConstantDelay{Delay: 50 * time.Millisecond},
		MaxAttempts: 5,
		RetryIf:     retry.RetryIf(errRetryable),
	})

	if err != nil {
		log.Printf("Stopped on non-retryable error: %v", err)
	}
}

var (
	errRetryable    = errors.New("retryable error")
	errNotRetryable = errors.New("not retryable")
)
