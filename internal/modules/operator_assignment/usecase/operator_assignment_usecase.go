package usecase

import (
	"context"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/operator_assignment/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/operator_assignment/model"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLogger interface {
	LogActivity(ctx context.Context, req auditModel.CreateAuditLogRequest) error
}

type OperatorAssignmentUseCase interface {
	Create(ctx context.Context, req *model.OperatorAssignmentRequest) (*model.OperatorAssignmentResponse, error)
	GetAll(ctx context.Context) ([]model.OperatorAssignmentResponse, error)
	Delete(ctx context.Context, id string) error
}

type operatorAssignmentUseCase struct {
	db    *gorm.DB
	audit AuditLogger
}

func NewOperatorAssignmentUseCase(db *gorm.DB, audit ...AuditLogger) OperatorAssignmentUseCase {
	var auditLogger AuditLogger
	if len(audit) > 0 {
		auditLogger = audit[0]
	}
	return &operatorAssignmentUseCase{db: db, audit: auditLogger}
}

func (u *operatorAssignmentUseCase) Create(ctx context.Context, req *model.OperatorAssignmentRequest) (*model.OperatorAssignmentResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || req == nil || req.BranchID == "" || req.UserID == "" || req.CounterID == "" {
		return nil, exception.ErrBadRequest
	}
	if err := u.ensureBranchCounterUser(ctx, tenantID, req.BranchID, req.CounterID, req.UserID); err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	assignment := &entity.OperatorCounterAssignment{ID: uuid.New().String(), TenantID: tenantID, BranchID: req.BranchID, UserID: req.UserID, CounterID: req.CounterID, AssignedAt: now}
	if err := u.db.WithContext(ctx).Create(assignment).Error; err != nil {
		return nil, err
	}
	u.tryAudit(ctx, "OPERATOR_ASSIGNMENT_CREATE", assignment.ID, map[string]string{"branch_id": assignment.BranchID, "user_id": assignment.UserID, "counter_id": assignment.CounterID})
	return toAssignmentResponse(assignment), nil
}

func (u *operatorAssignmentUseCase) GetAll(ctx context.Context) ([]model.OperatorAssignmentResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return nil, exception.ErrBadRequest
	}
	var rows []entity.OperatorCounterAssignment
	if err := u.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("assigned_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	res := make([]model.OperatorAssignmentResponse, 0, len(rows))
	for i := range rows {
		res = append(res, *toAssignmentResponse(&rows[i]))
	}
	return res, nil
}

func (u *operatorAssignmentUseCase) Delete(ctx context.Context, id string) error {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || id == "" {
		return exception.ErrBadRequest
	}
	var assignment entity.OperatorCounterAssignment
	if err := u.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&assignment).Error; err != nil {
		return exception.ErrNotFound
	}
	if assignment.UnassignedAt != nil {
		return nil
	}
	now := time.Now().UnixMilli()
	assignment.UnassignedAt = &now
	if err := u.db.WithContext(ctx).Save(&assignment).Error; err != nil {
		return err
	}
	u.tryAudit(ctx, "OPERATOR_ASSIGNMENT_UNASSIGN", assignment.ID, map[string]string{"unassigned": "true"})
	return nil
}

func (u *operatorAssignmentUseCase) tryAudit(ctx context.Context, action, entityID string, values map[string]string) {
	if u.audit == nil {
		return
	}
	userID, _ := authcontext.UserIDFromContext(ctx)
	_ = u.audit.LogActivity(ctx, auditModel.CreateAuditLogRequest{OrganizationID: database.GetTenantID(ctx), UserID: userID, Action: action, Entity: "operator_assignment", EntityID: entityID, NewValues: values})
}

func toAssignmentResponse(row *entity.OperatorCounterAssignment) *model.OperatorAssignmentResponse {
	return &model.OperatorAssignmentResponse{ID: row.ID, TenantID: row.TenantID, BranchID: row.BranchID, UserID: row.UserID, CounterID: row.CounterID, AssignedAt: row.AssignedAt, UnassignedAt: row.UnassignedAt}
}

func (u *operatorAssignmentUseCase) ensureBranchCounterUser(ctx context.Context, tenantID, branchID, counterID, userID string) error {
	var branchCount int64
	if err := u.db.WithContext(ctx).Model(&branchEntity.Branch{}).Where("id = ? AND tenant_id = ?", branchID, tenantID).Count(&branchCount).Error; err != nil {
		return err
	}
	if branchCount == 0 {
		return exception.ErrNotFound
	}

	var counterCount int64
	if err := u.db.WithContext(ctx).Model(&counterEntity.Counter{}).Where("id = ? AND tenant_id = ? AND branch_id = ?", counterID, tenantID, branchID).Count(&counterCount).Error; err != nil {
		return err
	}
	if counterCount == 0 {
		return exception.ErrNotFound
	}

	var userCount int64
	if err := u.db.WithContext(ctx).Model(&userEntity.User{}).Where("id = ? AND organization_id = ?", userID, tenantID).Count(&userCount).Error; err != nil {
		return err
	}
	if userCount == 0 {
		return exception.ErrNotFound
	}
	return nil
}
