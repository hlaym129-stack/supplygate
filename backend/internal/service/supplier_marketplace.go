package service

import (
	"context"
	"sort"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrSupplierMarketplaceSupplierNotFound = infraerrors.NotFound("SUPPLIER_MARKETPLACE_SUPPLIER_NOT_FOUND", "supplier marketplace entry not found")
	ErrSupplierMarketplaceGroupInvalid     = infraerrors.BadRequest("SUPPLIER_MARKETPLACE_GROUP_INVALID", "group is not a supplier marketplace subscription group")
)

const supplierMarketplaceSubscriptionNote = "supplier_marketplace"

type SupplierMarketplaceSupplier struct {
	SupplierID   int64
	CompanyName  string
	Notes        string
	IsSubscribed bool
	Groups       []SupplierMarketplaceGroupSummary
}

type SupplierMarketplaceGroupSummary struct {
	GroupID          int64
	Name             string
	Platform         string
	SubscriptionType string
	ModelCount       int
	Models           []string
	IsSubscribed     bool
}

type SupplierMarketplaceSupplierDetail struct {
	SupplierID   int64
	CompanyName  string
	Notes        string
	IsSubscribed bool
	Groups       []SupplierMarketplaceGroupDetail
}

type SupplierMarketplaceGroupDetail struct {
	GroupID          int64
	Name             string
	Platform         string
	SubscriptionType string
	IsSubscribed     bool
	Models           []SupportedModel
}

type SupplierMarketplaceService struct {
	accountRepo         SupplierAvailableAccountRepository
	supplierRepo        SupplierRepository
	groupRepo           GroupRepository
	userSubRepo         UserSubscriptionRepository
	subscriptionService *SubscriptionService
}

func NewSupplierMarketplaceService(
	accountRepo SupplierAvailableAccountRepository,
	supplierRepo SupplierRepository,
	groupRepo GroupRepository,
	userSubRepo UserSubscriptionRepository,
	subscriptionService *SubscriptionService,
) *SupplierMarketplaceService {
	return &SupplierMarketplaceService{
		accountRepo:         accountRepo,
		supplierRepo:        supplierRepo,
		groupRepo:           groupRepo,
		userSubRepo:         userSubRepo,
		subscriptionService: subscriptionService,
	}
}

func (s *SupplierMarketplaceService) ListSuppliers(ctx context.Context, userID int64) ([]SupplierMarketplaceSupplier, error) {
	aggregates, err := s.loadMarketplaceAggregates(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]SupplierMarketplaceSupplier, 0, len(aggregates))
	for _, aggregate := range aggregates {
		groups := make([]SupplierMarketplaceGroupSummary, 0, len(aggregate.groups))
		for _, group := range aggregate.sortedGroups() {
			modelNames := group.sortedModelNames()
			groups = append(groups, SupplierMarketplaceGroupSummary{
				GroupID:          group.GroupID,
				Name:             group.Name,
				Platform:         group.Platform,
				SubscriptionType: group.SubscriptionType,
				ModelCount:       len(modelNames),
				Models:           modelNames,
				IsSubscribed:     group.IsSubscribed,
			})
		}
		out = append(out, SupplierMarketplaceSupplier{
			SupplierID:   aggregate.supplier.ID,
			CompanyName:  strings.TrimSpace(aggregate.supplier.CompanyName),
			Notes:        strings.TrimSpace(aggregate.supplier.Notes),
			IsSubscribed: aggregate.isSubscribed(),
			Groups:       groups,
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].CompanyName) < strings.ToLower(out[j].CompanyName)
	})
	return out, nil
}

