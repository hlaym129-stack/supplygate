//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type supplierProfileRepoStub struct {
	SupplierRepository
	existing *SupplierProfile
	getErr   error
	saved    *SupplierProfile
}

func (r *supplierProfileRepoStub) GetProfileByUserID(ctx context.Context, userID int64) (*SupplierProfile, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if r.existing == nil || r.existing.UserID != userID {
		return nil, ErrSupplierProfileNotFound
	}
	cp := *r.existing
	return &cp, nil
}

func (r *supplierProfileRepoStub) UpsertProfile(ctx context.Context, profile *SupplierProfile) error {
	cp := *profile
	if cp.ID == 0 {
		cp.ID = 10
	}
	r.saved = &cp
	*profile = cp
	return nil
}

func TestApplyProfileCreatesPendingProfile(t *testing.T) {
	repo := &supplierProfileRepoStub{}
	svc := NewSupplierService(repo, nil, nil, nil, nil, nil)

	profile, err := svc.ApplyProfile(context.Background(), 42, SupplierProfileInput{
		CompanyName:  " Acme ",
		ContactName:  " Alice ",
		ContactEmail: " alice@example.com ",
		ContactPhone: " 123 ",
		Notes:        " public intro ",
	})

	require.NoError(t, err)
	require.Equal(t, SupplierStatusPending, profile.Status)
	require.Equal(t, SupplierStatusPending, repo.saved.Status)
	require.Equal(t, "Acme", profile.CompanyName)
	require.Equal(t, "Alice", profile.ContactName)
	require.Equal(t, "alice@example.com", profile.ContactEmail)
	require.Equal(t, "123", profile.ContactPhone)
	require.Equal(t, "public intro", profile.Notes)
}

func TestApplyProfileKeepsApprovedSupplierAccess(t *testing.T) {
	repo := &supplierProfileRepoStub{
		existing: &SupplierProfile{
			ID:                       7,
			UserID:                   42,
			Status:                   SupplierStatusApproved,
			AccountSubmissionEnabled: true,
		},
	}
	svc := NewSupplierService(repo, nil, nil, nil, nil, nil)

	profile, err := svc.ApplyProfile(context.Background(), 42, SupplierProfileInput{
		CompanyName: "Acme Updated",
	})

	require.NoError(t, err)
	require.Equal(t, SupplierStatusPending, profile.Status)
	require.True(t, profile.AccountSubmissionEnabled)
	require.Equal(t, SupplierStatusPending, repo.saved.Status)
	require.True(t, repo.saved.AccountSubmissionEnabled)
}

func TestApplyProfileResubmitsRejectedProfileForReview(t *testing.T) {
	repo := &supplierProfileRepoStub{
		existing: &SupplierProfile{
			ID:     7,
			UserID: 42,
			Status: SupplierStatusRejected,
		},
	}
	svc := NewSupplierService(repo, nil, nil, nil, nil, nil)

	profile, err := svc.ApplyProfile(context.Background(), 42, SupplierProfileInput{
		CompanyName: "Acme Resubmitted",
	})

	require.NoError(t, err)
	require.Equal(t, SupplierStatusPending, profile.Status)
	require.Equal(t, SupplierStatusPending, repo.saved.Status)
}

func TestApplyProfilePropagatesProfileLookupError(t *testing.T) {
	lookupErr := errors.New("database unavailable")
	repo := &supplierProfileRepoStub{getErr: lookupErr}
	svc := NewSupplierService(repo, nil, nil, nil, nil, nil)

	profile, err := svc.ApplyProfile(context.Background(), 42, SupplierProfileInput{
		CompanyName: "Acme",
	})

	require.Nil(t, profile)
	require.ErrorIs(t, err, lookupErr)
}
