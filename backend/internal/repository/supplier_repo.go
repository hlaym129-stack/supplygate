package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbsupplier "github.com/Wei-Shaw/sub2api/ent/supplierprofile"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

const supplierSettlementCostExpr = "COALESCE(account_stats_cost, total_cost) * COALESCE(account_rate_multiplier, 1)"

type supplierRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

func NewSupplierRepository(client *dbent.Client, sqlDB *sql.DB) service.SupplierRepository {
	return &supplierRepository{client: client, sql: sqlDB}
}

func (r *supplierRepository) UpsertProfile(ctx context.Context, profile *service.SupplierProfile) error {
	if profile == nil {
		return nil
	}
	client := r.client
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}
	settlement := profile.SettlementConfig
	if settlement == nil {
		settlement = map[string]any{}
	}
	existing, err := client.SupplierProfile.Query().Where(dbsupplier.UserIDEQ(profile.UserID)).Only(ctx)
	if err != nil && !dbent.IsNotFound(err) {
		return err
	}
	if dbent.IsNotFound(err) {
		created, err := client.SupplierProfile.Create().
			SetUserID(profile.UserID).
			SetCompanyName(profile.CompanyName).
			SetContactName(profile.ContactName).
			SetContactEmail(profile.ContactEmail).
			SetContactPhone(profile.ContactPhone).
			SetStatus(normalizeSupplierProfileStatus(profile.Status)).
			SetAccountSubmissionEnabled(profile.AccountSubmissionEnabled).
			SetSettlementConfig(settlement).
			SetNotes(profile.Notes).
			Save(ctx)
		if err != nil {
			return err
		}
		applySupplierEntity(profile, created)
		return nil
	}
	updated, err := client.SupplierProfile.UpdateOneID(existing.ID).
		SetCompanyName(profile.CompanyName).
		SetContactName(profile.ContactName).
		SetContactEmail(profile.ContactEmail).
		SetContactPhone(profile.ContactPhone).
		SetStatus(normalizeSupplierProfileStatus(profile.Status)).
		SetAccountSubmissionEnabled(profile.AccountSubmissionEnabled).
		SetSettlementConfig(settlement).
		SetNotes(profile.Notes).
		SetReviewNote("").
		ClearReviewedAt().
		ClearReviewedBy().
		Save(ctx)
	if err != nil {
		return err
	}
	applySupplierEntity(profile, updated)
	return nil
}

func normalizeSupplierProfileStatus(status string) string {
	switch status {
	case service.SupplierStatusApproved:
		return service.SupplierStatusApproved
	case service.SupplierStatusRejected:
		return service.SupplierStatusRejected
	default:
		return service.SupplierStatusPending
	}
}

func (r *supplierRepository) GetProfileByUserID(ctx context.Context, userID int64) (*service.SupplierProfile, error) {
	m, err := r.client.SupplierProfile.Query().
		Where(dbsupplier.UserIDEQ(userID)).
		WithUser().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrSupplierProfileNotFound, service.ErrSupplierProfileExists)
	}
	return supplierEntityToService(m), nil
}

func (r *supplierRepository) GetProfileByID(ctx context.Context, id int64) (*service.SupplierProfile, error) {
	m, err := r.client.SupplierProfile.Query().
		Where(dbsupplier.IDEQ(id)).
		WithUser().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrSupplierProfileNotFound, nil)
	}
	return supplierEntityToService(m), nil
}

func (r *supplierRepository) ListProfiles(ctx context.Context, params pagination.PaginationParams, status string) ([]service.SupplierProfile, *pagination.PaginationResult, error) {
	q := r.client.SupplierProfile.Query().WithUser()
	if status != "" {
		q = q.Where(dbsupplier.StatusEQ(status))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	items, err := q.Order(dbent.Desc(dbsupplier.FieldID)).
		Limit(params.Limit()).
		Offset(params.Offset()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}
	out := make([]service.SupplierProfile, 0, len(items))
	for _, item := range items {
		if svc := supplierEntityToService(item); svc != nil {
			out = append(out, *svc)
		}
	}
	return out, paginationResultFromTotal(int64(total), params), nil
}

func (r *supplierRepository) UpdateProfileStatus(ctx context.Context, id int64, status string, reviewerID int64, reviewNote string) (*service.SupplierProfile, error) {
	now := time.Now()
	update := r.client.SupplierProfile.UpdateOneID(id).
		SetStatus(status).
		SetReviewNote(reviewNote).
		SetReviewedAt(now).
		SetReviewedBy(reviewerID)
	if status == service.SupplierStatusApproved {
		update.SetAccountSubmissionEnabled(true)
	}
	updated, err := update.Save(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrSupplierProfileNotFound, nil)
	}
	loaded, err := r.client.SupplierProfile.Query().
		Where(dbsupplier.IDEQ(updated.ID)).
		WithUser().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrSupplierProfileNotFound, nil)
	}
	return supplierEntityToService(loaded), nil
}

