//go:build unit

package service

import "testing"

func TestValidateSupplierSettlementPricingAllowsEmptyPriceFields(t *testing.T) {
	pricing, err := validateSupplierSettlementPricing([]ChannelModelPricing{
		{
			Platform:         PlatformOpenAI,
			Models:           []string{"gpt-5.5"},
			BillingMode:      BillingModeToken,
			InputPrice:       testPtrFloat64(5e-6),
			OutputPrice:      testPtrFloat64(30e-6),
			CacheWritePrice:  nil,
			CacheReadPrice:   testPtrFloat64(0.5e-6),
			ImageOutputPrice: nil,
		},
	}, []string{"gpt-5.5"}, PlatformOpenAI)
	if err != nil {
		t.Fatalf("expected pricing with empty fields to be valid, got error: %v", err)
	}
	if len(pricing) != 1 {
		t.Fatalf("expected 1 pricing entry, got %d", len(pricing))
	}
	if pricing[0].CacheWritePrice != nil {
		t.Fatalf("expected empty cache write price to remain nil, got %#v", pricing[0].CacheWritePrice)
	}
	if pricing[0].ImageOutputPrice != nil {
		t.Fatalf("expected empty image output price to remain nil, got %#v", pricing[0].ImageOutputPrice)
	}
}
