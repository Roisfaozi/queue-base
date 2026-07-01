package usecase

import (
	"context"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/settings/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type stubSettingsRepo struct {
	settings map[string]*entity.Setting
}

func (s *stubSettingsRepo) Create(ctx context.Context, setting *entity.Setting) error {
	return nil
}
func (s *stubSettingsRepo) FindByID(ctx context.Context, tenantID, settingID string) (*entity.Setting, error) {
	return nil, nil
}
func (s *stubSettingsRepo) FindByScope(ctx context.Context, tenantID, scopeType, scopeID, key string) (*entity.Setting, error) {
	if val, ok := s.settings[scopeType+":"+scopeID+":"+key]; ok {
		return val, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *stubSettingsRepo) FindAllByKey(ctx context.Context, tenantID, key string) ([]*entity.Setting, error) {
	return nil, nil
}
func (s *stubSettingsRepo) Update(ctx context.Context, setting *entity.Setting) error    { return nil }
func (s *stubSettingsRepo) Delete(ctx context.Context, tenantID, settingID string) error { return nil }

// =============================================================================
// TestResolveSetting
// =============================================================================

func TestResolveSetting(t *testing.T) {
	tests := []struct {
		name     string
		category string
		repo     *stubSettingsRepo
		req      *model.ResolveSettingRequest
		tenantID string
		wantErr  bool
		wantVal  string
	}{
		// --- Inheritance chain ---
		{
			name:     "Positive_FallbackToTenantWhenBranchNotProvided",
			category: "positive",
			repo: &stubSettingsRepo{settings: map[string]*entity.Setting{
				"tenant:t-1:reset_time": {ID: "1", Value: "00:00"},
			}},
			req:      &model.ResolveSettingRequest{Key: "reset_time"},
			tenantID: "t-1",
			wantVal:  "00:00",
		},
		{
			name:     "Positive_ResolvesBranchOverrideWhenBranchProvided",
			category: "positive",
			repo: &stubSettingsRepo{settings: map[string]*entity.Setting{
				"tenant:t-1:reset_time": {ID: "1", Value: "00:00"},
				"branch:b-1:reset_time": {ID: "2", Value: "04:00"},
			}},
			req:      &model.ResolveSettingRequest{Key: "reset_time", BranchID: "b-1"},
			tenantID: "t-1",
			wantVal:  "04:00",
		},
		{
			name:     "Positive_ResolvesServiceOverrideBeforeBranch",
			category: "positive",
			repo: &stubSettingsRepo{settings: map[string]*entity.Setting{
				"tenant:t-1:reset_time":    {ID: "1", Value: "00:00"},
				"branch:b-1:reset_time":    {ID: "2", Value: "04:00"},
				"service:svc-1:reset_time": {ID: "3", Value: "05:00"},
			}},
			req:      &model.ResolveSettingRequest{Key: "reset_time", BranchID: "b-1", ServiceID: "svc-1"},
			tenantID: "t-1",
			wantVal:  "05:00",
		},
		{
			name:     "Positive_ResolvesCounterOverrideBeforeService",
			category: "positive",
			repo: &stubSettingsRepo{settings: map[string]*entity.Setting{
				"tenant:t-1:reset_time":        {ID: "1", Value: "00:00"},
				"branch:b-1:reset_time":        {ID: "2", Value: "04:00"},
				"service:svc-1:reset_time":     {ID: "3", Value: "05:00"},
				"counter:counter-1:reset_time": {ID: "4", Value: "06:00"},
			}},
			req:      &model.ResolveSettingRequest{Key: "reset_time", BranchID: "b-1", ServiceID: "svc-1", CounterID: "counter-1"},
			tenantID: "t-1",
			wantVal:  "06:00",
		},
		// --- Workflow settings ---
		{
			name:     "Positive_BranchOverridesTenantForPharmacyFlow",
			category: "positive",
			repo: &stubSettingsRepo{settings: map[string]*entity.Setting{
				"tenant:t-1:pharmacy_flow_enabled": {ID: "1", Value: "false"},
				"branch:b-1:pharmacy_flow_enabled": {ID: "2", Value: "true"},
			}},
			req:      &model.ResolveSettingRequest{Key: model.SettingKeyPharmacyFlowEnabled, BranchID: "b-1"},
			tenantID: "t-1",
			wantVal:  "true",
		},
		{
			name:     "Positive_ServiceOverridesTenantForRequireCounter",
			category: "positive",
			repo: &stubSettingsRepo{settings: map[string]*entity.Setting{
				"tenant:t-1:require_counter_for_service":    {ID: "4", Value: "false"},
				"service:svc-1:require_counter_for_service": {ID: "3", Value: "true"},
			}},
			req:      &model.ResolveSettingRequest{Key: model.SettingKeyRequireCounterForService, ServiceID: "svc-1"},
			tenantID: "t-1",
			wantVal:  "true",
		},
		{
			name:     "Positive_CounterOverridesServiceForRequireCounter",
			category: "positive",
			repo: &stubSettingsRepo{settings: map[string]*entity.Setting{
				"service:svc-1:require_counter_for_service":     {ID: "3", Value: "true"},
				"counter:counter-1:require_counter_for_service": {ID: "5", Value: "false"},
			}},
			req:      &model.ResolveSettingRequest{Key: model.SettingKeyRequireCounterForService, ServiceID: "svc-1", CounterID: "counter-1"},
			tenantID: "t-1",
			wantVal:  "false",
		},
		// --- Fallbacks ---
		{
			name:     "Edge_FallsBackFromCounterToService",
			category: "edge",
			repo: &stubSettingsRepo{settings: map[string]*entity.Setting{
				"service:svc-1:reset_time": {ID: "1", Value: "05:00"},
			}},
			req:      &model.ResolveSettingRequest{Key: "reset_time", ServiceID: "svc-1", CounterID: "no-counter-override"},
			tenantID: "t-1",
			wantVal:  "05:00",
		},
		{
			name:     "Edge_FallsBackFromServiceToBranch",
			category: "edge",
			repo: &stubSettingsRepo{settings: map[string]*entity.Setting{
				"branch:b-1:reset_time": {ID: "1", Value: "04:00"},
			}},
			req:      &model.ResolveSettingRequest{Key: "reset_time", BranchID: "b-1", ServiceID: "no-svc-override"},
			tenantID: "t-1",
			wantVal:  "04:00",
		},
		{
			name:     "Edge_PharmacyFlowFallsBackToTenant",
			category: "edge",
			repo: &stubSettingsRepo{settings: map[string]*entity.Setting{
				"tenant:t-1:pharmacy_flow_enabled": {ID: "1", Value: "false"},
			}},
			req:      &model.ResolveSettingRequest{Key: model.SettingKeyPharmacyFlowEnabled, BranchID: "b-no-override", ServiceID: "svc-no-override"},
			tenantID: "t-1",
			wantVal:  "false",
		},
		// --- Negative ---
		{
			name:     "Negative_MissingTenantReturnsError",
			category: "negative",
			repo:     &stubSettingsRepo{settings: map[string]*entity.Setting{}},
			req:      &model.ResolveSettingRequest{Key: "reset_time"},
			wantErr:  true,
		},
		{
			name:     "Negative_KeyNotFoundReturnsError",
			category: "negative",
			repo:     &stubSettingsRepo{settings: map[string]*entity.Setting{}},
			req:      &model.ResolveSettingRequest{Key: "nonexistent_key"},
			tenantID: "t-1",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewSettingsUseCase(tt.repo)
			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.ResolveSetting(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantVal, res.Value)
		})
	}
}
