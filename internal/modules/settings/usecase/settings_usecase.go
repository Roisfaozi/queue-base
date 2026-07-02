package usecase

import (
	"context"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/settings/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	"github.com/Roisfaozi/queue-base/internal/modules/settings/repository"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/google/uuid"
)

type AuditLogger interface {
	LogActivity(ctx context.Context, req auditModel.CreateAuditLogRequest) error
}

type SettingsUseCase interface {
	CreateSetting(ctx context.Context, req *model.CreateSettingRequest) (*model.SettingResponse, error)
	GetSetting(ctx context.Context, settingID string) (*model.SettingResponse, error)
	UpdateSetting(ctx context.Context, settingID string, req *model.UpdateSettingRequest) (*model.SettingResponse, error)
	DeleteSetting(ctx context.Context, settingID string) error
	ResolveSetting(ctx context.Context, req *model.ResolveSettingRequest) (*model.SettingResponse, error)
}

type settingsUseCase struct {
	repo  repository.SettingsRepository
	audit AuditLogger
}

func NewSettingsUseCase(repo repository.SettingsRepository, audit ...AuditLogger) SettingsUseCase {
	var auditLogger AuditLogger
	if len(audit) > 0 {
		auditLogger = audit[0]
	}
	return &settingsUseCase{repo: repo, audit: auditLogger}
}

func (u *settingsUseCase) CreateSetting(ctx context.Context, req *model.CreateSettingRequest) (*model.SettingResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return nil, exception.ErrBadRequest
	}
	req.Sanitize()

	vtype := "string"
	if req.ValueType != "" {
		vtype = req.ValueType
	}

	now := time.Now().UnixMilli()
	setting := &entity.Setting{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		ScopeType: req.ScopeType,
		ScopeID:   req.ScopeID,
		Key:       req.Key,
		Value:     req.Value,
		ValueType: vtype,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := u.repo.Create(ctx, setting); err != nil {
		return nil, err
	}
	u.tryAudit(ctx, "SETTING_CREATE", setting.ID, map[string]any{"scope_type": setting.ScopeType, "scope_id": setting.ScopeID, "key": setting.Key})
	return u.mapToResponse(setting), nil
}

func (u *settingsUseCase) GetSetting(ctx context.Context, settingID string) (*model.SettingResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || settingID == "" {
		return nil, exception.ErrBadRequest
	}
	setting, err := u.repo.FindByID(ctx, tenantID, settingID)
	if err != nil {
		return nil, exception.ErrNotFound
	}
	return u.mapToResponse(setting), nil
}

func (u *settingsUseCase) UpdateSetting(ctx context.Context, settingID string, req *model.UpdateSettingRequest) (*model.SettingResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || settingID == "" {
		return nil, exception.ErrBadRequest
	}
	setting, err := u.repo.FindByID(ctx, tenantID, settingID)
	if err != nil {
		return nil, exception.ErrNotFound
	}
	if req.Value != nil {
		setting.Value = *req.Value
	}
	if req.IsActive != nil {
		setting.IsActive = *req.IsActive
	}
	setting.UpdatedAt = time.Now().UnixMilli()
	if err := u.repo.Update(ctx, setting); err != nil {
		return nil, err
	}
	u.tryAudit(ctx, "SETTING_UPDATE", setting.ID, map[string]any{"scope_type": setting.ScopeType, "scope_id": setting.ScopeID, "key": setting.Key})
	return u.mapToResponse(setting), nil
}

func (u *settingsUseCase) DeleteSetting(ctx context.Context, settingID string) error {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || settingID == "" {
		return exception.ErrBadRequest
	}
	if err := u.repo.Delete(ctx, tenantID, settingID); err != nil {
		return err
	}
	u.tryAudit(ctx, "SETTING_DELETE", settingID, nil)
	return nil
}

// ResolveSetting walks the inheritance chain: Counter -> Service -> Branch -> Tenant
func (u *settingsUseCase) ResolveSetting(ctx context.Context, req *model.ResolveSettingRequest) (*model.SettingResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return nil, exception.ErrBadRequest
	}

	// 1. Check Counter Scope
	if req.CounterID != "" {
		s, err := u.repo.FindByScope(ctx, tenantID, entity.ScopeTypeCounter, req.CounterID, req.Key)
		if err == nil && s != nil {
			return u.mapToResponseWithMeta(s, "counter", false), nil
		}
	}

	// 2. Check Service Scope
	if req.ServiceID != "" {
		s, err := u.repo.FindByScope(ctx, tenantID, entity.ScopeTypeService, req.ServiceID, req.Key)
		if err == nil && s != nil {
			return u.mapToResponseWithMeta(s, "service", false), nil
		}
	}

	// 3. Check Branch Scope
	if req.BranchID != "" {
		s, err := u.repo.FindByScope(ctx, tenantID, entity.ScopeTypeBranch, req.BranchID, req.Key)
		if err == nil && s != nil {
			return u.mapToResponseWithMeta(s, "branch", false), nil
		}
	}

	// 4. Check Tenant Scope
	s, err := u.repo.FindByScope(ctx, tenantID, entity.ScopeTypeTenant, tenantID, req.Key)
	if err == nil && s != nil {
		inherited := req.BranchID != "" || req.ServiceID != "" || req.CounterID != ""
		return u.mapToResponseWithMeta(s, "tenant", inherited), nil
	}

	return nil, exception.ErrNotFound
}

func (u *settingsUseCase) mapToResponse(s *entity.Setting) *model.SettingResponse {
	return u.mapToResponseWithMeta(s, s.ScopeType, false)
}

func (u *settingsUseCase) mapToResponseWithMeta(s *entity.Setting, source string, inherited bool) *model.SettingResponse {
	return &model.SettingResponse{
		ID:        s.ID,
		TenantID:  s.TenantID,
		ScopeType: s.ScopeType,
		ScopeID:   s.ScopeID,
		Source:    source,
		Inherited: inherited,
		Key:       s.Key,
		Value:     s.Value,
		ValueType: s.ValueType,
		IsActive:  s.IsActive,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func (u *settingsUseCase) tryAudit(ctx context.Context, action, entityID string, values map[string]any) {
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
		Entity:         "setting",
		EntityID:       entityID,
		NewValues:      values,
	})
}
