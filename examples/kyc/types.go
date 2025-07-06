package kyc

import (
	"time"
)

// ProviderStatus tracks the state and history of a KYC verification attempt for a single provider.
type ProviderStatus struct {
	// Provider is the KYC provider implementation.
	Provider KYCProvider

	// Status indicates current state: "pending", "success", or "failed".
	Status string

	// KYCResponse contains the verification result when Status is "success".
	KYCResponse KYCStatus

	// Error stores the last error when Status is "failed".
	Error error

	// Attempts counts verification attempts for this provider.
	Attempts int

	// LastAttempt records the timestamp of the most recent attempt.
	LastAttempt time.Time

	// TotalTime measures duration from start to final result.
	TotalTime time.Duration
}

// KYCStatus represents the standardized response from any KYC provider.
type KYCStatus struct {
	// Status indicates the KYC verification result.
	Status string

	// OtherInfo contains additional data from provider.
	OtherInfo map[string]interface{}
}

// KYCProvider defines the interface that all KYC providers must implement.
type KYCProvider interface {
	// CheckKYC performs KYC verification and returns standardized status.
	CheckKYC(panDetails PanDetails) (KYCStatus, error)
}

// PanDetails contains the input data needed for KYC verification.
type PanDetails struct {
	// PAN is the 10-character Permanent Account Number.
	PAN string

	// AdditionalInfo contains provider-specific fields.
	AdditionalInfo map[string]string
}
