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

func TestResolveSettingInheritance(t *testing.T) {
	repo := &stubSettingsRepo{
		settings: map[string]*entity.Setting{
			"tenant:t-1:reset_time":        {ID: "1", Value: "00:00"},
			"branch:b-1:reset_time":        {ID: "2", Value: "04:00"},
			"service:svc-1:reset_time":     {ID: "3", Value: "05:00"},
			"counter:counter-1:reset_time": {ID: "4", Value: "06:00"},
		},
	}
	uc := NewSettingsUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	t.Run("Fallbacks to Tenant when Branch not provided", func(t *testing.T) {
		res, err := uc.ResolveSetting(ctx, &model.ResolveSettingRequest{
			Key: "reset_time",
		})
		assert.NoError(t, err)
		assert.Equal(t, "00:00", res.Value)
	})

	t.Run("Resolves Branch override when Branch provided", func(t *testing.T) {
		res, err := uc.ResolveSetting(ctx, &model.ResolveSettingRequest{
			Key:      "reset_time",
			BranchID: "b-1",
		})
		assert.NoError(t, err)
		assert.Equal(t, "04:00", res.Value)
	})

	t.Run("Resolves Service override before Branch", func(t *testing.T) {
		res, err := uc.ResolveSetting(ctx, &model.ResolveSettingRequest{
			Key:       "reset_time",
			BranchID:  "b-1",
			ServiceID: "svc-1",
		})
		assert.NoError(t, err)
		assert.Equal(t, "05:00", res.Value)
	})

	t.Run("Resolves Counter override before Service", func(t *testing.T) {
		res, err := uc.ResolveSetting(ctx, &model.ResolveSettingRequest{
			Key:       "reset_time",
			BranchID:  "b-1",
			ServiceID: "svc-1",
			CounterID: "counter-1",
		})
		assert.NoError(t, err)
		assert.Equal(t, "06:00", res.Value)
	})
}

func TestResolveWorkflowSettingInheritance(t *testing.T) {
	repo := &stubSettingsRepo{
		settings: map[string]*entity.Setting{
			"tenant:t-1:pharmacy_flow_enabled":              {ID: "1", Value: "false"},
			"branch:b-1:pharmacy_flow_enabled":              {ID: "2", Value: "true"},
			"service:svc-1:require_counter_for_service":     {ID: "3", Value: "true"},
			"tenant:t-1:require_counter_for_service":        {ID: "4", Value: "false"},
			"counter:counter-1:require_counter_for_service": {ID: "5", Value: "false"},
		},
	}
	uc := NewSettingsUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	t.Run("Branch overrides tenant for pharmacy flow", func(t *testing.T) {
		res, err := uc.ResolveSetting(ctx, &model.ResolveSettingRequest{
			Key:      model.SettingKeyPharmacyFlowEnabled,
			BranchID: "b-1",
		})
		assert.NoError(t, err)
		assert.Equal(t, "true", res.Value)
	})

	t.Run("Service overrides tenant for require counter", func(t *testing.T) {
		res, err := uc.ResolveSetting(ctx, &model.ResolveSettingRequest{
			Key:       model.SettingKeyRequireCounterForService,
			ServiceID: "svc-1",
		})
		assert.NoError(t, err)
		assert.Equal(t, "true", res.Value)
	})

	t.Run("Counter overrides service for require counter", func(t *testing.T) {
		res, err := uc.ResolveSetting(ctx, &model.ResolveSettingRequest{
			Key:       model.SettingKeyRequireCounterForService,
			ServiceID: "svc-1",
			CounterID: "counter-1",
		})
		assert.NoError(t, err)
		assert.Equal(t, "false", res.Value)
	})
}

func TestResolveSettingMissingTenantReturnsError(t *testing.T) {
	repo := &stubSettingsRepo{settings: map[string]*entity.Setting{}}
	uc := NewSettingsUseCase(repo)

	_, err := uc.ResolveSetting(context.Background(), &model.ResolveSettingRequest{Key: "reset_time"})
	assert.Error(t, err)
}

func TestResolveSettingKeyNotFoundReturnsError(t *testing.T) {
	repo := &stubSettingsRepo{settings: map[string]*entity.Setting{}}
	uc := NewSettingsUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	_, err := uc.ResolveSetting(ctx, &model.ResolveSettingRequest{Key: "nonexistent_key"})
	assert.Error(t, err)
}

func TestResolveSettingFallsBackFromCounterToService(t *testing.T) {
	repo := &stubSettingsRepo{
		settings: map[string]*entity.Setting{
			"service:svc-1:reset_time": {ID: "1", Value: "05:00"},
		},
	}
	uc := NewSettingsUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.ResolveSetting(ctx, &model.ResolveSettingRequest{
		Key:       "reset_time",
		ServiceID: "svc-1",
		CounterID: "no-counter-override",
	})
	assert.NoError(t, err)
	assert.Equal(t, "05:00", res.Value)
}

func TestResolveSettingFallsBackFromServiceToBranch(t *testing.T) {
	repo := &stubSettingsRepo{
		settings: map[string]*entity.Setting{
			"branch:b-1:reset_time": {ID: "1", Value: "04:00"},
		},
	}
	uc := NewSettingsUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.ResolveSetting(ctx, &model.ResolveSettingRequest{
		Key:       "reset_time",
		BranchID:  "b-1",
		ServiceID: "no-svc-override",
	})
	assert.NoError(t, err)
	assert.Equal(t, "04:00", res.Value)
}

func TestResolveSettingPharmacyFlowFallsBackToTenant(t *testing.T) {
	repo := &stubSettingsRepo{
		settings: map[string]*entity.Setting{
			"tenant:t-1:pharmacy_flow_enabled": {ID: "1", Value: "false"},
		},
	}
	uc := NewSettingsUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.ResolveSetting(ctx, &model.ResolveSettingRequest{
		Key:       model.SettingKeyPharmacyFlowEnabled,
		BranchID:  "b-no-override",
		ServiceID: "svc-no-override",
	})
	assert.NoError(t, err)
	assert.Equal(t, "false", res.Value)
}
