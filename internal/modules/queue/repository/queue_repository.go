package repository

import (
	"context"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/entity"
)

type QueueRepository interface {
	NextQueueNumber(ctx context.Context, tenantID, branchID string, date time.Time, prefix string) (int, error)
	CreateRegistration(ctx context.Context, queue *entity.Queue, journey *entity.QueueJourney, visit *entity.VisitJourney) error
}
