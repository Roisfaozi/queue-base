package repository

import (
	"context"

	"github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	txpkg "github.com/Roisfaozi/queue-base/pkg/tx"
	"gorm.io/gorm"
)

type CounterRepository interface {
	Create(ctx context.Context, counter *entity.Counter) error
	FindByID(ctx context.Context, tenantID, counterID string) (*entity.Counter, error)
	FindAll(ctx context.Context, tenantID string) ([]*entity.Counter, error)
	Update(ctx context.Context, counter *entity.Counter) error
	Delete(ctx context.Context, tenantID, counterID string) error
}

type counterRepository struct {
	db *gorm.DB
}

func NewCounterRepository(db *gorm.DB) CounterRepository {
	return &counterRepository{db: db}
}

func (r *counterRepository) getDB(ctx context.Context) *gorm.DB {
	if txDB, ok := txpkg.DBFromContext(ctx); ok {
		return txDB
	}
	return r.db.WithContext(ctx)
}

func (r *counterRepository) Create(ctx context.Context, counter *entity.Counter) error {
	return r.getDB(ctx).Create(counter).Error
}

func (r *counterRepository) FindByID(ctx context.Context, tenantID, counterID string) (*entity.Counter, error) {
	var counter entity.Counter
	if err := r.getDB(ctx).Where("tenant_id = ? AND id = ?", tenantID, counterID).First(&counter).Error; err != nil {
		return nil, err
	}
	return &counter, nil
}

func (r *counterRepository) FindAll(ctx context.Context, tenantID string) ([]*entity.Counter, error) {
	var counters []*entity.Counter
	if err := r.getDB(ctx).Where("tenant_id = ?", tenantID).Find(&counters).Error; err != nil {
		return nil, err
	}
	return counters, nil
}

func (r *counterRepository) Update(ctx context.Context, counter *entity.Counter) error {
	res := r.getDB(ctx).
		Model(&entity.Counter{}).
		Where("tenant_id = ? AND id = ?", counter.TenantID, counter.ID).
		Updates(map[string]interface{}{
			"code":       counter.Code,
			"name":       counter.Name,
			"status":     counter.Status,
			"settings":   counter.Settings,
			"updated_at": counter.UpdatedAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}

func (r *counterRepository) Delete(ctx context.Context, tenantID, counterID string) error {
	res := r.getDB(ctx).Where("tenant_id = ? AND id = ?", tenantID, counterID).Delete(&entity.Counter{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}
