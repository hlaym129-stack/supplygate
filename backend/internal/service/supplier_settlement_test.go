package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type supplierSettlementRepoStub struct {
	SupplierRepository
	profile         *SupplierProfile
	existing        *SupplierSettlementStatement
	statement       *SupplierSettlementStatement
	lastGenerated   *SupplierSettlementStatement
	lastFilters     SupplierSettlementListFilters
	confirmedAdjust float64
	confirmedReason string
	markedPaidInput SupplierSettlementPaymentInput
}

func (r *supplierSettlementRepoStub) GetProfileByID(ctx context.Context, id int64) (*SupplierProfile, error) {
	if r.profile == nil || r.profile.ID != id {
		return nil, ErrSupplierProfileNotFound
	}
	cp := *r.profile
	return &cp, nil
}

func (r *supplierSettlementRepoStub) GetProfileByUserID(ctx context.Context, userID int64) (*SupplierProfile, error) {
	if r.profile == nil || r.profile.UserID != userID {
		return nil, ErrSupplierProfileNotFound
	}
	cp := *r.profile
	return &cp, nil
}

func (r *supplierSettlementRepoStub) GetOpenSettlementStatementByPeriod(ctx context.Context, supplierID int64, periodStart time.Time) (*SupplierSettlementStatement, error) {
	if r.existing == nil || r.existing.SupplierID != supplierID || !r.existing.PeriodStart.Equal(periodStart) {
		return nil, ErrSupplierSettlementNotFound
	}
	cp := *r.existing
	return &cp, nil
}

func (r *supplierSettlementRepoStub) UpsertDraftSettlementStatement(ctx context.Context, statement *SupplierSettlementStatement) (*SupplierSettlementStatement, error) {
	cp := *statement
	r.lastGenerated = &cp
	cp.ID = 10
	cp.UsageAmount = 12.5
	cp.PayableAmount = 12.5
	return &cp, nil
}

func (r *supplierSettlementRepoStub) ListSettlementStatements(ctx context.Context, params pagination.PaginationParams, filters SupplierSettlementListFilters) ([]SupplierSettlementStatement, *pagination.PaginationResult, error) {
	r.lastFilters = filters
	items := []SupplierSettlementStatement{}
	if r.statement != nil {
		cp := *r.statement
		items = append(items, cp)
	}
	return items, &pagination.PaginationResult{Total: int64(len(items)), Page: params.Page, PageSize: params.Limit(), Pages: 1}, nil
}

func (r *supplierSettlementRepoStub) GetSettlementStatementByID(ctx context.Context, id int64) (*SupplierSettlementStatement, error) {
	if r.statement == nil || r.statement.ID != id {
		return nil, ErrSupplierSettlementNotFound
	}
	cp := *r.statement
	return &cp, nil
}

func (r *supplierSettlementRepoStub) ConfirmSettlementStatement(ctx context.Context, id int64, reviewerID int64, adjustmentAmount float64, adjustmentReason string) (*SupplierSettlementStatement, error) {
	r.confirmedAdjust = adjustmentAmount
	r.confirmedReason = adjustmentReason
	cp := *r.statement
	cp.Status = SupplierSettlementStatusConfirmed
	cp.AdjustmentAmount = adjustmentAmount
	cp.AdjustmentReason = adjustmentReason
	cp.PayableAmount = cp.UsageAmount + adjustmentAmount
	return &cp, nil
}

func (r *supplierSettlementRepoStub) MarkSettlementStatementPaid(ctx context.Context, id int64, paidBy int64, payment SupplierSettlementPaymentInput) (*SupplierSettlementStatement, error) {
	r.markedPaidInput = payment
	cp := *r.statement
	cp.Status = SupplierSettlementStatusPaid
	cp.Payment = &SupplierSettlementPayment{PaidAmount: payment.PaidAmount, PaidAt: *payment.PaidAt}
	return &cp, nil
}

func TestGenerateSettlementStatementRejectsLockedPeriod(t *testing.T) {
	periodStart, _ := ParseSupplierSettlementPeriodMonth("2026-05")
	repo := &supplierSettlementRepoStub{
		profile:  &SupplierProfile{ID: 9},
		existing: &SupplierSettlementStatement{ID: 3, SupplierID: 9, PeriodStart: *periodStart, Status: SupplierSettlementStatusConfirmed},
	}
	svc := NewSupplierService(repo, nil, nil, nil, nil, nil, nil)

	_, err := svc.GenerateSettlementStatement(context.Background(), SupplierSettlementGenerateInput{SupplierID: 9, PeriodMonth: "2026-05"}, 1)

	require.ErrorIs(t, err, ErrSupplierSettlementLocked)
	require.Nil(t, repo.lastGenerated)
}

func TestConfirmSettlementStatementAppliesAdjustment(t *testing.T) {
	repo := &supplierSettlementRepoStub{
		statement: &SupplierSettlementStatement{ID: 5, Status: SupplierSettlementStatusDraft, UsageAmount: 10},
	}
	svc := NewSupplierService(repo, nil, nil, nil, nil, nil, nil)

	out, err := svc.ConfirmSettlementStatement(context.Background(), 5, 1, SupplierSettlementConfirmInput{
		AdjustmentAmount: -1.25,
		AdjustmentReason: " rebate ",
	})

	require.NoError(t, err)
	require.Equal(t, SupplierSettlementStatusConfirmed, out.Status)
	require.Equal(t, 8.75, out.PayableAmount)
	require.Equal(t, -1.25, repo.confirmedAdjust)
	require.Equal(t, "rebate", repo.confirmedReason)
}

func TestMarkSettlementStatementPaidRequiresConfirmed(t *testing.T) {
	repo := &supplierSettlementRepoStub{
		statement: &SupplierSettlementStatement{ID: 5, Status: SupplierSettlementStatusPaid},
	}
	svc := NewSupplierService(repo, nil, nil, nil, nil, nil, nil)

	_, err := svc.MarkSettlementStatementPaid(context.Background(), 5, 1, SupplierSettlementPaymentInput{PaidAmount: 10})

	require.ErrorIs(t, err, ErrSupplierSettlementLocked)
}

func TestListMySettlementStatementsUsesSupplierVisibility(t *testing.T) {
	repo := &supplierSettlementRepoStub{
		profile:   &SupplierProfile{ID: 9, UserID: 77},
		statement: &SupplierSettlementStatement{ID: 5, SupplierID: 9, Status: SupplierSettlementStatusConfirmed},
	}
	svc := NewSupplierService(repo, nil, nil, nil, nil, nil, nil)

	out, _, err := svc.ListMySettlementStatements(context.Background(), 77, pagination.PaginationParams{Page: 1, PageSize: 20})

	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, int64(9), repo.lastFilters.SupplierID)
	require.True(t, repo.lastFilters.VisibleToSupplier)
}
