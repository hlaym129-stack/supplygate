//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type supplierApprovalAccountRepo struct {
	SupplierAccountRepository
	account        *Account
	updated        *Account
	boundAccountID int64
	boundGroupIDs  []int64
	bindCalls      int
}

func (r *supplierApprovalAccountRepo) GetByID(ctx context.Context, id int64) (*Account, error) {
	if r.account == nil || r.account.ID != id {
		return nil, ErrAccountNotFound
	}
	cp := *r.account
	cp.GroupIDs = append([]int64(nil), r.account.GroupIDs...)
	return &cp, nil
}

func (r *supplierApprovalAccountRepo) Update(ctx context.Context, account *Account) error {
	cp := *account
	cp.GroupIDs = append([]int64(nil), account.GroupIDs...)
	r.updated = &cp
	return nil
}

func (r *supplierApprovalAccountRepo) BindGroups(ctx context.Context, accountID int64, groupIDs []int64) error {
	r.bindCalls++
	r.boundAccountID = accountID
	r.boundGroupIDs = append([]int64(nil), groupIDs...)
	return nil
}

type supplierApprovalSupplierRepo struct {
	SupplierRepository
	pending      *SupplierAccountPricingRevision
	reviewCalls  int
	reviewStatus string
}

func (r *supplierApprovalSupplierRepo) GetPendingPricingRevision(ctx context.Context, accountID int64) (*SupplierAccountPricingRevision, error) {
	if r.pending == nil || r.pending.AccountID != accountID {
		return nil, ErrSupplierPricingNotFound
	}
	cp := *r.pending
	return &cp, nil
}

func (r *supplierApprovalSupplierRepo) ReviewPricingRevision(ctx context.Context, revisionID int64, status string, reviewerID int64, reviewNote string, effectiveAt *time.Time) (*SupplierAccountPricingRevision, error) {
	r.reviewCalls++
	r.reviewStatus = status
	cp := *r.pending
	cp.ID = revisionID
	cp.Status = status
	cp.ReviewedBy = &reviewerID
	cp.ReviewNote = reviewNote
	cp.EffectiveAt = effectiveAt
	return &cp, nil
}

type supplierApprovalGroupRepo struct {
	GroupRepository
	byID              map[int64]*Group
	byPlatform        map[string][]Group
	getByIDCalls      []int64
	listPlatformCalls []string
}

func (r *supplierApprovalGroupRepo) GetByID(ctx context.Context, id int64) (*Group, error) {
	r.getByIDCalls = append(r.getByIDCalls, id)
	group, ok := r.byID[id]
	if !ok {
		return nil, ErrGroupNotFound
	}
	cp := *group
	return &cp, nil
}

func (r *supplierApprovalGroupRepo) ListActiveByPlatform(ctx context.Context, platform string) ([]Group, error) {
	r.listPlatformCalls = append(r.listPlatformCalls, platform)
	return append([]Group(nil), r.byPlatform[platform]...), nil
}

func newSupplierApprovalService(accountRepo *supplierApprovalAccountRepo, supplierRepo *supplierApprovalSupplierRepo, groupRepo *supplierApprovalGroupRepo) *SupplierService {
	return NewSupplierService(supplierRepo, accountRepo, groupRepo, nil, nil, nil)
}

func supplierApprovalAccount(platform string) *Account {
	supplierID := int64(99)
	return &Account{
		ID:             7,
		Name:           "supplier account",
		Platform:       platform,
		Type:           AccountTypeAPIKey,
		OwnerType:      AccountOwnerTypeSupplier,
		SupplierID:     &supplierID,
		Status:         StatusActive,
		ApprovalStatus: AccountApprovalStatusPending,
	}
}

func supplierApprovalPendingRevision(accountID int64) *SupplierAccountPricingRevision {
	return &SupplierAccountPricingRevision{
		ID:           101,
		AccountID:    accountID,
		SupplierID:   99,
		RevisionKind: SupplierPricingRevisionKindInitial,
		Status:       SupplierPricingRevisionStatusPending,
	}
}

func TestApproveAccount_DefaultGroupByPlatform(t *testing.T) {
	account := supplierApprovalAccount(PlatformOpenAI)
	accountRepo := &supplierApprovalAccountRepo{account: account}
	supplierRepo := &supplierApprovalSupplierRepo{pending: supplierApprovalPendingRevision(account.ID)}
	groupRepo := &supplierApprovalGroupRepo{
		byPlatform: map[string][]Group{
			PlatformOpenAI: {
				{ID: 8, Platform: PlatformOpenAI, Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription},
				{ID: 2, Platform: PlatformOpenAI, Status: StatusActive, SubscriptionType: SubscriptionTypeStandard},
			},
		},
	}

	out, err := newSupplierApprovalService(accountRepo, supplierRepo, groupRepo).
		ApproveAccount(context.Background(), account.ID, 501, SupplierAccountApprovalInput{})

	require.NoError(t, err)
	require.Equal(t, []int64{2}, out.GroupIDs)
	require.Equal(t, []string{PlatformOpenAI}, groupRepo.listPlatformCalls)
	require.NotNil(t, accountRepo.updated)
	require.Equal(t, AccountApprovalStatusApproved, accountRepo.updated.ApprovalStatus)
	require.True(t, accountRepo.updated.Schedulable)
	require.Equal(t, []int64{2}, accountRepo.updated.GroupIDs)
	require.Equal(t, 1, accountRepo.bindCalls)
	require.Equal(t, int64(7), accountRepo.boundAccountID)
	require.Equal(t, []int64{2}, accountRepo.boundGroupIDs)
	require.Equal(t, 1, supplierRepo.reviewCalls)
	require.Equal(t, SupplierPricingRevisionStatusApproved, supplierRepo.reviewStatus)
}

