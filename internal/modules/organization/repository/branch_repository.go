package repository

import (
	"context"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	txpkg "github.com/Roisfaozi/queue-base/pkg/tx"
	"gorm.io/gorm"
)

type BranchRepository interface {
	Create(ctx context.Context, branch *entity.Branch) error
	FindByID(ctx context.Context, tenantID, branchID string) (*entity.Branch, error)
	FindAll(ctx context.Context, tenantID string) ([]*entity.Branch, error)
	Update(ctx context.Context, branch *entity.Branch) error
	Delete(ctx context.Context, tenantID, branchID string) error
}

type branchRepository struct {
	db *gorm.DB
}

func NewBranchRepository(db *gorm.DB) BranchRepository {
	return &branchRepository{db: db}
}

func (r *branchRepository) getDB(ctx context.Context) *gorm.DB {
	if txDB, ok := txpkg.DBFromContext(ctx); ok {
		return txDB
	}
	return r.db.WithContext(ctx)
}

func (r *branchRepository) Create(ctx context.Context, branch *entity.Branch) error {
	return r.getDB(ctx).Create(branch).Error
}

func (r *branchRepository) FindByID(ctx context.Context, tenantID, branchID string) (*entity.Branch, error) {
	var branch entity.Branch
	if err := r.getDB(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, branchID).
		First(&branch).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

func (r *branchRepository) FindAll(ctx context.Context, tenantID string) ([]*entity.Branch, error) {
	var branches []*entity.Branch
	if err := r.getDB(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&branches).Error; err != nil {
		return nil, err
	}
	return branches, nil
}

func (r *branchRepository) Update(ctx context.Context, branch *entity.Branch) error {
	res := r.getDB(ctx).
		Model(&entity.Branch{}).
		Where("tenant_id = ? AND id = ?", branch.TenantID, branch.ID).
		Updates(map[string]interface{}{
			"code":       branch.Code,
			"name":       branch.Name,
			"status":     branch.Status,
			"settings":   branch.Settings,
			"updated_at": branch.UpdatedAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}

func (r *branchRepository) Delete(ctx context.Context, tenantID, branchID string) error {
	res := r.getDB(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, branchID).
		Delete(&entity.Branch{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return exception.ErrNotFound
	}
	return nil
}
