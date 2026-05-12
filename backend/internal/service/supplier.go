package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/gin-gonic/gin"
)

var (
	ErrSupplierProfileNotFound     = infraerrors.NotFound("SUPPLIER_PROFILE_NOT_FOUND", "supplier profile not found")
	ErrSupplierProfileExists       = infraerrors.Conflict("SUPPLIER_PROFILE_EXISTS", "supplier profile already exists")
	ErrSupplierInvalidStatus       = infraerrors.BadRequest("SUPPLIER_INVALID_STATUS", "invalid supplier status")
	ErrSupplierAccountDenied       = infraerrors.Forbidden("SUPPLIER_ACCOUNT_DENIED", "supplier account access denied")
	ErrSupplierAccountInvalid      = infraerrors.BadRequest("SUPPLIER_ACCOUNT_INVALID", "invalid supplier account")
	ErrSupplierDefaultGroupMissing = infraerrors.BadRequest("SUPPLIER_DEFAULT_GROUP_MISSING", "未找到对应平台的活跃分组")
	ErrSupplierPricingNotFound     = infraerrors.NotFound("SUPPLIER_PRICING_NOT_FOUND", "supplier pricing revision not found")
	ErrSupplierPricingInvalid      = infraerrors.BadRequest("SUPPLIER_PRICING_INVALID", "invalid supplier pricing")
	ErrSupplierSettlementNotFound  = infraerrors.NotFound("SUPPLIER_SETTLEMENT_NOT_FOUND", "supplier settlement statement not found")
	ErrSupplierSettlementInvalid   = infraerrors.BadRequest("SUPPLIER_SETTLEMENT_INVALID", "invalid supplier settlement statement")
	ErrSupplierSettlementLocked    = infraerrors.Conflict("SUPPLIER_SETTLEMENT_LOCKED", "supplier settlement statement cannot be modified")
)

var supplierMarketplacePlatforms = []string{
	PlatformAnthropic,
	PlatformOpenAI,
	PlatformGemini,
	PlatformAntigravity,
}

type SupplierProfile struct {
	ID                       int64
	UserID                   int64
	CompanyName              string
	ContactName              string
	ContactEmail             string
	ContactPhone             string
	Status                   string
	AccountSubmissionEnabled bool
	SettlementConfig         map[string]any
	Notes                    string
	ReviewNote               string
	ReviewedAt               *time.Time
	ReviewedBy               *int64
	CreatedAt                time.Time
	UpdatedAt                time.Time
	User                     *User
}

func isSupplierPricingNotFound(err error) bool {
	return errors.Is(err, ErrSupplierPricingNotFound)
}

type SupplierProfileInput struct {
	CompanyName      string         `json:"company_name"`
	ContactName      string         `json:"contact_name"`
	ContactEmail     string         `json:"contact_email"`
	ContactPhone     string         `json:"contact_phone"`
	Notes            string         `json:"notes"`
	SettlementConfig map[string]any `json:"settlement_config"`
}

type SupplierAccountInput struct {
	Name               string                `json:"name"`
	Notes              *string               `json:"notes"`
	Platform           string                `json:"platform"`
	Type               string                `json:"type"`
	Credentials        map[string]any        `json:"credentials"`
	Extra              map[string]any        `json:"extra"`
	ProxyID            *int64                `json:"proxy_id"`
	Concurrency        int                   `json:"concurrency"`
	LoadFactor         *int                  `json:"load_factor"`
	Priority           int                   `json:"priority"`
	GroupIDs           []int64               `json:"group_ids"`
	ExpiresAt          *int64                `json:"expires_at"`
	AutoPauseOnExpired *bool                 `json:"auto_pause_on_expired"`
	SupportedModels    []string              `json:"supported_models"`
	SettlementPricing  []ChannelModelPricing `json:"settlement_pricing"`
	TestToken          string                `json:"test_token"`
	SubmitNote         string                `json:"submit_note"`
}

type SupplierAccountPretestInput struct {
	SupplierAccountInput
	Prompt string `json:"prompt"`
	Mode   string `json:"mode"`
}

type SupplierAccountPretestResult struct {
	TestToken string    `json:"test_token"`
	ExpiresAt time.Time `json:"expires_at"`
	Models    []string  `json:"models"`
	TestedAt  time.Time `json:"tested_at"`
}

type SupplierAccountApprovalInput struct {
	GroupIDs       []int64  `json:"group_ids"`
	Priority       *int     `json:"priority"`
	RateMultiplier *float64 `json:"rate_multiplier"`
	Schedulable    *bool    `json:"schedulable"`
	RejectReason   string   `json:"reject_reason"`
}

const (
	SupplierEditRequestStatusPending  = "pending"
	SupplierEditRequestStatusRejected = "rejected"
	SupplierEditRequestStatusApproved = "approved"
	supplierTestTokenTTL              = 30 * time.Minute
)

type supplierTestTokenPayload struct {
	SupplierID      int64    `json:"supplier_id"`
	Operation       string   `json:"operation"`
	AccountID       int64    `json:"account_id,omitempty"`
	Platform        string   `json:"platform"`
	Type            string   `json:"type"`
	CredentialsHash string   `json:"credentials_hash"`
	SupportedModels []string `json:"supported_models"`
	IssuedAtUnix    int64    `json:"iat"`
	ExpiresAtUnix   int64    `json:"exp"`
}

type SupplierAccountListFilters struct {
	SupplierID     int64
	ApprovalStatus string
	OwnerType      string
}

type SupplierUsageSummary struct {
	Requests       int64   `json:"requests"`
	TotalCost      float64 `json:"total_cost"`
	ActualCost     float64 `json:"actual_cost"`
	SettlementCost float64 `json:"settlement_cost"`
	InputTokens    int64   `json:"input_tokens"`
	OutputTokens   int64   `json:"output_tokens"`
}

type SupplierDashboardStats struct {
	TotalAccounts    int64 `json:"total_accounts"`
	ActiveAccounts   int64 `json:"active_accounts"`
	PendingAccounts  int64 `json:"pending_accounts"`
	ReturnedAccounts int64 `json:"returned_accounts"`
	RejectedAccounts int64 `json:"rejected_accounts"`

	TotalRequests            int64   `json:"total_requests"`
	TotalInputTokens         int64   `json:"total_input_tokens"`
	TotalOutputTokens        int64   `json:"total_output_tokens"`
	TotalCacheCreationTokens int64   `json:"total_cache_creation_tokens"`
	TotalCacheReadTokens     int64   `json:"total_cache_read_tokens"`
	TotalTokens              int64   `json:"total_tokens"`
	TotalCost                float64 `json:"total_cost"`
	TotalActualCost          float64 `json:"total_actual_cost"`
	TotalSettlementCost      float64 `json:"total_settlement_cost"`

	TodayRequests            int64   `json:"today_requests"`
	TodayInputTokens         int64   `json:"today_input_tokens"`
	TodayOutputTokens        int64   `json:"today_output_tokens"`
	TodayCacheCreationTokens int64   `json:"today_cache_creation_tokens"`
	TodayCacheReadTokens     int64   `json:"today_cache_read_tokens"`
	TodayTokens              int64   `json:"today_tokens"`
	TodayCost                float64 `json:"today_cost"`
	TodayActualCost          float64 `json:"today_actual_cost"`
	TodaySettlementCost      float64 `json:"today_settlement_cost"`

	AverageDurationMs float64 `json:"average_duration_ms"`
	Rpm               int64   `json:"rpm"`
	Tpm               int64   `json:"tpm"`
}

const (
	SupplierPricingRevisionStatusPending  = "pending"
	SupplierPricingRevisionStatusApproved = "approved"
	SupplierPricingRevisionStatusRejected = "rejected"
	SupplierPricingRevisionKindInitial    = "initial"
	SupplierPricingRevisionKindChange     = "change"
)

