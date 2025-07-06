# await

A Promise-like async task patterns library for Go, providing JavaScript-style async utilities with Go's concurrency model.

## Features

- **Core Functions**: `All`, `Any`, `Race`, `AllSettled`
- **Utilities**: `Retry`
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
        log.Fatal(err)
    }
    fmt.Println(results) // [1, 2]
}
```

## API Reference

### Core Functions

#### All
Waits for all tasks to complete. Returns all results (including partial results from successful tasks) and an AggregateError if any task fails.

```go
results, err := await.All(ctx, task1, task2, task3)
// If task2 fails: results = [result1, zero_value, result3], err = AggregateError
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

#### AllSettled
Returns all results regardless of success/failure.

```go
results := await.AllSettled(ctx, task1, task2, task3)
for _, result := range results {
    if result.Err != nil {
        // Handle error
    } else {
        // Use result.Value
    }
}
```


### Mental Model

Understanding when to use each function:

- **`All`**: "Execute all, succeed together or fail together"
  - Use when you need all results and any failure should be treated as total failure
  - Returns partial results on error (with zero values for failed tasks)

- **`AllSettled`**: "Execute all, return all outcomes"
  - Use when you need to know the outcome of every task, regardless of failures
  - Returns `Result[T]` for each task with both value and error

- **`Any`**: "Execute all, return first success"
  - Use when you just need one successful result (e.g., trying multiple mirrors)

- **`Race`**: "Execute all, return first completion"
  - Use when you want the fastest response, regardless of success/failure


### Function Behavior Comparison

| Function | Philosophy | Returns When | Cancels Others | Error Behavior | Return Value |
|----------|------------|--------------|----------------|----------------|--------------|
| `All` | Succeed together or fail together | All tasks complete | No | Returns AggregateError if any fail | `([]T, error)` - Partial results with zero values for failures |
| `AllSettled` | Return all outcomes | All tasks complete | No | Never returns error | `[]Result[T]` - Each task's value and error paired together |
| `Any` | Return first success | First success OR all fail | Yes (on success) | Only fails if all tasks fail | `(T, error)` - First success or aggregate error |
| `Race` | Return first completion | First completion | Yes | Returns first result as-is | `(T, error)` - First result (success or failure) |

### Differences

- **Error Handling**: `All` returns partial results with AggregateError, `AllSettled` never fails, and `Any` only fails if all tasks fail
- **Concurrency**: All functions run tasks concurrently
- **Cancellation**: Only `Any` and `Race` cancel remaining tasks; `Any` only on success, `Race` on any completion
- **Use Cases**:
  - `All`: When you need all results and any failure is critical
  - `AllSettled`: When you want all results regardless of individual failures
  - `Any`: When you need just one successful result (e.g., fallback servers)
  - `Race`: When you want the fastest result, regardless of success/failure

### Understanding All vs AllSettled

#### All - Partial Results Behavior
```go
// Example: Mixed success and failure
task1 := func(ctx context.Context) (string, error) { return "success1", nil }
task2 := func(ctx context.Context) (string, error) { return "", errors.New("failed") }
task3 := func(ctx context.Context) (string, error) { return "success3", nil }

results, err := await.All(ctx, task1, task2, task3)
// results = ["success1", "", "success3"]  // Note: empty string for failed task
// err = &AggregateError{Errors: [error("failed")]}

// Problem: Can't tell which task failed or if "" is a real value vs failure
```

#### AllSettled - Complete Information
```go
// Same tasks as above
settled := await.AllSettled(ctx, task1, task2, task3)
// settled[0] = Result{Value: "success1", Err: nil}
// settled[1] = Result{Value: "", Err: error("failed")}
// settled[2] = Result{Value: "success3", Err: nil}

// Now you can handle each result individually
for i, result := range settled {
    if result.Err != nil {
        log.Printf("Task %d failed: %v", i, result.Err)
    } else {
        log.Printf("Task %d succeeded: %v", i, result.Value)
    }
}
```

### Utility Functions

#### Retry
The retry package provides configurable retry logic with various strategies.

See the [retry package documentation](retry/README.md) for details.



## Error Types

- `ErrNoTasks`: Returned when no tasks are provided
- `ErrMaxRetriesExceeded`: Returned when retry limit is reached
- `AggregateError`: Contains multiple errors from failed tasks
- `RetryError`: Contains retry attempt information

## Examples

See the [example/main.go](example/main.go) file for usage examples.

## Running Tests

```bash
go test -v ./...
```

## License

MIT
