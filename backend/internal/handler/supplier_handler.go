package handler

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type SupplierHandler struct {
	supplierService *service.SupplierService
}

type supplierPricingIntervalRequest struct {
	MinTokens       int      `json:"min_tokens"`
	MaxTokens       *int     `json:"max_tokens"`
	TierLabel       string   `json:"tier_label"`
	InputPrice      *float64 `json:"input_price"`
	OutputPrice     *float64 `json:"output_price"`
	CacheWritePrice *float64 `json:"cache_write_price"`
	CacheReadPrice  *float64 `json:"cache_read_price"`
	PerRequestPrice *float64 `json:"per_request_price"`
	SortOrder       int      `json:"sort_order"`
}

type supplierChannelModelPricingRequest struct {
	Platform         string                           `json:"platform"`
	Models           []string                         `json:"models"`
	BillingMode      string                           `json:"billing_mode"`
	InputPrice       *float64                         `json:"input_price"`
	OutputPrice      *float64                         `json:"output_price"`
	CacheWritePrice  *float64                         `json:"cache_write_price"`
	CacheReadPrice   *float64                         `json:"cache_read_price"`
	ImageOutputPrice *float64                         `json:"image_output_price"`
	PerRequestPrice  *float64                         `json:"per_request_price"`
	Intervals        []supplierPricingIntervalRequest `json:"intervals"`
}

func NewSupplierHandler(supplierService *service.SupplierService) *SupplierHandler {
	return &SupplierHandler{supplierService: supplierService}
}

func supplierPricingRequestToService(reqs []supplierChannelModelPricingRequest) []service.ChannelModelPricing {
	result := make([]service.ChannelModelPricing, 0, len(reqs))
	for _, r := range reqs {
		billingMode := service.BillingMode(r.BillingMode)
		if billingMode == "" {
			billingMode = service.BillingModeToken
		}
		intervals := make([]service.PricingInterval, 0, len(r.Intervals))
		for _, iv := range r.Intervals {
			intervals = append(intervals, service.PricingInterval{
				MinTokens:       iv.MinTokens,
				MaxTokens:       iv.MaxTokens,
				TierLabel:       iv.TierLabel,
				InputPrice:      iv.InputPrice,
				OutputPrice:     iv.OutputPrice,
				CacheWritePrice: iv.CacheWritePrice,
				CacheReadPrice:  iv.CacheReadPrice,
				PerRequestPrice: iv.PerRequestPrice,
				SortOrder:       iv.SortOrder,
			})
		}
		result = append(result, service.ChannelModelPricing{
			Platform:         r.Platform,
			Models:           r.Models,
			BillingMode:      billingMode,
			InputPrice:       r.InputPrice,
			OutputPrice:      r.OutputPrice,
			CacheWritePrice:  r.CacheWritePrice,
			CacheReadPrice:   r.CacheReadPrice,
			ImageOutputPrice: r.ImageOutputPrice,
			PerRequestPrice:  r.PerRequestPrice,
			Intervals:        intervals,
		})
	}
	return result
}

func (h *SupplierHandler) ApplyProfile(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	var req service.SupplierProfileInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	profile, err := h.supplierService.ApplyProfile(c.Request.Context(), subject.UserID, req)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, supplierProfileResponse(profile))
}

func (h *SupplierHandler) GetProfile(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	profile, err := h.supplierService.GetProfileByUserID(c.Request.Context(), subject.UserID)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, supplierProfileResponse(profile))
}

func (h *SupplierHandler) ListAccounts(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	params := parsePaginationParams(c)
	accounts, page, err := h.supplierService.ListMyAccounts(c.Request.Context(), subject.UserID, params)
	if response.ErrorFrom(c, err) {
		return
	}
	items := make([]*dto.Account, 0, len(accounts))
	for i := range accounts {
		items = append(items, dto.AccountFromService(&accounts[i]))
	}
	response.Success(c, response.PaginatedData{
		Items:    items,
		Total:    page.Total,
		Page:     page.Page,
		PageSize: page.PageSize,
		Pages:    page.Pages,
	})
}

func (h *SupplierHandler) ListGroups(c *gin.Context) {
	groups, err := h.supplierService.ListSuggestionGroups(c.Request.Context(), c.Query("platform"))
	if response.ErrorFrom(c, err) {
		return
	}
	items := make([]*dto.Group, 0, len(groups))
	for i := range groups {
		items = append(items, dto.GroupFromService(&groups[i]))
	}
	response.Success(c, items)
}

func (h *SupplierHandler) ListProxies(c *gin.Context) {
	proxies, err := h.supplierService.ListSuggestionProxies(c.Request.Context())
	if response.ErrorFrom(c, err) {
		return
	}
	items := make([]*dto.Proxy, 0, len(proxies))
	for i := range proxies {
		items = append(items, dto.ProxyFromService(&proxies[i]))
	}
	response.Success(c, items)
}

func (h *SupplierHandler) ListModels(c *gin.Context) {
	models, err := h.supplierService.ListDefaultModels(c.Query("platform"), c.Query("type"))
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, models)
}

func (h *SupplierHandler) ModelPricing(c *gin.Context) {
	pricing, err := h.supplierService.GetModelDefaultPricing(c.Query("model"))
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.ChannelModelPricingFromService(pricing))
}

func (h *SupplierHandler) TestAccount(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	var req service.SupplierAccountPretestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.supplierService.TestSupplierAccount(c, subject.UserID, req)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, result)
}

func (h *SupplierHandler) TestAccountUpdate(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "invalid account id")
		return
	}
	var req service.SupplierAccountPretestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.supplierService.TestSupplierAccountUpdate(c, subject.UserID, accountID, req)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, result)
}