type SupplierAccountPricingRevision struct {
	ID           int64                 `json:"id"`
	AccountID    int64                 `json:"account_id"`
	SupplierID   int64                 `json:"supplier_id"`
	RevisionKind string                `json:"revision_kind"`
	Status       string                `json:"status"`
	Pricing      []ChannelModelPricing `json:"pricing"`
	SubmitNote   string                `json:"submit_note"`
	ReviewNote   string                `json:"review_note"`
	ReviewedBy   *int64                `json:"reviewed_by"`
	ReviewedAt   *time.Time            `json:"reviewed_at"`
	EffectiveAt  *time.Time            `json:"effective_at"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

const (
	SupplierSettlementStatusDraft     = "draft"
	SupplierSettlementStatusConfirmed = "confirmed"
	SupplierSettlementStatusPaid      = "paid"
	SupplierSettlementStatusVoided    = "voided"
)

type SupplierSettlementStatement struct {
	ID               int64                      `json:"id"`
	SupplierID       int64                      `json:"supplier_id"`
	PeriodStart      time.Time                  `json:"period_start"`
	PeriodEnd        time.Time                  `json:"period_end"`
	Status           string                     `json:"status"`
	UsageAmount      float64                    `json:"usage_amount"`
	AdjustmentAmount float64                    `json:"adjustment_amount"`
	AdjustmentReason string                     `json:"adjustment_reason"`
	PayableAmount    float64                    `json:"payable_amount"`
	RequestCount     int64                      `json:"request_count"`
	InputTokens      int64                      `json:"input_tokens"`
	OutputTokens     int64                      `json:"output_tokens"`
	TotalTokens      int64                      `json:"total_tokens"`
	CreatedBy        int64                      `json:"created_by"`
	ConfirmedBy      *int64                     `json:"confirmed_by"`
	ConfirmedAt      *time.Time                 `json:"confirmed_at"`
	PaidBy           *int64                     `json:"paid_by"`
	PaidAt           *time.Time                 `json:"paid_at"`
	VoidedBy         *int64                     `json:"voided_by"`
	VoidedAt         *time.Time                 `json:"voided_at"`
	CreatedAt        time.Time                  `json:"created_at"`
	UpdatedAt        time.Time                  `json:"updated_at"`
	Supplier         *SupplierProfile           `json:"supplier,omitempty"`
	Payment          *SupplierSettlementPayment `json:"payment,omitempty"`
}

type SupplierSettlementPayment struct {
	ID               int64     `json:"id"`
	StatementID      int64     `json:"statement_id"`
	SupplierID       int64     `json:"supplier_id"`
	PaidAmount       float64   `json:"paid_amount"`
	PaidAt           time.Time `json:"paid_at"`
	PaymentReference string    `json:"payment_reference"`
	PaymentNote      string    `json:"payment_note"`
	CreatedBy        int64     `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
}

type SupplierSettlementListFilters struct {
	SupplierID        int64
	Status            string
	PeriodStart       *time.Time
	VisibleToSupplier bool
}

type SupplierSettlementGenerateInput struct {
	SupplierID  int64  `json:"supplier_id"`
	PeriodMonth string `json:"period_month"`
}

type SupplierSettlementConfirmInput struct {
	AdjustmentAmount float64 `json:"adjustment_amount"`
	AdjustmentReason string  `json:"adjustment_reason"`
}

type SupplierSettlementPaymentInput struct {
	PaidAmount       float64    `json:"paid_amount"`
	PaidAt           *time.Time `json:"paid_at"`
	PaymentReference string     `json:"payment_reference"`
	PaymentNote      string     `json:"payment_note"`
}

type SupplierRepository interface {
	UpsertProfile(ctx context.Context, profile *SupplierProfile) error
	GetProfileByUserID(ctx context.Context, userID int64) (*SupplierProfile, error)
	GetProfileByID(ctx context.Context, id int64) (*SupplierProfile, error)
	ListProfiles(ctx context.Context, params pagination.PaginationParams, status string) ([]SupplierProfile, *pagination.PaginationResult, error)
	UpdateProfileStatus(ctx context.Context, id int64, status string, reviewerID int64, reviewNote string) (*SupplierProfile, error)
	GetUsageSummary(ctx context.Context, supplierID int64) (*SupplierUsageSummary, error)
	GetDashboardStats(ctx context.Context, supplierID int64) (*SupplierDashboardStats, error)
	GetDashboardTrend(ctx context.Context, supplierID int64, startTime, endTime time.Time, granularity string) ([]usagestats.TrendDataPoint, error)
	GetDashboardModels(ctx context.Context, supplierID int64, startTime, endTime time.Time) ([]usagestats.ModelStat, error)
	ListRecentUsage(ctx context.Context, supplierID int64, params pagination.PaginationParams, startTime, endTime time.Time) ([]UsageLog, *pagination.PaginationResult, error)
	CreatePricingRevision(ctx context.Context, revision *SupplierAccountPricingRevision) error
	ListPricingRevisions(ctx context.Context, accountID int64) ([]SupplierAccountPricingRevision, error)
	GetPricingRevisionByID(ctx context.Context, revisionID int64) (*SupplierAccountPricingRevision, error)
	GetPendingPricingRevision(ctx context.Context, accountID int64) (*SupplierAccountPricingRevision, error)
	GetEffectivePricingRevisionAt(ctx context.Context, accountID int64, at time.Time) (*SupplierAccountPricingRevision, error)
	ListApprovedPricingRevisionsForAccounts(ctx context.Context, accountIDs []int64) (map[int64][]SupplierAccountPricingRevision, error)
	GetLatestPricingRevisionBySupplierAndKind(ctx context.Context, supplierID int64, kind string, since time.Time) (*SupplierAccountPricingRevision, error)
	CountPricingRevisionsBySupplierAndKindSince(ctx context.Context, supplierID int64, kind string, since time.Time) (int, error)
	ReviewPricingRevision(ctx context.Context, revisionID int64, status string, reviewerID int64, reviewNote string, effectiveAt *time.Time) (*SupplierAccountPricingRevision, error)
	GetSettlementStatementByID(ctx context.Context, id int64) (*SupplierSettlementStatement, error)
	GetOpenSettlementStatementByPeriod(ctx context.Context, supplierID int64, periodStart time.Time) (*SupplierSettlementStatement, error)
	UpsertDraftSettlementStatement(ctx context.Context, statement *SupplierSettlementStatement) (*SupplierSettlementStatement, error)
	ListSettlementStatements(ctx context.Context, params pagination.PaginationParams, filters SupplierSettlementListFilters) ([]SupplierSettlementStatement, *pagination.PaginationResult, error)
	ConfirmSettlementStatement(ctx context.Context, id int64, reviewerID int64, adjustmentAmount float64, adjustmentReason string) (*SupplierSettlementStatement, error)
	MarkSettlementStatementPaid(ctx context.Context, id int64, paidBy int64, payment SupplierSettlementPaymentInput) (*SupplierSettlementStatement, error)
	VoidSettlementStatement(ctx context.Context, id int64, voidedBy int64) (*SupplierSettlementStatement, error)
}

type SupplierAccountRepository interface {
	AccountRepository
	ListBySupplier(ctx context.Context, supplierID int64, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error)
	ListSupplierAccounts(ctx context.Context, params pagination.PaginationParams, filters SupplierAccountListFilters) ([]Account, *pagination.PaginationResult, error)
}

type SupplierService struct {
	supplierRepo       SupplierRepository
	accountRepo        SupplierAccountRepository
	groupRepo          GroupRepository
	proxyRepo          ProxyRepository
	userService        *UserService
	accountTestService *AccountTestService
	billingService     *BillingService
	testTokenSecret    []byte
}

