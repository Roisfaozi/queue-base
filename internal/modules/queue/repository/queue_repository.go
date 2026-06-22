package repository

import (
	"context"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/entity"
)

type QueueRepository interface {
	NextQueueNumber(ctx context.Context, tenantID, branchID string, date time.Time, prefix string) (int, error)
	CreateRegistration(ctx context.Context, queue *entity.Queue, journey *entity.QueueJourney, visit *entity.VisitJourney) error
	FindQueueByID(ctx context.Context, tenantID, queueID string) (*entity.Queue, error)
	FindCurrentJourney(ctx context.Context, queueID, journeyID string) (*entity.QueueJourney, error)
	NextJourneySequence(ctx context.Context, queueID string) (int, error)
	CreateForwarding(ctx context.Context, queue *entity.Queue, currentJourney *entity.QueueJourney, nextJourney *entity.QueueJourney, visit *entity.VisitJourney) error
}