func (s *SupplierMarketplaceService) GetSupplierDetail(ctx context.Context, userID, supplierID int64) (*SupplierMarketplaceSupplierDetail, error) {
	aggregates, err := s.loadMarketplaceAggregates(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, aggregate := range aggregates {
		if aggregate.supplier.ID != supplierID {
			continue
		}
		detail := &SupplierMarketplaceSupplierDetail{
			SupplierID:   aggregate.supplier.ID,
			CompanyName:  strings.TrimSpace(aggregate.supplier.CompanyName),
			Notes:        strings.TrimSpace(aggregate.supplier.Notes),
			IsSubscribed: aggregate.isSubscribed(),
			Groups:       make([]SupplierMarketplaceGroupDetail, 0, len(aggregate.groups)),
		}
		for _, group := range aggregate.sortedGroups() {
			detail.Groups = append(detail.Groups, SupplierMarketplaceGroupDetail{
				GroupID:          group.GroupID,
				Name:             group.Name,
				Platform:         group.Platform,
				SubscriptionType: group.SubscriptionType,
				IsSubscribed:     group.IsSubscribed,
				Models:           group.sortedModels(),
			})
		}
		return detail, nil
	}

	return nil, ErrSupplierMarketplaceSupplierNotFound
}

func (s *SupplierMarketplaceService) SubscribeGroup(ctx context.Context, userID, groupID int64) (*UserSubscription, error) {
	group, err := s.validateMarketplaceGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	existing, err := s.userSubRepo.GetByUserIDAndGroupID(ctx, userID, groupID)
	if err == nil && existing != nil {
		if existing.ExpiresAt.Before(MaxExpiresAt) {
			if err := s.userSubRepo.ExtendExpiry(ctx, existing.ID, MaxExpiresAt); err != nil {
				return nil, err
			}
		}
		if existing.Status != SubscriptionStatusActive {
			if err := s.userSubRepo.UpdateStatus(ctx, existing.ID, SubscriptionStatusActive); err != nil {
				return nil, err
			}
		}
		s.invalidateSubscription(userID, groupID)
		return s.userSubRepo.GetByID(ctx, existing.ID)
	}
	if err != nil && !infraerrors.IsNotFound(err) {
		return nil, err
	}

	now := time.Now()
	sub := &UserSubscription{
		UserID:     userID,
		GroupID:    group.ID,
		StartsAt:   now,
		ExpiresAt:  MaxExpiresAt,
		Status:     SubscriptionStatusActive,
		AssignedAt: now,
		Notes:      supplierMarketplaceSubscriptionNote,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.userSubRepo.Create(ctx, sub); err != nil {
		return nil, err
	}
	s.invalidateSubscription(userID, groupID)
	return s.userSubRepo.GetByID(ctx, sub.ID)
}

func (s *SupplierMarketplaceService) UnsubscribeGroup(ctx context.Context, userID, groupID int64) error {
	if _, err := s.validateMarketplaceGroup(ctx, groupID); err != nil {
		return err
	}

	existing, err := s.userSubRepo.GetByUserIDAndGroupID(ctx, userID, groupID)
	if err != nil {
		if infraerrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	now := time.Now()
	if err := s.userSubRepo.UpdateStatus(ctx, existing.ID, SubscriptionStatusExpired); err != nil {
		return err
	}
	if err := s.userSubRepo.ExtendExpiry(ctx, existing.ID, now); err != nil {
		return err
	}
	s.invalidateSubscription(userID, groupID)
	return nil
}

func (s *SupplierMarketplaceService) validateMarketplaceGroup(ctx context.Context, groupID int64) (*Group, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.SupplierProfileID == nil || !group.IsSubscriptionType() || !group.IsActive() {
		return nil, ErrSupplierMarketplaceGroupInvalid
	}
	return group, nil
}

func (s *SupplierMarketplaceService) invalidateSubscription(userID, groupID int64) {
	if s.subscriptionService == nil {
		return
	}
	s.subscriptionService.InvalidateSubCache(userID, groupID)
	if s.subscriptionService.billingCacheService != nil {
		_ = s.subscriptionService.billingCacheService.InvalidateSubscription(context.Background(), userID, groupID)
	}
}

type supplierMarketplaceAggregate struct {
	supplier *SupplierProfile
	groups   map[int64]*supplierMarketplaceGroupAggregate
}

func (a *supplierMarketplaceAggregate) isSubscribed() bool {
	for _, group := range a.groups {
		if group.IsSubscribed {
			return true
		}
	}
	return false
}

func (a *supplierMarketplaceAggregate) sortedGroups() []*supplierMarketplaceGroupAggregate {
	out := make([]*supplierMarketplaceGroupAggregate, 0, len(a.groups))
	for _, group := range a.groups {
		out = append(out, group)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Platform == out[j].Platform {
			return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
		}
		return strings.ToLower(out[i].Platform) < strings.ToLower(out[j].Platform)
	})
	return out
}

type supplierMarketplaceGroupAggregate struct {
	GroupID          int64
	Name             string
	Platform         string
	SubscriptionType string
	IsSubscribed     bool
	modelsByName     map[string]SupportedModel
}

func (g *supplierMarketplaceGroupAggregate) addModel(model SupportedModel) {
	if g.modelsByName == nil {
		g.modelsByName = make(map[string]SupportedModel)
	}
	key := strings.ToLower(strings.TrimSpace(model.Name))
	if key == "" {
		return
	}
	existing, ok := g.modelsByName[key]
	if !ok || shouldReplaceMarketplaceModel(existing, model) {
		g.modelsByName[key] = model
	}
}

func (g *supplierMarketplaceGroupAggregate) sortedModelNames() []string {
	models := g.sortedModels()
	out := make([]string, 0, len(models))
	for _, model := range models {
		out = append(out, model.Name)
	}
	return out
}

func (g *supplierMarketplaceGroupAggregate) sortedModels() []SupportedModel {
	out := make([]SupportedModel, 0, len(g.modelsByName))
	for _, model := range g.modelsByName {
		out = append(out, model)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

func shouldReplaceMarketplaceModel(current SupportedModel, candidate SupportedModel) bool {
	currentScore, currentOK := marketplacePricingScore(current.Pricing)
	candidateScore, candidateOK := marketplacePricingScore(candidate.Pricing)
	switch {
	case !currentOK && candidateOK:
		return true
	case currentOK && !candidateOK:
		return false
	case !currentOK && !candidateOK:
		return false
	default:
		return candidateScore < currentScore
	}
}

func marketplacePricingScore(pricing *ChannelModelPricing) (float64, bool) {
	if pricing == nil {
		return 0, false
	}
	switch pricing.BillingMode {
	case BillingModePerRequest:
		if pricing.PerRequestPrice != nil {
			return *pricing.PerRequestPrice, true
		}
	case BillingModeImage:
		if pricing.ImageOutputPrice != nil {
			return *pricing.ImageOutputPrice, true
		}
	default:
		switch {
		case pricing.InputPrice != nil && pricing.OutputPrice != nil:
			return *pricing.InputPrice + *pricing.OutputPrice, true
		case pricing.InputPrice != nil:
			return *pricing.InputPrice, true
		case pricing.OutputPrice != nil:
			return *pricing.OutputPrice, true
		}
	}
	return 0, false
}

func (s *SupplierMarketplaceService) loadMarketplaceAggregates(ctx context.Context, userID int64) ([]*supplierMarketplaceAggregate, error) {
	groups, err := s.groupRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	marketGroups := make(map[int64]*Group)
	for i := range groups {
		group := groups[i]
		if group.SupplierProfileID == nil || !group.IsSubscriptionType() || !group.IsActive() {
			continue
		}
		cp := group
		marketGroups[group.ID] = &cp
	}
	if len(marketGroups) == 0 {
		return []*supplierMarketplaceAggregate{}, nil
	}

	subscribedGroupIDs := make(map[int64]bool)
	if userID > 0 {
		subs, err := s.userSubRepo.ListActiveByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		for i := range subs {
			subscribedGroupIDs[subs[i].GroupID] = true
		}
	}

	accounts, err := s.accountRepo.ListAvailableSupplierAccounts(ctx)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return []*supplierMarketplaceAggregate{}, nil
	}

	accountIDs := make([]int64, 0, len(accounts))
	for i := range accounts {
		accountIDs = append(accountIDs, accounts[i].ID)
	}
	revisionsByAccount, err := s.supplierRepo.ListApprovedPricingRevisionsForAccounts(ctx, accountIDs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	bySupplier := make(map[int64]*supplierMarketplaceAggregate)
	for i := range accounts {
		account := accounts[i]
		if account.SupplierID == nil || account.Supplier == nil {
			continue
		}

		current, scheduled := selectSupplierPricingRevisions(revisionsByAccount[account.ID], now)
		if current == nil {
			continue
		}
		models := supplierSupportedModels(account, current, scheduled)
		if len(models) == 0 {
			continue
		}

		aggregate, ok := bySupplier[*account.SupplierID]
		if !ok {
			supplierCopy := *account.Supplier
			aggregate = &supplierMarketplaceAggregate{
				supplier: &supplierCopy,
				groups:   make(map[int64]*supplierMarketplaceGroupAggregate),
			}
			bySupplier[*account.SupplierID] = aggregate
		}

		for _, groupID := range supplierAccountMarketplaceGroupIDs(account, marketGroups) {
			group := marketGroups[groupID]
			if group == nil {
				continue
			}
			groupAggregate, ok := aggregate.groups[groupID]
			if !ok {
				groupAggregate = &supplierMarketplaceGroupAggregate{
					GroupID:          group.ID,
					Name:             group.Name,
					Platform:         group.Platform,
					SubscriptionType: group.SubscriptionType,
					IsSubscribed:     subscribedGroupIDs[group.ID],
					modelsByName:     make(map[string]SupportedModel),
				}
				aggregate.groups[groupID] = groupAggregate
			}
			for _, model := range models {
				groupAggregate.addModel(model)
			}
		}
	}

	out := make([]*supplierMarketplaceAggregate, 0, len(bySupplier))
	for _, aggregate := range bySupplier {
		if len(aggregate.groups) == 0 {
			continue
		}
		out = append(out, aggregate)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].supplier.CompanyName) < strings.ToLower(out[j].supplier.CompanyName)
	})
	return out, nil
}

func supplierAccountMarketplaceGroupIDs(account Account, marketGroups map[int64]*Group) []int64 {
	ids := make([]int64, 0, len(account.GroupIDs)+len(account.AccountGroups))
	ids = append(ids, account.GroupIDs...)
	for _, group := range account.AccountGroups {
		ids = append(ids, group.GroupID)
	}

	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}

		group := marketGroups[id]
		if group == nil || group.SupplierProfileID == nil || account.SupplierID == nil {
			continue
		}
		if *group.SupplierProfileID != *account.SupplierID {
			continue
		}
		if account.Platform != "" && group.Platform != account.Platform {
			continue
		}
		out = append(out, id)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