func NewSupplierService(supplierRepo SupplierRepository, accountRepo SupplierAccountRepository, groupRepo GroupRepository, proxyRepo ProxyRepository, userService *UserService, accountTestService *AccountTestService, optionalBillingService ...*BillingService) *SupplierService {
	var billingService *BillingService
	if len(optionalBillingService) > 0 {
		billingService = optionalBillingService[0]
	}
	return &SupplierService{
		supplierRepo:       supplierRepo,
		accountRepo:        accountRepo,
		groupRepo:          groupRepo,
		proxyRepo:          proxyRepo,
		userService:        userService,
		accountTestService: accountTestService,
		billingService:     billingService,
		testTokenSecret:    []byte("sub2api-supplier-account-test-token-v1"),
	}
}

func (s *SupplierService) ApplyProfile(ctx context.Context, userID int64, input SupplierProfileInput) (*SupplierProfile, error) {
	accountSubmissionEnabled := false
	existing, err := s.supplierRepo.GetProfileByUserID(ctx, userID)
	if err != nil && !infraerrors.IsNotFound(err) {
		return nil, err
	}
	if existing != nil {
		accountSubmissionEnabled = existing.AccountSubmissionEnabled || existing.Status == SupplierStatusApproved
	}
	profile := &SupplierProfile{
		UserID:                   userID,
		CompanyName:              strings.TrimSpace(input.CompanyName),
		ContactName:              strings.TrimSpace(input.ContactName),
		ContactEmail:             strings.TrimSpace(input.ContactEmail),
		ContactPhone:             strings.TrimSpace(input.ContactPhone),
		Notes:                    strings.TrimSpace(input.Notes),
		SettlementConfig:         input.SettlementConfig,
		Status:                   SupplierStatusPending,
		AccountSubmissionEnabled: accountSubmissionEnabled,
	}
	if profile.CompanyName == "" {
		return nil, infraerrors.BadRequest("SUPPLIER_COMPANY_REQUIRED", "company name is required")
	}
	if err := s.supplierRepo.UpsertProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (s *SupplierService) GetProfileByUserID(ctx context.Context, userID int64) (*SupplierProfile, error) {
	return s.supplierRepo.GetProfileByUserID(ctx, userID)
}

func (s *SupplierService) ListProfiles(ctx context.Context, params pagination.PaginationParams, status string) ([]SupplierProfile, *pagination.PaginationResult, error) {
	return s.supplierRepo.ListProfiles(ctx, params, strings.TrimSpace(status))
}

func (s *SupplierService) ReviewProfile(ctx context.Context, supplierID int64, reviewerID int64, status string, reviewNote string) (*SupplierProfile, error) {
	status = strings.TrimSpace(status)
	if status != SupplierStatusApproved && status != SupplierStatusRejected && status != SupplierStatusPending {
		return nil, ErrSupplierInvalidStatus
	}
	profile, err := s.supplierRepo.UpdateProfileStatus(ctx, supplierID, status, reviewerID, strings.TrimSpace(reviewNote))
	if err != nil {
		return nil, err
	}
	if profile.Status == SupplierStatusApproved {
		if err := s.ensureSupplierMarketplaceGroups(ctx, profile); err != nil {
			return nil, err
		}
	}
	return profile, nil
}

func (s *SupplierService) GetApprovedProfileByUserID(ctx context.Context, userID int64) (*SupplierProfile, bool, error) {
	profile, err := s.supplierRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if infraerrors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if profile == nil || !supplierAccountSubmissionEnabled(profile) {
		return profile, false, nil
	}
	return profile, true, nil
}

func supplierAccountSubmissionEnabled(profile *SupplierProfile) bool {
	if profile == nil {
		return false
	}
	return profile.AccountSubmissionEnabled || profile.Status == SupplierStatusApproved
}

func (s *SupplierService) ListMyAccounts(ctx context.Context, userID int64, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	profile, err := s.supplierRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	return s.accountRepo.ListBySupplier(ctx, profile.ID, params)
}

func (s *SupplierService) ListSupplierAccounts(ctx context.Context, params pagination.PaginationParams, filters SupplierAccountListFilters) ([]Account, *pagination.PaginationResult, error) {
	filters.OwnerType = AccountOwnerTypeSupplier
	filters.ApprovalStatus = normalizeSupplierAccountApprovalStatus(filters.ApprovalStatus)
	return s.accountRepo.ListSupplierAccounts(ctx, params, filters)
}

func (s *SupplierService) ListSuggestionGroups(ctx context.Context, platform string) ([]Group, error) {
	platform = strings.TrimSpace(platform)
	if platform != "" {
		return s.groupRepo.ListActiveByPlatform(ctx, platform)
	}
	return s.groupRepo.ListActive(ctx)
}

func (s *SupplierService) ListSuggestionProxies(ctx context.Context) ([]Proxy, error) {
	if s.proxyRepo == nil {
		return []Proxy{}, nil
	}
	return s.proxyRepo.ListActive(ctx)
}

func (s *SupplierService) ListDefaultModels(platform string, accountType string) ([]string, error) {
	switch strings.TrimSpace(platform) {
	case PlatformOpenAI:
		models := make([]string, 0, len(openai.DefaultModels))
		for _, model := range openai.DefaultModels {
			models = append(models, model.ID)
		}
		return models, nil
	case PlatformGemini:
		models := make([]string, 0, len(geminicli.DefaultModels))
		for _, model := range geminicli.DefaultModels {
			models = append(models, model.ID)
		}
		return models, nil
	case PlatformAntigravity:
		defaultModels := antigravity.DefaultModels()
		models := make([]string, 0, len(defaultModels))
		for _, model := range defaultModels {
			models = append(models, model.ID)
		}
		return models, nil
	case PlatformAnthropic:
		if strings.TrimSpace(accountType) == AccountTypeBedrock {
			models := make([]string, 0, len(domain.DefaultBedrockModelMapping))
			for model := range domain.DefaultBedrockModelMapping {
				models = append(models, model)
			}
			return normalizeSupportedModels(models), nil
		}
		models := make([]string, 0, len(claude.DefaultModels))
		for _, model := range claude.DefaultModels {
			models = append(models, model.ID)
		}
		return models, nil
	default:
		return nil, supplierAccountInvalid("unsupported account platform")
	}
}

func (s *SupplierService) TestSupplierAccount(c *gin.Context, userID int64, input SupplierAccountPretestInput) (*SupplierAccountPretestResult, error) {
	profile, err := s.supplierRepo.GetProfileByUserID(c.Request.Context(), userID)
	if err != nil {
		return nil, err
	}
	if !supplierAccountSubmissionEnabled(profile) {
		return nil, infraerrors.Forbidden("SUPPLIER_NOT_APPROVED", "supplier account submission is not approved")
	}
	input.SupplierAccountInput = sanitizeSupplierAccountInput(input.SupplierAccountInput)
	if err := validateSupplierAccountInput(input.SupplierAccountInput); err != nil {
		return nil, err
	}
	models, err := validateAndNormalizeSupplierModels(input.SupportedModels)
	if err != nil {
		return nil, err
	}
	if s.accountTestService == nil {
		return nil, infraerrors.ServiceUnavailable("SUPPLIER_ACCOUNT_TEST_UNAVAILABLE", "account test service is unavailable")
	}
	account := supplierInputToTransientAccount(input.SupplierAccountInput, profile.ID)
	for _, model := range models {
		if err := s.runSupplierPretest(c, account, model, input.Prompt, input.Mode); err != nil {
			return nil, err
		}
	}
	now := time.Now()
	expiresAt := now.Add(supplierTestTokenTTL)
	operation := "create"
	accountID := int64(0)
	token, err := s.signSupplierTestToken(supplierTestTokenPayload{
		SupplierID:      profile.ID,
		Operation:       operation,
		AccountID:       accountID,
		Platform:        strings.TrimSpace(input.Platform),
		Type:            strings.TrimSpace(input.Type),
		CredentialsHash: supplierCredentialsHash(input.Credentials),
		SupportedModels: models,
		IssuedAtUnix:    now.Unix(),
		ExpiresAtUnix:   expiresAt.Unix(),
	})
	if err != nil {
		return nil, err
	}
	return &SupplierAccountPretestResult{TestToken: token, ExpiresAt: expiresAt, Models: models, TestedAt: now}, nil
}

func (s *SupplierService) runSupplierPretest(c *gin.Context, account *Account, model string, prompt string, mode string) error {
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = (&http.Request{}).WithContext(c.Request.Context())
	err := s.accountTestService.TestAccountConnectionWithAccount(ginCtx, account, model, prompt, mode)
	_, errMsg := parseTestSSEOutput(w.Body.String())
	if err != nil {
		return infraerrors.BadRequest("SUPPLIER_ACCOUNT_TEST_FAILED", fmt.Sprintf("model %s test failed: %s", model, err.Error()))
	}
	if strings.TrimSpace(errMsg) != "" {
		return infraerrors.BadRequest("SUPPLIER_ACCOUNT_TEST_FAILED", fmt.Sprintf("model %s test failed: %s", model, errMsg))
	}
	return nil
}

func (s *SupplierService) TestSupplierAccountUpdate(c *gin.Context, userID int64, accountID int64, input SupplierAccountPretestInput) (*SupplierAccountPretestResult, error) {
	result, err := s.TestSupplierAccount(c, userID, input)
	if err != nil {
		return nil, err
	}
	models, _ := validateAndNormalizeSupplierModels(input.SupportedModels)
	profile, err := s.supplierRepo.GetProfileByUserID(c.Request.Context(), userID)
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(supplierTestTokenTTL)
	token, err := s.signSupplierTestToken(supplierTestTokenPayload{
		SupplierID:      profile.ID,
		Operation:       "update",
		AccountID:       accountID,
		Platform:        strings.TrimSpace(input.Platform),
		Type:            strings.TrimSpace(input.Type),
		CredentialsHash: supplierCredentialsHash(input.Credentials),
		SupportedModels: models,
		IssuedAtUnix:    result.TestedAt.Unix(),
		ExpiresAtUnix:   expiresAt.Unix(),
	})
	if err != nil {
		return nil, err
	}
	result.TestToken = token
	result.ExpiresAt = expiresAt
	return result, nil
}

func (s *SupplierService) CreateSupplierAccount(ctx context.Context, userID int64, input SupplierAccountInput) (*Account, error) {
	profile, err := s.supplierRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !supplierAccountSubmissionEnabled(profile) {
		return nil, infraerrors.Forbidden("SUPPLIER_NOT_APPROVED", "supplier account submission is not approved")
	}
	input = sanitizeSupplierAccountInput(input)
	if err := validateSupplierAccountInput(input); err != nil {
		return nil, err
	}
	models, err := validateAndNormalizeSupplierModels(input.SupportedModels)
	if err != nil {
		return nil, err
	}
	pricing, err := validateSupplierSettlementPricing(input.SettlementPricing, models, input.Platform)
	if err != nil {
		return nil, err
	}
	testedAt, err := s.validateSupplierTestToken(profile.ID, "create", 0, input, models)
	if err != nil {
		return nil, err
	}
	concurrency := input.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	priority := 50
	if err := s.validateSupplierAccountSuggestions(ctx, input); err != nil {
		return nil, err
	}
	expiresAt := unixSecondsPtrToTime(input.ExpiresAt)
	schedulable := false
	account := &Account{
		Name:                 strings.TrimSpace(input.Name),
		Notes:                normalizeAccountNotes(input.Notes),
		Platform:             strings.TrimSpace(input.Platform),
		Type:                 strings.TrimSpace(input.Type),
		Credentials:          input.Credentials,
		Extra:                input.Extra,
		ProxyID:              input.ProxyID,
		Concurrency:          concurrency,
		LoadFactor:           nil,
		Priority:             priority,
		Status:               StatusActive,
		OwnerType:            AccountOwnerTypeSupplier,
		SupplierID:           &profile.ID,
		ApprovalStatus:       AccountApprovalStatusPending,
		Schedulable:          schedulable,
		AutoPauseOnExpired:   true,
		ExpiresAt:            expiresAt,
		GroupIDs:             input.GroupIDs,
		SupportedModels:      models,
		SupplierTestedAt:     &testedAt,
		SupplierTestedModels: models,
	}
	if input.AutoPauseOnExpired != nil {
		account.AutoPauseOnExpired = *input.AutoPauseOnExpired
	}
	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}
	if err := s.supplierRepo.CreatePricingRevision(ctx, &SupplierAccountPricingRevision{
		AccountID:    account.ID,
		SupplierID:   profile.ID,
		RevisionKind: SupplierPricingRevisionKindInitial,
		Status:       SupplierPricingRevisionStatusPending,
		Pricing:      pricing,
		SubmitNote:   strings.TrimSpace(input.SubmitNote),
	}); err != nil {
		return nil, err
	}
	if len(input.GroupIDs) > 0 {
		if err := s.accountRepo.BindGroups(ctx, account.ID, input.GroupIDs); err != nil {
			return nil, err
		}
	}
	return account, nil
}

