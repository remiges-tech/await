package kyc

import (
	"errors"
	"fmt"
)

// Common errors used across KYC providers.
var (
	// ErrAuthentication indicates authentication failure with provider.
	ErrAuthentication = errors.New("authentication failed")

	// ErrInvalidPAN indicates invalid PAN format.
	ErrInvalidPAN = errors.New("invalid PAN format")

	// ErrProviderUnavailable is returned when the KYC provider service is down.
	// Currently unused but kept for future provider implementations.
	ErrProviderUnavailable = errors.New("provider service unavailable")

	// ErrRateLimitExceeded is returned when too many requests are made to a provider.
	// Currently unused but kept for future provider implementations.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrInvalidResponse indicates malformed response from provider.
	ErrInvalidResponse = errors.New("invalid response from provider")

	// ErrTimeout is returned when a provider request times out.
	// Currently unused but kept for future provider implementations.
	ErrTimeout = errors.New("request timeout")
)

// ProviderError wraps provider-specific errors with additional context.
type ProviderError struct {
	Provider string
	Code     string
	Message  string
	Err      error
}

// Error returns the formatted error message.
func (e *ProviderError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s provider error (code: %s): %s - %v", e.Provider, e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s provider error (code: %s): %s", e.Provider, e.Code, e.Message)
}

// Unwrap returns the wrapped error.
func (e *ProviderError) Unwrap() error {
	return e.Err
}
