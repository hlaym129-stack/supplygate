package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

const supplierSettlementStatementColumns = `
	s.id, s.supplier_id, s.period_start, s.period_end, s.status,
	s.usage_amount, s.adjustment_amount, s.adjustment_reason, s.payable_amount,
	s.request_count, s.input_tokens, s.output_tokens, s.total_tokens,
	s.created_by, s.confirmed_by, s.confirmed_at, s.paid_by, s.paid_at, s.voided_by, s.voided_at,
	s.created_at, s.updated_at,
	sp.company_name, sp.contact_email, u.email,
	p.id, p.statement_id, p.supplier_id, p.paid_amount, p.paid_at, p.payment_reference, p.payment_note, p.created_by, p.created_at
`

func (r *supplierRepository) GetSettlementStatementByID(ctx context.Context, id int64) (*service.SupplierSettlementStatement, error) {
	row, err := querySingleSupplierSettlementStatement(ctx, r.sql, fmt.Sprintf(`
		SELECT %s
		FROM supplier_settlement_statements s
		JOIN supplier_profiles sp ON sp.id = s.supplier_id
		LEFT JOIN users u ON u.id = sp.user_id
		LEFT JOIN supplier_settlement_payments p ON p.statement_id = s.id
		WHERE s.id = $1
		LIMIT 1
	`, supplierSettlementStatementColumns), id)
	return translateSupplierSettlementNotFound(row, err)
}

func (r *supplierRepository) GetOpenSettlementStatementByPeriod(ctx context.Context, supplierID int64, periodStart time.Time) (*service.SupplierSettlementStatement, error) {
	row, err := querySingleSupplierSettlementStatement(ctx, r.sql, fmt.Sprintf(`
		SELECT %s
		FROM supplier_settlement_statements s
		JOIN supplier_profiles sp ON sp.id = s.supplier_id
		LEFT JOIN users u ON u.id = sp.user_id
		LEFT JOIN supplier_settlement_payments p ON p.statement_id = s.id
		WHERE s.supplier_id = $1 AND s.period_start = $2 AND s.status <> $3
		ORDER BY s.id DESC
		LIMIT 1
	`, supplierSettlementStatementColumns), supplierID, periodStart, service.SupplierSettlementStatusVoided)
	return translateSupplierSettlementNotFound(row, err)
}

func (r *supplierRepository) UpsertDraftSettlementStatement(ctx context.Context, statement *service.SupplierSettlementStatement) (*service.SupplierSettlementStatement, error) {
	if statement == nil {
		return nil, service.ErrSupplierSettlementInvalid
	}
	statementID := statement.ID
	if statementID == 0 {
		if existing, err := r.GetOpenSettlementStatementByPeriod(ctx, statement.SupplierID, statement.PeriodStart); err == nil && existing != nil {
			statementID = existing.ID
		} else if err != nil && !errors.Is(err, service.ErrSupplierSettlementNotFound) {
			return nil, err
		}
	}

	var id int64
	err := scanSingleRow(ctx, r.sql, `
		WITH agg AS (
			SELECT
				COALESCE(SUM(`+supplierSettlementCostExpr+`), 0) AS usage_amount,
				COUNT(*) AS request_count,
				COALESCE(SUM(input_tokens), 0) AS input_tokens,
				COALESCE(SUM(output_tokens), 0) AS output_tokens,
				COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) AS total_tokens
			FROM usage_logs
			WHERE supplier_id = $1 AND created_at >= $2 AND created_at < $3
		),
		updated AS (
			UPDATE supplier_settlement_statements s
			SET usage_amount = agg.usage_amount,
				adjustment_amount = 0,
				adjustment_reason = '',
				payable_amount = agg.usage_amount,
				request_count = agg.request_count,
				input_tokens = agg.input_tokens,
				output_tokens = agg.output_tokens,
				total_tokens = agg.total_tokens,
				updated_at = NOW()
			FROM agg
			WHERE s.id = $5 AND s.status = $6
			RETURNING s.id
		),
		inserted AS (
			INSERT INTO supplier_settlement_statements (
				supplier_id, period_start, period_end, status, usage_amount, adjustment_amount,
				adjustment_reason, payable_amount, request_count, input_tokens, output_tokens, total_tokens, created_by
			)
			SELECT $1, $2, $3, $6, agg.usage_amount, 0, '', agg.usage_amount,
				agg.request_count, agg.input_tokens, agg.output_tokens, agg.total_tokens, $4
			FROM agg
			WHERE $5 = 0
			RETURNING id
		)
		SELECT id FROM updated
		UNION ALL
		SELECT id FROM inserted
		LIMIT 1
	`, []any{
		statement.SupplierID,
		statement.PeriodStart,
		statement.PeriodEnd,
		statement.CreatedBy,
		statementID,
		service.SupplierSettlementStatusDraft,
	}, &id)
	if err != nil {
		return nil, err
	}
	return r.GetSettlementStatementByID(ctx, id)
}

