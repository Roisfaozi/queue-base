package settings

import (
	"context"
	"fmt"

	"github.com/Roisfaozi/queue-base/internal/modules/settings/entity"
	settingsModel "github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	settingsUsecase "github.com/Roisfaozi/queue-base/internal/modules/settings/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"gorm.io/gorm"
)

// typedConfigKeys are core QMS keys resolved from typed tables.
// All other keys fall back to generic settings.
var typedConfigKeys = map[string]bool{
	"queue_reset_time":           true,
	"reset_time":                 true,
	"ticket_prefix":              true,
	"numbering_strategy":         true,
	"default_estimated_duration": true,
}

type QueueSettingsResolver struct {
	useCase           settingsUsecase.SettingsUseCase
	db                *gorm.DB
	fallbackToGeneric bool
}

func NewQueueSettingsResolver(db *gorm.DB, useCase settingsUsecase.SettingsUseCase) *QueueSettingsResolver {
	return &QueueSettingsResolver{useCase: useCase, db: db, fallbackToGeneric: true}
}

func (r *QueueSettingsResolver) Resolve(ctx context.Context, key string, branchID string, serviceID string, counterID string) (string, error) {
	resolved, err := r.ResolveDetailed(ctx, key, branchID, serviceID, counterID)
	if err != nil {
		return "", err
	}
	return resolved.Value, nil
}

func (r *QueueSettingsResolver) ResolveDetailed(ctx context.Context, key string, branchID string, serviceID string, counterID string) (*settingsModel.ResolvedQueueSetting, error) {
	// Step 1: try typed tables for core QMS keys
	if typedConfigKeys[key] {
		if resolved, err := r.resolveTypedDetailed(ctx, key, branchID, serviceID, counterID); err == nil && resolved != nil && resolved.Value != "" {
			return resolved, nil
		}
	}

	// Step 2: fall back to generic settings
	if !r.fallbackToGeneric {
		return nil, fmt.Errorf("not found")
	}
	res, err := r.useCase.ResolveSetting(ctx, &settingsModel.ResolveSettingRequest{
		Key:       key,
		BranchID:  branchID,
		ServiceID: serviceID,
		CounterID: counterID,
	})
	if err != nil {
		return nil, err
	}
	return &settingsModel.ResolvedQueueSetting{
		Key:       key,
		Value:     res.Value,
		Source:    res.Source,
		Inherited: res.Inherited,
	}, nil
}

func (r *QueueSettingsResolver) resolveTypedDetailed(ctx context.Context, key string, branchID, serviceID, counterID string) (*settingsModel.ResolvedQueueSetting, error) {
	if r.db == nil {
		return nil, fmt.Errorf("no db")
	}
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return nil, fmt.Errorf("no tenant")
	}

	// Resolve hierarchy: counter -> service -> branch -> tenant
	if counterID != "" {
		if val, err := readTypedMap(r.db, entity.ScopeTypeCounter, tenantID, counterID, key); err == nil && val != nil {
			return &settingsModel.ResolvedQueueSetting{Key: key, Value: *val, Source: entity.ScopeTypeCounter, Inherited: false}, nil
		}
	}
	if serviceID != "" {
		if val, err := readTypedService(r.db, tenantID, serviceID, key); err == nil && val != nil {
			return &settingsModel.ResolvedQueueSetting{Key: key, Value: *val, Source: entity.ScopeTypeService, Inherited: false}, nil
		}
	}
	if branchID != "" {
		if val, err := readTypedBranch(r.db, tenantID, branchID, key); err == nil && val != nil {
			return &settingsModel.ResolvedQueueSetting{Key: key, Value: *val, Source: entity.ScopeTypeBranch, Inherited: false}, nil
		}
	}
	// Tenant default
	if val, err := readTypedTenant(r.db, tenantID, key); err == nil && val != nil {
		inherited := branchID != "" || serviceID != "" || counterID != ""
		return &settingsModel.ResolvedQueueSetting{Key: key, Value: *val, Source: entity.ScopeTypeTenant, Inherited: inherited}, nil
	}
	return nil, fmt.Errorf("not found")
}

// typed table helpers
func readTypedTenant(db *gorm.DB, tenantID, key string) (*string, error) {
	var row entity.TenantQueueSetting
	if err := db.Where("tenant_id = ?", tenantID).First(&row).Error; err != nil {
		return nil, err
	}
	return typedField(&row, key), nil
}

func readTypedBranch(db *gorm.DB, tenantID, branchID, key string) (*string, error) {
	var row entity.BranchQueueSetting
	if err := db.Where("tenant_id = ? AND branch_id = ?", tenantID, branchID).First(&row).Error; err != nil {
		return nil, err
	}
	return typedFieldNullable(&row, key), nil
}

func readTypedService(db *gorm.DB, tenantID, serviceID, key string) (*string, error) {
	var row entity.ServiceQueueSetting
	if err := db.Where("tenant_id = ? AND service_id = ?", tenantID, serviceID).First(&row).Error; err != nil {
		return nil, err
	}
	return typedFieldNullable(&row, key), nil
}

func readTypedMap(db *gorm.DB, scopeType, tenantID, scopeID, key string) (*string, error) {
	switch scopeType {
	case entity.ScopeTypeCounter:
		var row entity.CounterQueueSetting
		if err := db.Where("tenant_id = ? AND counter_id = ?", tenantID, scopeID).First(&row).Error; err != nil {
			return nil, err
		}
		return typedFieldNullable(&row, key), nil
	default:
		return nil, nil
	}
}

func typedField(row *entity.TenantQueueSetting, key string) *string {
	switch key {
	case "queue_reset_time", "reset_time":
		return &row.QueueResetTime
	case "default_ticket_prefix", "ticket_prefix":
		return &row.DefaultTicketPrefix
	case "numbering_strategy":
		return &row.NumberingStrategy
	default:
		return nil
	}
}

func typedFieldNullable(row any, key string) *string {
	switch r := row.(type) {
	case *entity.BranchQueueSetting:
		switch key {
		case "queue_reset_time", "reset_time":
			return r.QueueResetTime
		case "ticket_prefix":
			return r.TicketPrefix
		case "numbering_strategy":
			return r.NumberingStrategy
		}
	case *entity.ServiceQueueSetting:
		return nil
	case *entity.CounterQueueSetting:
		switch key {
		case "queue_reset_time", "reset_time":
			return r.QueueResetTime
		case "ticket_prefix":
			return r.TicketPrefix
		case "numbering_strategy":
			return r.NumberingStrategy
		}
	}
	return nil
}
