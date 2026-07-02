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
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Code:        req.Code,
		Name:        req.Name,
		Address:     req.Address,
		City:        req.City,
		Province:    req.Province,
		Phone:       req.Phone,
		RunningText: req.RunningText,
		Timezone:    req.Timezone,
		Status:      entity.BranchStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
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
	if req.Address != nil {
		branch.Address = *req.Address
	}
	if req.City != nil {
		branch.City = *req.City
	}
	if req.Province != nil {
		branch.Province = *req.Province
	}
	if req.PostalCode != nil {
		branch.PostalCode = *req.PostalCode
	}
	if req.Phone != nil {
		branch.Phone = *req.Phone
	}
	if req.Email != nil {
		branch.Email = *req.Email
	}
	if req.LogoAssetID != nil {
		branch.LogoAssetID = *req.LogoAssetID
	}
	if req.RunningText != nil {
		branch.RunningText = *req.RunningText
	}
	if req.Timezone != nil {
		branch.Timezone = *req.Timezone
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
		ID:          branch.ID,
		TenantID:    branch.TenantID,
		Code:        branch.Code,
		Name:        branch.Name,
		Address:     branch.Address,
		City:        branch.City,
		Province:    branch.Province,
		PostalCode:  branch.PostalCode,
		Phone:       branch.Phone,
		Email:       branch.Email,
		LogoAssetID: branch.LogoAssetID,
		RunningText: branch.RunningText,
		Timezone:    branch.Timezone,
		Status:      branch.Status,
		CreatedAt:   branch.CreatedAt,
		UpdatedAt:   branch.UpdatedAt,
	}
}
