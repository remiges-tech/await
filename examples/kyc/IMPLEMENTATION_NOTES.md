# KYC Implementation Notes

## Overview

This implementation demonstrates a KYC verification system using the go-await library for concurrent processing and retry functionality.

## Features Implemented

### 1. Core Architecture
- **Provider Interface**: Abstraction for KYC providers
- **Coordinator**: Manages concurrent execution using go-await
- **Status Tracking**: Real-time monitoring of each provider's progress

### 2. Provider Implementations
- **NDML**: Token-based authentication with caching
- **CAMS**: API key authentication
- **CVL**: XML-based requests (SOAP-style)
- **KARVY**: Access/Secret key authentication

### 3. Concurrency & Retry
- Uses go-await's `Any` for race-based execution
- Integrated with retry package for automatic retries
- Exponential backoff with configurable parameters
- Error-based retry logic

### 4. Additional Features
- **Race Pattern**: Stop as soon as ONE provider succeeds (using go-await's `Any`)
- **Timeout Handling**: Per-request timeout protection
- **Error Classification**: Differentiate retryable vs permanent errors
- **Context Propagation**: Proper cancellation support
- **Early Cancellation**: Other providers are cancelled once one succeeds

## Usage Pattern

### Race (First Success Wins)
```go
status, providerName, allStatuses, err := coordinator.CheckKYC(ctx, panDetails)
```

## go-await Integration Features

1. **Concurrency Management**: No manual goroutine management required
2. **Type Safety**: Generic types provide compile-time type checking
3. **Error Handling**: Aggregate errors and results from multiple providers
4. **Context Support**: Automatic cancellation and timeout handling

## Testing

The implementation includes tests demonstrating:
- Basic concurrent verification
- Retry on failure
- Race-based early termination (first success wins)
- Timeout handling
- All-providers-fail scenario

## Production Considerations

1. **Authentication**: Token caching reduces API calls
2. **Error Handling**: Proper classification for retry logic
3. **Monitoring**: Status tracking for observability
4. **Performance**: Concurrent execution with early termination
