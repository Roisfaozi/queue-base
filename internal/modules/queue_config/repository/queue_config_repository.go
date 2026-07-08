package repository

import (
	"context"
	"fmt"

	entity "github.com/Roisfaozi/queue-base/internal/modules/queue_config/entity"
	"gorm.io/gorm"
)

type QueueConfigRepository interface {
	UpsertTenantQueueSetting(ctx context.Context, tenantID string, values map[string]any) error
	UpsertBranchQueueSetting(ctx context.Context, tenantID, branchID string, values map[string]any) error
	UpsertBranchServiceQueueSetting(ctx context.Context, tenantID, branchID, branchServiceID string, values map[string]any) error
	UpsertCounterQueueSetting(ctx context.Context, tenantID, counterID string, values map[string]any) error
}

type queueConfigRepository struct {
	db *gorm.DB
}

func NewQueueConfigRepository(db *gorm.DB) QueueConfigRepository {
	return &queueConfigRepository{db: db}
}

func (r *queueConfigRepository) UpsertTenantQueueSetting(ctx context.Context, tenantID string, values map[string]any) error {
	row := &entity.TenantQueueSetting{TenantID: tenantID}
	return r.upsert(ctx, row.TableName(), "tenant_id", tenantID, values)
}

func (r *queueConfigRepository) UpsertBranchQueueSetting(ctx context.Context, tenantID, branchID string, values map[string]any) error {
	row := &entity.BranchQueueSetting{TenantID: tenantID, BranchID: branchID}
	return r.upsert(ctx, row.TableName(), "tenant_id", tenantID, values, "branch_id", branchID)
}

func (r *queueConfigRepository) UpsertBranchServiceQueueSetting(ctx context.Context, tenantID, branchID, branchServiceID string, values map[string]any) error {
	row := &entity.BranchServiceQueueSetting{TenantID: tenantID, BranchID: branchID, BranchServiceID: branchServiceID}
	return r.upsert(ctx, row.TableName(), "tenant_id", tenantID, values, "branch_id", branchID, "branch_service_id", branchServiceID)
}

func (r *queueConfigRepository) UpsertCounterQueueSetting(ctx context.Context, tenantID, counterID string, values map[string]any) error {
	row := &entity.CounterQueueSetting{TenantID: tenantID, CounterID: counterID}
	return r.upsert(ctx, row.TableName(), "tenant_id", tenantID, values, "counter_id", counterID)
}

func (r *queueConfigRepository) upsert(ctx context.Context, table, firstKey string, firstValue any, values map[string]any, extraWhere ...any) error {
	if r.db == nil {
		return gorm.ErrInvalidDB
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Table(table).Where(fmt.Sprintf("%s = ?", firstKey), firstValue)
		for i := 0; i < len(extraWhere); i += 2 {
			query = query.Where(fmt.Sprintf("%s = ?", extraWhere[i]), extraWhere[i+1])
		}
		if err := query.Take(map[string]any{}).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				payload := make(map[string]any, len(values)+len(extraWhere)/2+1)
				payload[firstKey] = firstValue
				for i := 0; i < len(extraWhere); i += 2 {
					payload[extraWhere[i].(string)] = extraWhere[i+1]
				}
				for key, value := range values {
					payload[key] = value
				}
				return tx.Table(table).Create(payload).Error
			}
			return err
		}
		return query.Updates(values).Error
	})
}
