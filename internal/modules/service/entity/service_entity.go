package entity

import "gorm.io/plugin/soft_delete"

const (
	ServiceStatusActive   = "active"
	ServiceStatusInactive = "inactive"

	ServiceTypeRegistration = "registration"
	ServiceTypeDoctor       = "doctor"
	ServiceTypePharmacy     = "pharmacy"
	ServiceTypeCashier      = "cashier"
	ServiceTypeLaboratory   = "laboratory"
	ServiceTypeGeneral      = "general"
)

type Service struct {
	ID                       string                 `gorm:"column:id;primaryKey;type:varchar(36)"`
	TenantID                 string                 `gorm:"column:tenant_id;type:varchar(36);not null;index:idx_service_tenant_deleted;index"`
	Code                     string                 `gorm:"column:code;type:varchar(50);not null;uniqueIndex:uk_service_tenant_code"`
	Name                     string                 `gorm:"column:name;type:varchar(255);not null"`
	Type                     string                 `gorm:"column:type;type:varchar(50);not null;default:'general';index"`
	DefaultEstimatedDuration int                    `gorm:"column:default_estimated_duration;type:int;not null;default:5"`
	Status                   string                 `gorm:"column:status;type:varchar(20);default:'active';index"`
	IsPharmacy               bool                   `gorm:"column:is_pharmacy;not null;default:false"`
	IsPharmacyReception      bool                   `gorm:"column:is_pharmacy_reception;not null;default:false"`
	Settings                 map[string]interface{} `gorm:"column:settings;serializer:json"`
	CreatedAt                int64                  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt                int64                  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt                soft_delete.DeletedAt  `gorm:"column:deleted_at;softDelete:milli;index;index:idx_service_tenant_deleted"`
}

func (Service) TableName() string {
	return "services"
}

type BranchService struct {
	ID         string `gorm:"column:id;primaryKey;type:varchar(36)"`
	TenantID   string `gorm:"column:tenant_id;type:varchar(36);not null;uniqueIndex:uk_branch_service_tenant_branch_service;index:idx_branch_service_tenant"`
	BranchID   string `gorm:"column:branch_id;type:varchar(36);not null;uniqueIndex:uk_branch_service_tenant_branch_service;index:idx_branch_service_branch"`
	ServiceID  string `gorm:"column:service_id;type:varchar(36);not null;uniqueIndex:uk_branch_service_tenant_branch_service;index:idx_branch_service_service"`
	CustomName string `gorm:"column:custom_name;type:varchar(255)"`
	IsActive   bool   `gorm:"column:is_active;not null;default:true;index"`
	SortOrder  int    `gorm:"column:sort_order;type:int;not null;default:0"`
	CreatedAt  int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt  int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}

func (BranchService) TableName() string {
	return "branch_services"
}
