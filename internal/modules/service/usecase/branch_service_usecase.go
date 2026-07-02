package usecase

import (
	"context"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	branchRepository "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/service/model"
	"github.com/Roisfaozi/queue-base/internal/modules/service/repository"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/google/uuid"
)

type BranchServiceUseCase interface {
	CreateBranchService(ctx context.Context, branchID string, req *model.CreateBranchServiceRequest) (*model.BranchServiceResponse, error)
	ListBranchServices(ctx context.Context, branchID string) ([]model.BranchServiceResponse, error)
	UpdateBranchService(ctx context.Context, branchID, id string, req *model.UpdateBranchServiceRequest) (*model.BranchServiceResponse, error)
	DeleteBranchService(ctx context.Context, branchID, id string) error
	EnsureActiveBranchService(ctx context.Context, tenantID, branchID, serviceID string) (*entity.BranchService, error)
}

type branchServiceUseCase struct {
	repo        repository.BranchServiceRepository
	serviceRepo repository.ServiceRepository
	branchRepo  branchRepository.BranchRepository
	audit       AuditLogger
}

func NewBranchServiceUseCase(repo repository.BranchServiceRepository, serviceRepo repository.ServiceRepository, branchRepo branchRepository.BranchRepository, audit ...AuditLogger) BranchServiceUseCase {
	var auditLogger AuditLogger
	if len(audit) > 0 {
		auditLogger = audit[0]
	}
	return &branchServiceUseCase{repo: repo, serviceRepo: serviceRepo, branchRepo: branchRepo, audit: auditLogger}
}

func (u *branchServiceUseCase) CreateBranchService(ctx context.Context, branchID string, req *model.CreateBranchServiceRequest) (*model.BranchServiceResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || branchID == "" || req == nil || req.ServiceID == "" {
		return nil, exception.ErrBadRequest
	}
	if err := u.validateTenantBranchService(ctx, tenantID, branchID, req.ServiceID); err != nil {
		return nil, err
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	now := time.Now().UnixMilli()
	bs := &entity.BranchService{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		BranchID:   branchID,
		ServiceID:  req.ServiceID,
		CustomName: req.CustomName,
		IsActive:   isActive,
		SortOrder:  req.SortOrder,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := u.repo.Create(ctx, bs); err != nil {
		return nil, err
	}
	u.tryAudit(ctx, "BRANCH_SERVICE_CREATE", bs.ID, map[string]any{"branch_id": bs.BranchID, "service_id": bs.ServiceID, "is_active": bs.IsActive})
	return u.mapToResponse(bs), nil
}

func (u *branchServiceUseCase) ListBranchServices(ctx context.Context, branchID string) ([]model.BranchServiceResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || branchID == "" {
		return nil, exception.ErrBadRequest
	}
	if _, err := u.branchRepo.FindByID(ctx, tenantID, branchID); err != nil {
		return nil, exception.ErrForbidden
	}
	items, err := u.repo.FindAll(ctx, tenantID, branchID)
	if err != nil {
		return nil, err
	}
	res := make([]model.BranchServiceResponse, len(items))
	for i, item := range items {
		res[i] = *u.mapToResponse(item)
	}
	return res, nil
}

func (u *branchServiceUseCase) UpdateBranchService(ctx context.Context, branchID, id string, req *model.UpdateBranchServiceRequest) (*model.BranchServiceResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || branchID == "" || id == "" || req == nil {
		return nil, exception.ErrBadRequest
	}
	bs, err := u.repo.FindByID(ctx, tenantID, branchID, id)
	if err != nil {
		return nil, exception.ErrNotFound
	}
	if req.CustomName != nil {
		bs.CustomName = *req.CustomName
	}
	if req.IsActive != nil {
		bs.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		bs.SortOrder = *req.SortOrder
	}
	bs.UpdatedAt = time.Now().UnixMilli()
	if err := u.repo.Update(ctx, bs); err != nil {
		return nil, err
	}
	u.tryAudit(ctx, "BRANCH_SERVICE_UPDATE", bs.ID, map[string]any{"branch_id": bs.BranchID, "service_id": bs.ServiceID, "is_active": bs.IsActive})
	return u.mapToResponse(bs), nil
}

func (u *branchServiceUseCase) DeleteBranchService(ctx context.Context, branchID, id string) error {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || branchID == "" || id == "" {
		return exception.ErrBadRequest
	}
	if err := u.repo.Delete(ctx, tenantID, branchID, id); err != nil {
		return err
	}
	u.tryAudit(ctx, "BRANCH_SERVICE_DELETE", id, map[string]any{"branch_id": branchID})
	return nil
}

func (u *branchServiceUseCase) tryAudit(ctx context.Context, action, entityID string, values map[string]any) {
	if u.audit == nil {
		return
	}
	userID, ok := authcontext.UserIDFromContext(ctx)
	if !ok || userID == "" {
		userID = "system"
	}
	_ = u.audit.LogActivity(ctx, auditModel.CreateAuditLogRequest{
		OrganizationID: database.GetTenantID(ctx),
		UserID:         userID,
		Action:         action,
		Entity:         "branch_service",
		EntityID:       entityID,
		NewValues:      values,
	})
}

func (u *branchServiceUseCase) EnsureActiveBranchService(ctx context.Context, tenantID, branchID, serviceID string) (*entity.BranchService, error) {
	if tenantID == "" || branchID == "" || serviceID == "" {
		return nil, exception.ErrBadRequest
	}
	bs, err := u.repo.FindByService(ctx, tenantID, branchID, serviceID)
	if err != nil {
		return nil, exception.ErrForbidden
	}
	if !bs.IsActive {
		return nil, exception.ErrForbidden
	}
	return bs, nil
}

func (u *branchServiceUseCase) validateTenantBranchService(ctx context.Context, tenantID, branchID, serviceID string) error {
	if _, err := u.branchRepo.FindByID(ctx, tenantID, branchID); err != nil {
		return exception.ErrForbidden
	}
	if _, err := u.serviceRepo.FindByID(ctx, tenantID, serviceID); err != nil {
		return exception.ErrForbidden
	}
	return nil
}

func (u *branchServiceUseCase) mapToResponse(bs *entity.BranchService) *model.BranchServiceResponse {
	return &model.BranchServiceResponse{
		ID:         bs.ID,
		TenantID:   bs.TenantID,
		BranchID:   bs.BranchID,
		ServiceID:  bs.ServiceID,
		CustomName: bs.CustomName,
		IsActive:   bs.IsActive,
		SortOrder:  bs.SortOrder,
		CreatedAt:  bs.CreatedAt,
		UpdatedAt:  bs.UpdatedAt,
	}
}
