package usecase

import (
	"context"

	"github.com/Roisfaozi/queue-base/internal/modules/caller/model"
	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	queueUsecase "github.com/Roisfaozi/queue-base/internal/modules/queue/usecase"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"gorm.io/gorm"
)

type CallerUseCase interface {
	ExecuteAction(ctx context.Context, clientID, journeyID, action string) (*model.CallerActionResponse, error)
}

type callerUseCase struct {
	db      *gorm.DB
	queueUC queueUsecase.QueueUseCase
}

func NewCallerUseCase(db *gorm.DB, qu queueUsecase.QueueUseCase) CallerUseCase {
	return &callerUseCase{db: db, queueUC: qu}
}

func (u *callerUseCase) ExecuteAction(ctx context.Context, clientID, journeyID, action string) (*model.CallerActionResponse, error) {
	if clientID == "" {
		return nil, exception.ErrUnauthorized
	}
	type journeyRow struct {
		QueueID   string
		TenantID  string
		BranchID  string
		CounterID *string
	}
	var jr journeyRow
	if err := u.db.WithContext(ctx).Table("queue_journeys").
		Select("queue_id, tenant_id, branch_id, counter_id").
		Where("id = ?", journeyID).
		First(&jr).Error; err != nil {
		return nil, exception.ErrNotFound
	}

	type clientRow struct {
		TenantID        string
		BranchID        string
		BranchServiceID *string
		CounterID       *string
	}
	var client clientRow
	if err := u.db.WithContext(ctx).Table("qms_clients").
		Select("tenant_id, branch_id, branch_service_id, counter_id").
		Where("id = ? AND is_active = ?", clientID, true).
		First(&client).Error; err != nil {
		return nil, exception.ErrUnauthorized
	}

	if client.TenantID != jr.TenantID || client.BranchID != jr.BranchID {
		return nil, exception.ErrForbidden
	}
	if client.CounterID != nil && *client.CounterID != "" {
		if jr.CounterID == nil || *jr.CounterID == "" || *client.CounterID != *jr.CounterID {
			return nil, exception.ErrForbidden
		}
	}

	if userID, ok := authcontext.UserIDFromContext(ctx); ok && userID != "" && client.CounterID != nil && *client.CounterID != "" {
		var count int64
		if err := u.db.WithContext(ctx).Table("operator_counter_assignments").
			Where("tenant_id = ? AND branch_id = ? AND user_id = ? AND counter_id = ? AND unassigned_at IS NULL", jr.TenantID, jr.BranchID, userID, *client.CounterID).
			Count(&count).Error; err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, exception.ErrForbidden
		}
	}

	ctx = database.SetOrganizationContext(ctx, jr.TenantID)
	ctx = database.SetBranchContext(ctx, jr.BranchID)

	transitionReq := &queueModel.QueueTransitionRequest{Action: action}
	res, err := u.queueUC.TransitionQueue(ctx, jr.QueueID, transitionReq)
	if err != nil {
		return nil, err
	}
	return &model.CallerActionResponse{
		Success:   true,
		TrackNo:   res.TicketNo,
		QueueNo:   res.QueueNo,
		Status:    res.Status,
		JourneyID: journeyID,
	}, nil
}
