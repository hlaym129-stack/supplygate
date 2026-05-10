package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type SupplierMarketplaceHandler struct {
	marketplaceService *service.SupplierMarketplaceService
}

func NewSupplierMarketplaceHandler(marketplaceService *service.SupplierMarketplaceService) *SupplierMarketplaceHandler {
	return &SupplierMarketplaceHandler{marketplaceService: marketplaceService}
}

type supplierMarketplaceGroupSummaryResponse struct {
	ID               int64    `json:"id"`
	Name             string   `json:"name"`
	Platform         string   `json:"platform"`
	SubscriptionType string   `json:"subscription_type"`
	ModelCount       int      `json:"model_count"`
	Models           []string `json:"models"`
	IsSubscribed     bool     `json:"is_subscribed"`
}

type supplierMarketplaceSupplierResponse struct {
	ID           int64                                     `json:"id"`
	CompanyName  string                                    `json:"company_name"`
	Notes        string                                    `json:"notes"`
	IsSubscribed bool                                      `json:"is_subscribed"`
	Groups       []supplierMarketplaceGroupSummaryResponse `json:"groups"`
}

type supplierMarketplaceModelResponse struct {
	Name                 string                   `json:"name"`
	Platform             string                   `json:"platform"`
	Pricing              *dto.ChannelModelPricing `json:"pricing"`
	PricingEffectiveAt   *string                  `json:"pricing_effective_at,omitempty"`
	ScheduledPricing     *dto.ChannelModelPricing `json:"scheduled_pricing,omitempty"`
	ScheduledEffectiveAt *string                  `json:"scheduled_effective_at,omitempty"`
}

type supplierMarketplaceGroupDetailResponse struct {
	ID               int64                              `json:"id"`
	Name             string                             `json:"name"`
	Platform         string                             `json:"platform"`
	SubscriptionType string                             `json:"subscription_type"`
	IsSubscribed     bool                               `json:"is_subscribed"`
	Models           []supplierMarketplaceModelResponse `json:"models"`
}

type supplierMarketplaceSupplierDetailResponse struct {
	ID           int64                                    `json:"id"`
	CompanyName  string                                   `json:"company_name"`
	Notes        string                                   `json:"notes"`
	IsSubscribed bool                                     `json:"is_subscribed"`
	Groups       []supplierMarketplaceGroupDetailResponse `json:"groups"`
}

func (h *SupplierMarketplaceHandler) ListSuppliers(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	suppliers, err := h.marketplaceService.ListSuppliers(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]supplierMarketplaceSupplierResponse, 0, len(suppliers))
	for i := range suppliers {
		groups := make([]supplierMarketplaceGroupSummaryResponse, 0, len(suppliers[i].Groups))
		for _, group := range suppliers[i].Groups {
			groups = append(groups, supplierMarketplaceGroupSummaryResponse{
				ID:               group.GroupID,
				Name:             group.Name,
				Platform:         group.Platform,
				SubscriptionType: group.SubscriptionType,
				ModelCount:       group.ModelCount,
				Models:           group.Models,
				IsSubscribed:     group.IsSubscribed,
			})
		}
		out = append(out, supplierMarketplaceSupplierResponse{
			ID:           suppliers[i].SupplierID,
			CompanyName:  suppliers[i].CompanyName,
			Notes:        suppliers[i].Notes,
			IsSubscribed: suppliers[i].IsSubscribed,
			Groups:       groups,
		})
	}

	response.Success(c, out)
}

func (h *SupplierMarketplaceHandler) GetSupplierDetail(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	supplierID, err := strconv.ParseInt(c.Param("supplierId"), 10, 64)
	if err != nil || supplierID <= 0 {
		response.BadRequest(c, "invalid supplier id")
		return
	}

	detail, err := h.marketplaceService.GetSupplierDetail(c.Request.Context(), subject.UserID, supplierID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := supplierMarketplaceSupplierDetailResponse{
		ID:           detail.SupplierID,
		CompanyName:  detail.CompanyName,
		Notes:        detail.Notes,
		IsSubscribed: detail.IsSubscribed,
		Groups:       make([]supplierMarketplaceGroupDetailResponse, 0, len(detail.Groups)),
	}
	for _, group := range detail.Groups {
		models := make([]supplierMarketplaceModelResponse, 0, len(group.Models))
		for _, model := range group.Models {
			item := supplierMarketplaceModelResponse{
				Name:     model.Name,
				Platform: model.Platform,
			}
			if model.Pricing != nil {
				pricing := dto.ChannelModelPricingFromService(model.Pricing)
				item.Pricing = &pricing
			}
			if model.PricingEffectiveAt != nil {
				formatted := model.PricingEffectiveAt.Format(timeFormatRFC3339)
				item.PricingEffectiveAt = &formatted
			}
			if model.ScheduledPricing != nil {
				pricing := dto.ChannelModelPricingFromService(model.ScheduledPricing)
				item.ScheduledPricing = &pricing
			}
			if model.ScheduledEffectiveAt != nil {
				formatted := model.ScheduledEffectiveAt.Format(timeFormatRFC3339)
				item.ScheduledEffectiveAt = &formatted
			}
			models = append(models, item)
		}
		out.Groups = append(out.Groups, supplierMarketplaceGroupDetailResponse{
			ID:               group.GroupID,
			Name:             group.Name,
			Platform:         group.Platform,
			SubscriptionType: group.SubscriptionType,
			IsSubscribed:     group.IsSubscribed,
			Models:           models,
		})
	}

	response.Success(c, out)
}

func (h *SupplierMarketplaceHandler) SubscribeGroup(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	groupID, err := strconv.ParseInt(c.Param("groupId"), 10, 64)
	if err != nil || groupID <= 0 {
		response.BadRequest(c, "invalid group id")
		return
	}

	subscription, err := h.marketplaceService.SubscribeGroup(c.Request.Context(), subject.UserID, groupID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"group_id":        groupID,
		"subscription_id": subscription.ID,
		"is_subscribed":   true,
	})
}

func (h *SupplierMarketplaceHandler) UnsubscribeGroup(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	groupID, err := strconv.ParseInt(c.Param("groupId"), 10, 64)
	if err != nil || groupID <= 0 {
		response.BadRequest(c, "invalid group id")
		return
	}

	if err := h.marketplaceService.UnsubscribeGroup(c.Request.Context(), subject.UserID, groupID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"group_id":      groupID,
		"is_subscribed": false,
	})
}

const timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"
