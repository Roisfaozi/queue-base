package entity

import "gorm.io/plugin/soft_delete"

const (
	BranchStatusActive   = "active"
	BranchStatusInactive = "inactive"
)

type Branch struct {
	ID        string                `gorm:"column:id;primaryKey;type:varchar(36)"`
	TenantID  string                `gorm:"column:tenant_id;type:varchar(36);not null;index:idx_branch_tenant_deleted;index"`
	Code      string                `gorm:"column:code;type:varchar(50);not null;uniqueIndex:uk_branch_tenant_code"`
	Name      string                `gorm:"column:name;type:varchar(255);not null"`
	Status    string                `gorm:"column:status;type:varchar(20);default:'active';index"`
	CreatedAt int64                 `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index;index:idx_branch_tenant_deleted"`
}

func (Branch) TableName() string {
	return "branches"
}
