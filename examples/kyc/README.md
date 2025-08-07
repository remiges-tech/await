# KYC Verification System Example

This example demonstrates how to use the go-await library to build a concurrent KYC (Know Your Customer) verification system that integrates with multiple KRA (KYC Registration Agency) providers.

## Overview

The system implements:
- Concurrent KYC verification across multiple providers (NDML, CAMS, CVL, KARVY)
- First successful provider wins, others are automatically cancelled
- Retry logic with exponential backoff for handling transient failures
- Provider-specific authentication and request/response handling
- Error handling and status tracking

## Architecture

```
Application Layer
    ↓
Coordination Layer (uses go-await)
    ↓
Provider Interface
    ↓
HTTP Calls to KRAs
```

### Components

1. **Provider Interface (`KYCProvider`)**: Common interface that all providers implement
2. **Coordinator**: Manages concurrent execution using go-await's `Any` function and retry functionality
3. **Provider Implementations**: Each KRA has its own implementation handling auth, request format, etc.
4. **Status Tracking**: Real-time tracking of attempts, errors, and timing for the winning provider

## Usage

```go
// Create providers
providers := map[string]kyc.KYCProvider{
    "NDML":  providers.NewNDMLProvider(ndmlConfig),
    "CAMS":  providers.NewCAMSProvider(camsConfig),
    "CVL":   providers.NewCVLProvider(cvlConfig),
    "KARVY": providers.NewKARVYProvider(karvyConfig),
}

// Create coordinator
coordinator := kyc.NewCoordinator(providers, config)

// Run KYC check - returns as soon as one provider succeeds
// Also returns tracking info for all providers
status, providerName, allStatuses, err := coordinator.CheckKYC(ctx, panDetails)

if err != nil {
    // All providers failed
    fmt.Printf("KYC verification failed: %v\n", err)
} else {
    // Success - other providers were cancelled
    fmt.Printf("KYC verified by %s in %v\n", providerName, status.TotalTime)
}

// Check status of all providers (for monitoring/debugging)
for name, provStatus := range allStatuses {
    fmt.Printf("%s: %s\n", name, provStatus.Status)
}

## Running the Example

1. Set environment variables for provider credentials:
```bash
export NDML_PASSWORD=your_password
export NDML_PASSKEY=your_passkey
# ... other provider credentials
```

2. Run the example:
```bash
cd examples/kyc
go run ./cmd/kyc-demo
```

The example includes:
- First-success verification (stop when first provider succeeds)
- Mock provider demonstration showing how the fastest successful provider wins

## Features Demonstrated

### 1. First-Success Pattern
All providers run concurrently using go-await's `Any`:
```go
// Use Any to get the first successful result
result, err := await.Any(ctx, tasks...)
// Other providers are automatically cancelled
```

### 2. Retry Logic
Each provider call is wrapped with retry logic:
```go
import "github.com/remiges-tech/await/retry"

result, err := retry.Do(ctx, checkKYC, retry.Options{
    MaxAttempts: 3,
    Strategy: &retry.ConstantDelay{
        Delay: 2 * time.Second,
    },
    OnRetry: func(attempt int, err error) {
        log.Printf("Attempt %d failed: %v", attempt, err)
    },
    RetryIf: IsRetryable, // Custom function to determine if error is retryable
})
```

### 3. Status Tracking
Real-time tracking of each provider's status:
```go
type ProviderStatus struct {
    Provider     KYCProvider
    Status       string        // "pending", "success", "failed"
    KYCResponse  KYCStatus
    Error        error
    Attempts     int
    LastAttempt  time.Time
    TotalTime    time.Duration
}
```

### 4. Error Classification
Error-based retry logic:
- Don't retry: Authentication errors, invalid PAN
- Do retry: Timeouts, rate limits, service unavailable

## Provider Implementations

Each provider demonstrates different patterns:
- **NDML**: Token-based auth with caching, JSON API
- **CAMS**: API key auth, JSON API
- **CVL**: Basic auth, XML-based API (SOAP-like)
- **KARVY**: Access/Secret key auth, nested JSON responses

## Implementation with go-await

1. **Concurrency Management**: No manual goroutine management required
2. **Retry Functionality**: Configurable retry with exponential backoff
3. **Type Safety**: Generic types provide compile-time type checking
4. **Context Support**: Automatic cancellation and timeout handling
5. **Error Handling**: Aggregate errors from multiple provider attempts
