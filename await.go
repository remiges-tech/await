// Package await provides Promise-like async task patterns for Go, bringing JavaScript-style
// async utilities (All, Any, Race, AllSettled) to Go's concurrency model with type safety
// via generics.
package await

import (
	"context"
	"sync"
)

// Result holds either a value or an error from an async operation.
// Used by AllSettled to return both successful and failed results.
type Result[T any] struct {
	Value T     // The successful result value
	Err   error // The error if the operation failed
}

// Task represents an async operation that returns a value of type T or an error.
// All async functions in this package operate on Task functions.
type Task[T any] func(ctx context.Context) (T, error)

// All executes all tasks concurrently and waits for all to complete.
// Philosophy: "Execute all, succeed together or fail together."
// Returns all results in order if all succeed, or an AggregateError containing
// all failures if any task fails. Results array always preserves task order,
// with zero values in positions where tasks failed.
// Similar to Promise.all in JavaScript.
func All[T any](ctx context.Context, tasks ...Task[T]) ([]T, error) {
	if len(tasks) == 0 {
		return []T{}, nil
	}

	results := make([]T, len(tasks))
	errs := make([]error, len(tasks))
	var wg sync.WaitGroup

	for i, t := range tasks {
		wg.Add(1)
		go func(idx int, task Task[T]) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				errs[idx] = ctx.Err()
				return
			default:
				results[idx], errs[idx] = task(ctx)
			}
		}(i, t)
	}

	wg.Wait()

	// Collect errors
	var errors []error
	for _, err := range errs {
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return results, &AggregateError{Errors: errors}
	}

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

// AllSettled executes all tasks concurrently and returns all results,
// regardless of success or failure. Never returns an error.
// Philosophy: "Execute all, give me everything."
// Returns a Result for each task containing both its value and error,
// preserving the original task order. This allows you to handle mixed
// success/failure scenarios where you need to know the outcome of each task.
// Similar to Promise.allSettled in JavaScript.
func AllSettled[T any](ctx context.Context, tasks ...Task[T]) []Result[T] {
	if len(tasks) == 0 {
		return []Result[T]{}
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
	return results
}
