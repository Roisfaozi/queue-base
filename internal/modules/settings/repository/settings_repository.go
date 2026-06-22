package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/settings/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	txpkg "github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"gorm.io/gorm"
)

type SettingsRepository interface {
	Create(ctx context.Context, setting *entity.Setting) error
	FindByID(ctx context.Context, tenantID, settingID string) (*entity.Setting, error)
	FindByScope(ctx context.Context, tenantID, scopeType, scopeID, key string) (*entity.Setting, error)
	FindAllByKey(ctx context.Context, tenantID, key string) ([]*entity.Setting, error)
	Update(ctx context.Context, setting *entity.Setting) error
	Delete(ctx context.Context, tenantID, settingID string) error
}

type settingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) SettingsRepository {
	return &settingsRepository{db: db}
}

func (r *settingsRepository) getDB(ctx context.Context) *gorm.DB {
	if txDB, ok := txpkg.DBFromContext(ctx); ok {
		return txDB
	}
	return r.db.WithContext(ctx)
}

func (r *settingsRepository) Create(ctx context.Context, setting *entity.Setting) error {
	return r.getDB(ctx).Create(setting).Error
}

func (r *settingsRepository) FindByID(ctx context.Context, tenantID, settingID string) (*entity.Setting, error) {
	var setting entity.Setting
	if err := r.getDB(ctx).Where("tenant_id = ? AND id = ?", tenantID, settingID).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *settingsRepository) FindByScope(ctx context.Context, tenantID, scopeType, scopeID, key string) (*entity.Setting, error) {
	var setting entity.Setting
	if err := r.getDB(ctx).Where("tenant_id = ? AND scope_type = ? AND scope_id = ? AND key = ? AND is_active = ?", tenantID, scopeType, scopeID, key, true).First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // return nil safely to allow inheritance chain to proceed
		}
		return nil, err
	}
	return &setting, nil
}

func (r *settingsRepository) FindAllByKey(ctx context.Context, tenantID, key string) ([]*entity.Setting, error) {
	var settings []*entity.Setting
	if err := r.getDB(ctx).Where("tenant_id = ? AND key = ? AND is_active = ?", tenantID, key, true).Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *settingsRepository) Update(ctx context.Context, setting *entity.Setting) error {
	res := r.getDB(ctx).
		Model(&entity.Setting{}).
		Where("tenant_id = ? AND id = ?", setting.TenantID, setting.ID).
		Updates(map[string]interface{}{
			"value":      setting.Value,
			"is_active":  setting.IsActive,
			"updated_at": setting.UpdatedAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}

func (r *settingsRepository) Delete(ctx context.Context, tenantID, settingID string) error {
	res := r.getDB(ctx).Where("tenant_id = ? AND id = ?", tenantID, settingID).Delete(&entity.Setting{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}
