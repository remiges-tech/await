package await

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAll(t *testing.T) {
	ctx := context.Background()

	t.Run("successful tasks", func(t *testing.T) {
		t1 := Task[int](func(ctx context.Context) (int, error) {
			return 1, nil
		})
		t2 := Task[int](func(ctx context.Context) (int, error) {
			return 2, nil
		})

		results, err := All(ctx, t1, t2)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		if results[0].Err != nil || results[0].Value != 1 {
			t.Fatalf("expected results[0] = {1, nil}, got %v", results[0])
		}
		if results[1].Err != nil || results[1].Value != 2 {
			t.Fatalf("expected results[1] = {2, nil}, got %v", results[1])
		}
	})

	t.Run("with task errors", func(t *testing.T) {
		t1 := Task[int](func(ctx context.Context) (int, error) {
			return 0, errors.New("task failed")
		})
		t2 := Task[int](func(ctx context.Context) (int, error) {
			return 2, nil
		})

		results, err := All(ctx, t1, t2)
		if err != nil {
			t.Fatalf("expected no function error, got %v", err)
		}

		// Check results contain task errors
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		if results[0].Err == nil || results[0].Err.Error() != "task failed" {
			t.Fatalf("expected results[0].Err = 'task failed', got %v", results[0].Err)
		}
		if results[0].Value != 0 {
			t.Fatalf("expected results[0].Value = 0, got %d", results[0].Value)
		}
		if results[1].Err != nil {
			t.Fatalf("expected results[1].Err = nil, got %v", results[1].Err)
		}
		if results[1].Value != 2 {
			t.Fatalf("expected results[1].Value = 2, got %d", results[1].Value)
		}
	})

	t.Run("multiple task errors", func(t *testing.T) {
		t1 := Task[int](func(ctx context.Context) (int, error) {
			return 10, errors.New("error 1")
		})
		t2 := Task[int](func(ctx context.Context) (int, error) {
			return 20, errors.New("error 2")
		})
		t3 := Task[int](func(ctx context.Context) (int, error) {
			return 30, nil
		})

		results, err := All(ctx, t1, t2, t3)
		if err != nil {
			t.Fatalf("expected no function error, got %v", err)
		}

		// Check all results were returned with their errors
		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d", len(results))
		}
		if results[0].Err == nil || results[0].Err.Error() != "error 1" {
			t.Fatalf("expected results[0].Err = 'error 1', got %v", results[0].Err)
		}
		if results[0].Value != 10 {
			t.Fatalf("expected results[0].Value = 10, got %d", results[0].Value)
		}
		if results[1].Err == nil || results[1].Err.Error() != "error 2" {
			t.Fatalf("expected results[1].Err = 'error 2', got %v", results[1].Err)
		}
		if results[1].Value != 20 {
			t.Fatalf("expected results[1].Value = 20, got %d", results[1].Value)
		}
		if results[2].Err != nil {
			t.Fatalf("expected results[2].Err = nil, got %v", results[2].Err)
		}
		if results[2].Value != 30 {
			t.Fatalf("expected results[2].Value = 30, got %d", results[2].Value)
		}
	})

	t.Run("empty tasks", func(t *testing.T) {
		results, err := All[int](ctx)
		if err != ErrNoTasks {
			t.Fatalf("expected ErrNoTasks, got %v", err)
		}
		if results != nil {
			t.Fatalf("expected nil results, got %v", results)
		}
	})

	t.Run("context cancellation before execution", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		t1 := Task[int](func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			return 1, nil
		})

		results, err := All(ctx, t1)
		if err != context.Canceled {
			t.Fatalf("expected context.Canceled, got %v", err)
		}

		// Results should be nil for function-level error
		if results != nil {
			t.Fatalf("expected nil results, got %v", results)
		}
	})

	t.Run("context cancellation during execution", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		t1 := Task[int](func(ctx context.Context) (int, error) {
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return 1, nil
			}
		})

		results, err := All(ctx, t1)
		if err != nil {
			t.Fatalf("expected no function error, got %v", err)
		}

		// Task should have context error
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		if results[0].Err != context.DeadlineExceeded {
			t.Fatalf("expected context.DeadlineExceeded, got %v", results[0].Err)
		}
	})
}

func TestAny(t *testing.T) {
	ctx := context.Background()

	t.Run("first succeeds", func(t *testing.T) {
		t1 := Task[int](func(ctx context.Context) (int, error) {
			return 1, nil
		})
		t2 := Task[int](func(ctx context.Context) (int, error) {
			time.Sleep(50 * time.Millisecond)
			return 2, nil
		})

		result, err := Any(ctx, t1, t2)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != 1 {
			t.Fatalf("expected 1, got %d", result)
		}
	})

	t.Run("all fail", func(t *testing.T) {
		t1 := Task[int](func(ctx context.Context) (int, error) {
			return 0, errors.New("error 1")
		})
		t2 := Task[int](func(ctx context.Context) (int, error) {
			return 0, errors.New("error 2")
		})

		_, err := Any(ctx, t1, t2)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		aggErr, ok := err.(*AggregateError)
		if !ok {
			t.Fatalf("expected AggregateError, got %T", err)
		}
		if len(aggErr.Errors) != 2 {
			t.Fatalf("expected 2 errors, got %d", len(aggErr.Errors))
		}
	})

	t.Run("empty tasks", func(t *testing.T) {
		_, err := Any[int](ctx)
		if err != ErrNoTasks {
			t.Fatalf("expected ErrNoTasks, got %v", err)
		}
	})
}

func TestRace(t *testing.T) {
	ctx := context.Background()

	t.Run("first completes", func(t *testing.T) {
		t1 := Task[int](func(ctx context.Context) (int, error) {
			time.Sleep(10 * time.Millisecond)
			return 1, nil
		})
		t2 := Task[int](func(ctx context.Context) (int, error) {
			time.Sleep(50 * time.Millisecond)
			return 2, nil
		})

		result, err := Race(ctx, t1, t2)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != 1 {
			t.Fatalf("expected 1, got %d", result)
		}
	})

	t.Run("first fails", func(t *testing.T) {
		t1 := Task[int](func(ctx context.Context) (int, error) {
			return 0, errors.New("quick error")
		})
		t2 := Task[int](func(ctx context.Context) (int, error) {
			time.Sleep(50 * time.Millisecond)
			return 2, nil
		})

		_, err := Race(ctx, t1, t2)
		if err == nil || err.Error() != "quick error" {
			t.Fatalf("expected quick error, got %v", err)
		}
	})
}
