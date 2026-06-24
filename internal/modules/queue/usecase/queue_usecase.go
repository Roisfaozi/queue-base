package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/queue/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/repository"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/google/uuid"
)

type SettingsResolver interface {
	Resolve(ctx context.Context, key string, branchID string, serviceID string, counterID string) (string, error)
}

type RelationValidator interface {
	Validate(ctx context.Context, tenantID, branchID, serviceID, counterID string) error
}

type QueueUseCase interface {
	RegisterQueue(ctx context.Context, req *model.RegisterQueueRequest) (*model.QueueResponse, error)
	ListQueues(ctx context.Context, req model.ListQueuesRequest) ([]model.QueueResponse, error)
	GetQueueByID(ctx context.Context, queueID string) (*model.QueueResponse, error)
	ForwardQueue(ctx context.Context, queueID string, req *model.ForwardQueueRequest) (*model.QueueResponse, error)
	TransitionQueue(ctx context.Context, queueID string, req *model.QueueTransitionRequest) (*model.QueueResponse, error)
	ListActiveJourneys(ctx context.Context, req model.QueueJourneyListRequest) ([]model.QueueJourneyResponse, error)
	GetVisitJourneys(ctx context.Context, queueID string) ([]model.VisitJourneyResponse, error)
	GetQueueStats(ctx context.Context) (*model.QueueStatsResponse, error)
}

type queueUseCase struct {
	repo             repository.QueueRepository
	settingsResolver SettingsResolver
	validator        RelationValidator
}

func NewQueueUseCase(repo repository.QueueRepository, settingsResolver SettingsResolver, validator RelationValidator) QueueUseCase {
	return &queueUseCase{repo: repo, settingsResolver: settingsResolver, validator: validator}
}

func (u *queueUseCase) ListQueues(ctx context.Context, req model.ListQueuesRequest) ([]model.QueueResponse, error) {
	tenantID := database.GetTenantID(ctx)
	branchID := database.GetBranchID(ctx)
	if tenantID == "" || branchID == "" {
		return nil, exception.ErrBadRequest
	}
	queues, err := u.repo.ListQueues(ctx, tenantID, branchID, req)
	if err != nil {
		return nil, err
	}
	res := make([]model.QueueResponse, len(queues))
	for i, q := range queues {
		res[i] = mapQueueResponse(q)
	}
	return res, nil
}