func TestApproveAccount_DefaultAnthropicGroupUsesFirstSortedStandardGroup(t *testing.T) {
	account := supplierApprovalAccount(PlatformAnthropic)
	accountRepo := &supplierApprovalAccountRepo{account: account}
	supplierRepo := &supplierApprovalSupplierRepo{pending: supplierApprovalPendingRevision(account.ID)}
	groupRepo := &supplierApprovalGroupRepo{
		byPlatform: map[string][]Group{
			PlatformAnthropic: {
				{ID: 3, Name: "Claude 系列-余额", Platform: PlatformAnthropic, Status: StatusActive, SubscriptionType: SubscriptionTypeStandard, SortOrder: 0},
				{ID: 1, Name: "default", Platform: PlatformAnthropic, Status: StatusActive, SubscriptionType: SubscriptionTypeStandard, SortOrder: 1},
			},
		},
	}

	out, err := newSupplierApprovalService(accountRepo, supplierRepo, groupRepo).
		ApproveAccount(context.Background(), account.ID, 501, SupplierAccountApprovalInput{})

	require.NoError(t, err)
	require.Equal(t, []int64{3}, out.GroupIDs)
	require.Equal(t, []int64{3}, accountRepo.boundGroupIDs)
}

func TestApproveAccount_ManualGroupsOverrideDefaultGroup(t *testing.T) {
	account := supplierApprovalAccount(PlatformOpenAI)
	accountRepo := &supplierApprovalAccountRepo{account: account}
	supplierRepo := &supplierApprovalSupplierRepo{pending: supplierApprovalPendingRevision(account.ID)}
	groupRepo := &supplierApprovalGroupRepo{
		byID: map[int64]*Group{
			9: {ID: 9, Platform: PlatformAnthropic, Status: StatusActive, SubscriptionType: SubscriptionTypeStandard},
		},
		byPlatform: map[string][]Group{
			PlatformOpenAI: {
				{ID: 2, Platform: PlatformOpenAI, Status: StatusActive, SubscriptionType: SubscriptionTypeStandard},
			},
		},
	}

	out, err := newSupplierApprovalService(accountRepo, supplierRepo, groupRepo).
		ApproveAccount(context.Background(), account.ID, 501, SupplierAccountApprovalInput{GroupIDs: []int64{9}})

	require.NoError(t, err)
	require.Equal(t, []int64{9}, out.GroupIDs)
	require.Equal(t, []int64{9}, accountRepo.boundGroupIDs)
	require.Equal(t, []int64{9}, groupRepo.getByIDCalls)
	require.Empty(t, groupRepo.listPlatformCalls)
}

func TestApproveAccount_NoDefaultGroupFailsWithoutMutation(t *testing.T) {
	account := supplierApprovalAccount(PlatformOpenAI)
	accountRepo := &supplierApprovalAccountRepo{account: account}
	supplierRepo := &supplierApprovalSupplierRepo{pending: supplierApprovalPendingRevision(account.ID)}
	groupRepo := &supplierApprovalGroupRepo{
		byPlatform: map[string][]Group{
			PlatformOpenAI: {
				{ID: 8, Platform: PlatformOpenAI, Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription},
			},
		},
	}

	out, err := newSupplierApprovalService(accountRepo, supplierRepo, groupRepo).
		ApproveAccount(context.Background(), account.ID, 501, SupplierAccountApprovalInput{})

	require.Nil(t, out)
	require.ErrorIs(t, err, ErrSupplierDefaultGroupMissing)
	require.Nil(t, accountRepo.updated)
	require.Zero(t, accountRepo.bindCalls)
	require.Zero(t, supplierRepo.reviewCalls)
}

var _ SupplierAccountRepository = (*supplierApprovalAccountRepo)(nil)
var _ SupplierRepository = (*supplierApprovalSupplierRepo)(nil)
var _ GroupRepository = (*supplierApprovalGroupRepo)(nil)

func (r *supplierApprovalGroupRepo) Create(context.Context, *Group) error { return nil }
func (r *supplierApprovalGroupRepo) GetByIDLite(context.Context, int64) (*Group, error) {
	return nil, nil
}
func (r *supplierApprovalGroupRepo) Update(context.Context, *Group) error { return nil }
func (r *supplierApprovalGroupRepo) Delete(context.Context, int64) error  { return nil }
func (r *supplierApprovalGroupRepo) DeleteCascade(context.Context, int64) ([]int64, error) {
	return nil, nil
}
func (r *supplierApprovalGroupRepo) List(context.Context, pagination.PaginationParams) ([]Group, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *supplierApprovalGroupRepo) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string, *bool) ([]Group, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *supplierApprovalGroupRepo) ListActive(context.Context) ([]Group, error) {
	return nil, nil
}
func (r *supplierApprovalGroupRepo) ExistsByName(context.Context, string) (bool, error) {
	return false, nil
}
func (r *supplierApprovalGroupRepo) GetAccountCount(context.Context, int64) (int64, int64, error) {
	return 0, 0, nil
}
func (r *supplierApprovalGroupRepo) DeleteAccountGroupsByGroupID(context.Context, int64) (int64, error) {
	return 0, nil
}
func (r *supplierApprovalGroupRepo) GetAccountIDsByGroupIDs(context.Context, []int64) ([]int64, error) {
	return nil, nil
}
func (r *supplierApprovalGroupRepo) BindAccountsToGroup(context.Context, int64, []int64) error {
	return nil
}
func (r *supplierApprovalGroupRepo) UpdateSortOrders(context.Context, []GroupSortOrderUpdate) error {
	return nil
}
