package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/google/uuid"
)

type QueueUseCase interface {
	RegisterQueue(ctx context.Context, req *model.RegisterQueueRequest) (*model.QueueResponse, error)
	ForwardQueue(ctx context.Context, queueID string, req *model.ForwardQueueRequest) (*model.QueueResponse, error)
}

type queueUseCase struct {
	repo repository.QueueRepository
}

func NewQueueUseCase(repo repository.QueueRepository) QueueUseCase {
	return &queueUseCase{repo: repo}
}

func (u *queueUseCase) RegisterQueue(ctx context.Context, req *model.RegisterQueueRequest) (*model.QueueResponse, error) {
	tenantID := database.GetTenantID(ctx)
	branchID := database.GetBranchID(ctx)

	if tenantID == "" || branchID == "" {
		return nil, exception.ErrBadRequest
	}

	req.Sanitize()

	now := time.Now()
	// FIXME: should compute queue_date based on branch reset_time
	queueDateStr := now.Format("2006-01-02")
	prefix := "A" // FIXME: dynamic by service

	qNo, err := u.repo.NextQueueNumber(ctx, tenantID, branchID, now, prefix)
	if err != nil {
		return nil, err
	}

	ticketNo := fmt.Sprintf("%s%03d", prefix, qNo)
	nowMs := now.UnixMilli()

	queueID := uuid.New().String()
	journeyID := uuid.New().String()
	visitID := uuid.New().String()

	q := &entity.Queue{
		ID:               queueID,
		TenantID:         tenantID,
		BranchID:         branchID,
		QueueDate:        queueDateStr,
		TicketNo:         ticketNo,
		QueueNo:          qNo,
		PatientID:        req.PatientID,
		PatientName:      req.PatientName,
		Status:           entity.QueueStatusWaiting,
		CurrentJourneyID: journeyID,
		CreatedAt:        nowMs,
		UpdatedAt:        nowMs,
	}

	j := &entity.QueueJourney{
		ID:        journeyID,
		QueueID:   queueID,
		TenantID:  tenantID,
		ServiceID: req.ServiceID,
		SeqNo:     1,
		Status:    entity.JourneyStatusPending,
		CreatedAt: nowMs,
		UpdatedAt: nowMs,
	}

	v := &entity.VisitJourney{
		ID:        visitID,
		QueueID:   queueID,
		TenantID:  tenantID,
		EventType: "registration",
		Payload:   fmt.Sprintf(`{"service_id":"%s"}`, req.ServiceID),
		CreatedAt: nowMs,
	}

	if err := u.repo.CreateRegistration(ctx, q, j, v); err != nil {
		return nil, err
	}

	return &model.QueueResponse{
		ID:               q.ID,
		TenantID:         q.TenantID,
		BranchID:         q.BranchID,
		QueueDate:        q.QueueDate,
		TicketNo:         q.TicketNo,
		QueueNo:          q.QueueNo,
		PatientID:        q.PatientID,
		PatientName:      q.PatientName,
		Status:           q.Status,
		CurrentJourneyID: q.CurrentJourneyID,
		CreatedAt:        q.CreatedAt,
		UpdatedAt:        q.UpdatedAt,
	}, nil
}

func (u *queueUseCase) ForwardQueue(ctx context.Context, queueID string, req *model.ForwardQueueRequest) (*model.QueueResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || queueID == "" || req == nil || req.DestinationServiceID == "" {
		return nil, exception.ErrBadRequest
	}

	queue, err := u.repo.FindQueueByID(ctx, tenantID, queueID)
	if err != nil || queue == nil {
		return nil, exception.ErrNotFound
	}

	currentJourney, err := u.repo.FindCurrentJourney(ctx, queue.ID, queue.CurrentJourneyID)
	if err != nil || currentJourney == nil {
		return nil, exception.ErrNotFound
	}

	seqNo, err := u.repo.NextJourneySequence(ctx, queue.ID)
	if err != nil {
		return nil, err
	}

	nowMs := time.Now().UnixMilli()
	nextJourneyID := uuid.New().String()
	visitID := uuid.New().String()

	currentJourney.Status = entity.JourneyStatusForwarded
	currentJourney.UpdatedAt = nowMs

	nextJourney := &entity.QueueJourney{
		ID:        nextJourneyID,
		QueueID:   queue.ID,
		TenantID:  tenantID,
		ServiceID: req.DestinationServiceID,
		CounterID: req.DestinationCounterID,
		SeqNo:     seqNo,
		Status:    entity.JourneyStatusPending,
		CreatedAt: nowMs,
		UpdatedAt: nowMs,
	}

	visit := &entity.VisitJourney{
		ID:        visitID,
		QueueID:   queue.ID,
		TenantID:  tenantID,
		EventType: "forward",
		Payload:   fmt.Sprintf(`{"from_journey_id":"%s","to_service_id":"%s"}`, currentJourney.ID, req.DestinationServiceID),
		CreatedAt: nowMs,
	}

	queue.CurrentJourneyID = nextJourneyID
	queue.UpdatedAt = nowMs

	if err := u.repo.CreateForwarding(ctx, queue, currentJourney, nextJourney, visit); err != nil {
		return nil, err
	}

	return &model.QueueResponse{
		ID:               queue.ID,
		TenantID:         queue.TenantID,
		BranchID:         queue.BranchID,
		QueueDate:        queue.QueueDate,
		TicketNo:         queue.TicketNo,
		QueueNo:          queue.QueueNo,
		PatientID:        queue.PatientID,
		PatientName:      queue.PatientName,
		Status:           queue.Status,
		CurrentJourneyID: queue.CurrentJourneyID,
		CreatedAt:        queue.CreatedAt,
		UpdatedAt:        queue.UpdatedAt,
	}, nil
}
