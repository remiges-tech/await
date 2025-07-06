package kyc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/remiges-tech/await"
	"github.com/remiges-tech/await/retry"
)

// CoordinatorConfig holds configuration for the KYC coordinator.
type CoordinatorConfig struct {
	MaxRetries     int
	RetryBackoff   time.Duration
	RequestTimeout time.Duration
}

// DefaultCoordinatorConfig returns default configuration.
func DefaultCoordinatorConfig() CoordinatorConfig {
	return CoordinatorConfig{
		MaxRetries:     3,
		RetryBackoff:   2 * time.Second,
		RequestTimeout: 30 * time.Second,
	}
}

// Coordinator manages concurrent KYC checks across multiple providers.
type Coordinator struct {
	providers map[string]KYCProvider
	config    CoordinatorConfig
}

// NewCoordinator creates a new KYC coordinator.
func NewCoordinator(providers map[string]KYCProvider, config CoordinatorConfig) *Coordinator {
	return &Coordinator{
		providers: providers,
		config:    config,
	}
}

// CheckKYC runs KYC checks and returns as soon as one provider succeeds.
// It also returns a map of all provider statuses for monitoring.
func (c *Coordinator) CheckKYC(ctx context.Context, panDetails PanDetails) (*ProviderStatus, string, map[string]*ProviderStatus, error) {
	if len(c.providers) == 0 {
		return nil, "", nil, fmt.Errorf("no providers configured")
	}

	tracking := make(map[string]*ProviderStatus)
	trackingMu := sync.Mutex{}

	type providerResult struct {
		status       *ProviderStatus
		providerName string
	}

	tasks := make([]await.Task[providerResult], 0, len(c.providers))

	for providerName, provider := range c.providers {
		name := providerName
		prov := provider

		task := func(ctx context.Context) (providerResult, error) {
			startTime := time.Now()
			status := &ProviderStatus{
				Provider: prov,
				Status:   "pending",
			}

			trackingMu.Lock()
			tracking[name] = status
			trackingMu.Unlock()

			checkKYC := func(ctx context.Context) (KYCStatus, error) {
				return prov.CheckKYC(panDetails)
			}

			retryOpts := retry.Options{
				MaxAttempts: c.config.MaxRetries,
				Strategy: &retry.ConstantDelay{
					Delay: c.config.RetryBackoff,
				},
				OnRetry: func(attempt int, err error) {
					trackingMu.Lock()
					status.Attempts = attempt
					status.LastAttempt = time.Now()
					trackingMu.Unlock()
					log.Printf("%s: Attempt %d failed: %v", name, attempt, err)
				},
				RetryIf: IsRetryable,
			}

			response, err := retry.Do(ctx, checkKYC, retryOpts)

			trackingMu.Lock()
			status.TotalTime = time.Since(startTime)
			if err != nil {
				status.Status = "failed"
				status.Error = err
				trackingMu.Unlock()
				return providerResult{}, err
			}
			status.Status = "success"
			status.KYCResponse = response
			trackingMu.Unlock()

			return providerResult{
				status:       status,
				providerName: name,
			}, nil
		}

		tasks = append(tasks, task)
	}

	result, err := await.Any(ctx, tasks...)
	if err != nil {
		return nil, "", tracking, fmt.Errorf("all providers failed: %w", err)
	}

	return result.status, result.providerName, tracking, nil
}

// IsRetryable determines if an error should trigger a retry.
func IsRetryable(err error) bool {
	switch {
	case errors.Is(err, ErrAuthentication):
		return false
	case errors.Is(err, ErrInvalidPAN):
		return false
	case errors.Is(err, ErrInvalidResponse):
		return false
	}

	var provErr *ProviderError
	if errors.As(err, &provErr) && len(provErr.Code) > 0 && provErr.Code[0] == '4' {
		return false
	}

	return true
}
