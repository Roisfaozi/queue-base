package repository

import (
	"context"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/model"
	txpkg "github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type QueueRepository interface {
	NextQueueNumber(ctx context.Context, tenantID, branchID string, date time.Time, prefix string) (int, error)
	ExistsRegistration(ctx context.Context, tenantID, branchID, queueDate, patientID, patientName string) (bool, error)
	CreateRegistration(ctx context.Context, queue *entity.Queue, journey *entity.QueueJourney, visit *entity.VisitJourney) error
	ListQueues(ctx context.Context, tenantID, branchID string, req model.ListQueuesRequest) ([]*entity.Queue, error)
	FindQueueByID(ctx context.Context, tenantID, queueID string) (*entity.Queue, error)
	FindCurrentJourney(ctx context.Context, queueID, journeyID string) (*entity.QueueJourney, error)
	NextJourneySequence(ctx context.Context, queueID string) (int, error)
	CreateForwarding(ctx context.Context, queue *entity.Queue, currentJourney *entity.QueueJourney, nextJourney *entity.QueueJourney, visit *entity.VisitJourney) error
	UpdateQueueState(ctx context.Context, queue *entity.Queue, currentJourney *entity.QueueJourney, visit *entity.VisitJourney) error
}

type queueRepository struct {
	db *gorm.DB
}

func NewQueueRepository(db *gorm.DB) QueueRepository {
	return &queueRepository{db: db}
}

func (r *queueRepository) getDB(ctx context.Context) *gorm.DB {
	if txDB, ok := txpkg.DBFromContext(ctx); ok {
		return txDB.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

type queueCounterRow struct {
	TenantID  string `gorm:"column:tenant_id"`
	BranchID  string `gorm:"column:branch_id"`
	QueueDate string `gorm:"column:queue_date"`
	Prefix    string `gorm:"column:prefix"`
	LastValue int    `gorm:"column:last_value"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
}

func (queueCounterRow) TableName() string { return "queue_counters" }

func (r *queueRepository) NextQueueNumber(ctx context.Context, tenantID, branchID string, date time.Time, prefix string) (int, error) {
	queueDate := date.Format("2006-01-02")
	nowMs := date.UnixMilli()
	var next int
	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		row := &queueCounterRow{}
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("tenant_id = ? AND branch_id = ? AND queue_date = ? AND prefix = ?", tenantID, branchID, queueDate, prefix).
			First(row).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				row = &queueCounterRow{TenantID: tenantID, BranchID: branchID, QueueDate: queueDate, Prefix: prefix, LastValue: 1, CreatedAt: nowMs, UpdatedAt: nowMs}
				if err := tx.Table(row.TableName()).Create(row).Error; err != nil {
					return err
				}
				next = 1
				return nil
			}
			return err
		}
		row.LastValue++
		row.UpdatedAt = nowMs
		if err := tx.Table(row.TableName()).Where("tenant_id = ? AND branch_id = ? AND queue_date = ? AND prefix = ?", tenantID, branchID, queueDate, prefix).Update("last_value", row.LastValue).Error; err != nil {
			return err
		}
		next = row.LastValue
		return nil
	})
	return next, err
}

func (r *queueRepository) ExistsRegistration(ctx context.Context, tenantID, branchID, queueDate, patientID, patientName string) (bool, error) {
	var count int64
	query := r.getDB(ctx).Model(&entity.Queue{}).Where("tenant_id = ? AND branch_id = ? AND queue_date = ?", tenantID, branchID, queueDate)
	if patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	} else {
		query = query.Where("patient_name = ?", patientName)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *queueRepository) CreateRegistration(ctx context.Context, queue *entity.Queue, journey *entity.QueueJourney, visit *entity.VisitJourney) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(queue).Error; err != nil {
			return err
		}
		if err := tx.Create(journey).Error; err != nil {
			return err
		}
		if err := tx.Create(visit).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Queue{}).Where("id = ?", queue.ID).Update("current_journey_id", journey.ID).Error
	})
}

func (r *queueRepository) ListQueues(ctx context.Context, tenantID, branchID string, req model.ListQueuesRequest) ([]*entity.Queue, error) {
	var queues []*entity.Queue
	query := r.getDB(ctx).Model(&entity.Queue{}).Where("queues.tenant_id = ? AND queues.branch_id = ?", tenantID, branchID)
	if req.Status != "" {
		query = query.Where("queues.status = ?", req.Status)
	}
	if req.QueueDate != "" {
		query = query.Where("queues.queue_date = ?", req.QueueDate)
	}
	if req.ServiceID != "" {
		query = query.Joins("JOIN queue_journeys ON queue_journeys.queue_id = queues.id").Where("queue_journeys.service_id = ?", req.ServiceID)
	}
	if err := query.Find(&queues).Error; err != nil {
		return nil, err
	}
	return queues, nil
}

func (r *queueRepository) FindQueueByID(ctx context.Context, tenantID, queueID string) (*entity.Queue, error) {
	var q entity.Queue
	if err := r.getDB(ctx).Where("tenant_id = ? AND id = ?", tenantID, queueID).First(&q).Error; err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *queueRepository) FindCurrentJourney(ctx context.Context, queueID, journeyID string) (*entity.QueueJourney, error) {
	var j entity.QueueJourney
	if err := r.getDB(ctx).Where("queue_id = ? AND id = ?", queueID, journeyID).First(&j).Error; err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *queueRepository) NextJourneySequence(ctx context.Context, queueID string) (int, error) {
	var maxSeq int
	if err := r.getDB(ctx).Model(&entity.QueueJourney{}).Where("queue_id = ?", queueID).Select("COALESCE(MAX(seq_no), 0)").Scan(&maxSeq).Error; err != nil {
		return 0, err
	}
	return maxSeq + 1, nil
}

func (r *queueRepository) CreateForwarding(ctx context.Context, queue *entity.Queue, currentJourney *entity.QueueJourney, nextJourney *entity.QueueJourney, visit *entity.VisitJourney) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&entity.QueueJourney{}).Where("id = ?", currentJourney.ID).Updates(map[string]any{"status": currentJourney.Status, "updated_at": currentJourney.UpdatedAt})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Create(nextJourney).Error; err != nil {
			return err
		}
		if err := tx.Create(visit).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Queue{}).Where("id = ?", queue.ID).Updates(map[string]any{"current_journey_id": queue.CurrentJourneyID, "updated_at": queue.UpdatedAt}).Error
	})
}

func (r *queueRepository) UpdateQueueState(ctx context.Context, queue *entity.Queue, currentJourney *entity.QueueJourney, visit *entity.VisitJourney) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		queueResult := tx.Model(&entity.Queue{}).Where("id = ?", queue.ID).Updates(map[string]any{"status": queue.Status, "updated_at": queue.UpdatedAt})
		if queueResult.Error != nil {
			return queueResult.Error
		}
		if queueResult.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		journeyResult := tx.Model(&entity.QueueJourney{}).Where("id = ?", currentJourney.ID).Updates(map[string]any{"status": currentJourney.Status, "updated_at": currentJourney.UpdatedAt})
		if journeyResult.Error != nil {
			return journeyResult.Error
		}
		if journeyResult.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Create(visit).Error
	})
}
