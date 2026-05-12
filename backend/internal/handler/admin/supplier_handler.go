package admin

import (
	"strconv"

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

func parseAdminPagination(c *gin.Context) pagination.PaginationParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	return pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}
}

func NewSupplierHandler(supplierService *service.SupplierService) *SupplierHandler {
	return &SupplierHandler{supplierService: supplierService}
}

func (h *SupplierHandler) ListProfiles(c *gin.Context) {
	profiles, page, err := h.supplierService.ListProfiles(c.Request.Context(), parseAdminPagination(c), c.Query("status"))
	if response.ErrorFrom(c, err) {
		return
	}
	items := make([]gin.H, 0, len(profiles))
	for i := range profiles {
		items = append(items, supplierProfileAdminResponse(&profiles[i]))
	}
	response.Success(c, response.PaginatedData{
		Items:    items,
		Total:    page.Total,
		Page:     page.Page,
		PageSize: page.PageSize,
		Pages:    page.Pages,
	})
}

func (h *SupplierHandler) ListAccounts(c *gin.Context) {
	params := parseAdminPagination(c)
	supplierID, _ := strconv.ParseInt(c.Query("supplier_id"), 10, 64)
	accounts, page, err := h.supplierService.ListSupplierAccounts(c.Request.Context(), params, service.SupplierAccountListFilters{
		SupplierID:     supplierID,
		ApprovalStatus: c.Query("approval_status"),
	})
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

func (h *SupplierHandler) ReviewProfile(c *gin.Context) {
	supplierID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || supplierID <= 0 {
		response.BadRequest(c, "invalid supplier id")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	var req struct {
		Status     string `json:"status"`
		ReviewNote string `json:"review_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	profile, err := h.supplierService.ReviewProfile(c.Request.Context(), supplierID, subject.UserID, req.Status, req.ReviewNote)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, supplierProfileAdminResponse(profile))
}

func (h *SupplierHandler) ApproveAccount(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "invalid account id")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	var req service.SupplierAccountApprovalInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	account, err := h.supplierService.ApproveAccount(c.Request.Context(), accountID, subject.UserID, req)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.AccountFromService(account))
}

func (h *SupplierHandler) RejectAccount(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "invalid account id")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	var req struct {
		RejectReason string `json:"reject_reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	account, err := h.supplierService.RejectAccount(c.Request.Context(), accountID, subject.UserID, req.RejectReason)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.AccountFromService(account))
}

func (h *SupplierHandler) ReturnAccountForEdit(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "invalid account id")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	var req struct {
		ReviewNote string `json:"review_note"`
	}
	_ = c.ShouldBindJSON(&req)
	account, err := h.supplierService.ReturnAccountForEdit(c.Request.Context(), accountID, subject.UserID, req.ReviewNote)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.AccountFromService(account))
}

func (h *SupplierHandler) RejectAccountEditRequest(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "invalid account id")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	var req struct {
		ReviewNote string `json:"review_note"`
	}
	_ = c.ShouldBindJSON(&req)
	account, err := h.supplierService.RejectAccountEditRequest(c.Request.Context(), accountID, subject.UserID, req.ReviewNote)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.AccountFromService(account))
}

func (h *SupplierHandler) ListPricingRevisions(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "invalid account id")
		return
	}
	revisions, err := h.supplierService.ListAccountPricingRevisions(c.Request.Context(), accountID)
	if response.ErrorFrom(c, err) {
		return
	}
	items := make([]*dto.SupplierAccountPricingRevision, 0, len(revisions))
	for i := range revisions {
		items = append(items, dto.SupplierAccountPricingRevisionFromService(&revisions[i]))
	}
	response.Success(c, items)
}

func (h *SupplierHandler) ApprovePricingRevision(c *gin.Context) {
	revisionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || revisionID <= 0 {
		response.BadRequest(c, "invalid revision id")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	var req struct {
		ReviewNote string `json:"review_note"`
	}
	_ = c.ShouldBindJSON(&req)
	revision, err := h.supplierService.ApprovePricingRevision(c.Request.Context(), revisionID, subject.UserID, req.ReviewNote)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.SupplierAccountPricingRevisionFromService(revision))
}

func (h *SupplierHandler) RejectPricingRevision(c *gin.Context) {
	revisionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || revisionID <= 0 {
		response.BadRequest(c, "invalid revision id")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	var req struct {
		ReviewNote string `json:"review_note"`
	}
	_ = c.ShouldBindJSON(&req)
	revision, err := h.supplierService.RejectPricingRevision(c.Request.Context(), revisionID, subject.UserID, req.ReviewNote)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.SupplierAccountPricingRevisionFromService(revision))
}

func (h *SupplierHandler) ListSettlementStatements(c *gin.Context) {
	supplierID, _ := strconv.ParseInt(c.Query("supplier_id"), 10, 64)
	periodStart, err := service.ParseSupplierSettlementPeriodMonth(c.Query("period_month"))
	if response.ErrorFrom(c, err) {
		return
	}
	items, page, err := h.supplierService.ListSettlementStatements(c.Request.Context(), parseAdminPagination(c), service.SupplierSettlementListFilters{
		SupplierID:  supplierID,
		Status:      c.Query("status"),
		PeriodStart: periodStart,
	})
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, paginatedSupplierSettlementStatements(items, page))
}

func (h *SupplierHandler) GenerateSettlementStatement(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	var req service.SupplierSettlementGenerateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	statement, err := h.supplierService.GenerateSettlementStatement(c.Request.Context(), req, subject.UserID)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, statement)
}

func (h *SupplierHandler) ConfirmSettlementStatement(c *gin.Context) {
	statementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || statementID <= 0 {
		response.BadRequest(c, "invalid statement id")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	var req service.SupplierSettlementConfirmInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	statement, err := h.supplierService.ConfirmSettlementStatement(c.Request.Context(), statementID, subject.UserID, req)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, statement)
}

func (h *SupplierHandler) MarkSettlementStatementPaid(c *gin.Context) {
	statementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || statementID <= 0 {
		response.BadRequest(c, "invalid statement id")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	var req service.SupplierSettlementPaymentInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	statement, err := h.supplierService.MarkSettlementStatementPaid(c.Request.Context(), statementID, subject.UserID, req)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, statement)
}

func (h *SupplierHandler) VoidSettlementStatement(c *gin.Context) {
	statementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || statementID <= 0 {
		response.BadRequest(c, "invalid statement id")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	statement, err := h.supplierService.VoidSettlementStatement(c.Request.Context(), statementID, subject.UserID)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, statement)
}

func paginatedSupplierSettlementStatements(items []service.SupplierSettlementStatement, page *pagination.PaginationResult) response.PaginatedData {
	return response.PaginatedData{
		Items:    items,
		Total:    page.Total,
		Page:     page.Page,
		PageSize: page.PageSize,
		Pages:    page.Pages,
	}
}

func supplierProfileAdminResponse(profile *service.SupplierProfile) gin.H {
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
		"user":                       supplierAdminUserResponse(profile.User),
	}
}

func supplierAdminUserResponse(user *service.User) any {
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