func (u *queueUseCase) GetQueueStats(ctx context.Context) (*model.QueueStatsResponse, error) {
	tenantID := database.GetTenantID(ctx)
	branchID := database.GetBranchID(ctx)
	if tenantID == "" || branchID == "" {
		return nil, exception.ErrBadRequest
	}
	if u.validator != nil {
		if err := u.validator.Validate(ctx, tenantID, branchID, "", ""); err != nil {
			return nil, err
		}
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	resetTime := "04:00"
	if u.settingsResolver != nil {
		if resolved, err := u.settingsResolver.Resolve(ctx, "queue_reset_time", branchID, "", ""); err == nil && resolved != "" {
			resetTime = resolved
		} else if resolved, err := u.settingsResolver.Resolve(ctx, "reset_time", branchID, "", ""); err == nil && resolved != "" {
			resetTime = resolved
		}
	}
	queueDateStr := computeBusinessQueueDate(now, resetTime)

	stats, err := u.repo.GetQueueStats(ctx, tenantID, branchID, queueDateStr)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (u *queueUseCase) ListActiveJourneys(ctx context.Context, req model.QueueJourneyListRequest) ([]model.QueueJourneyResponse, error) {
	tenantID := database.GetTenantID(ctx)
	branchID := database.GetBranchID(ctx)
	if tenantID == "" || branchID == "" {
		return nil, exception.ErrBadRequest
	}
	if u.validator != nil {
		if err := u.validator.Validate(ctx, tenantID, branchID, req.ServiceID, req.CounterID); err != nil {
			return nil, err
		}
	}
	journeys, err := u.repo.ListActiveJourneys(ctx, tenantID, branchID, req)
	if err != nil {
		return nil, err
	}
	res := make([]model.QueueJourneyResponse, len(journeys))
	for i, journey := range journeys {
		res[i] = model.QueueJourneyResponse{
			ID:        journey.ID,
			QueueID:   journey.QueueID,
			ServiceID: journey.ServiceID,
			CounterID: journey.CounterID,
			SeqNo:     journey.SeqNo,
			Status:    journey.Status,
			CreatedAt: journey.CreatedAt,
			UpdatedAt: journey.UpdatedAt,
		}
	}
	return res, nil
}

func (u *queueUseCase) GetVisitJourneys(ctx context.Context, queueID string) ([]model.VisitJourneyResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || queueID == "" {
		return nil, exception.ErrBadRequest
	}
	// Ensure queue exists and belongs to tenant
	if _, err := u.repo.FindQueueByID(ctx, tenantID, queueID); err != nil {
		return nil, exception.ErrNotFound
	}
	visits, err := u.repo.FindVisitJourneysByQueueID(ctx, tenantID, queueID)
	if err != nil {
		return nil, err
	}
	res := make([]model.VisitJourneyResponse, len(visits))
	for i, v := range visits {
		res[i] = model.VisitJourneyResponse{
			ID:        v.ID,
			QueueID:   v.QueueID,
			TenantID:  v.TenantID,
			EventType: v.EventType,
			Payload:   v.Payload,
			CreatedAt: v.CreatedAt,
		}
	}
	return res, nil
}

func (u *queueUseCase) GetQueueByID(ctx context.Context, queueID string) (*model.QueueResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || queueID == "" {
		return nil, exception.ErrBadRequest
	}
	queue, err := u.repo.FindQueueByID(ctx, tenantID, queueID)
	if err != nil {
		return nil, exception.ErrNotFound
	}
	res := mapQueueResponse(queue)
	return &res, nil
}

func (u *queueUseCase) RegisterQueue(ctx context.Context, req *model.RegisterQueueRequest) (*model.QueueResponse, error) {
	tenantID := database.GetTenantID(ctx)
	branchID := database.GetBranchID(ctx)

	if tenantID == "" || branchID == "" {
		return nil, exception.ErrBadRequest
	}

	req.Sanitize()

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	resetTime := "04:00"
	if u.settingsResolver != nil {
		if resolved, err := u.settingsResolver.Resolve(ctx, "queue_reset_time", branchID, req.ServiceID, ""); err == nil && resolved != "" {
			resetTime = resolved
		} else if resolved, err := u.settingsResolver.Resolve(ctx, "reset_time", branchID, req.ServiceID, ""); err == nil && resolved != "" {
			resetTime = resolved
		}
	}
	queueDateStr := computeBusinessQueueDate(now, resetTime)
	prefix := resolveTicketPrefix(ctx, u.settingsResolver, branchID, req.ServiceID)
	_ = resolveNumberingStrategy(ctx, u.settingsResolver, branchID, req.ServiceID)

	exists, err := u.repo.ExistsRegistration(ctx, tenantID, branchID, queueDateStr, req.PatientID, req.PatientName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, exception.ErrConflict
	}

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

	res := mapQueueResponse(q)
	return &res, nil
}

func resolveTicketPrefix(ctx context.Context, resolver SettingsResolver, branchID, serviceID string) string {
	if resolver == nil {
		return "A"
	}
	if resolved, err := resolver.Resolve(ctx, "ticket_prefix", branchID, serviceID, ""); err == nil && resolved != "" {
		return resolved
	}
	if resolved, err := resolver.Resolve(ctx, "prefix", branchID, serviceID, ""); err == nil && resolved != "" {
		return resolved
	}
	return "A"
}

func resolveNumberingStrategy(ctx context.Context, resolver SettingsResolver, branchID, serviceID string) string {
	if resolver == nil {
		return "sequential"
	}
	if resolved, err := resolver.Resolve(ctx, "numbering_strategy", branchID, serviceID, ""); err == nil && resolved != "" {
		if resolved == "sequential" {
			return resolved
		}
		return "sequential"
	}
	if resolved, err := resolver.Resolve(ctx, "numbering", branchID, serviceID, ""); err == nil && resolved != "" {
		if resolved == "sequential" {
			return resolved
		}
	}
	return "sequential"
}

func computeBusinessQueueDate(now time.Time, resetTime string) string {
	parts := strings.Split(resetTime, ":")
	if len(parts) != 2 {
		return now.Format("2006-01-02")
	}
	hour, errHour := strconv.Atoi(parts[0])
	minute, errMinute := strconv.Atoi(parts[1])
	if errHour != nil || errMinute != nil {
		return now.Format("2006-01-02")
	}
	resetAt := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if now.Before(resetAt) {
		return now.AddDate(0, 0, -1).Format("2006-01-02")
	}
	return now.Format("2006-01-02")
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

	if u.validator != nil {
		if err := u.validator.Validate(ctx, tenantID, queue.BranchID, req.DestinationServiceID, req.DestinationCounterID); err != nil {
			return nil, err
		}
	}

	currentJourney, err := u.repo.FindCurrentJourney(ctx, queue.ID, queue.CurrentJourneyID)
	if err != nil || currentJourney == nil {
		return nil, exception.ErrNotFound
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

	res := mapQueueResponse(queue)
	return &res, nil
}

func (u *queueUseCase) TransitionQueue(ctx context.Context, queueID string, req *model.QueueTransitionRequest) (*model.QueueResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || queueID == "" || req == nil {
		return nil, exception.ErrBadRequest
	}
	if strings.TrimSpace(req.Action) == "" {
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

	nowMs := time.Now().UnixMilli()
	visit := &entity.VisitJourney{ID: uuid.New().String(), QueueID: queue.ID, TenantID: tenantID, CreatedAt: nowMs}

	switch req.Action {
	case model.QueueActionCall:
		if queue.Status != entity.QueueStatusWaiting && queue.Status != entity.QueueStatusSkipped {
			return nil, exception.ErrBadRequest
		}
		queue.Status = entity.QueueStatusCalling
		currentJourney.Status = entity.JourneyStatusCalling
		visit.EventType = "call"
	case model.QueueActionServe:
		if queue.Status != entity.QueueStatusCalling {
			return nil, exception.ErrBadRequest
		}
		queue.Status = entity.QueueStatusServing
		currentJourney.Status = entity.JourneyStatusServing
		visit.EventType = "serve"
	case model.QueueActionComplete:
		if queue.Status != entity.QueueStatusServing {
			return nil, exception.ErrBadRequest
		}
		queue.Status = entity.QueueStatusCompleted
		currentJourney.Status = entity.JourneyStatusCompleted
		visit.EventType = "complete"
	case model.QueueActionSkip:
		if queue.Status != entity.QueueStatusWaiting && queue.Status != entity.QueueStatusCalling {
			return nil, exception.ErrBadRequest
		}
		queue.Status = entity.QueueStatusSkipped
		currentJourney.Status = entity.JourneyStatusSkipped
		visit.EventType = "skip"
	case model.QueueActionCancel:
		if queue.Status == entity.QueueStatusCompleted || queue.Status == entity.QueueStatusCanceled {
			return nil, exception.ErrBadRequest
		}
		queue.Status = entity.QueueStatusCanceled
		currentJourney.Status = entity.JourneyStatusCanceled
		visit.EventType = "cancel"
	default:
		return nil, exception.ErrBadRequest
	}

	queue.UpdatedAt = nowMs
	currentJourney.UpdatedAt = nowMs

	if err := u.repo.UpdateQueueState(ctx, queue, currentJourney, visit); err != nil {
		return nil, err
	}

	res := mapQueueResponse(queue)
	return &res, nil
}

func mapQueueResponse(queue *entity.Queue) model.QueueResponse {
	return model.QueueResponse{
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
	}
}
