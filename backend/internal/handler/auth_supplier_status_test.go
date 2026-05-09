//go:build unit

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type authSupplierRepoStub struct {
	service.SupplierRepository
	profile *service.SupplierProfile
}

func (s *authSupplierRepoStub) GetProfileByUserID(_ context.Context, userID int64) (*service.SupplierProfile, error) {
	if s.profile == nil || s.profile.UserID != userID {
		return nil, service.ErrSupplierProfileNotFound
	}
	cloned := *s.profile
	return &cloned, nil
}

func TestAuthHandlerRespondWithTokenPairIncludesSupplierStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		status            string
		hasSupplierAccess bool
	}{
		{
			name:              "pending supplier",
			status:            service.SupplierStatusPending,
			hasSupplierAccess: false,
		},
		{
			name:              "approved supplier",
			status:            service.SupplierStatusApproved,
			hasSupplierAccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &service.User{
				ID:        88,
				Email:     "supplier@example.com",
				Username:  "pending-supplier",
				Role:      service.RoleUser,
				Status:    service.StatusActive,
				CreatedAt: time.Date(2026, 5, 4, 8, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 5, 4, 8, 0, 0, 0, time.UTC),
			}
			authService := service.NewAuthService(
				nil,
				nil,
				nil,
				&userHandlerRefreshTokenCacheStub{},
				&config.Config{
					JWT: config.JWTConfig{
						Secret:                 "test-secret",
						ExpireHour:             1,
						RefreshTokenExpireDays: 30,
					},
				},
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)
			supplierService := service.NewSupplierService(
				&authSupplierRepoStub{
					profile: &service.SupplierProfile{
						ID:          7,
						UserID:      user.ID,
						CompanyName: "Supplier Co",
						Status:      tt.status,
					},
				},
				nil,
				nil,
				nil,
				nil,
				nil,
			)
			handler := &AuthHandler{
				authService:     authService,
				supplierService: supplierService,
			}

			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)

			handler.respondWithTokenPair(c, user)

			require.Equal(t, http.StatusOK, recorder.Code)

			var resp struct {
				Code int          `json:"code"`
				Data AuthResponse `json:"data"`
			}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
			require.Equal(t, 0, resp.Code)
			require.NotNil(t, resp.Data.User)
			require.Equal(t, tt.status, resp.Data.User.SupplierStatus)
			require.Equal(t, tt.hasSupplierAccess, resp.Data.User.HasSupplierAccess)
		})
	}
}

func TestEnrichSupplierAccessClearsStaleStatusWhenProfileMissing(t *testing.T) {
	user := &service.User{
		ID:                99,
		HasSupplierAccess: true,
		SupplierStatus:    service.SupplierStatusApproved,
	}
	supplierService := service.NewSupplierService(&authSupplierRepoStub{}, nil, nil, nil, nil, nil)

	err := enrichSupplierAccess(context.Background(), supplierService, user)

	require.NoError(t, err)
	require.False(t, user.HasSupplierAccess)
	require.Empty(t, user.SupplierStatus)
}

func TestApplySupplierSignupProfileRequiresCompanyName(t *testing.T) {
	handler := &AuthHandler{
		authService:     &service.AuthService{},
		supplierService: service.NewSupplierService(nil, nil, nil, nil, nil, nil),
	}

	err := handler.applySupplierSignupProfile(
		context.Background(),
		&service.User{ID: 101},
		"supplier",
		service.SupplierProfileInput{},
		"",
	)

	require.Error(t, err)
	require.True(t, infraerrors.IsBadRequest(err))
	require.Equal(t, "SUPPLIER_COMPANY_REQUIRED", infraerrors.Reason(err))
}
