package entity

import "gorm.io/plugin/soft_delete"

const (
	ScopeTypeTenant  = "tenant"
	ScopeTypeBranch  = "branch"
	ScopeTypeService = "service"
	ScopeTypeCounter = "counter"
)

type Setting struct {
	ID        string                `gorm:"column:id;primaryKey;type:varchar(36)"`
	TenantID  string                `gorm:"column:tenant_id;type:varchar(36);not null;index:idx_setting_tenant_deleted;index"`
	ScopeType string                `gorm:"column:scope_type;type:varchar(20);not null;index"`
	ScopeID   string                `gorm:"column:scope_id;type:varchar(36);not null;index"`
	Key       string                `gorm:"column:key;type:varchar(100);not null"`
	Value     string                `gorm:"column:value;type:text"`
	ValueType string                `gorm:"column:value_type;type:varchar(20);default:'string'"`
	IsActive  bool                  `gorm:"column:is_active;default:true"`
	CreatedAt int64                 `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index;index:idx_setting_tenant_deleted"`
}

func (Setting) TableName() string {
	return "settings"
}
