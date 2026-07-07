package entity

import "gorm.io/plugin/soft_delete"

const (
	OrgStatusDraft     = "draft"
	OrgStatusActive    = "active"
	OrgStatusSuspended = "suspended"
	OrgStatusInactive  = "inactive"
)

// Organization represents a tenant/workspace in the multi-tenant system.
// Users can belong to multiple organizations with different roles.
type Organization struct {
	ID          string                 `gorm:"column:id;primaryKey;type:varchar(36)"`
	Code        string                 `gorm:"column:code;type:varchar(50);uniqueIndex:uk_org_code_deleted"`
	Name        string                 `gorm:"column:name;type:varchar(255);not null"`
	LegalName   string                 `gorm:"column:legal_name;type:varchar(255)"`
	Slug        string                 `gorm:"column:slug;type:varchar(100);uniqueIndex;not null"`
	OwnerID     string                 `gorm:"column:owner_id;type:varchar(36);not null;index:idx_org_owner_deleted;index"`
	Address     string                 `gorm:"column:address;type:text"`
	City        string                 `gorm:"column:city;type:varchar(100)"`
	Province    string                 `gorm:"column:province;type:varchar(100)"`
	PostalCode  string                 `gorm:"column:postal_code;type:varchar(20)"`
	Phone       string                 `gorm:"column:phone;type:varchar(50)"`
	Email       string                 `gorm:"column:email;type:varchar(255)"`
	LogoAssetID string                 `gorm:"column:logo_asset_id;type:varchar(36)"`
	Timezone    string                 `gorm:"column:timezone;type:varchar(100)"`
	Settings    map[string]interface{} `gorm:"serializer:json"`
	Status      string                 `gorm:"column:status;type:varchar(20);default:'active';index:idx_org_status_deleted;index"`
	CreatedAt   int64                  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt   int64                  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt   soft_delete.DeletedAt  `gorm:"column:deleted_at;softDelete:milli;index;index:idx_org_owner_deleted;index:idx_org_status_deleted"`

	// Relations
	Members []OrganizationMember `gorm:"foreignKey:OrganizationID;references:ID"`
}

func (Organization) TableName() string {
	return "organizations"
}
