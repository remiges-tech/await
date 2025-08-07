# Retry Package

A retry package for Go with customizable strategies and conditions.

## Features

- Multiple built-in retry strategies (Exponential, Linear, Constant)
- Custom retry strategies
- Conditional retry based on error types
- Context support with cancellation
- Retry callbacks for monitoring
- Permanent error marking to prevent unnecessary retries
- Type-safe with generics

## Installation

```bash
go get github.com/remiges-tech/await/retry
```

## Usage

### Basic Usage

```go
import "github.com/remiges-tech/await/retry"

// Retry with default options
result, err := retry.Do(ctx, func(ctx context.Context) (string, error) {
    return fetchData()
}, retry.DefaultOptions())
```

### Exponential Backoff

```go
result, err := retry.Do(ctx, fetchData, retry.Options{
    Strategy: &retry.ExponentialBackoff{
        InitialDelay: 100 * time.Millisecond,
        Multiplier:   2,
        MaxDelay:     10 * time.Second,
    },
    MaxAttempts: 5,
})
```

### Linear Backoff

```go
result, err := retry.Do(ctx, fetchData, retry.Options{
    Strategy: &retry.LinearBackoff{
        InitialDelay: 100 * time.Millisecond,
        Increment:    100 * time.Millisecond,
    },
    MaxAttempts: 3,
})
```

### Custom Strategy

```go
customStrategy := &retry.CustomStrategy{
    DelayFunc: func(attempt int) time.Duration {
        // Your custom delay logic
        return time.Duration(attempt*attempt) * 100 * time.Millisecond
    },
    ShouldRetryFunc: func(attempt int, err error) bool {
        // Your custom retry logic
        return attempt < 5 && !errors.Is(err, ErrPermanent)
    },
}

result, err := retry.Do(ctx, fetchData, retry.Options{
    Strategy:    customStrategy,
    MaxAttempts: 10,
})
```

### Conditional Retry

```go
// Retry only on specific errors
result, err := retry.Do(ctx, fetchData, retry.Options{
    Strategy:    &retry.ConstantDelay{Delay: 1 * time.Second},
    MaxAttempts: 3,
    RetryIf:     retry.RetryIf(io.EOF, net.ErrClosed),
})

```

### Permanent Errors

```go
func fetchData(ctx context.Context) (string, error) {
    if invalidConfig {
        // This error won't be retried
        return "", retry.Permanent(errors.New("invalid configuration"))
    }
    // ... rest of the logic
}
```

### Monitoring Retries

```go
result, err := retry.Do(ctx, fetchData, retry.Options{
    Strategy: &retry.ExponentialBackoff{
        InitialDelay: 100 * time.Millisecond,
        Multiplier:   2,
        MaxDelay:     30 * time.Second,
    },
    MaxAttempts: 5,
    OnRetry: func(attempt int, err error) {
        log.Printf("Retry attempt %d after error: %v", attempt, err)
    },
})
```

## Built-in Strategies

### ExponentialBackoff
Delays increase exponentially with each attempt.

### LinearBackoff
Delays increase linearly by a fixed increment.

### ConstantDelay
Same delay between all attempts.

### NoDelay
Retry immediately without any delay.

## Error Types

- `RetryError`: Returned when all retry attempts fail
- `ErrMaxAttemptsInvalid`: Returned when MaxAttempts is <= 0
- `ErrPermanent`: Used to mark errors that should not be retried

## License

Same as the parent go-await project.