func (s *SupplierService) UpdateSupplierAccount(ctx context.Context, userID int64, accountID int64, input SupplierAccountInput) (*Account, error) {
	profile, err := s.supplierRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	input = sanitizeSupplierAccountInput(input)
	if err := validateSupplierAccountInput(input); err != nil {
		return nil, err
	}
	models, err := validateAndNormalizeSupplierModels(input.SupportedModels)
	if err != nil {
		return nil, err
	}
	pricing, err := validateSupplierSettlementPricing(input.SettlementPricing, models, input.Platform)
	if err != nil {
		return nil, err
	}
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.SupplierID == nil || *account.SupplierID != profile.ID {
		return nil, ErrSupplierAccountDenied
	}
	if account.ApprovalStatus != AccountApprovalStatusReturned {
		return nil, infraerrors.Forbidden("SUPPLIER_ACCOUNT_EDIT_LOCKED", "supplier account must be returned before it can be edited")
	}
	testedAt, err := s.validateSupplierTestToken(profile.ID, "update", accountID, input, models)
	if err != nil {
		return nil, err
	}
	account.Name = strings.TrimSpace(input.Name)
	account.Notes = normalizeAccountNotes(input.Notes)
	account.Platform = strings.TrimSpace(input.Platform)
	account.Type = strings.TrimSpace(input.Type)
	account.Credentials = input.Credentials
	account.Extra = input.Extra
	account.ProxyID = input.ProxyID
	if input.Concurrency > 0 {
		account.Concurrency = input.Concurrency
	}
	account.LoadFactor = nil
	account.Priority = 50
	if err := s.validateSupplierAccountSuggestions(ctx, input); err != nil {
		return nil, err
	}
	account.ExpiresAt = unixSecondsPtrToTime(input.ExpiresAt)
	if input.AutoPauseOnExpired != nil {
		account.AutoPauseOnExpired = *input.AutoPauseOnExpired
	}
	account.GroupIDs = input.GroupIDs
	account.ApprovalStatus = AccountApprovalStatusPending
	account.Schedulable = false
	account.RejectReason = ""
	account.SupportedModels = models
	account.SupplierTestedAt = &testedAt
	account.SupplierTestedModels = models
	account.SupplierEditRequestStatus = ""
	account.SupplierEditRequestReason = ""
	account.SupplierEditRequestReviewNote = ""
	account.SupplierEditRequestedAt = nil
	account.SupplierEditRequestReviewedAt = nil
	account.SupplierEditRequestReviewedBy = nil
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}
	if err := s.supplierRepo.CreatePricingRevision(ctx, &SupplierAccountPricingRevision{
		AccountID:    account.ID,
		SupplierID:   profile.ID,
		RevisionKind: SupplierPricingRevisionKindInitial,
		Status:       SupplierPricingRevisionStatusPending,
		Pricing:      pricing,
		SubmitNote:   strings.TrimSpace(input.SubmitNote),
	}); err != nil {
		return nil, err
	}
	if err := s.accountRepo.BindGroups(ctx, account.ID, input.GroupIDs); err != nil {
		return nil, err
	}
	return account, nil
}

