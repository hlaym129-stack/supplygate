package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	AvailableChannelSourceConfigured      = "configured"
	AvailableChannelSourceSupplierAccount = "supplier_account"
)

// AvailableGroupRef 渠道视图中关联分组的简要信息。
//
// 用户侧「可用渠道」页面据此展示：专属分组 vs 公开分组（IsExclusive）、
// 订阅 vs 标准（SubscriptionType）、默认倍率（RateMultiplier）。用户专属倍率
// 不在这里暴露，前端自己通过 /groups/rates 拉取，和 API 密钥页面保持一致。
type AvailableGroupRef struct {
	ID               int64
	Name             string
	Platform         string
	SubscriptionType string
	RateMultiplier   float64
	IsExclusive      bool
}

// AvailableChannel 可用渠道视图：用于「可用渠道」页面展示渠道基础信息 +
// 关联的分组 + 推导出的支持模型列表（无通配符）。
type AvailableChannel struct {
	ID                 int64
	Name               string
	Description        string
	Source             string
	Status             string
	BillingModelSource string
	RestrictModels     bool
	Groups             []AvailableGroupRef
	SupportedModels    []SupportedModel
}

// ListAvailable 返回所有渠道的可用视图：每个渠道附带关联分组信息与支持模型列表。
//
// 支持模型通过 (*Channel).SupportedModels() 计算（mapping ∪ pricing 并联）。
// 对于渠道未配置定价的模型，进一步用 PricingService 的全局 LiteLLM 数据合成
// 一份展示用定价，让用户看到默认价格而非"未配置"。
//
// 关联分组信息通过 groupRepo.ListActive 查询后按 ID 映射；渠道 GroupIDs 中未在活跃列表中
// 的分组（已停用或删除）会被忽略。
//
// 前置条件：s.groupRepo 必须非 nil（由 wire DI 保证）。直接 nil-deref 用于 fail-fast，
// 避免静默掩盖注入缺失。
func (s *ChannelService) ListAvailable(ctx context.Context) ([]AvailableChannel, error) {
	channels, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}

	groups, err := s.groupRepo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active groups: %w", err)
	}
	groupByID := make(map[int64]AvailableGroupRef, len(groups))
	for i := range groups {
		g := groups[i]
		groupByID[g.ID] = AvailableGroupRef{
			ID:               g.ID,
			Name:             g.Name,
			Platform:         g.Platform,
			SubscriptionType: g.SubscriptionType,
			RateMultiplier:   g.RateMultiplier,
			IsExclusive:      g.IsExclusive,
		}
	}

	out := make([]AvailableChannel, 0, len(channels))
	for i := range channels {
		ch := &channels[i]
		groups := make([]AvailableGroupRef, 0, len(ch.GroupIDs))
		for _, gid := range ch.GroupIDs {
			if ref, ok := groupByID[gid]; ok {
				groups = append(groups, ref)
			}
		}
		sort.SliceStable(groups, func(i, j int) bool { return groups[i].Name < groups[j].Name })

		ch.normalizeBillingModelSource()

		supported := ch.SupportedModels()
		s.fillGlobalPricingFallback(supported)

		out = append(out, AvailableChannel{
			ID:                 ch.ID,
			Name:               ch.Name,
			Description:        ch.Description,
			Source:             AvailableChannelSourceConfigured,
			Status:             ch.Status,
			BillingModelSource: ch.BillingModelSource,
			RestrictModels:     ch.RestrictModels,
			Groups:             groups,
			SupportedModels:    supported,
		})
	}

	supplierChannels, err := s.listAvailableSupplierChannels(ctx, groupByID)
	if err != nil {
		return nil, err
	}
	out = append(out, supplierChannels...)

	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

