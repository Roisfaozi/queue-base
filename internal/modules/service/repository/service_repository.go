package repository

import (
	"context"

	"github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	txpkg "github.com/Roisfaozi/queue-base/pkg/tx"
	"gorm.io/gorm"
)

type ServiceRepository interface {
	Create(ctx context.Context, service *entity.Service) error
	FindByID(ctx context.Context, tenantID, serviceID string) (*entity.Service, error)
	FindAll(ctx context.Context, tenantID string) ([]*entity.Service, error)
	Update(ctx context.Context, service *entity.Service) error
	Delete(ctx context.Context, tenantID, serviceID string) error
}

type serviceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) ServiceRepository {
	return &serviceRepository{db: db}
}

func (r *serviceRepository) getDB(ctx context.Context) *gorm.DB {
	if txDB, ok := txpkg.DBFromContext(ctx); ok {
		return txDB
	}
	return r.db.WithContext(ctx)
}

func (r *serviceRepository) Create(ctx context.Context, service *entity.Service) error {
	return r.getDB(ctx).Create(service).Error
}

func (r *serviceRepository) FindByID(ctx context.Context, tenantID, serviceID string) (*entity.Service, error) {
	var service entity.Service
	if err := r.getDB(ctx).Where("tenant_id = ? AND id = ?", tenantID, serviceID).First(&service).Error; err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *serviceRepository) FindAll(ctx context.Context, tenantID string) ([]*entity.Service, error) {
	var services []*entity.Service
	if err := r.getDB(ctx).Where("tenant_id = ?", tenantID).Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

func (r *serviceRepository) Update(ctx context.Context, service *entity.Service) error {
	res := r.getDB(ctx).
		Model(&entity.Service{}).
		Where("tenant_id = ? AND id = ?", service.TenantID, service.ID).
		Select("Code", "Name", "Type", "DefaultEstimatedDuration", "Status", "IsPharmacy", "IsPharmacyReception", "Settings", "UpdatedAt").
		Updates(service)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}

func (r *serviceRepository) Delete(ctx context.Context, tenantID, serviceID string) error {
	res := r.getDB(ctx).Where("tenant_id = ? AND id = ?", tenantID, serviceID).Delete(&entity.Service{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}