func normalizeSupplierAccountApprovalStatus(status string) string {
	switch strings.TrimSpace(status) {
	case AccountApprovalStatusPending:
		return AccountApprovalStatusPending
	case AccountApprovalStatusApproved:
		return AccountApprovalStatusApproved
	case AccountApprovalStatusRejected:
		return AccountApprovalStatusRejected
	case AccountApprovalStatusReturned:
		return AccountApprovalStatusReturned
	default:
		return ""
	}
}

func validateSupplierAccountInput(input SupplierAccountInput) error {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Platform) == "" || strings.TrimSpace(input.Type) == "" {
		return supplierAccountInvalid("account name, platform, and type are required")
	}
	credentials := input.Credentials
	if credentials == nil {
		credentials = map[string]any{}
	}
	required, err := requiredCredentialKeys(input.Platform, input.Type, credentials)
	if err != nil {
		return err
	}
	for _, key := range required {
		if !hasCredentialValue(credentials, key) {
			return supplierAccountInvalid("missing required credential: " + key)
		}
	}
	return nil
}

func validateSupplierSettlementPricing(pricing []ChannelModelPricing, supportedModels []string, platform string) ([]ChannelModelPricing, error) {
	if len(pricing) == 0 {
		return nil, supplierPricingInvalid("settlement pricing is required")
	}
	required := make(map[string]struct{}, len(supportedModels))
	for _, model := range supportedModels {
		model = strings.ToLower(strings.TrimSpace(model))
		if model != "" {
			required[model] = struct{}{}
		}
	}
	covered := make(map[string]struct{}, len(required))
	out := make([]ChannelModelPricing, 0, len(pricing))
	hasEffectivePrice := false
	for _, entry := range pricing {
		entry.ID = 0
		entry.ChannelID = 0
		entry.Platform = strings.TrimSpace(entry.Platform)
		if entry.Platform == "" {
			entry.Platform = strings.TrimSpace(platform)
		}
		if entry.Platform != "" && strings.TrimSpace(platform) != "" && entry.Platform != strings.TrimSpace(platform) {
			return nil, supplierPricingInvalid("pricing platform must match account platform")
		}
		if !entry.BillingMode.IsValid() {
			return nil, supplierPricingInvalid("invalid billing mode")
		}
		if entry.BillingMode == "" {
			entry.BillingMode = BillingModeToken
		}
		models := make([]string, 0, len(entry.Models))
		for _, model := range entry.Models {
			model = strings.TrimSpace(model)
			if model == "" {
				continue
			}
			models = append(models, model)
			if _, ok := required[strings.ToLower(model)]; ok {
				covered[strings.ToLower(model)] = struct{}{}
			}
		}
		entry.Models = normalizeSupplierPricingModels(models)
		if len(entry.Models) == 0 {
			return nil, supplierPricingInvalid("pricing models are required")
		}
		if err := validateSupplierPricingEntryPrices(entry); err != nil {
			return nil, err
		}
		if err := ValidateIntervals(entry.Intervals); err != nil {
			return nil, supplierPricingInvalid(err.Error())
		}
		if supplierPricingEntryHasPrice(entry) {
			hasEffectivePrice = true
		}
		out = append(out, entry)
	}
	for model := range required {
		if _, ok := covered[model]; !ok {
			return nil, supplierPricingInvalid("settlement pricing must cover all supported models")
		}
	}
	if !hasEffectivePrice {
		return nil, supplierPricingInvalid("settlement pricing must contain at least one positive price")
	}
	return out, nil
}

func validateSupplierPricingEntryPrices(entry ChannelModelPricing) error {
	prices := []*float64{
		entry.InputPrice,
		entry.OutputPrice,
		entry.CacheWritePrice,
		entry.CacheReadPrice,
		entry.ImageOutputPrice,
		entry.PerRequestPrice,
	}
	for _, value := range prices {
		if value != nil && *value < 0 {
			return supplierPricingInvalid("pricing values must be >= 0")
		}
	}
	return nil
}

func supplierPricingInvalid(message string) error {
	return infraerrors.BadRequest("SUPPLIER_PRICING_INVALID", message)
}

func normalizeSupplierPricingModels(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	return out
}

func supplierPricingEntryHasPrice(entry ChannelModelPricing) bool {
	values := []*float64{
		entry.InputPrice,
		entry.OutputPrice,
		entry.CacheWritePrice,
		entry.CacheReadPrice,
		entry.ImageOutputPrice,
		entry.PerRequestPrice,
	}
	for _, value := range values {
		if value != nil && *value > 0 {
			return true
		}
	}
	for _, interval := range entry.Intervals {
		for _, value := range []*float64{
			interval.InputPrice,
			interval.OutputPrice,
			interval.CacheWritePrice,
			interval.CacheReadPrice,
			interval.PerRequestPrice,
		} {
			if value != nil && *value > 0 {
				return true
			}
		}
	}
	return false
}

func requiredCredentialKeys(platform string, accountType string, credentials map[string]any) ([]string, error) {
	switch strings.TrimSpace(platform) {
	case PlatformAnthropic:
		switch strings.TrimSpace(accountType) {
		case AccountTypeOAuth, AccountTypeSetupToken:
			return []string{"access_token"}, nil
		case AccountTypeAPIKey:
			return []string{"api_key"}, nil
		case AccountTypeBedrock:
			if strings.TrimSpace(credentialString(credentials, "auth_mode")) == "apikey" {
				return []string{"auth_mode", "api_key"}, nil
			}
			return []string{"aws_access_key_id", "aws_secret_access_key"}, nil
		case AccountTypeServiceAccount:
			return []string{"service_account_json"}, nil
		default:
			return nil, supplierAccountInvalid("unsupported anthropic account type")
		}
	case PlatformOpenAI:
		switch strings.TrimSpace(accountType) {
		case AccountTypeOAuth:
			return []string{"access_token"}, nil
		case AccountTypeAPIKey:
			return []string{"api_key"}, nil
		default:
			return nil, supplierAccountInvalid("unsupported openai account type")
		}
	case PlatformGemini:
		switch strings.TrimSpace(accountType) {
		case AccountTypeAPIKey:
			return []string{"api_key"}, nil
		case AccountTypeOAuth:
			return []string{"access_token"}, nil
		case AccountTypeServiceAccount:
			return []string{"service_account_json"}, nil
		default:
			return nil, supplierAccountInvalid("unsupported gemini account type")
		}
	case PlatformAntigravity:
		if strings.TrimSpace(accountType) != AccountTypeOAuth {
			return nil, supplierAccountInvalid("unsupported antigravity account type")
		}
		return []string{"access_token"}, nil
	default:
		return nil, supplierAccountInvalid("unsupported account platform")
	}
}

func supplierAccountInvalid(message string) error {
	return infraerrors.BadRequest("SUPPLIER_ACCOUNT_INVALID", message)
}