func (h *SupplierHandler) CreateAccount(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	var req struct {
		service.SupplierAccountInput
		SettlementPricing []supplierChannelModelPricingRequest `json:"settlement_pricing"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	input := req.SupplierAccountInput
	input.SettlementPricing = supplierPricingRequestToService(req.SettlementPricing)
	account, err := h.supplierService.CreateSupplierAccount(c.Request.Context(), subject.UserID, input)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Created(c, dto.AccountFromService(account))
}

func (h *SupplierHandler) UpdateAccount(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "invalid account id")
		return
	}
	var req struct {
		service.SupplierAccountInput
		SettlementPricing []supplierChannelModelPricingRequest `json:"settlement_pricing"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	input := req.SupplierAccountInput
	input.SettlementPricing = supplierPricingRequestToService(req.SettlementPricing)
	account, err := h.supplierService.UpdateSupplierAccount(c.Request.Context(), subject.UserID, accountID, input)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.AccountFromService(account))
}

func (h *SupplierHandler) RequestAccountEdit(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "invalid account id")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	account, err := h.supplierService.RequestAccountEdit(c.Request.Context(), subject.UserID, accountID, req.Reason)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.AccountFromService(account))
}

func (h *SupplierHandler) ListPricingRevisions(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "invalid account id")
		return
	}
	revisions, err := h.supplierService.ListMyPricingRevisions(c.Request.Context(), subject.UserID, accountID)
	if response.ErrorFrom(c, err) {
		return
	}
	items := make([]*dto.SupplierAccountPricingRevision, 0, len(revisions))
	for i := range revisions {
		items = append(items, dto.SupplierAccountPricingRevisionFromService(&revisions[i]))
	}
	response.Success(c, items)
}

func (h *SupplierHandler) SubmitPricingChange(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "invalid account id")
		return
	}
	var req struct {
		Pricing    []supplierChannelModelPricingRequest `json:"settlement_pricing"`
		SubmitNote string                               `json:"submit_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	revision, err := h.supplierService.SubmitPricingChange(
		c.Request.Context(),
		subject.UserID,
		accountID,
		supplierPricingRequestToService(req.Pricing),
		req.SubmitNote,
	)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Created(c, dto.SupplierAccountPricingRevisionFromService(revision))
}

func (h *SupplierHandler) UsageSummary(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	summary, err := h.supplierService.GetUsageSummary(c.Request.Context(), subject.UserID)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, summary)
}

func (h *SupplierHandler) DashboardStats(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	stats, err := h.supplierService.GetDashboardStats(c.Request.Context(), subject.UserID)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, stats)
}

func (h *SupplierHandler) DashboardTrend(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	startTime, endTime := parseUserTimeRange(c)
	granularity := c.DefaultQuery("granularity", "day")

	trend, err := h.supplierService.GetDashboardTrend(c.Request.Context(), subject.UserID, startTime, endTime, granularity)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, gin.H{
		"trend":       trend,
		"start_date":  startTime.Format("2006-01-02"),
		"end_date":    endTime.Add(-24 * time.Hour).Format("2006-01-02"),
		"granularity": granularity,
	})
}

func (h *SupplierHandler) DashboardModels(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	startTime, endTime := parseUserTimeRange(c)

	models, err := h.supplierService.GetDashboardModels(c.Request.Context(), subject.UserID, startTime, endTime)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, gin.H{
		"models":     models,
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   endTime.Add(-24 * time.Hour).Format("2006-01-02"),
	})
}

func (h *SupplierHandler) DashboardRecent(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	params := parsePaginationParams(c)
	startTime, endTime := parseUserTimeRange(c)

	logs, page, err := h.supplierService.ListRecentUsage(c.Request.Context(), subject.UserID, params, startTime, endTime)
	if response.ErrorFrom(c, err) {
		return
	}
	items := make([]*dto.UsageLog, 0, len(logs))
	for i := range logs {
		items = append(items, dto.UsageLogFromService(&logs[i]))
	}
	response.Success(c, response.PaginatedData{
		Items:    items,
		Total:    page.Total,
		Page:     page.Page,
		PageSize: page.PageSize,
		Pages:    page.Pages,
	})
}

func (h *SupplierHandler) ListSettlementStatements(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	params := parsePaginationParams(c)
	statements, page, err := h.supplierService.ListMySettlementStatements(c.Request.Context(), subject.UserID, params)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, response.PaginatedData{
		Items:    statements,
		Total:    page.Total,
		Page:     page.Page,
		PageSize: page.PageSize,
		Pages:    page.Pages,
	})
}

func parsePaginationParams(c *gin.Context) pagination.PaginationParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	return pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}
}

func supplierProfileResponse(profile *service.SupplierProfile) gin.H {
	if profile == nil {
		return gin.H{}
	}
	return gin.H{
		"id":                         profile.ID,
		"user_id":                    profile.UserID,
		"company_name":               profile.CompanyName,
		"contact_name":               profile.ContactName,
		"contact_email":              profile.ContactEmail,
		"contact_phone":              profile.ContactPhone,
		"status":                     profile.Status,
		"account_submission_enabled": profile.AccountSubmissionEnabled,
		"settlement_config":          profile.SettlementConfig,
		"notes":                      profile.Notes,
		"review_note":                profile.ReviewNote,
		"reviewed_at":                profile.ReviewedAt,
		"reviewed_by":                profile.ReviewedBy,
		"created_at":                 profile.CreatedAt,
		"updated_at":                 profile.UpdatedAt,
		"user":                       supplierUserResponse(profile.User),
	}
}

func supplierUserResponse(user *service.User) any {
	if user == nil {
		return nil
	}
	return gin.H{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
		"role":     user.Role,
		"status":   user.Status,
	}
}
