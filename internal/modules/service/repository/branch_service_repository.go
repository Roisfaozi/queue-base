package repository

import (
	"context"

	"github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	txpkg "github.com/Roisfaozi/queue-base/pkg/tx"
	"gorm.io/gorm"
)

type BranchServiceRepository interface {
	Create(ctx context.Context, branchService *entity.BranchService) error
	FindByID(ctx context.Context, tenantID, branchID, id string) (*entity.BranchService, error)
	FindByService(ctx context.Context, tenantID, branchID, serviceID string) (*entity.BranchService, error)
	FindAll(ctx context.Context, tenantID, branchID string) ([]*entity.BranchService, error)
	Update(ctx context.Context, branchService *entity.BranchService) error
	Delete(ctx context.Context, tenantID, branchID, id string) error
}

type branchServiceRepository struct {
	db *gorm.DB
}

func NewBranchServiceRepository(db *gorm.DB) BranchServiceRepository {
	return &branchServiceRepository{db: db}
}

func (r *branchServiceRepository) getDB(ctx context.Context) *gorm.DB {
	if txDB, ok := txpkg.DBFromContext(ctx); ok {
		return txDB
	}
	return r.db.WithContext(ctx)
}

func (r *branchServiceRepository) Create(ctx context.Context, branchService *entity.BranchService) error {
	return r.getDB(ctx).Create(branchService).Error
}

func (r *branchServiceRepository) FindByID(ctx context.Context, tenantID, branchID, id string) (*entity.BranchService, error) {
	var bs entity.BranchService
	if err := r.getDB(ctx).Where("tenant_id = ? AND branch_id = ? AND id = ?", tenantID, branchID, id).First(&bs).Error; err != nil {
		return nil, err
	}
	return &bs, nil
}

func (r *branchServiceRepository) FindByService(ctx context.Context, tenantID, branchID, serviceID string) (*entity.BranchService, error) {
	var bs entity.BranchService
	if err := r.getDB(ctx).Where("tenant_id = ? AND branch_id = ? AND service_id = ?", tenantID, branchID, serviceID).First(&bs).Error; err != nil {
		return nil, err
	}
	return &bs, nil
}

func (r *branchServiceRepository) FindAll(ctx context.Context, tenantID, branchID string) ([]*entity.BranchService, error) {
	var bss []*entity.BranchService
	if err := r.getDB(ctx).Where("tenant_id = ? AND branch_id = ?", tenantID, branchID).Order("sort_order asc").Find(&bss).Error; err != nil {
		return nil, err
	}
	return bss, nil
}

func (r *branchServiceRepository) Update(ctx context.Context, branchService *entity.BranchService) error {
	res := r.getDB(ctx).
		Model(&entity.BranchService{}).
		Where("tenant_id = ? AND branch_id = ? AND id = ?", branchService.TenantID, branchService.BranchID, branchService.ID).
		Select("CustomName", "IsActive", "SortOrder", "UpdatedAt").
		Updates(branchService)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}

func (r *branchServiceRepository) Delete(ctx context.Context, tenantID, branchID, id string) error {
	res := r.getDB(ctx).Where("tenant_id = ? AND branch_id = ? AND id = ?", tenantID, branchID, id).Delete(&entity.BranchService{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}
