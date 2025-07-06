package providers

import (
	"github.com/remiges-tech/await/examples/kyc"
)

// NDMLProvider implements KYC verification for NDML.
type NDMLProvider struct{}

// NewNDMLProvider creates a new NDML provider instance.
func NewNDMLProvider() *NDMLProvider {
	return &NDMLProvider{}
}

// CheckKYC implements the KYCProvider interface for NDML.
func (n *NDMLProvider) CheckKYC(panDetails kyc.PanDetails) (kyc.KYCStatus, error) {
	return kyc.KYCStatus{
		Status:    "VERIFIED",
		OtherInfo: nil,
	}, nil
}
