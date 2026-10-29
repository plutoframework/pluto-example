package config

import "github.com/plutoframework/pluto-example/internal/features"

type FeaturesConfig struct {
	PurchaseExport *bool
	PurchaseReview *bool
}

func (c FeaturesConfig) Overrides() map[features.Key]bool {
	values := make(map[features.Key]bool)
	if c.PurchaseReview != nil {
		values[features.PurchaseReview] = *c.PurchaseReview
	}

	if c.PurchaseExport != nil {
		values[features.PurchaseExport] = *c.PurchaseExport
	}

	return values
}
