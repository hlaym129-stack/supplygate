package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

const supplierPricingRevisionColumns = `
	id, account_id, supplier_id, revision_kind, status, pricing, submit_note, review_note,
	reviewed_by, reviewed_at, effective_at, created_at, updated_at
`

func (r *supplierRepository) CreatePricingRevision(ctx context.Context, revision *service.SupplierAccountPricingRevision) error {
	if revision == nil {
		return nil
	}
	pricingJSON, err := json.Marshal(revision.Pricing)
	if err != nil {
		return fmt.Errorf("marshal supplier pricing: %w", err)
	}
	row, err := querySingleSupplierPricingRevision(ctx, r.sql, fmt.Sprintf(`
		INSERT INTO supplier_account_pricing_revisions
			(account_id, supplier_id, revision_kind, status, pricing, submit_note)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6)
		RETURNING %s
	`, supplierPricingRevisionColumns),
		revision.AccountID,
		revision.SupplierID,
		defaultSupplierPricingKind(revision.RevisionKind),
		defaultSupplierPricingStatus(revision.Status),
		string(pricingJSON),
		revision.SubmitNote,
	)
	if err != nil {
		return err
	}
	*revision = *row
	return nil
}

func (r *supplierRepository) ListPricingRevisions(ctx context.Context, accountID int64) ([]service.SupplierAccountPricingRevision, error) {
	rows, err := r.sql.QueryContext(ctx, fmt.Sprintf(`
		SELECT %s
		FROM supplier_account_pricing_revisions
		WHERE account_id = $1
		ORDER BY id DESC
	`, supplierPricingRevisionColumns), accountID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]service.SupplierAccountPricingRevision, 0)
	for rows.Next() {
		item, err := scanSupplierPricingRevision(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *supplierRepository) GetPendingPricingRevision(ctx context.Context, accountID int64) (*service.SupplierAccountPricingRevision, error) {
	row, err := querySingleSupplierPricingRevision(ctx, r.sql, fmt.Sprintf(`
		SELECT %s
		FROM supplier_account_pricing_revisions
		WHERE account_id = $1 AND status = $2
		ORDER BY id DESC
		LIMIT 1
	`, supplierPricingRevisionColumns), accountID, service.SupplierPricingRevisionStatusPending)
	return translateSupplierPricingNotFound(row, err)
}

func (r *supplierRepository) GetPricingRevisionByID(ctx context.Context, revisionID int64) (*service.SupplierAccountPricingRevision, error) {
	row, err := querySingleSupplierPricingRevision(ctx, r.sql, fmt.Sprintf(`
		SELECT %s
		FROM supplier_account_pricing_revisions
		WHERE id = $1
		LIMIT 1
	`, supplierPricingRevisionColumns), revisionID)
	return translateSupplierPricingNotFound(row, err)
}

func (r *supplierRepository) GetEffectivePricingRevisionAt(ctx context.Context, accountID int64, at time.Time) (*service.SupplierAccountPricingRevision, error) {
	row, err := querySingleSupplierPricingRevision(ctx, r.sql, fmt.Sprintf(`
		SELECT %s
		FROM supplier_account_pricing_revisions
		WHERE account_id = $1
		  AND status = $2
		  AND effective_at IS NOT NULL
		  AND effective_at <= $3
		ORDER BY effective_at DESC, id DESC
		LIMIT 1
	`, supplierPricingRevisionColumns), accountID, service.SupplierPricingRevisionStatusApproved, at)
	return translateSupplierPricingNotFound(row, err)
}

func (r *supplierRepository) ListApprovedPricingRevisionsForAccounts(ctx context.Context, accountIDs []int64) (map[int64][]service.SupplierAccountPricingRevision, error) {
	out := make(map[int64][]service.SupplierAccountPricingRevision, len(accountIDs))
	if len(accountIDs) == 0 {
		return out, nil
	}
	rows, err := r.sql.QueryContext(ctx, fmt.Sprintf(`
		SELECT %s
		FROM supplier_account_pricing_revisions
		WHERE account_id = ANY($1)
		  AND status = $2
		  AND effective_at IS NOT NULL
		ORDER BY account_id ASC, effective_at ASC, id ASC
	`, supplierPricingRevisionColumns), pq.Array(accountIDs), service.SupplierPricingRevisionStatusApproved)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		item, err := scanSupplierPricingRevision(rows)
		if err != nil {
			return nil, err
		}
		out[item.AccountID] = append(out[item.AccountID], *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *supplierRepository) GetLatestPricingRevisionBySupplierAndKind(ctx context.Context, supplierID int64, kind string, since time.Time) (*service.SupplierAccountPricingRevision, error) {
	row, err := querySingleSupplierPricingRevision(ctx, r.sql, fmt.Sprintf(`
		SELECT %s
		FROM supplier_account_pricing_revisions
		WHERE supplier_id = $1
		  AND revision_kind = $2
		  AND created_at >= $3
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, supplierPricingRevisionColumns), supplierID, kind, since)
	return translateSupplierPricingNotFound(row, err)
}

func (r *supplierRepository) CountPricingRevisionsBySupplierAndKindSince(ctx context.Context, supplierID int64, kind string, since time.Time) (int, error) {
	var count int
	if err := scanSingleRow(ctx, r.sql, `
		SELECT COUNT(*)
		FROM supplier_account_pricing_revisions
		WHERE supplier_id = $1
		  AND revision_kind = $2
		  AND created_at >= $3
	`, []any{supplierID, kind, since}, &count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *supplierRepository) ReviewPricingRevision(ctx context.Context, revisionID int64, status string, reviewerID int64, reviewNote string, effectiveAt *time.Time) (*service.SupplierAccountPricingRevision, error) {
	var effectiveValue any
	if effectiveAt != nil {
		effectiveValue = *effectiveAt
	}
	return querySingleSupplierPricingRevision(ctx, r.sql, fmt.Sprintf(`
		UPDATE supplier_account_pricing_revisions
		SET status = $2,
		    review_note = $3,
		    reviewed_by = $4,
		    reviewed_at = NOW(),
		    effective_at = $5,
		    updated_at = NOW()
		WHERE id = $1
		  AND status = '`+service.SupplierPricingRevisionStatusPending+`'
		RETURNING %s
	`, supplierPricingRevisionColumns), revisionID, status, reviewNote, reviewerID, effectiveValue)
}

func querySingleSupplierPricingRevision(ctx context.Context, q sqlQueryer, query string, args ...any) (*service.SupplierAccountPricingRevision, error) {
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
	item, err := scanSupplierPricingRevision(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return item, nil
}

type supplierPricingScanner interface {
	Scan(dest ...any) error
}

func scanSupplierPricingRevision(scanner supplierPricingScanner) (*service.SupplierAccountPricingRevision, error) {
	var item service.SupplierAccountPricingRevision
	var pricingJSON []byte
	var revisionKind sql.NullString
	var reviewedBy sql.NullInt64
	var reviewedAt sql.NullTime
	var effectiveAt sql.NullTime
	if err := scanner.Scan(
		&item.ID,
		&item.AccountID,
		&item.SupplierID,
		&revisionKind,
		&item.Status,
		&pricingJSON,
		&item.SubmitNote,
		&item.ReviewNote,
		&reviewedBy,
		&reviewedAt,
		&effectiveAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(pricingJSON, &item.Pricing); err != nil {
		item.Pricing = []service.ChannelModelPricing{}
	}
	if revisionKind.Valid {
		item.RevisionKind = revisionKind.String
	}
	if item.RevisionKind == "" {
		item.RevisionKind = service.SupplierPricingRevisionKindInitial
	}
	if reviewedBy.Valid {
		item.ReviewedBy = &reviewedBy.Int64
	}
	if reviewedAt.Valid {
		item.ReviewedAt = &reviewedAt.Time
	}
	if effectiveAt.Valid {
		item.EffectiveAt = &effectiveAt.Time
	}
	return &item, nil
}

func translateSupplierPricingNotFound(row *service.SupplierAccountPricingRevision, err error) (*service.SupplierAccountPricingRevision, error) {
	if err == nil {
		return row, nil
	}
	if err == sql.ErrNoRows {
		return nil, service.ErrSupplierPricingNotFound
	}
	return nil, err
}

func defaultSupplierPricingStatus(status string) string {
	if status == "" {
		return service.SupplierPricingRevisionStatusPending
	}
	return status
}

func defaultSupplierPricingKind(kind string) string {
	if kind == "" {
		return service.SupplierPricingRevisionKindInitial
	}
	return kind
}