func (s *ChannelService) listAvailableSupplierChannels(ctx context.Context, groupByID map[int64]AvailableGroupRef) ([]AvailableChannel, error) {
	if s.accountRepo == nil || s.supplierRepo == nil {
		return nil, nil
	}

	accounts, err := s.accountRepo.ListAvailableSupplierAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list supplier available accounts: %w", err)
	}
	if len(accounts) == 0 {
		return nil, nil
	}

	accountIDs := make([]int64, 0, len(accounts))
	for i := range accounts {
		if accounts[i].ID > 0 {
			accountIDs = append(accountIDs, accounts[i].ID)
		}
	}
	revisionsByAccount, err := s.supplierRepo.ListApprovedPricingRevisionsForAccounts(ctx, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("list supplier approved pricing revisions: %w", err)
	}

	now := time.Now()
	out := make([]AvailableChannel, 0, len(accounts))
	for i := range accounts {
		account := accounts[i]
		if !supplierAccountEligibleForAvailable(account) {
			continue
		}
		groups := supplierAccountAvailableGroups(account, groupByID)
		if len(groups) == 0 {
			continue
		}
		current, scheduled := selectSupplierPricingRevisions(revisionsByAccount[account.ID], now)
		if current == nil {
			continue
		}
		supported := supplierSupportedModels(account, current, scheduled)
		if len(supported) == 0 {
			continue
		}
		out = append(out, AvailableChannel{
			Name:            fmt.Sprintf("供应商：%s / %s", supplierDisplayName(account), supplierAccountDisplayName(account)),
			Description:     "报价来自供应商审核通过的账号；若存在待生效报价，生效前仍按当前价格结算。",
			Source:          AvailableChannelSourceSupplierAccount,
			Status:          StatusActive,
			Groups:          groups,
			SupportedModels: supported,
		})
	}
	return out, nil
}

func supplierAccountEligibleForAvailable(account Account) bool {
	return account.OwnerType == AccountOwnerTypeSupplier &&
		account.ApprovalStatus == AccountApprovalStatusApproved &&
		account.Status == StatusActive &&
		account.Schedulable
}

func supplierAccountAvailableGroups(account Account, groupByID map[int64]AvailableGroupRef) []AvailableGroupRef {
	ids := make([]int64, 0, len(account.GroupIDs)+len(account.AccountGroups))
	ids = append(ids, account.GroupIDs...)
	for _, ag := range account.AccountGroups {
		ids = append(ids, ag.GroupID)
	}

	seen := make(map[int64]struct{}, len(ids))
	groups := make([]AvailableGroupRef, 0, len(ids))
	for _, gid := range ids {
		if gid <= 0 {
			continue
		}
		if _, ok := seen[gid]; ok {
			continue
		}
		seen[gid] = struct{}{}
		ref, ok := groupByID[gid]
		if !ok {
			continue
		}
		if account.Platform != "" && ref.Platform != account.Platform {
			continue
		}
		groups = append(groups, ref)
	}
	sort.SliceStable(groups, func(i, j int) bool { return groups[i].Name < groups[j].Name })
	return groups
}

func selectSupplierPricingRevisions(revisions []SupplierAccountPricingRevision, now time.Time) (*SupplierAccountPricingRevision, *SupplierAccountPricingRevision) {
	var current *SupplierAccountPricingRevision
	var scheduled *SupplierAccountPricingRevision
	for i := range revisions {
		rev := &revisions[i]
		if rev.Status != SupplierPricingRevisionStatusApproved || rev.EffectiveAt == nil {
			continue
		}
		if !rev.EffectiveAt.After(now) {
			if current == nil ||
				rev.EffectiveAt.After(*current.EffectiveAt) ||
				(rev.EffectiveAt.Equal(*current.EffectiveAt) && rev.ID > current.ID) {
				current = rev
			}
			continue
		}
		if scheduled == nil ||
			rev.EffectiveAt.Before(*scheduled.EffectiveAt) ||
			(rev.EffectiveAt.Equal(*scheduled.EffectiveAt) && rev.ID > scheduled.ID) {
			scheduled = rev
		}
	}
	return current, scheduled
}

