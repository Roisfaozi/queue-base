package entity

import "gorm.io/plugin/soft_delete"

const (
	CounterStatusActive   = "active"
	CounterStatusInactive = "inactive"
)

type Counter struct {
	ID              string                 `gorm:"column:id;primaryKey;type:varchar(36)"`
	TenantID        string                 `gorm:"column:tenant_id;type:varchar(36);not null;index:idx_counter_tenant_deleted;index"`
	BranchID        string                 `gorm:"column:branch_id;type:varchar(36);not null;index:idx_counter_branch_deleted;index"`
	BranchServiceID string                 `gorm:"column:branch_service_id;type:varchar(36);index:idx_counter_branch_service_deleted"`
	Code            string                 `gorm:"column:code;type:varchar(50);not null;uniqueIndex:uk_counter_tenant_branch_code"`
	Name            string                 `gorm:"column:name;type:varchar(255);not null"`
	DisplayName     string                 `gorm:"column:display_name;type:varchar(255)"`
	Status          string                 `gorm:"column:status;type:varchar(20);default:'active';index"`
	Settings        map[string]interface{} `gorm:"column:settings;serializer:json"`
	CreatedAt       int64                  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt       int64                  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt       soft_delete.DeletedAt  `gorm:"column:deleted_at;softDelete:milli;index;index:idx_counter_tenant_deleted;index:idx_counter_branch_deleted"`
}

func (Counter) TableName() string {
	return "counters"
}