func (r *supplierRepository) GetUsageSummary(ctx context.Context, supplierID int64) (*service.SupplierUsageSummary, error) {
	var out service.SupplierUsageSummary
	err := scanSingleRow(ctx, r.sql, `
		SELECT
			COUNT(*),
			COALESCE(SUM(total_cost), 0),
			COALESCE(SUM(actual_cost), 0),
			COALESCE(SUM(`+supplierSettlementCostExpr+`), 0),
			COALESCE(SUM(input_tokens), 0),
			COALESCE(SUM(output_tokens), 0)
		FROM usage_logs
		WHERE supplier_id = $1
	`, []any{supplierID}, &out.Requests, &out.TotalCost, &out.ActualCost, &out.SettlementCost, &out.InputTokens, &out.OutputTokens)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *supplierRepository) GetDashboardStats(ctx context.Context, supplierID int64) (*service.SupplierDashboardStats, error) {
	stats := &service.SupplierDashboardStats{}
	today := timezone.Today()

	if err := scanSingleRow(ctx, r.sql, `
		SELECT
			COUNT(*),
			COALESCE(COUNT(*) FILTER (WHERE status = $2 AND approval_status = $3), 0),
			COALESCE(COUNT(*) FILTER (WHERE approval_status = $4), 0),
			COALESCE(COUNT(*) FILTER (WHERE approval_status = $5), 0),
			COALESCE(COUNT(*) FILTER (WHERE approval_status = $6), 0)
		FROM accounts
		WHERE supplier_id = $1 AND owner_type = $7 AND deleted_at IS NULL
	`, []any{
		supplierID,
		service.StatusActive,
		service.AccountApprovalStatusApproved,
		service.AccountApprovalStatusPending,
		service.AccountApprovalStatusReturned,
		service.AccountApprovalStatusRejected,
		service.AccountOwnerTypeSupplier,
	}, &stats.TotalAccounts, &stats.ActiveAccounts, &stats.PendingAccounts, &stats.ReturnedAccounts, &stats.RejectedAccounts); err != nil {
		return nil, err
	}

	if err := scanSingleRow(ctx, r.sql, `
		SELECT
			COUNT(*),
			COALESCE(SUM(input_tokens), 0),
			COALESCE(SUM(output_tokens), 0),
			COALESCE(SUM(cache_creation_tokens), 0),
			COALESCE(SUM(cache_read_tokens), 0),
			COALESCE(SUM(total_cost), 0),
			COALESCE(SUM(actual_cost), 0),
			COALESCE(SUM(`+supplierSettlementCostExpr+`), 0),
			COALESCE(AVG(duration_ms), 0)
		FROM usage_logs
		WHERE supplier_id = $1
	`, []any{supplierID},
		&stats.TotalRequests,
		&stats.TotalInputTokens,
		&stats.TotalOutputTokens,
		&stats.TotalCacheCreationTokens,
		&stats.TotalCacheReadTokens,
		&stats.TotalCost,
		&stats.TotalActualCost,
		&stats.TotalSettlementCost,
		&stats.AverageDurationMs,
	); err != nil {
		return nil, err
	}
	stats.TotalTokens = stats.TotalInputTokens + stats.TotalOutputTokens + stats.TotalCacheCreationTokens + stats.TotalCacheReadTokens

	if err := scanSingleRow(ctx, r.sql, `
		SELECT
			COUNT(*),
			COALESCE(SUM(input_tokens), 0),
			COALESCE(SUM(output_tokens), 0),
			COALESCE(SUM(cache_creation_tokens), 0),
			COALESCE(SUM(cache_read_tokens), 0),
			COALESCE(SUM(total_cost), 0),
			COALESCE(SUM(actual_cost), 0),
			COALESCE(SUM(`+supplierSettlementCostExpr+`), 0)
		FROM usage_logs
		WHERE supplier_id = $1 AND created_at >= $2
	`, []any{supplierID, today},
		&stats.TodayRequests,
		&stats.TodayInputTokens,
		&stats.TodayOutputTokens,
		&stats.TodayCacheCreationTokens,
		&stats.TodayCacheReadTokens,
		&stats.TodayCost,
		&stats.TodayActualCost,
		&stats.TodaySettlementCost,
	); err != nil {
		return nil, err
	}
	stats.TodayTokens = stats.TodayInputTokens + stats.TodayOutputTokens + stats.TodayCacheCreationTokens + stats.TodayCacheReadTokens

	if err := scanSingleRow(ctx, r.sql, `
		SELECT
			COUNT(*),
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0)
		FROM usage_logs
		WHERE supplier_id = $1 AND created_at >= $2
	`, []any{supplierID, time.Now().Add(-5 * time.Minute)}, &stats.Rpm, &stats.Tpm); err != nil {
		return nil, err
	}
	stats.Rpm = stats.Rpm / 5
	stats.Tpm = stats.Tpm / 5

	return stats, nil
}

