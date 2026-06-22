package usecase

import (
	"context"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/settings/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/settings/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/settings/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/google/uuid"
)

type SettingsUseCase interface {
	CreateSetting(ctx context.Context, req *model.CreateSettingRequest) (*model.SettingResponse, error)
	GetSetting(ctx context.Context, settingID string) (*model.SettingResponse, error)
	UpdateSetting(ctx context.Context, settingID string, req *model.UpdateSettingRequest) (*model.SettingResponse, error)
	DeleteSetting(ctx context.Context, settingID string) error
	ResolveSetting(ctx context.Context, req *model.ResolveSettingRequest) (*model.SettingResponse, error)
}

type settingsUseCase struct {
	repo repository.SettingsRepository
}

func NewSettingsUseCase(repo repository.SettingsRepository) SettingsUseCase {
	return &settingsUseCase{repo: repo}
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
	return u.mapToResponse(setting), nil
}

func (u *settingsUseCase) DeleteSetting(ctx context.Context, settingID string) error {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || settingID == "" {
		return exception.ErrBadRequest
	}
	return u.repo.Delete(ctx, tenantID, settingID)
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
			return u.mapToResponse(s), nil
		}
	}

	// 2. Check Service Scope
	if req.ServiceID != "" {
		s, err := u.repo.FindByScope(ctx, tenantID, entity.ScopeTypeService, req.ServiceID, req.Key)
		if err == nil && s != nil {
			return u.mapToResponse(s), nil
		}
	}

	// 3. Check Branch Scope
	if req.BranchID != "" {
		s, err := u.repo.FindByScope(ctx, tenantID, entity.ScopeTypeBranch, req.BranchID, req.Key)
		if err == nil && s != nil {
			return u.mapToResponse(s), nil
		}
	}

	// 4. Check Tenant Scope
	s, err := u.repo.FindByScope(ctx, tenantID, entity.ScopeTypeTenant, tenantID, req.Key)
	if err == nil && s != nil {
		return u.mapToResponse(s), nil
	}

	return nil, exception.ErrNotFound
}

func (u *settingsUseCase) mapToResponse(s *entity.Setting) *model.SettingResponse {
	return &model.SettingResponse{
		ID:        s.ID,
		TenantID:  s.TenantID,
		ScopeType: s.ScopeType,
		ScopeID:   s.ScopeID,
		Key:       s.Key,
		Value:     s.Value,
		ValueType: s.ValueType,
		IsActive:  s.IsActive,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}