func (r *supplierRepository) ListSettlementStatements(ctx context.Context, params pagination.PaginationParams, filters service.SupplierSettlementListFilters) ([]service.SupplierSettlementStatement, *pagination.PaginationResult, error) {
	where, args := supplierSettlementFiltersSQL(filters)
	var total int64
	if err := scanSingleRow(ctx, r.sql, `SELECT COUNT(*) FROM supplier_settlement_statements s `+where, args, &total); err != nil {
		return nil, nil, err
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM supplier_settlement_statements s
		JOIN supplier_profiles sp ON sp.id = s.supplier_id
		LEFT JOIN users u ON u.id = sp.user_id
		LEFT JOIN supplier_settlement_payments p ON p.statement_id = s.id
		%s
		ORDER BY s.period_start DESC, s.id DESC
		LIMIT $%d OFFSET $%d
	`, supplierSettlementStatementColumns, where, len(args)+1, len(args)+2)
	args = append(args, params.Limit(), params.Offset())
	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]service.SupplierSettlementStatement, 0)
	for rows.Next() {
		item, err := scanSupplierSettlementStatement(rows)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return out, paginationResultFromTotal(total, params), nil
}

func (r *supplierRepository) ConfirmSettlementStatement(ctx context.Context, id int64, reviewerID int64, adjustmentAmount float64, adjustmentReason string) (*service.SupplierSettlementStatement, error) {
	var updatedID int64
	err := scanSingleRow(ctx, r.sql, `
		UPDATE supplier_settlement_statements
		SET status = $2,
			adjustment_amount = $3,
			adjustment_reason = $4,
			payable_amount = usage_amount + $3,
			confirmed_by = $5,
			confirmed_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND status = $6
		RETURNING id
	`, []any{id, service.SupplierSettlementStatusConfirmed, adjustmentAmount, adjustmentReason, reviewerID, service.SupplierSettlementStatusDraft}, &updatedID)
	if err != nil {
		return nil, err
	}
	return r.GetSettlementStatementByID(ctx, updatedID)
}

func (r *supplierRepository) MarkSettlementStatementPaid(ctx context.Context, id int64, paidBy int64, payment service.SupplierSettlementPaymentInput) (*service.SupplierSettlementStatement, error) {
	var updatedID int64
	paidAt := time.Now()
	if payment.PaidAt != nil {
		paidAt = *payment.PaidAt
	}
	err := scanSingleRow(ctx, r.sql, `
		WITH statement AS (
			SELECT id, supplier_id
			FROM supplier_settlement_statements
			WHERE id = $1 AND status = $2
		),
		payment AS (
			INSERT INTO supplier_settlement_payments (
				statement_id, supplier_id, paid_amount, paid_at, payment_reference, payment_note, created_by
			)
			SELECT id, supplier_id, $3, $4, $5, $6, $7
			FROM statement
			RETURNING statement_id, paid_at
		)
		UPDATE supplier_settlement_statements s
		SET status = $8,
			paid_by = $7,
			paid_at = payment.paid_at,
			updated_at = NOW()
		FROM payment
		WHERE s.id = payment.statement_id
		RETURNING s.id
	`, []any{
		id,
		service.SupplierSettlementStatusConfirmed,
		payment.PaidAmount,
		paidAt,
		payment.PaymentReference,
		payment.PaymentNote,
		paidBy,
		service.SupplierSettlementStatusPaid,
	}, &updatedID)
	if err != nil {
		return nil, err
	}
	return r.GetSettlementStatementByID(ctx, updatedID)
}

func (r *supplierRepository) VoidSettlementStatement(ctx context.Context, id int64, voidedBy int64) (*service.SupplierSettlementStatement, error) {
	var updatedID int64
	err := scanSingleRow(ctx, r.sql, `
		UPDATE supplier_settlement_statements
		SET status = $2,
			voided_by = $3,
			voided_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND status <> $4
		RETURNING id
	`, []any{id, service.SupplierSettlementStatusVoided, voidedBy, service.SupplierSettlementStatusPaid}, &updatedID)
	if err != nil {
		return nil, err
	}
	return r.GetSettlementStatementByID(ctx, updatedID)
}

func supplierSettlementFiltersSQL(filters service.SupplierSettlementListFilters) (string, []any) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	add := func(clause string, arg any) {
		args = append(args, arg)
		clauses = append(clauses, fmt.Sprintf(clause, len(args)))
	}
	if filters.SupplierID > 0 {
		add("s.supplier_id = $%d", filters.SupplierID)
	}
	if filters.Status != "" {
		add("s.status = $%d", filters.Status)
	}
	if filters.PeriodStart != nil {
		add("s.period_start = $%d", *filters.PeriodStart)
	}
	if filters.VisibleToSupplier {
		clauses = append(clauses, "s.status IN ('confirmed', 'paid')")
	}
	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

type supplierSettlementScanner interface {
	Scan(dest ...any) error
}

func querySingleSupplierSettlementStatement(ctx context.Context, q sqlQueryer, query string, args ...any) (*service.SupplierSettlementStatement, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}
	item, err := scanSupplierSettlementStatement(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return item, nil
}

func scanSupplierSettlementStatement(scanner supplierSettlementScanner) (*service.SupplierSettlementStatement, error) {
	var item service.SupplierSettlementStatement
	var supplierCompany, supplierContactEmail sql.NullString
	var supplierUserEmail sql.NullString
	var confirmedBy, paidBy, voidedBy sql.NullInt64
	var confirmedAt, paidAt, voidedAt sql.NullTime
	var paymentID, paymentStatementID, paymentSupplierID, paymentCreatedBy sql.NullInt64
	var paymentPaidAmount sql.NullFloat64
	var paymentPaidAt, paymentCreatedAt sql.NullTime
	var paymentReference, paymentNote sql.NullString

	if err := scanner.Scan(
		&item.ID,
		&item.SupplierID,
		&item.PeriodStart,
		&item.PeriodEnd,
		&item.Status,
		&item.UsageAmount,
		&item.AdjustmentAmount,
		&item.AdjustmentReason,
		&item.PayableAmount,
		&item.RequestCount,
		&item.InputTokens,
		&item.OutputTokens,
		&item.TotalTokens,
		&item.CreatedBy,
		&confirmedBy,
		&confirmedAt,
		&paidBy,
		&paidAt,
		&voidedBy,
		&voidedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&supplierCompany,
		&supplierContactEmail,
		&supplierUserEmail,
		&paymentID,
		&paymentStatementID,
		&paymentSupplierID,
		&paymentPaidAmount,
		&paymentPaidAt,
		&paymentReference,
		&paymentNote,
		&paymentCreatedBy,
		&paymentCreatedAt,
	); err != nil {
		return nil, err
	}
	item.ConfirmedBy = nullableInt64Ptr(confirmedBy)
	item.ConfirmedAt = nullableTimePtr(confirmedAt)
	item.PaidBy = nullableInt64Ptr(paidBy)
	item.PaidAt = nullableTimePtr(paidAt)
	item.VoidedBy = nullableInt64Ptr(voidedBy)
	item.VoidedAt = nullableTimePtr(voidedAt)
	item.Supplier = &service.SupplierProfile{
		ID:           item.SupplierID,
		CompanyName:  supplierCompany.String,
		ContactEmail: supplierContactEmail.String,
	}
	if supplierUserEmail.Valid {
		item.Supplier.User = &service.User{Email: supplierUserEmail.String}
	}
	if paymentID.Valid {
		item.Payment = &service.SupplierSettlementPayment{
			ID:               paymentID.Int64,
			StatementID:      paymentStatementID.Int64,
			SupplierID:       paymentSupplierID.Int64,
			PaidAmount:       paymentPaidAmount.Float64,
			PaidAt:           paymentPaidAt.Time,
			PaymentReference: paymentReference.String,
			PaymentNote:      paymentNote.String,
			CreatedBy:        paymentCreatedBy.Int64,
			CreatedAt:        paymentCreatedAt.Time,
		}
	}
	return &item, nil
}

func nullableInt64Ptr(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}

func nullableTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	v := value.Time
	return &v
}

func translateSupplierSettlementNotFound(row *service.SupplierSettlementStatement, err error) (*service.SupplierSettlementStatement, error) {
	if err == nil {
		return row, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrSupplierSettlementNotFound
	}
	return nil, err
}