func sanitizeSupplierAccountInput(input SupplierAccountInput) SupplierAccountInput {
	input.ProxyID = nil
	input.GroupIDs = nil
	input.LoadFactor = nil
	input.Priority = 50
	if input.Credentials != nil {
		delete(input.Credentials, "model_mapping")
		delete(input.Credentials, "compact_model_mapping")
	}
	return input
}

func (s *SupplierService) validateSupplierAccountSuggestions(ctx context.Context, input SupplierAccountInput) error {
	if input.LoadFactor != nil && *input.LoadFactor > 10000 {
		return infraerrors.BadRequest("SUPPLIER_LOAD_FACTOR_INVALID", "load_factor must be <= 10000")
	}
	if len(input.GroupIDs) > 0 {
		if err := validateSupplierGroupIDs(ctx, s.groupRepo, input.GroupIDs); err != nil {
			return err
		}
		for _, groupID := range input.GroupIDs {
			group, err := s.groupRepo.GetByID(ctx, groupID)
			if err != nil {
				return err
			}
			if group == nil || group.Status != StatusActive {
				return infraerrors.BadRequest("SUPPLIER_GROUP_INVALID", "group is not active")
			}
		}
	}
	if input.ProxyID != nil {
		if *input.ProxyID <= 0 {
			return infraerrors.BadRequest("SUPPLIER_PROXY_INVALID", "invalid proxy id")
		}
		if s.proxyRepo == nil {
			return infraerrors.BadRequest("SUPPLIER_PROXY_INVALID", "proxy is not available")
		}
		proxy, err := s.proxyRepo.GetByID(ctx, *input.ProxyID)
		if err != nil {
			return err
		}
		if proxy == nil || !proxy.IsActive() {
			return infraerrors.BadRequest("SUPPLIER_PROXY_INVALID", "proxy is not active")
		}
	}
	return nil
}

func validateAndNormalizeSupplierModels(models []string) ([]string, error) {
	normalized := normalizeSupportedModels(models)
	if len(normalized) == 0 {
		return nil, infraerrors.BadRequest("SUPPLIER_MODELS_REQUIRED", "at least one supported model is required")
	}
	if len(normalized) > 50 {
		return nil, infraerrors.BadRequest("SUPPLIER_MODELS_TOO_MANY", "supported_models cannot exceed 50")
	}
	return normalized, nil
}

func normalizeSupportedModels(models []string) []string {
	if len(models) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		trimmed := strings.TrimSpace(model)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func supplierInputToTransientAccount(input SupplierAccountInput, supplierID int64) *Account {
	concurrency := input.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	return &Account{
		ID:                 0,
		Name:               strings.TrimSpace(input.Name),
		Notes:              normalizeAccountNotes(input.Notes),
		Platform:           strings.TrimSpace(input.Platform),
		Type:               strings.TrimSpace(input.Type),
		Credentials:        input.Credentials,
		Extra:              input.Extra,
		Concurrency:        concurrency,
		Priority:           50,
		Status:             StatusActive,
		OwnerType:          AccountOwnerTypeSupplier,
		SupplierID:         &supplierID,
		ApprovalStatus:     AccountApprovalStatusPending,
		Schedulable:        false,
		AutoPauseOnExpired: true,
		ExpiresAt:          unixSecondsPtrToTime(input.ExpiresAt),
	}
}

func supplierCredentialsHash(credentials map[string]any) string {
	raw, _ := json.Marshal(normalizeJSONForSigning(credentials))
	sum := sha256.Sum256(raw)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func normalizeJSONForSigning(value any) any {
	raw, err := json.Marshal(value)
	if err != nil {
		return value
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return value
	}
	return decoded
}

func (s *SupplierService) signSupplierTestToken(payload supplierTestTokenPayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, s.testTokenSecret)
	_, _ = mac.Write([]byte(body))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return body + "." + sig, nil
}

func (s *SupplierService) validateSupplierTestToken(supplierID int64, operation string, accountID int64, input SupplierAccountInput, models []string) (time.Time, error) {
	token := strings.TrimSpace(input.TestToken)
	if token == "" {
		return time.Time{}, infraerrors.BadRequest("SUPPLIER_TEST_TOKEN_REQUIRED", "test_token is required")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return time.Time{}, infraerrors.BadRequest("SUPPLIER_TEST_TOKEN_INVALID", "invalid test_token")
	}
	mac := hmac.New(sha256.New, s.testTokenSecret)
	_, _ = mac.Write([]byte(parts[0]))
	expected := mac.Sum(nil)
	actual, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(expected, actual) {
		return time.Time{}, infraerrors.BadRequest("SUPPLIER_TEST_TOKEN_INVALID", "invalid test_token")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return time.Time{}, infraerrors.BadRequest("SUPPLIER_TEST_TOKEN_INVALID", "invalid test_token")
	}
	var payload supplierTestTokenPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return time.Time{}, infraerrors.BadRequest("SUPPLIER_TEST_TOKEN_INVALID", "invalid test_token")
	}
	now := time.Now()
	if payload.ExpiresAtUnix <= now.Unix() {
		return time.Time{}, infraerrors.BadRequest("SUPPLIER_TEST_TOKEN_EXPIRED", "test_token has expired")
	}
	if payload.SupplierID != supplierID || payload.Operation != operation || payload.AccountID != accountID {
		return time.Time{}, infraerrors.BadRequest("SUPPLIER_TEST_TOKEN_MISMATCH", "test_token does not match this operation")
	}
	if payload.Platform != strings.TrimSpace(input.Platform) || payload.Type != strings.TrimSpace(input.Type) {
		return time.Time{}, infraerrors.BadRequest("SUPPLIER_TEST_TOKEN_MISMATCH", "test_token does not match account type")
	}
	if payload.CredentialsHash != supplierCredentialsHash(input.Credentials) {
		return time.Time{}, infraerrors.BadRequest("SUPPLIER_TEST_TOKEN_MISMATCH", "test_token does not match credentials")
	}
	if !sameStringSlice(payload.SupportedModels, models) {
		return time.Time{}, infraerrors.BadRequest("SUPPLIER_TEST_TOKEN_MISMATCH", "test_token does not match supported models")
	}
	return time.Unix(payload.IssuedAtUnix, 0), nil
}

func sameStringSlice(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func unixSecondsPtrToTime(value *int64) *time.Time {
	if value == nil || *value <= 0 {
		return nil
	}
	t := time.Unix(*value, 0)
	return &t
}

func hasCredentialValue(credentials map[string]any, key string) bool {
	value, ok := credentials[key]
	if !ok || value == nil {
		return false
	}
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) != ""
	case map[string]any:
		return len(v) > 0
	default:
		return true
	}
}

