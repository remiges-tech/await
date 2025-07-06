package kyc_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/remiges-tech/await/examples/kyc"
)

// MockProvider for testing
type MockProvider struct {
	name         string
	shouldFail   bool
	failCount    int
	attemptCount int
	delay        time.Duration
}

func (m *MockProvider) CheckKYC(panDetails kyc.PanDetails) (kyc.KYCStatus, error) {
	m.attemptCount++

	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if m.shouldFail && m.attemptCount <= m.failCount {
		return kyc.KYCStatus{}, fmt.Errorf("mock provider %s failed on attempt %d", m.name, m.attemptCount)
	}

	return kyc.KYCStatus{
		Status:    "VERIFIED",
		OtherInfo: nil,
	}, nil
}

func TestCoordinatorWithRetry(t *testing.T) {
	providers := map[string]kyc.KYCProvider{
		"FastFlaky":  &MockProvider{name: "FastFlaky", shouldFail: true, failCount: 1, delay: 5 * time.Millisecond},
		"SlowStable": &MockProvider{name: "SlowStable", shouldFail: false, delay: 100 * time.Millisecond},
		"AlwaysFail": &MockProvider{name: "AlwaysFail", shouldFail: true, failCount: 10, delay: 10 * time.Millisecond},
	}

	config := kyc.CoordinatorConfig{
		MaxRetries:     3,
		RetryBackoff:   10 * time.Millisecond,
		RequestTimeout: 1 * time.Second,
	}
	coordinator := kyc.NewCoordinator(providers, config)

	ctx := context.Background()
	panDetails := kyc.PanDetails{
		PAN: "RETRY123X",
	}

	status, providerName, allStatuses, err := coordinator.CheckKYC(ctx, panDetails)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if providerName != "FastFlaky" {
		t.Errorf("Expected FastFlaky to win after retry, got %s", providerName)
	}

	if status.Attempts != 1 {
		t.Errorf("Expected 1 retry attempt for FastFlaky, got %d", status.Attempts)
	}

	provider := providers["FastFlaky"].(*MockProvider)
	if provider.attemptCount != 2 {
		t.Errorf("Expected 2 total attempts for FastFlaky, got %d", provider.attemptCount)
	}

	if len(allStatuses) != len(providers) {
		t.Errorf("Expected %d providers in tracking map, got %d", len(providers), len(allStatuses))
	}

	for name, status := range allStatuses {
		t.Logf("%s: Status=%s, Attempts=%d", name, status.Status, status.Attempts)
		if name == "FastFlaky" && status.Status != "success" {
			t.Errorf("Expected FastFlaky to have success status")
		}
	}
}

func TestCoordinatorFirstSuccess(t *testing.T) {
	providers := map[string]kyc.KYCProvider{
		"Fast1": &MockProvider{name: "Fast1", delay: 5 * time.Millisecond},
		"Slow1": &MockProvider{name: "Slow1", delay: 100 * time.Millisecond},
		"Slow2": &MockProvider{name: "Slow2", delay: 200 * time.Millisecond},
		"Slow3": &MockProvider{name: "Slow3", delay: 300 * time.Millisecond},
	}

	config := kyc.DefaultCoordinatorConfig()
	coordinator := kyc.NewCoordinator(providers, config)

	ctx := context.Background()
	panDetails := kyc.PanDetails{
		PAN: "RACE123X",
	}

	startTime := time.Now()
	status, providerName, allStatuses, err := coordinator.CheckKYC(ctx, panDetails)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if duration > 50*time.Millisecond {
		t.Errorf("Race check took too long: %v", duration)
	}

	if providerName != "Fast1" {
		t.Errorf("Expected Fast1 to win the race, got %s", providerName)
	}

	if status.Status != "success" {
		t.Errorf("Expected success status, got %s", status.Status)
	}

	t.Logf("Race completed in %v, winner: %s", duration, providerName)

	successCount := 0
	pendingCount := 0
	failedCount := 0
	for name, status := range allStatuses {
		switch status.Status {
		case "success":
			successCount++
		case "pending":
			pendingCount++
		case "failed":
			failedCount++
		}
		t.Logf("%s: Status=%s, Time=%v", name, status.Status, status.TotalTime)
	}

	if successCount < 1 {
		t.Errorf("Expected at least 1 success in tracking")
	}
	t.Logf("Tracking summary: %d success, %d pending, %d failed", successCount, pendingCount, failedCount)
}

func TestCoordinatorAllFail(t *testing.T) {
	providers := map[string]kyc.KYCProvider{
		"Fail1": &MockProvider{name: "Fail1", shouldFail: true, failCount: 10, delay: 5 * time.Millisecond},
		"Fail2": &MockProvider{name: "Fail2", shouldFail: true, failCount: 10, delay: 10 * time.Millisecond},
		"Fail3": &MockProvider{name: "Fail3", shouldFail: true, failCount: 10, delay: 15 * time.Millisecond},
	}

	config := kyc.DefaultCoordinatorConfig()
	config.MaxRetries = 2
	coordinator := kyc.NewCoordinator(providers, config)

	ctx := context.Background()
	panDetails := kyc.PanDetails{
		PAN: "ALLFAIL1",
	}

	status, providerName, allStatuses, err := coordinator.CheckKYC(ctx, panDetails)

	if err == nil {
		t.Errorf("Expected error when all providers fail")
	}

	if len(allStatuses) != len(providers) {
		t.Errorf("Expected %d providers in tracking, got %d", len(providers), len(allStatuses))
	}

	if status != nil {
		t.Errorf("Expected nil status when all fail, got %v", status)
	}

	if providerName != "" {
		t.Errorf("Expected empty provider name when all fail, got %s", providerName)
	}

	if !strings.Contains(err.Error(), "all providers failed") {
		t.Errorf("Expected 'all providers failed' error, got: %v", err)
	}
}

func TestCoordinatorTimeout(t *testing.T) {
	providers := map[string]kyc.KYCProvider{
		"TimeoutProvider": &MockProvider{name: "TimeoutProvider", delay: 2 * time.Second},
	}

	config := kyc.CoordinatorConfig{
		MaxRetries:     1,
		RetryBackoff:   10 * time.Millisecond,
		RequestTimeout: 50 * time.Millisecond,
	}
	coordinator := kyc.NewCoordinator(providers, config)

	ctx := context.Background()
	panDetails := kyc.PanDetails{
		PAN: "TIMEOUT99",
	}

	status, providerName, allStatuses, err := coordinator.CheckKYC(ctx, panDetails)

	if err == nil {
		t.Errorf("Expected error due to timeout, got success")
	}

	if allStatuses["TimeoutProvider"] == nil {
		t.Errorf("Expected TimeoutProvider in tracking")
	} else if allStatuses["TimeoutProvider"].Status != "failed" {
		t.Errorf("Expected TimeoutProvider to have failed status")
	}

	if providerName != "" {
		t.Errorf("Expected empty provider name on failure, got %s", providerName)
	}

	if status != nil {
		t.Errorf("Expected nil status on failure, got %v", status)
	}

	if err != nil && !contains(err.Error(), "request timeout") {
		t.Errorf("Expected timeout error, got %v", err)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
