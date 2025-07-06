package providers

import (
	"github.com/remiges-tech/await/examples/kyc"
)

// CVLProvider implements KYC verification for CVL.
type CVLProvider struct{}

// NewCVLProvider creates a new CVL provider instance.
func NewCVLProvider() *CVLProvider {
	return &CVLProvider{}
}

// CheckKYC implements the KYCProvider interface for CVL.
func (c *CVLProvider) CheckKYC(panDetails kyc.PanDetails) (kyc.KYCStatus, error) {
	return kyc.KYCStatus{
		Status:    "VERIFIED",
		OtherInfo: nil,
	}, nil
}
