// Package await provides Promise-like async task patterns for Go, bringing JavaScript-style
// async utilities (All, Any, Race) to Go's concurrency model with type safety
// via generics.
package await

import (
	"context"
	"sync"
)

// Result holds either a value or an error from an async operation.
// Used by All to return both successful and failed results for each task.
type Result[T any] struct {
	Value T     // The successful result value (or zero value if failed)
	Err   error // The error if the operation failed (nil if succeeded)
}

// Task represents an async operation that returns a value of type T or an error.
// All async functions in this package operate on Task functions.
type Task[T any] func(ctx context.Context) (T, error)

// All executes all tasks concurrently and waits for all to complete.
// Philosophy: "Execute all tasks and return complete information about each task's outcome."
// Returns a Result for each task containing both its value and error,
// preserving the original task order. This allows you to handle mixed
// success/failure scenarios with complete information about each task.
// Function-level errors (returned as second parameter) only occur for operational issues
// like empty task list or context cancellation before execution.
// Task-level errors are captured in each Result[T].Err field.
func All[T any](ctx context.Context, tasks ...Task[T]) ([]Result[T], error) {
	// Validate inputs - function-level errors
	if len(tasks) == 0 {
		return nil, ErrNoTasks
	}

	// Check context before starting
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	results := make([]Result[T], len(tasks))
	var wg sync.WaitGroup

	for i, t := range tasks {
		wg.Add(1)
		go func(idx int, task Task[T]) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				results[idx] = Result[T]{Err: ctx.Err()}
				return
			default:
				val, err := task(ctx)
				results[idx] = Result[T]{Value: val, Err: err}
			}
		}(i, t)
	}

	wg.Wait()
	return results, nil
}

// Any executes all tasks concurrently and returns when the first task succeeds.
// Returns the value from the first successful task, or an AggregateError
// if all tasks fail. Similar to Promise.any in JavaScript.
func Any[T any](ctx context.Context, tasks ...Task[T]) (T, error) {
	var zero T
	if len(tasks) == 0 {
		return zero, ErrNoTasks
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		val T
		err error
	}

	ch := make(chan result, len(tasks))

	for _, t := range tasks {
		go func(task Task[T]) {
			select {
			case <-ctx.Done():
				ch <- result{err: ctx.Err()}
				return
			default:
				val, err := task(ctx)
				ch <- result{val, err}
			}
		}(t)
	}

	errors := make([]error, 0, len(tasks))
	for i := 0; i < len(tasks); i++ {
		res := <-ch
		if res.err == nil {
			cancel() // Cancel remaining
			return res.val, nil
		}
		errors = append(errors, res.err)
	}

	return zero, &AggregateError{Errors: errors}
}

// Race executes all tasks concurrently and returns the first to complete,
// whether it succeeds or fails. Similar to Promise.race in JavaScript.
func Race[T any](ctx context.Context, tasks ...Task[T]) (T, error) {
	var zero T
	if len(tasks) == 0 {
		return zero, ErrNoTasks
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		val T
		err error
	}

	ch := make(chan result, len(tasks))

	for _, t := range tasks {
		go func(task Task[T]) {
			select {
			case <-ctx.Done():
				ch <- result{err: ctx.Err()}
				return
			default:
				val, err := task(ctx)
				select {
				case ch <- result{val, err}:
				case <-ctx.Done():
				}
			}
		}(t)
	}

	res := <-ch
	cancel() // Cancel remaining
	return res.val, res.err
}
