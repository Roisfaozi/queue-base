package entity

import "gorm.io/plugin/soft_delete"

const (
	QueueStatusWaiting   = "waiting"
	QueueStatusCalling   = "calling"
	QueueStatusServing   = "serving"
	QueueStatusSkipped   = "skipped"
	QueueStatusCanceled  = "canceled"
	QueueStatusCompleted = "completed"

	JourneyStatusPending   = "pending"
	JourneyStatusCalling   = "calling"
	JourneyStatusServing   = "serving"
	JourneyStatusSkipped   = "skipped"
	JourneyStatusCanceled  = "canceled"
	JourneyStatusCompleted = "completed"
	JourneyStatusForwarded = "forwarded"
)

type Queue struct {
	ID               string                `gorm:"column:id;primaryKey;type:varchar(36)"`
	TenantID         string                `gorm:"column:tenant_id;type:varchar(36);not null"`
	BranchID         string                `gorm:"column:branch_id;type:varchar(36);not null"`
	QueueDate        string                `gorm:"column:queue_date;type:date;not null"`
	TicketNo         string                `gorm:"column:ticket_no;type:varchar(50);not null"`
	QueueNo          int                   `gorm:"column:queue_no;type:int;not null"`
	PatientID        string                `gorm:"column:patient_id;type:varchar(36)"`
	PatientName      string                `gorm:"column:patient_name;type:varchar(255)"`
	Status           string                `gorm:"column:status;type:varchar(50);not null"`
	CurrentJourneyID string                `gorm:"column:current_journey_id;type:varchar(36)"`
	CreatedAt        int64                 `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt        int64                 `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt        soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli"`
}

type QueueJourney struct {
	ID        string `gorm:"column:id;primaryKey;type:varchar(36)"`
	QueueID   string `gorm:"column:queue_id;type:varchar(36);not null"`
	TenantID  string `gorm:"column:tenant_id;type:varchar(36);not null"`
	ServiceID string `gorm:"column:service_id;type:varchar(36);not null"`
	CounterID string `gorm:"column:counter_id;type:varchar(36)"`
	SeqNo     int    `gorm:"column:seq_no;type:int;not null"`
	Status    string `gorm:"column:status;type:varchar(50);not null"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}

type VisitJourney struct {
	ID        string `gorm:"column:id;primaryKey;type:varchar(36)"`
	QueueID   string `gorm:"column:queue_id;type:varchar(36);not null"`
	TenantID  string `gorm:"column:tenant_id;type:varchar(36);not null"`
	EventType string `gorm:"column:event_type;type:varchar(100);not null"`
	Payload   string `gorm:"column:payload;type:text"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:milli"`
}

func (Queue) TableName() string        { return "queues" }
func (QueueJourney) TableName() string { return "queue_journeys" }
func (VisitJourney) TableName() string { return "visit_journeys" }