func credentialString(credentials map[string]any, key string) string {
	if credentials == nil {
		return ""
	}
	value, ok := credentials[key]
	if !ok || value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func (s *SupplierService) ApproveAccount(ctx context.Context, accountID int64, reviewerID int64, input SupplierAccountApprovalInput) (*Account, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.OwnerType != AccountOwnerTypeSupplier || account.SupplierID == nil {
		return nil, infraerrors.BadRequest("SUPPLIER_ACCOUNT_REQUIRED", "account is not a supplier account")
	}
	profile, err := s.supplierRepo.GetProfileByID(ctx, *account.SupplierID)
	if err != nil {
		return nil, err
	}
	if !supplierAccountSubmissionEnabled(profile) {
		return nil, infraerrors.BadRequest("SUPPLIER_PROFILE_NOT_APPROVED", "supplier account submission must be approved before approving supplier account")
	}
	if err := s.ensureSupplierMarketplaceGroups(ctx, profile); err != nil {
		return nil, err
	}
	groupIDs, err := supplierOwnedGroupIDByPlatform(ctx, s.groupRepo, *account.SupplierID, account.Platform)
	if err != nil {
		return nil, err
	}
	if input.Priority != nil {
		account.Priority = *input.Priority
	}
	if input.RateMultiplier != nil {
		account.RateMultiplier = input.RateMultiplier
	}
	pendingPricing, err := s.supplierRepo.GetPendingPricingRevision(ctx, account.ID)
	if err != nil && !isSupplierPricingNotFound(err) {
		return nil, err
	}
	if pendingPricing == nil {
		return nil, ErrSupplierPricingNotFound
	}
	now := time.Now()
	account.ApprovalStatus = AccountApprovalStatusApproved
	account.ApprovedAt = &now
	account.ApprovedBy = &reviewerID
	account.RejectReason = ""
	account.Schedulable = true
	if input.Schedulable != nil {
		account.Schedulable = *input.Schedulable
	}
	account.GroupIDs = groupIDs
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}
	if len(groupIDs) > 0 {
		if err := s.accountRepo.BindGroups(ctx, account.ID, groupIDs); err != nil {
			return nil, err
		}
	}
	if _, err := s.supplierRepo.ReviewPricingRevision(ctx, pendingPricing.ID, SupplierPricingRevisionStatusApproved, reviewerID, "", &now); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *SupplierService) RejectAccount(ctx context.Context, accountID int64, reviewerID int64, reason string) (*Account, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.OwnerType != AccountOwnerTypeSupplier || account.SupplierID == nil {
		return nil, infraerrors.BadRequest("SUPPLIER_ACCOUNT_REQUIRED", "account is not a supplier account")
	}
	now := time.Now()
	account.ApprovalStatus = AccountApprovalStatusRejected
	account.ApprovedAt = &now
	account.ApprovedBy = &reviewerID
	account.RejectReason = strings.TrimSpace(reason)
	account.Schedulable = false
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}
	if pendingPricing, err := s.supplierRepo.GetPendingPricingRevision(ctx, account.ID); err == nil && pendingPricing != nil {
		if _, reviewErr := s.supplierRepo.ReviewPricingRevision(ctx, pendingPricing.ID, SupplierPricingRevisionStatusRejected, reviewerID, strings.TrimSpace(reason), nil); reviewErr != nil {
			return nil, reviewErr
		}
	} else if err != nil && !isSupplierPricingNotFound(err) {
		return nil, err
	}
	return account, nil
}

func (s *SupplierService) RequestAccountEdit(ctx context.Context, userID int64, accountID int64, reason string) (*Account, error) {
	profile, err := s.supplierRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.SupplierID == nil || *account.SupplierID != profile.ID {
		return nil, ErrSupplierAccountDenied
	}
	if account.ApprovalStatus == AccountApprovalStatusReturned {
		return nil, infraerrors.Conflict("SUPPLIER_ACCOUNT_ALREADY_RETURNED", "account is already returned for editing")
	}
	if account.SupplierEditRequestStatus == SupplierEditRequestStatusPending {
		return nil, infraerrors.Conflict("SUPPLIER_EDIT_REQUEST_EXISTS", "edit request is already pending")
	}
	now := time.Now()
	account.SupplierEditRequestStatus = SupplierEditRequestStatusPending
	account.SupplierEditRequestReason = strings.TrimSpace(reason)
	account.SupplierEditRequestReviewNote = ""
	account.SupplierEditRequestedAt = &now
	account.SupplierEditRequestReviewedAt = nil
	account.SupplierEditRequestReviewedBy = nil
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *SupplierService) ReturnAccountForEdit(ctx context.Context, accountID int64, reviewerID int64, reviewNote string) (*Account, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.OwnerType != AccountOwnerTypeSupplier || account.SupplierID == nil {
		return nil, infraerrors.BadRequest("SUPPLIER_ACCOUNT_REQUIRED", "account is not a supplier account")
	}
	now := time.Now()
	account.ApprovalStatus = AccountApprovalStatusReturned
	account.Schedulable = false
	account.SupplierEditRequestStatus = SupplierEditRequestStatusApproved
	account.SupplierEditRequestReviewNote = strings.TrimSpace(reviewNote)
	account.SupplierEditRequestReviewedAt = &now
	account.SupplierEditRequestReviewedBy = &reviewerID
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *SupplierService) RejectAccountEditRequest(ctx context.Context, accountID int64, reviewerID int64, reviewNote string) (*Account, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.OwnerType != AccountOwnerTypeSupplier || account.SupplierID == nil {
		return nil, infraerrors.BadRequest("SUPPLIER_ACCOUNT_REQUIRED", "account is not a supplier account")
	}
	if account.SupplierEditRequestStatus != SupplierEditRequestStatusPending {
		return nil, infraerrors.BadRequest("SUPPLIER_EDIT_REQUEST_NOT_PENDING", "edit request is not pending")
	}
	now := time.Now()
	account.SupplierEditRequestStatus = SupplierEditRequestStatusRejected
	account.SupplierEditRequestReviewNote = strings.TrimSpace(reviewNote)
	account.SupplierEditRequestReviewedAt = &now
	account.SupplierEditRequestReviewedBy = &reviewerID
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *SupplierService) GetUsageSummary(ctx context.Context, userID int64) (*SupplierUsageSummary, error) {
	profile, err := s.supplierRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.supplierRepo.GetUsageSummary(ctx, profile.ID)
}

func (s *SupplierService) GetDashboardStats(ctx context.Context, userID int64) (*SupplierDashboardStats, error) {
	profile, err := s.supplierRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.supplierRepo.GetDashboardStats(ctx, profile.ID)
}

func (s *SupplierService) GetDashboardTrend(ctx context.Context, userID int64, startTime, endTime time.Time, granularity string) ([]usagestats.TrendDataPoint, error) {
	profile, err := s.supplierRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.supplierRepo.GetDashboardTrend(ctx, profile.ID, startTime, endTime, granularity)
}

func (s *SupplierService) GetDashboardModels(ctx context.Context, userID int64, startTime, endTime time.Time) ([]usagestats.ModelStat, error) {
	profile, err := s.supplierRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.supplierRepo.GetDashboardModels(ctx, profile.ID, startTime, endTime)
}

func (s *SupplierService) ListRecentUsage(ctx context.Context, userID int64, params pagination.PaginationParams, startTime, endTime time.Time) ([]UsageLog, *pagination.PaginationResult, error) {
	profile, err := s.supplierRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	return s.supplierRepo.ListRecentUsage(ctx, profile.ID, params, startTime, endTime)
}

