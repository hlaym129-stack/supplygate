//go:build unit

package service

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
)

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

func TestSupplierPricingChangeEffectiveAt(t *testing.T) {
	if err := timezone.Init("Asia/Shanghai"); err != nil {
		t.Fatalf("init timezone: %v", err)
	}

	tests := []struct {
		name      string
		submitted string
		want      string
	}{
		{
			name:      "before release uses same day release",
			submitted: "2026-05-05T06:59:59+08:00",
			want:      "2026-05-05T07:30:00+08:00",
		},
		{
			name:      "between seven and seven thirty uses same day release",
			submitted: "2026-05-05T07:15:00+08:00",
			want:      "2026-05-05T07:30:00+08:00",
		},
		{
			name:      "at release cutoff uses next day release",
			submitted: "2026-05-05T07:30:00+08:00",
			want:      "2026-05-06T07:30:00+08:00",
		},
		{
			name:      "after release cutoff uses next day release",
			submitted: "2026-05-05T19:45:00+08:00",
			want:      "2026-05-06T07:30:00+08:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			submitted, err := time.Parse(time.RFC3339, tt.submitted)
			if err != nil {
				t.Fatalf("parse submitted: %v", err)
			}
			want, err := time.Parse(time.RFC3339, tt.want)
			if err != nil {
				t.Fatalf("parse want: %v", err)
			}
			got := supplierPricingChangeEffectiveAt(submitted)
			if !got.Equal(want) {
				t.Fatalf("effectiveAt = %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
			}
		})
	}
}

func TestSupplierPricingChangeApprovedEffectiveAtDoesNotBackdate(t *testing.T) {
	if err := timezone.Init("Asia/Shanghai"); err != nil {
		t.Fatalf("init timezone: %v", err)
	}
	submitted, err := time.Parse(time.RFC3339, "2026-05-05T06:59:59+08:00")
	if err != nil {
		t.Fatalf("parse submitted: %v", err)
	}
	approved, err := time.Parse(time.RFC3339, "2026-05-05T08:00:00+08:00")
	if err != nil {
		t.Fatalf("parse approved: %v", err)
	}
	got := supplierPricingChangeApprovedEffectiveAt(submitted, approved)
	if !got.Equal(approved) {
		t.Fatalf("effectiveAt = %s, want approval time %s", got.Format(time.RFC3339), approved.Format(time.RFC3339))
	}
}
