package usecase

import (
	"context"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/google/uuid"
)

type BranchUseCase interface {
	CreateBranch(ctx context.Context, req *model.CreateBranchRequest) (*model.BranchResponse, error)
	ResolveBranch(ctx context.Context, branchID string) (*model.BranchResponse, error)
	ListBranches(ctx context.Context) ([]model.BranchResponse, error)
	UpdateBranch(ctx context.Context, branchID string, req *model.UpdateBranchRequest) (*model.BranchResponse, error)
	DeleteBranch(ctx context.Context, branchID string) error
}

type branchUseCase struct {
	repo repository.BranchRepository
}

func NewBranchUseCase(repo repository.BranchRepository) BranchUseCase {
	return &branchUseCase{repo: repo}
}

func (u *branchUseCase) CreateBranch(ctx context.Context, req *model.CreateBranchRequest) (*model.BranchResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return nil, exception.ErrBadRequest
	}
	req.Sanitize()
	now := time.Now().UnixMilli()
	branch := &entity.Branch{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Code:      req.Code,
		Name:      req.Name,
		Status:    entity.BranchStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := u.repo.Create(ctx, branch); err != nil {
		return nil, err
	}
	return u.mapToResponse(branch), nil
}

func (u *branchUseCase) ResolveBranch(ctx context.Context, branchID string) (*model.BranchResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || branchID == "" {
		return nil, exception.ErrBadRequest
	}

	branch, err := u.repo.FindByID(ctx, tenantID, branchID)
	if err != nil {
		return nil, exception.ErrNotFound
	}

	return u.mapToResponse(branch), nil
}

func (u *branchUseCase) ListBranches(ctx context.Context) ([]model.BranchResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return nil, exception.ErrBadRequest
	}
	branches, err := u.repo.FindAll(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	res := make([]model.BranchResponse, len(branches))
	for i, branch := range branches {
		res[i] = *u.mapToResponse(branch)
	}
	return res, nil
}

func (u *branchUseCase) UpdateBranch(ctx context.Context, branchID string, req *model.UpdateBranchRequest) (*model.BranchResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || branchID == "" {
		return nil, exception.ErrBadRequest
	}
	req.Sanitize()
	branch, err := u.repo.FindByID(ctx, tenantID, branchID)
	if err != nil {
		return nil, exception.ErrNotFound
	}
	if req.Code != nil {
		branch.Code = *req.Code
	}
	if req.Name != nil {
		branch.Name = *req.Name
	}
	if req.Status != nil {
		branch.Status = *req.Status
	}
	branch.UpdatedAt = time.Now().UnixMilli()
	if err := u.repo.Update(ctx, branch); err != nil {
		return nil, err
	}
	return u.mapToResponse(branch), nil
}

func (u *branchUseCase) DeleteBranch(ctx context.Context, branchID string) error {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || branchID == "" {
		return exception.ErrBadRequest
	}
	return u.repo.Delete(ctx, tenantID, branchID)
}

func (u *branchUseCase) mapToResponse(branch *entity.Branch) *model.BranchResponse {
	return &model.BranchResponse{
		ID:        branch.ID,
		TenantID:  branch.TenantID,
		Code:      branch.Code,
		Name:      branch.Name,
		Status:    branch.Status,
		CreatedAt: branch.CreatedAt,
		UpdatedAt: branch.UpdatedAt,
	}
}
