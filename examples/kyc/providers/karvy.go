package providers

import (
	"github.com/remiges-tech/await/examples/kyc"
)

// KARVYProvider implements KYC verification for KARVY.
type KARVYProvider struct{}

// NewKARVYProvider creates a new KARVY provider instance.
func NewKARVYProvider() *KARVYProvider {
	return &KARVYProvider{}
}

// CheckKYC implements the KYCProvider interface for KARVY.
func (k *KARVYProvider) CheckKYC(panDetails kyc.PanDetails) (kyc.KYCStatus, error) {
	return kyc.KYCStatus{
		Status:    "VERIFIED",
		OtherInfo: nil,
	}, nil
}