func (s *SupplierService) ListMyPricingRevisions(ctx context.Context, userID int64, accountID int64) ([]SupplierAccountPricingRevision, error) {
	profile, account, err := s.getOwnedSupplierAccount(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	if account.SupplierID == nil || *account.SupplierID != profile.ID {
		return nil, ErrSupplierAccountDenied
	}
	return s.supplierRepo.ListPricingRevisions(ctx, accountID)
}

func (s *SupplierService) SubmitPricingChange(ctx context.Context, userID int64, accountID int64, pricing []ChannelModelPricing, submitNote string) (*SupplierAccountPricingRevision, error) {
	profile, account, err := s.getOwnedSupplierAccount(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	if account.ApprovalStatus != AccountApprovalStatusApproved {
		return nil, infraerrors.BadRequest("SUPPLIER_ACCOUNT_NOT_APPROVED", "pricing changes require an approved account")
	}
	if err := s.ensureSupplierPricingChangeDailyLimit(ctx, profile.ID); err != nil {
		return nil, err
	}
	if pending, err := s.supplierRepo.GetPendingPricingRevision(ctx, accountID); err == nil && pending != nil {
		return nil, infraerrors.Conflict("SUPPLIER_PRICING_CHANGE_EXISTS", "pricing change is already pending")
	} else if err != nil && !isSupplierPricingNotFound(err) {
		return nil, err
	}
	normalizedPricing, err := validateSupplierSettlementPricing(pricing, account.SupportedModels, account.Platform)
	if err != nil {
		return nil, err
	}
	revision := &SupplierAccountPricingRevision{
		AccountID:    account.ID,
		SupplierID:   profile.ID,
		RevisionKind: SupplierPricingRevisionKindChange,
		Status:       SupplierPricingRevisionStatusPending,
		Pricing:      normalizedPricing,
		SubmitNote:   strings.TrimSpace(submitNote),
	}
	if err := s.supplierRepo.CreatePricingRevision(ctx, revision); err != nil {
		return nil, err
	}
	return revision, nil
}

func (s *SupplierService) ensureSupplierPricingChangeDailyLimit(ctx context.Context, supplierID int64) error {
	now := timezone.Now()
	count, err := s.supplierRepo.CountPricingRevisionsBySupplierAndKindSince(ctx, supplierID, SupplierPricingRevisionKindChange, timezone.StartOfDay(now))
	if err != nil {
		return err
	}
	if count >= 3 {
		return infraerrors.Forbidden("SUPPLIER_PRICING_CHANGE_DAILY_LIMIT", "pricing changes are limited to three times per day")
	}
	return nil
}

func supplierPricingChangeEffectiveAt(submittedAt time.Time) time.Time {
	loc := timezone.Location()
	submittedAt = submittedAt.In(loc)
	sameDay := time.Date(submittedAt.Year(), submittedAt.Month(), submittedAt.Day(), 7, 30, 0, 0, loc)
	if submittedAt.Before(sameDay) {
		return sameDay
	}
	return sameDay.AddDate(0, 0, 1)
}

func supplierPricingChangeApprovedEffectiveAt(submittedAt time.Time, approvedAt time.Time) time.Time {
	scheduledAt := supplierPricingChangeEffectiveAt(submittedAt)
	if approvedAt.After(scheduledAt) {
		return approvedAt
	}
	return scheduledAt
}

func (s *SupplierService) ListAccountPricingRevisions(ctx context.Context, accountID int64) ([]SupplierAccountPricingRevision, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.OwnerType != AccountOwnerTypeSupplier || account.SupplierID == nil {
		return nil, infraerrors.BadRequest("SUPPLIER_ACCOUNT_REQUIRED", "account is not a supplier account")
	}
	return s.supplierRepo.ListPricingRevisions(ctx, accountID)
}

func (s *SupplierService) ApprovePricingRevision(ctx context.Context, revisionID int64, reviewerID int64, reviewNote string) (*SupplierAccountPricingRevision, error) {
	revision, err := s.supplierRepo.GetPricingRevisionByID(ctx, revisionID)
	if err != nil {
		return nil, err
	}
	effectiveAt := time.Now()
	if revision.RevisionKind == SupplierPricingRevisionKindChange {
		effectiveAt = supplierPricingChangeApprovedEffectiveAt(revision.CreatedAt, effectiveAt)
	}
	return s.supplierRepo.ReviewPricingRevision(ctx, revisionID, SupplierPricingRevisionStatusApproved, reviewerID, strings.TrimSpace(reviewNote), &effectiveAt)
}

func (s *SupplierService) RejectPricingRevision(ctx context.Context, revisionID int64, reviewerID int64, reviewNote string) (*SupplierAccountPricingRevision, error) {
	return s.supplierRepo.ReviewPricingRevision(ctx, revisionID, SupplierPricingRevisionStatusRejected, reviewerID, strings.TrimSpace(reviewNote), nil)
}

func (s *SupplierService) GetModelDefaultPricing(model string) (*ChannelModelPricing, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, infraerrors.BadRequest("MODEL_REQUIRED", "model is required")
	}
	if s.billingService == nil {
		return &ChannelModelPricing{Models: []string{model}, BillingMode: BillingModeToken}, nil
	}
	pricing, err := s.billingService.GetModelPricing(model)
	if err != nil || pricing == nil {
		return &ChannelModelPricing{Models: []string{model}, BillingMode: BillingModeToken}, nil
	}
	return &ChannelModelPricing{
		Models:           []string{model},
		BillingMode:      BillingModeToken,
		InputPrice:       floatPtr(pricing.InputPricePerToken),
		OutputPrice:      floatPtr(pricing.OutputPricePerToken),
		CacheWritePrice:  floatPtr(pricing.CacheCreationPricePerToken),
		CacheReadPrice:   floatPtr(pricing.CacheReadPricePerToken),
		ImageOutputPrice: floatPtr(pricing.ImageOutputPricePerToken),
	}, nil
}

func (s *SupplierService) getOwnedSupplierAccount(ctx context.Context, userID int64, accountID int64) (*SupplierProfile, *Account, error) {
	profile, err := s.supplierRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, nil, err
	}
	if account.SupplierID == nil || *account.SupplierID != profile.ID {
		return nil, nil, ErrSupplierAccountDenied
	}
	return profile, account, nil
}

func floatPtr(value float64) *float64 {
	v := value
	return &v
}

func supplierMarketplacePlatformLabel(platform string) string {
	switch strings.TrimSpace(platform) {
	case PlatformAnthropic:
		return "Anthropic"
	case PlatformOpenAI:
		return "OpenAI"
	case PlatformGemini:
		return "Gemini"
	case PlatformAntigravity:
		return "Antigravity"
	default:
		return strings.TrimSpace(platform)
	}
}

func supplierMarketplaceGroupName(profile *SupplierProfile, platform string) string {
	companyName := strings.TrimSpace(profile.CompanyName)
	if companyName == "" {
		companyName = "供应商"
	}
	return companyName + " " + supplierMarketplacePlatformLabel(platform)
}

func (s *SupplierService) ensureSupplierMarketplaceGroups(ctx context.Context, profile *SupplierProfile) error {
	if profile == nil || profile.ID <= 0 {
		return ErrSupplierProfileNotFound
	}

	groups, err := s.groupRepo.ListActive(ctx)
	if err != nil {
		return err
	}

	existingByPlatform := make(map[string]struct{}, len(supplierMarketplacePlatforms))
	for i := range groups {
		group := groups[i]
		if group.SupplierProfileID == nil || *group.SupplierProfileID != profile.ID {
			continue
		}
		if !group.IsSubscriptionType() || !group.IsActive() {
			continue
		}
		existingByPlatform[group.Platform] = struct{}{}
	}

	for _, platform := range supplierMarketplacePlatforms {
		if _, ok := existingByPlatform[platform]; ok {
			continue
		}
		supplierID := profile.ID
		group := &Group{
			Name:              supplierMarketplaceGroupName(profile, platform),
			Description:       strings.TrimSpace(profile.Notes),
			Platform:          platform,
			RateMultiplier:    1.0,
			IsExclusive:       true,
			Status:            StatusActive,
			SubscriptionType:  SubscriptionTypeSubscription,
			SupplierProfileID: &supplierID,
		}
		if err := s.groupRepo.Create(ctx, group); err != nil {
			return err
		}
	}

	return nil
}

func supplierOwnedGroupIDByPlatform(ctx context.Context, groupRepo GroupRepository, supplierID int64, platform string) ([]int64, error) {
	groups, err := groupRepo.ListActiveByPlatform(ctx, platform)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		if group.SupplierProfileID == nil || *group.SupplierProfileID != supplierID {
			continue
		}
		if group.IsSubscriptionType() && group.IsActive() {
			return []int64{group.ID}, nil
		}
	}
	return nil, ErrSupplierDefaultGroupMissing
}

func validateSupplierGroupIDs(ctx context.Context, groupRepo GroupRepository, ids []int64) error {
	for _, id := range ids {
		if id <= 0 {
			return infraerrors.BadRequest("SUPPLIER_GROUP_INVALID", "invalid group id")
		}
		if _, err := groupRepo.GetByID(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
