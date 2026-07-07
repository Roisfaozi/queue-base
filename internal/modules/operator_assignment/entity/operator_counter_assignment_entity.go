package entity

type OperatorCounterAssignment struct {
	ID           string `gorm:"column:id;primaryKey;type:varchar(36)"`
	TenantID     string `gorm:"column:tenant_id;type:varchar(36);not null;index"`
	BranchID     string `gorm:"column:branch_id;type:varchar(36);not null;index"`
	UserID       string `gorm:"column:user_id;type:varchar(36);not null;index"`
	CounterID    string `gorm:"column:counter_id;type:varchar(36);not null;index"`
	AssignedAt   int64  `gorm:"column:assigned_at;not null"`
	UnassignedAt *int64 `gorm:"column:unassigned_at"`
}

func (OperatorCounterAssignment) TableName() string { return "operator_counter_assignments" }
