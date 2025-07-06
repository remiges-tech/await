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

		if len(results) != 2 || results[0] != 1 || results[1] != 2 {
			t.Fatalf("expected [1, 2], got %v", results)
		}
	})

	t.Run("with error returns partial results", func(t *testing.T) {
		t1 := Task[int](func(ctx context.Context) (int, error) {
			return 0, errors.New("task failed")
		})
		t2 := Task[int](func(ctx context.Context) (int, error) {
			return 2, nil
		})

		results, err := All(ctx, t1, t2)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// Check we got AggregateError
		aggErr, ok := err.(*AggregateError)
		if !ok {
			t.Fatalf("expected AggregateError, got %T", err)
		}
		if len(aggErr.Errors) != 1 {
			t.Fatalf("expected 1 error, got %d", len(aggErr.Errors))
		}

		// Check partial results were returned
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		if results[0] != 0 {
			t.Fatalf("expected results[0] = 0, got %d", results[0])
		}
		if results[1] != 2 {
			t.Fatalf("expected results[1] = 2, got %d", results[1])
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
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
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// Check we got AggregateError with 2 errors
		aggErr, ok := err.(*AggregateError)
		if !ok {
			t.Fatalf("expected AggregateError, got %T", err)
		}
		if len(aggErr.Errors) != 2 {
			t.Fatalf("expected 2 errors, got %d", len(aggErr.Errors))
		}

		// Check all results were returned
		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d", len(results))
		}
		if results[0] != 10 || results[1] != 20 || results[2] != 30 {
			t.Fatalf("expected [10, 20, 30], got %v", results)
		}
	})

	t.Run("empty tasks", func(t *testing.T) {
		results, err := All[int](ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(results) != 0 {
			t.Fatalf("expected empty results, got %v", results)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		t1 := Task[int](func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			return 1, nil
		})

		results, err := All(ctx, t1)
		if err == nil {
			t.Fatal("expected context error")
		}

		// Results should still be returned (with zero value for cancelled task)
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
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

func TestAllSettled(t *testing.T) {
	ctx := context.Background()

	t.Run("mixed results", func(t *testing.T) {
		t1 := Task[int](func(ctx context.Context) (int, error) {
			return 1, nil
		})
		t2 := Task[int](func(ctx context.Context) (int, error) {
			return 0, errors.New("task 2 failed")
		})
		t3 := Task[int](func(ctx context.Context) (int, error) {
			return 3, nil
		})

		results := AllSettled(ctx, t1, t2, t3)

		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d", len(results))
		}

		if results[0].Err != nil || results[0].Value != 1 {
			t.Errorf("expected first result to be 1, got %v", results[0])
		}

		if results[1].Err == nil {
			t.Error("expected second result to have error")
		}

		if results[2].Err != nil || results[2].Value != 3 {
			t.Errorf("expected third result to be 3, got %v", results[2])
		}
	})
}