func supplierSupportedModels(account Account, current *SupplierAccountPricingRevision, scheduled *SupplierAccountPricingRevision) []SupportedModel {
	if current == nil {
		return nil
	}

	scheduledByModel := make(map[string]*ChannelModelPricing)
	if scheduled != nil {
		for i := range scheduled.Pricing {
			pricing := scheduled.Pricing[i]
			for _, model := range pricing.Models {
				model = strings.TrimSpace(model)
				if model == "" {
					continue
				}
				if _, wild := splitWildcardSuffix(model); wild {
					continue
				}
				platform := strings.TrimSpace(pricing.Platform)
				if platform == "" {
					platform = account.Platform
				}
				if account.Platform != "" && platform != account.Platform {
					continue
				}
				scheduledByModel[strings.ToLower(model)] = cloneSupplierModelPricing(pricing, platform, model)
			}
		}
	}

	seen := make(map[string]struct{})
	out := make([]SupportedModel, 0)
	for i := range current.Pricing {
		pricing := current.Pricing[i]
		platform := strings.TrimSpace(pricing.Platform)
		if platform == "" {
			platform = account.Platform
		}
		if account.Platform != "" && platform != account.Platform {
			continue
		}
		for _, model := range pricing.Models {
			model = strings.TrimSpace(model)
			if model == "" {
				continue
			}
			if _, wild := splitWildcardSuffix(model); wild {
				continue
			}
			key := strings.ToLower(model)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			supported := SupportedModel{
				Name:               model,
				Platform:           platform,
				Pricing:            cloneSupplierModelPricing(pricing, platform, model),
				PricingEffectiveAt: current.EffectiveAt,
			}
			if scheduledPricing, ok := scheduledByModel[key]; ok && scheduled != nil {
				supported.ScheduledPricing = scheduledPricing
				supported.ScheduledEffectiveAt = scheduled.EffectiveAt
			}
			out = append(out, supported)
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func cloneSupplierModelPricing(src ChannelModelPricing, platform, model string) *ChannelModelPricing {
	cp := src.Clone()
	cp.ID = 0
	cp.ChannelID = 0
	cp.Platform = platform
	cp.Models = []string{model}
	if cp.BillingMode == "" {
		cp.BillingMode = BillingModeToken
	}
	return &cp
}

func supplierDisplayName(account Account) string {
	if account.Supplier != nil {
		if name := strings.TrimSpace(account.Supplier.CompanyName); name != "" {
			return name
		}
	}
	return "供应商"
}

func supplierAccountDisplayName(account Account) string {
	if name := strings.TrimSpace(account.Name); name != "" {
		return name
	}
	return "未命名账号"
}

// fillGlobalPricingFallback 对未命中渠道定价的支持模型，从全局 LiteLLM 数据合成一份
// 展示用定价（按 token 计费）。仅用于「可用渠道」展示，不影响真实计费链路。
//
// 当 s.pricingService 为 nil（测试场景），跳过回落。
func (s *ChannelService) fillGlobalPricingFallback(models []SupportedModel) {
	if s.pricingService == nil {
		return
	}
	for i := range models {
		if models[i].Pricing != nil {
			continue
		}
		lp := s.pricingService.GetModelPricing(models[i].Name)
		if lp == nil {
			continue
		}
		models[i].Pricing = synthesizePricingFromLiteLLM(lp)
	}
}

// synthesizePricingFromLiteLLM 把 LiteLLM 的定价数据转成 ChannelModelPricing 形态，
// 仅用于展示。BillingMode 固定为 token；图片场景的 OutputCostPerImageToken 也归到
// ImageOutputPrice 字段（与渠道侧"图片输出按 token 计价"语义一致）。
//
// LiteLLM 中字段 0 视为未配置，不带入展示。
func synthesizePricingFromLiteLLM(lp *LiteLLMModelPricing) *ChannelModelPricing {
	if lp == nil {
		return nil
	}
	return &ChannelModelPricing{
		BillingMode:      BillingModeToken,
		InputPrice:       nonZeroPtr(lp.InputCostPerToken),
		OutputPrice:      nonZeroPtr(lp.OutputCostPerToken),
		CacheWritePrice:  nonZeroPtr(lp.CacheCreationInputTokenCost),
		CacheReadPrice:   nonZeroPtr(lp.CacheReadInputTokenCost),
		ImageOutputPrice: nonZeroPtr(lp.OutputCostPerImageToken),
	}
}

func nonZeroPtr(v float64) *float64 {
	if v == 0 {
		return nil
	}
	return &v
}
