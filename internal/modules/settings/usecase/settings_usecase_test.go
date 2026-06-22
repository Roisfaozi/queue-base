package usecase

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/settings/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/settings/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
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
			"tenant:t-1:reset_time": {ID: "1", Value: "00:00"},
			"branch:b-1:reset_time": {ID: "2", Value: "04:00"},
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
}
