package dto

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestChannelModelPricingFromService(t *testing.T) {
	inputPrice := 0.000002
	outputPrice := 0.000008
	cacheWritePrice := 0.000003
	cacheReadPrice := 0.000001
	imageOutputPrice := 0.000011
	perRequestPrice := 0.25
	maxTokens := 8192

	got := ChannelModelPricingFromService(&service.ChannelModelPricing{
		ID:               12,
		Platform:         service.PlatformOpenAI,
		Models:           []string{"gpt-5.5"},
		BillingMode:      service.BillingModeToken,
		InputPrice:       &inputPrice,
		OutputPrice:      &outputPrice,
		CacheWritePrice:  &cacheWritePrice,
		CacheReadPrice:   &cacheReadPrice,
		ImageOutputPrice: &imageOutputPrice,
		PerRequestPrice:  &perRequestPrice,
		Intervals: []service.PricingInterval{{
			ID:              3,
			MinTokens:       0,
			MaxTokens:       &maxTokens,
			TierLabel:       "default",
			InputPrice:      &inputPrice,
			OutputPrice:     &outputPrice,
			CacheWritePrice: &cacheWritePrice,
			CacheReadPrice:  &cacheReadPrice,
			PerRequestPrice: &perRequestPrice,
			SortOrder:       1,
		}},
	})

	if got.Platform != service.PlatformOpenAI {
		t.Fatalf("expected platform %q, got %q", service.PlatformOpenAI, got.Platform)
	}
	if got.BillingMode != string(service.BillingModeToken) {
		t.Fatalf("expected billing mode %q, got %q", service.BillingModeToken, got.BillingMode)
	}
	if len(got.Models) != 1 || got.Models[0] != "gpt-5.5" {
		t.Fatalf("unexpected models: %#v", got.Models)
	}
	if got.InputPrice == nil || *got.InputPrice != inputPrice {
		t.Fatalf("unexpected input price: %#v", got.InputPrice)
	}
	if got.OutputPrice == nil || *got.OutputPrice != outputPrice {
		t.Fatalf("unexpected output price: %#v", got.OutputPrice)
	}
	if got.CacheWritePrice == nil || *got.CacheWritePrice != cacheWritePrice {
		t.Fatalf("unexpected cache write price: %#v", got.CacheWritePrice)
	}
	if got.CacheReadPrice == nil || *got.CacheReadPrice != cacheReadPrice {
		t.Fatalf("unexpected cache read price: %#v", got.CacheReadPrice)
	}
	if got.ImageOutputPrice == nil || *got.ImageOutputPrice != imageOutputPrice {
		t.Fatalf("unexpected image output price: %#v", got.ImageOutputPrice)
	}
	if got.PerRequestPrice == nil || *got.PerRequestPrice != perRequestPrice {
		t.Fatalf("unexpected per request price: %#v", got.PerRequestPrice)
	}
	if len(got.Intervals) != 1 {
		t.Fatalf("expected 1 interval, got %d", len(got.Intervals))
	}
	if got.Intervals[0].MaxTokens == nil || *got.Intervals[0].MaxTokens != maxTokens {
		t.Fatalf("unexpected max tokens: %#v", got.Intervals[0].MaxTokens)
	}
}

func TestSupplierAccountPricingRevisionFromService(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	reviewedBy := int64(88)
	inputPrice := 0.000004

	got := SupplierAccountPricingRevisionFromService(&service.SupplierAccountPricingRevision{
		ID:           7,
		AccountID:    9,
		SupplierID:   11,
		RevisionKind: service.SupplierPricingRevisionKindChange,
		Status:       service.SupplierPricingRevisionStatusApproved,
		Pricing: []service.ChannelModelPricing{{
			Platform:    service.PlatformOpenAI,
			Models:      []string{"gpt-5.5"},
			BillingMode: service.BillingModeToken,
			InputPrice:  &inputPrice,
		}},
		SubmitNote:  "note",
		ReviewNote:  "ok",
		ReviewedBy:  &reviewedBy,
		ReviewedAt:  &now,
		EffectiveAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	if got == nil {
		t.Fatal("expected non-nil revision dto")
	}
	if got.AccountID != 9 || got.SupplierID != 11 {
		t.Fatalf("unexpected ids: %#v", got)
	}
	if got.RevisionKind != service.SupplierPricingRevisionKindChange {
		t.Fatalf("unexpected revision kind: %q", got.RevisionKind)
	}
	if got.Status != service.SupplierPricingRevisionStatusApproved {
		t.Fatalf("unexpected status: %q", got.Status)
	}
	if len(got.Pricing) != 1 || len(got.Pricing[0].Models) != 1 || got.Pricing[0].Models[0] != "gpt-5.5" {
		t.Fatalf("unexpected pricing payload: %#v", got.Pricing)
	}
	if got.Pricing[0].InputPrice == nil || *got.Pricing[0].InputPrice != inputPrice {
		t.Fatalf("unexpected input price: %#v", got.Pricing[0].InputPrice)
	}
}

func TestSupplierSettlementStatementFromServiceUsesJSONReadySupplier(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	got := SupplierSettlementStatementFromService(&service.SupplierSettlementStatement{
		ID:            5,
		SupplierID:    9,
		PeriodStart:   now,
		PeriodEnd:     now.Add(30 * 24 * time.Hour),
		Status:        service.SupplierSettlementStatusPaid,
		PayableAmount: 70,
		Supplier: &service.SupplierProfile{
			ID:           9,
			UserID:       12,
			CompanyName:  "Demo Settlement Supplier Ltd.",
			ContactEmail: "demo.supplier.settlement@supplygate.local",
			User:         &service.User{ID: 12, Email: "demo.supplier.settlement@supplygate.local"},
		},
		Payment: &service.SupplierSettlementPayment{
			ID:               1,
			StatementID:      5,
			SupplierID:       9,
			PaidAmount:       70,
			PaidAt:           now,
			PaymentReference: "DEMO-PAYOUT-202603-001",
			CreatedAt:        now,
		},
	})

	if got == nil {
		t.Fatal("expected non-nil statement dto")
	}
	supplier, ok := got.Supplier.(map[string]any)
	if !ok {
		t.Fatalf("expected supplier map dto, got %T", got.Supplier)
	}
	if supplier["company_name"] != "Demo Settlement Supplier Ltd." {
		t.Fatalf("expected snake_case supplier company, got %#v", supplier)
	}
	if supplier["contact_email"] != "demo.supplier.settlement@supplygate.local" {
		t.Fatalf("expected snake_case supplier contact email, got %#v", supplier)
	}
	if got.Payment == nil || got.Payment.PaymentReference != "DEMO-PAYOUT-202603-001" {
		t.Fatalf("unexpected payment dto: %#v", got.Payment)
	}
}
