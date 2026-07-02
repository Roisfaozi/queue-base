package usecase

import (
	"context"

	"github.com/Roisfaozi/queue-base/internal/modules/caller/model"
	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	queueUsecase "github.com/Roisfaozi/queue-base/internal/modules/queue/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"gorm.io/gorm"
)

type CallerUseCase interface {
	ExecuteAction(ctx context.Context, journeyID, action string) (*model.CallerActionResponse, error)
}

type callerUseCase struct {
	db      *gorm.DB
	queueUC queueUsecase.QueueUseCase
}

func NewCallerUseCase(db *gorm.DB, qu queueUsecase.QueueUseCase) CallerUseCase {
	return &callerUseCase{db: db, queueUC: qu}
}

func (u *callerUseCase) ExecuteAction(ctx context.Context, journeyID, action string) (*model.CallerActionResponse, error) {
	type journeyRow struct {
		QueueID  string
		TenantID string
		BranchID string
	}
	var jr journeyRow
	if err := u.db.WithContext(ctx).Table("queue_journeys").
		Select("queue_id, tenant_id, branch_id").
		Where("id = ?", journeyID).
		First(&jr).Error; err != nil {
		return nil, exception.ErrNotFound
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
