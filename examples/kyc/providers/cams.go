package providers

import (
	"github.com/remiges-tech/await/examples/kyc"
)

// CAMSProvider implements KYC verification for CAMS.
type CAMSProvider struct{}

// NewCAMSProvider creates a new CAMS provider instance.
func NewCAMSProvider() *CAMSProvider {
	return &CAMSProvider{}
}

// CheckKYC implements the KYCProvider interface for CAMS.
func (c *CAMSProvider) CheckKYC(panDetails kyc.PanDetails) (kyc.KYCStatus, error) {
	return kyc.KYCStatus{
		Status:    "VERIFIED",
		OtherInfo: nil,
	}, nil
}
