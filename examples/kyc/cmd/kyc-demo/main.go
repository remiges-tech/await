package main

import (
	"context"
	"log"

	"github.com/remiges-tech/await/examples/kyc"
	"github.com/remiges-tech/await/examples/kyc/providers"
)

func main() {
	providers := map[string]kyc.KYCProvider{
		"NDML":  providers.NewNDMLProvider(),
		"CAMS":  providers.NewCAMSProvider(),
		"CVL":   providers.NewCVLProvider(),
		"KARVY": providers.NewKARVYProvider(),
	}

	coordinator := kyc.NewCoordinator(providers, kyc.DefaultCoordinatorConfig())

	ctx := context.Background()
	_, provider, allStatuses, err := coordinator.CheckKYC(ctx, kyc.PanDetails{
		PAN: "AAAAA1111A",
	})

	if err != nil {
		log.Printf("KYC failed: %v\n", err)
		for name, s := range allStatuses {
			log.Printf("%s: %s (attempts: %d)\n", name, s.Status, s.Attempts)
		}
	} else {
		log.Printf("KYC verified by %s\n", provider)
	}
}
