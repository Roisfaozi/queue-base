package entity

type ClientType string

const (
	ClientTypeCaller  ClientType = "caller"
	ClientTypeSignage ClientType = "signage"
	ClientTypeScanner ClientType = "scanner"
	ClientTypeKiosk   ClientType = "kiosk"
)

type QMSClient struct {
	ID              string     `gorm:"column:id;primaryKey;type:varchar(36)"`
	TenantID        string     `gorm:"column:tenant_id;type:varchar(36);not null;index"`
	BranchID        string     `gorm:"column:branch_id;type:varchar(36);not null;index"`
	BranchServiceID *string    `gorm:"column:branch_service_id;type:varchar(36)"`
	CounterID       *string    `gorm:"column:counter_id;type:varchar(36)"`
	ClientType      ClientType `gorm:"column:client_type;type:varchar(50);not null"`
	Name            string     `gorm:"column:name;type:varchar(255);not null"`
	Description     *string    `gorm:"column:description;type:text"`
	IsActive        bool       `gorm:"column:is_active;not null;default:true"`
	CreatedAt       int64      `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt       int64      `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}

func (QMSClient) TableName() string { return "qms_clients" }

type QMSClientCredential struct {
	ID               string `gorm:"column:id;primaryKey;type:varchar(36)"`
	TenantID         string `gorm:"column:tenant_id;type:varchar(36);not null;index"`
	ClientID         string `gorm:"column:client_id;type:varchar(36);not null;index"`
	ClientSecretHash string `gorm:"column:client_secret_hash;type:varchar(255);not null"`
	ExpiresAt        *int64 `gorm:"column:expires_at"`
	CreatedAt        int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt        int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}

func (QMSClientCredential) TableName() string { return "qms_client_credentials" }
