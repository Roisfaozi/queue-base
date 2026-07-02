package entity

type TenantQueueSetting struct {
	ID                       string `gorm:"column:id;primaryKey;type:varchar(36)"`
	TenantID                 string `gorm:"column:tenant_id;type:varchar(36);not null;uniqueIndex:uk_tenant_queue_settings_tenant"`
	QueueResetTime           string `gorm:"column:queue_reset_time;type:varchar(10);not null;default:'04:00'"`
	DefaultTicketPrefix      string `gorm:"column:default_ticket_prefix;type:varchar(10);not null;default:'A'"`
	DefaultEstimatedDuration int    `gorm:"column:default_estimated_duration;type:int;not null;default:5"`
	AllowForward             bool   `gorm:"column:allow_forward;not null;default:true"`
	AllowSkip                bool   `gorm:"column:allow_skip;not null;default:true"`
	AllowRecall              bool   `gorm:"column:allow_recall;not null;default:true"`
	AllowCancel              bool   `gorm:"column:allow_cancel;not null;default:true"`
	NumberingStrategy        string `gorm:"column:numbering_strategy;type:varchar(50);not null;default:'daily_branch_sequence'"`
	CreatedAt                int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt                int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}

func (TenantQueueSetting) TableName() string { return "tenant_queue_settings" }

type BranchQueueSetting struct {
	ID                       string  `gorm:"column:id;primaryKey;type:varchar(36)"`
	TenantID                 string  `gorm:"column:tenant_id;type:varchar(36);not null;uniqueIndex:uk_branch_queue_settings_branch"`
	BranchID                 string  `gorm:"column:branch_id;type:varchar(36);not null;uniqueIndex:uk_branch_queue_settings_branch"`
	QueueResetTime           *string `gorm:"column:queue_reset_time;type:varchar(10)"`
	TicketPrefix             *string `gorm:"column:ticket_prefix;type:varchar(10)"`
	DefaultEstimatedDuration *int    `gorm:"column:default_estimated_duration;type:int"`
	AllowForward             *bool   `gorm:"column:allow_forward"`
	AllowSkip                *bool   `gorm:"column:allow_skip"`
	AllowRecall              *bool   `gorm:"column:allow_recall"`
	AllowCancel              *bool   `gorm:"column:allow_cancel"`
	NumberingStrategy        *string `gorm:"column:numbering_strategy;type:varchar(50)"`
	CreatedAt                int64   `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt                int64   `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}

func (BranchQueueSetting) TableName() string { return "branch_queue_settings" }

type ServiceQueueSetting struct {
	ID                       string `gorm:"column:id;primaryKey;type:varchar(36)"`
	TenantID                 string `gorm:"column:tenant_id;type:varchar(36);not null;uniqueIndex:uk_service_queue_settings_service"`
	ServiceID                string `gorm:"column:service_id;type:varchar(36);not null;uniqueIndex:uk_service_queue_settings_service"`
	DefaultEstimatedDuration *int   `gorm:"column:default_estimated_duration;type:int"`
	RequireCounter           *bool  `gorm:"column:require_counter"`
	AllowForwardFrom         *bool  `gorm:"column:allow_forward_from"`
	AllowForwardTo           *bool  `gorm:"column:allow_forward_to"`
	AllowSkip                *bool  `gorm:"column:allow_skip"`
	AllowRecall              *bool  `gorm:"column:allow_recall"`
	AllowCancel              *bool  `gorm:"column:allow_cancel"`
	CreatedAt                int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt                int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}

func (ServiceQueueSetting) TableName() string { return "service_queue_settings" }

type CounterQueueSetting struct {
	ID                       string  `gorm:"column:id;primaryKey;type:varchar(36)"`
	TenantID                 string  `gorm:"column:tenant_id;type:varchar(36);not null;uniqueIndex:uk_counter_queue_settings_counter"`
	CounterID                string  `gorm:"column:counter_id;type:varchar(36);not null;uniqueIndex:uk_counter_queue_settings_counter"`
	QueueResetTime           *string `gorm:"column:queue_reset_time;type:varchar(10)"`
	TicketPrefix             *string `gorm:"column:ticket_prefix;type:varchar(10)"`
	DefaultEstimatedDuration *int    `gorm:"column:default_estimated_duration;type:int"`
	AllowForward             *bool   `gorm:"column:allow_forward"`
	AllowSkip                *bool   `gorm:"column:allow_skip"`
	AllowRecall              *bool   `gorm:"column:allow_recall"`
	AllowCancel              *bool   `gorm:"column:allow_cancel"`
	NumberingStrategy        *string `gorm:"column:numbering_strategy;type:varchar(50)"`
	CreatedAt                int64   `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt                int64   `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}

func (CounterQueueSetting) TableName() string { return "counter_queue_settings" }
