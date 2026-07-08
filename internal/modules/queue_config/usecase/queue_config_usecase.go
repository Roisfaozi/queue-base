package usecase

import (
	"context"
	"fmt"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	auditUsecase "github.com/Roisfaozi/queue-base/internal/modules/audit/usecase"
	"github.com/Roisfaozi/queue-base/internal/modules/queue_config/repository"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/exception"
)

type QueueConfigUseCase interface {
	UpdateTenantQueueSetting(ctx context.Context, tenantID string, values map[string]any) error
	UpdateBranchQueueSetting(ctx context.Context, tenantID, branchID string, values map[string]any) error
	UpdateBranchServiceQueueSetting(ctx context.Context, tenantID, branchID, branchServiceID string, values map[string]any) error
	UpdateCounterQueueSetting(ctx context.Context, tenantID, counterID string, values map[string]any) error
	ResetQueueSetting(ctx context.Context, entity, entityID, field string) error
}

type queueConfigUseCase struct {
	repo  repository.QueueConfigRepository
	audit auditUsecase.AuditUseCase
}

func NewQueueConfigUseCase(repo repository.QueueConfigRepository, audit auditUsecase.AuditUseCase) QueueConfigUseCase {
	return &queueConfigUseCase{repo: repo, audit: audit}
}

func (u *queueConfigUseCase) UpdateTenantQueueSetting(ctx context.Context, tenantID string, values map[string]any) error {
	if tenantID == "" {
		return exception.ErrBadRequest
	}
	if err := u.repo.UpsertTenantQueueSetting(ctx, tenantID, values); err != nil {
		return err
	}
	u.tryAudit(ctx, "SETTING_UPDATE", "tenant_queue_settings:"+tenantID, values)
	return nil
}

func (u *queueConfigUseCase) UpdateBranchQueueSetting(ctx context.Context, tenantID, branchID string, values map[string]any) error {
	if tenantID == "" || branchID == "" {
		return exception.ErrBadRequest
	}
	if err := u.repo.UpsertBranchQueueSetting(ctx, tenantID, branchID, values); err != nil {
		return err
	}
	u.tryAudit(ctx, "SETTING_UPDATE", "branch_queue_settings:"+branchID, values)
	return nil
}

func (u *queueConfigUseCase) UpdateBranchServiceQueueSetting(ctx context.Context, tenantID, branchID, branchServiceID string, values map[string]any) error {
	if tenantID == "" || branchID == "" || branchServiceID == "" {
		return exception.ErrBadRequest
	}
	if err := u.repo.UpsertBranchServiceQueueSetting(ctx, tenantID, branchID, branchServiceID, values); err != nil {
		return err
	}
	u.tryAudit(ctx, "SETTING_UPDATE", "branch_service_queue_settings:"+branchServiceID, values)
	return nil
}

func (u *queueConfigUseCase) UpdateCounterQueueSetting(ctx context.Context, tenantID, counterID string, values map[string]any) error {
	if tenantID == "" || counterID == "" {
		return exception.ErrBadRequest
	}
	if err := u.repo.UpsertCounterQueueSetting(ctx, tenantID, counterID, values); err != nil {
		return err
	}
	u.tryAudit(ctx, "SETTING_UPDATE", "counter_queue_settings:"+counterID, values)
	return nil
}

func (u *queueConfigUseCase) ResetQueueSetting(ctx context.Context, entity, entityID, field string) error {
	if u.audit == nil {
		return nil
	}
	userID, _ := authcontext.UserIDFromContext(ctx)
	if userID == "" {
		userID = "system"
	}
	_ = u.audit.LogActivity(ctx, auditModel.CreateAuditLogRequest{UserID: userID, Action: "SETTING_RESET", Entity: entity, EntityID: entityID, NewValues: map[string]string{"field": field}})
	return nil
}

func (u *queueConfigUseCase) tryAudit(ctx context.Context, action, entityID string, values map[string]any) {
	if u.audit == nil {
		return
	}
	userID, _ := authcontext.UserIDFromContext(ctx)
	if userID == "" {
		userID = "system"
	}
	flat := map[string]string{}
	for key, value := range values {
		flat[key] = fmtAny(value)
	}
	_ = u.audit.LogActivity(ctx, auditModel.CreateAuditLogRequest{UserID: userID, Action: action, Entity: "queue_setting", EntityID: entityID, NewValues: flat})
}

func fmtAny(v any) string {
	switch value := v.(type) {
	case string:
		return value
	case *string:
		if value == nil {
			return ""
		}
		return *value
	case bool:
		if value {
			return "true"
		}
		return "false"
	case *bool:
		if value == nil {
			return ""
		}
		if *value {
			return "true"
		}
		return "false"
	case int:
		return fmt.Sprintf("%d", value)
	case *int:
		if value == nil {
			return ""
		}
		return fmt.Sprintf("%d", *value)
	default:
		return fmt.Sprintf("%v", value)
	}
}