func (r *supplierRepository) GetDashboardTrend(ctx context.Context, supplierID int64, startTime, endTime time.Time, granularity string) (results []usagestats.TrendDataPoint, err error) {
	dateFormat := safeDateFormat(granularity)
	query := fmt.Sprintf(`
		SELECT
			TO_CHAR(created_at, '%s') as date,
			COUNT(*) as requests,
			COALESCE(SUM(input_tokens), 0) as input_tokens,
			COALESCE(SUM(output_tokens), 0) as output_tokens,
			COALESCE(SUM(cache_creation_tokens), 0) as cache_creation_tokens,
			COALESCE(SUM(cache_read_tokens), 0) as cache_read_tokens,
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) as total_tokens,
			COALESCE(SUM(total_cost), 0) as cost,
			COALESCE(SUM(`+supplierSettlementCostExpr+`), 0) as actual_cost
		FROM usage_logs
		WHERE supplier_id = $1 AND created_at >= $2 AND created_at < $3
		GROUP BY date
		ORDER BY date ASC
	`, dateFormat)

	rows, err := r.sql.QueryContext(ctx, query, supplierID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			results = nil
		}
	}()

	return scanTrendRows(rows)
}

func (r *supplierRepository) GetDashboardModels(ctx context.Context, supplierID int64, startTime, endTime time.Time) (results []usagestats.ModelStat, err error) {
	query := `
		SELECT
			COALESCE(NULLIF(requested_model, ''), model) as model,
			COUNT(*) as requests,
			COALESCE(SUM(input_tokens), 0) as input_tokens,
			COALESCE(SUM(output_tokens), 0) as output_tokens,
			COALESCE(SUM(cache_creation_tokens), 0) as cache_creation_tokens,
			COALESCE(SUM(cache_read_tokens), 0) as cache_read_tokens,
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) as total_tokens,
			COALESCE(SUM(total_cost), 0) as cost,
			COALESCE(SUM(` + supplierSettlementCostExpr + `), 0) as actual_cost,
			COALESCE(SUM(COALESCE(account_stats_cost, total_cost) * COALESCE(account_rate_multiplier, 1)), 0) as account_cost
		FROM usage_logs
		WHERE supplier_id = $1 AND created_at >= $2 AND created_at < $3
		GROUP BY COALESCE(NULLIF(requested_model, ''), model)
		ORDER BY total_tokens DESC
	`

	rows, err := r.sql.QueryContext(ctx, query, supplierID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			results = nil
		}
	}()

	return scanModelStatsRows(rows)
}

func (r *supplierRepository) ListRecentUsage(ctx context.Context, supplierID int64, params pagination.PaginationParams, startTime, endTime time.Time) (results []service.UsageLog, page *pagination.PaginationResult, err error) {
	var total int64
	if err := scanSingleRow(ctx, r.sql, `
		SELECT COUNT(*)
		FROM usage_logs
		WHERE supplier_id = $1 AND created_at >= $2 AND created_at < $3
	`, []any{supplierID, startTime, endTime}, &total); err != nil {
		return nil, nil, err
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM usage_logs
		WHERE supplier_id = $1 AND created_at >= $2 AND created_at < $3
		ORDER BY created_at DESC, id DESC
		LIMIT $4 OFFSET $5
	`, usageLogSelectColumns)

	rows, err := r.sql.QueryContext(ctx, query, supplierID, startTime, endTime, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			results = nil
			page = nil
		}
	}()

	results = make([]service.UsageLog, 0)
	for rows.Next() {
		log, err := scanUsageLog(rows)
		if err != nil {
			return nil, nil, err
		}
		applySupplierSettlementViewCost(log)
		results = append(results, *log)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return results, paginationResultFromTotal(total, params), nil
}

func applySupplierSettlementViewCost(log *service.UsageLog) {
	if log == nil {
		return
	}
	settlementCost := log.TotalCost
	if log.AccountStatsCost != nil {
		settlementCost = *log.AccountStatsCost
	}
	multiplier := 1.0
	if log.AccountRateMultiplier != nil && *log.AccountRateMultiplier >= 0 {
		multiplier = *log.AccountRateMultiplier
	}
	log.ActualCost = settlementCost * multiplier
}

func supplierEntityToService(m *dbent.SupplierProfile) *service.SupplierProfile {
	return supplierProfileEntityToService(m)
}

func supplierProfileEntityToService(m *dbent.SupplierProfile) *service.SupplierProfile {
	if m == nil {
		return nil
	}
	out := &service.SupplierProfile{}
	applySupplierEntity(out, m)
	if m.Edges.User != nil {
		out.User = userEntityToService(m.Edges.User)
	}
	return out
}

func applySupplierEntity(out *service.SupplierProfile, m *dbent.SupplierProfile) {
	if out == nil || m == nil {
		return
	}
	out.ID = m.ID
	out.UserID = m.UserID
	out.CompanyName = m.CompanyName
	out.ContactName = m.ContactName
	out.ContactEmail = m.ContactEmail
	out.ContactPhone = m.ContactPhone
	out.Status = m.Status
	out.AccountSubmissionEnabled = m.AccountSubmissionEnabled
	out.SettlementConfig = copyJSONMap(m.SettlementConfig)
	out.Notes = m.Notes
	out.ReviewNote = m.ReviewNote
	out.ReviewedAt = m.ReviewedAt
	out.ReviewedBy = m.ReviewedBy
	out.CreatedAt = m.CreatedAt
	out.UpdatedAt = m.UpdatedAt
}
