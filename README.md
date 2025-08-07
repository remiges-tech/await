# await

A Promise-like async task patterns library for Go, providing JavaScript-style async utilities with Go's concurrency model.

## Features

- **Core Functions**: `All`, `Any`, `Race`
- **Retry Package**: Configurable retry with multiple strategies
- **Error Handling**: Aggregate errors, retry errors, context cancellation
- **Type-Safe**: Full generic support for type safety
- **Context-Aware**: All operations respect context cancellation

## Installation

```bash
go get github.com/remiges-tech/await
```

## Getting Started

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/remiges-tech/await"
)

func main() {
    ctx := context.Background()

    // Create async tasks
    t1 := await.Task[int](func(ctx context.Context) (int, error) {
        time.Sleep(100 * time.Millisecond)
        return 1, nil
    })

    t2 := await.Task[int](func(ctx context.Context) (int, error) {
        time.Sleep(200 * time.Millisecond)
        return 2, nil
    })

    // Wait for all tasks
    results, err := await.All(ctx, t1, t2)
    if err != nil {
        log.Fatal(err) // Function-level error (e.g., context cancelled)
    }

    // Check task results
    for i, result := range results {
        if result.Err != nil {
            fmt.Printf("Task %d failed: %v\n", i, result.Err)
        } else {
            fmt.Printf("Task %d result: %d\n", i, result.Value)
        }
    }
}
```

## API Reference

### Core Functions

#### All
Executes all tasks concurrently and returns complete information about each task's outcome. Returns a `Result[T]` for each task containing both value and error.

```go
results, err := await.All(ctx, task1, task2, task3)
// err is only for function-level errors (e.g., ErrNoTasks, context cancelled)
// Each result contains {Value: T, Err: error} for that specific task
```

#### Any
Returns when the first task succeeds. Returns error only if all tasks fail.

```go
result, err := await.Any(ctx, task1, task2, task3)
```

#### Race
Returns the first task to complete (success or failure).

```go
result, err := await.Race(ctx, task1, task2, task3)
```



### Mental Model

Understanding when to use each function:

- **`All`**: "Execute all, return complete information"
  - Use when you need to know the outcome of every task
  - Returns `Result[T]` for each task with both value and error
  - Function-level errors only for operational issues (empty tasks, context cancelled)

- **`Any`**: "Execute all, return first success"
  - Use when you just need one successful result (e.g., trying multiple mirrors)
  - Cancels remaining tasks on first success

- **`Race`**: "Execute all, return first completion"
  - Use when you want the fastest response, regardless of success/failure
  - Cancels remaining tasks on first completion


### Function Behavior Comparison

| Function | Philosophy | Returns When | Cancels Others | Error Behavior | Return Value |
|----------|------------|--------------|----------------|----------------|--------------|
| `All` | Execute all, return complete information | All tasks complete | No | Function errors for operational issues only | `([]Result[T], error)` - Each task's value and error paired |
| `Any` | Return first success | First success OR all fail | Yes (on success) | Only fails if all tasks fail | `(T, error)` - First success or aggregate error |
| `Race` | Return first completion | First completion | Yes | Returns first result as-is | `(T, error)` - First result (success or failure) |

### Differences

- **Error Handling**: `All` separates function errors from task errors, `Any` only fails if all tasks fail
- **Concurrency**: All functions run tasks concurrently
- **Cancellation**: Only `Any` and `Race` cancel remaining tasks; `Any` only on success, `Race` on any completion
- **Use Cases**:
  - `All`: When you need to know the outcome of every task
  - `Any`: When you need just one successful result (e.g., fallback servers)
  - `Race`: When you want the fastest result, regardless of success/failure

### Understanding the Unified All Function

#### Complete Information for Each Task
```go
// Example: Mixed success and failure
task1 := func(ctx context.Context) (string, error) { return "success1", nil }
task2 := func(ctx context.Context) (string, error) { return "", errors.New("failed") }
task3 := func(ctx context.Context) (string, error) { return "success3", nil }

results, err := await.All(ctx, task1, task2, task3)
// err = nil (no function-level error)
// results[0] = Result{Value: "success1", Err: nil}
// results[1] = Result{Value: "", Err: error("failed")}
// results[2] = Result{Value: "success3", Err: nil}

// Handle each result individually with complete information
for i, result := range results {
    if result.Err != nil {
        log.Printf("Task %d failed: %v", i, result.Err)
    } else {
        log.Printf("Task %d succeeded: %v", i, result.Value)
    }
}
```

#### Function-Level vs Task-Level Errors
```go
// Function-level error example (operational issue)
results, err := await.All(ctx) // No tasks provided
// err = ErrNoTasks
// results = nil

// Task-level errors (normal operation)
tasks := []await.Task[int]{
    func(ctx context.Context) (int, error) { return 0, errors.New("task error") },
}
results, err := await.All(ctx, tasks...)
// err = nil (function executed successfully)
// results[0].Err = error("task error") (task-specific error)
```

### Utility Functions

#### Retry
The retry package provides configurable retry logic with various strategies.

See the [retry package documentation](retry/README.md) for details.



## Error Types

- `ErrNoTasks`: Returned when no tasks are provided
- `ErrMaxAttemptsInvalid`: Returned when MaxAttempts is <= 0
- `AggregateError`: Contains multiple errors from failed tasks
- `RetryError`: Contains retry attempt information

## Examples

See the [examples/basic/main.go](examples/basic/main.go) file for usage examples.

## Running Tests

```bash
go test -v ./...
```

## License

MIT
