package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
)

func (s *SupplierService) GenerateSettlementStatement(ctx context.Context, input SupplierSettlementGenerateInput, operatorID int64) (*SupplierSettlementStatement, error) {
	if input.SupplierID <= 0 {
		return nil, infraerrors.BadRequest("SUPPLIER_SETTLEMENT_SUPPLIER_REQUIRED", "supplier is required")
	}
	if operatorID <= 0 {
		return nil, infraerrors.BadRequest("SUPPLIER_SETTLEMENT_OPERATOR_REQUIRED", "operator is required")
	}
	periodStart, periodEnd, err := supplierSettlementMonthRange(input.PeriodMonth)
	if err != nil {
		return nil, err
	}
	if _, err := s.supplierRepo.GetProfileByID(ctx, input.SupplierID); err != nil {
		return nil, err
	}
	existing, err := s.supplierRepo.GetOpenSettlementStatementByPeriod(ctx, input.SupplierID, periodStart)
	if err != nil && !errors.Is(err, ErrSupplierSettlementNotFound) {
		return nil, err
	}
	if existing != nil && existing.Status != SupplierSettlementStatusDraft {
		return nil, ErrSupplierSettlementLocked
	}
	statement := &SupplierSettlementStatement{
		SupplierID:  input.SupplierID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Status:      SupplierSettlementStatusDraft,
		CreatedBy:   operatorID,
	}
	return s.supplierRepo.UpsertDraftSettlementStatement(ctx, statement)
}

func (s *SupplierService) ListSettlementStatements(ctx context.Context, params pagination.PaginationParams, filters SupplierSettlementListFilters) ([]SupplierSettlementStatement, *pagination.PaginationResult, error) {
	if filters.Status != "" && !validSupplierSettlementStatus(filters.Status) {
		return nil, nil, ErrSupplierSettlementInvalid
	}
	return s.supplierRepo.ListSettlementStatements(ctx, params, filters)
}

func (s *SupplierService) ListMySettlementStatements(ctx context.Context, userID int64, params pagination.PaginationParams) ([]SupplierSettlementStatement, *pagination.PaginationResult, error) {
	profile, err := s.supplierRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	return s.supplierRepo.ListSettlementStatements(ctx, params, SupplierSettlementListFilters{
		SupplierID:        profile.ID,
		VisibleToSupplier: true,
	})
}

func (s *SupplierService) ConfirmSettlementStatement(ctx context.Context, id int64, operatorID int64, input SupplierSettlementConfirmInput) (*SupplierSettlementStatement, error) {
	if id <= 0 {
		return nil, ErrSupplierSettlementNotFound
	}
	if operatorID <= 0 {
		return nil, infraerrors.BadRequest("SUPPLIER_SETTLEMENT_OPERATOR_REQUIRED", "operator is required")
	}
	statement, err := s.supplierRepo.GetSettlementStatementByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if statement.Status != SupplierSettlementStatusDraft {
		return nil, ErrSupplierSettlementLocked
	}
	return s.supplierRepo.ConfirmSettlementStatement(ctx, id, operatorID, input.AdjustmentAmount, strings.TrimSpace(input.AdjustmentReason))
}

func (s *SupplierService) MarkSettlementStatementPaid(ctx context.Context, id int64, operatorID int64, input SupplierSettlementPaymentInput) (*SupplierSettlementStatement, error) {
	if id <= 0 {
		return nil, ErrSupplierSettlementNotFound
	}
	if operatorID <= 0 {
		return nil, infraerrors.BadRequest("SUPPLIER_SETTLEMENT_OPERATOR_REQUIRED", "operator is required")
	}
	statement, err := s.supplierRepo.GetSettlementStatementByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if statement.Status != SupplierSettlementStatusConfirmed {
		return nil, ErrSupplierSettlementLocked
	}
	if input.PaidAmount < 0 {
		return nil, infraerrors.BadRequest("SUPPLIER_SETTLEMENT_PAYMENT_INVALID", "paid amount must be >= 0")
	}
	if input.PaidAt == nil || input.PaidAt.IsZero() {
		now := timezone.Now()
		input.PaidAt = &now
	}
	input.PaymentReference = strings.TrimSpace(input.PaymentReference)
	input.PaymentNote = strings.TrimSpace(input.PaymentNote)
	return s.supplierRepo.MarkSettlementStatementPaid(ctx, id, operatorID, input)
}

func (s *SupplierService) VoidSettlementStatement(ctx context.Context, id int64, operatorID int64) (*SupplierSettlementStatement, error) {
	if id <= 0 {
		return nil, ErrSupplierSettlementNotFound
	}
	if operatorID <= 0 {
		return nil, infraerrors.BadRequest("SUPPLIER_SETTLEMENT_OPERATOR_REQUIRED", "operator is required")
	}
	statement, err := s.supplierRepo.GetSettlementStatementByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if statement.Status == SupplierSettlementStatusPaid || statement.Status == SupplierSettlementStatusVoided {
		return nil, ErrSupplierSettlementLocked
	}
	return s.supplierRepo.VoidSettlementStatement(ctx, id, operatorID)
}

func ParseSupplierSettlementPeriodMonth(month string) (*time.Time, error) {
	month = strings.TrimSpace(month)
	if month == "" {
		return nil, nil
	}
	start, _, err := supplierSettlementMonthRange(month)
	if err != nil {
		return nil, err
	}
	return &start, nil
}

func supplierSettlementMonthRange(month string) (time.Time, time.Time, error) {
	month = strings.TrimSpace(month)
	if month == "" {
		return time.Time{}, time.Time{}, infraerrors.BadRequest("SUPPLIER_SETTLEMENT_PERIOD_REQUIRED", "period month is required")
	}
	t, err := time.ParseInLocation("2006-01", month, timezone.Location())
	if err != nil {
		return time.Time{}, time.Time{}, infraerrors.BadRequest("SUPPLIER_SETTLEMENT_PERIOD_INVALID", "period month must use YYYY-MM format")
	}
	start := timezone.StartOfMonth(t)
	return start, start.AddDate(0, 1, 0), nil
}

func validSupplierSettlementStatus(status string) bool {
	switch status {
	case SupplierSettlementStatusDraft, SupplierSettlementStatusConfirmed, SupplierSettlementStatusPaid, SupplierSettlementStatusVoided:
		return true
	default:
		return false
	}
}

func isSupplierSettlementNotFound(err error) bool {
	return errors.Is(err, ErrSupplierSettlementNotFound) || errors.Is(err, sql.ErrNoRows)
}
