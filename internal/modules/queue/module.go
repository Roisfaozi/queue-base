package queue

import (
	"context"

	counterRepo "github.com/Roisfaozi/queue-base/internal/modules/counter/repository"
	branchRepo "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	queueHttp "github.com/Roisfaozi/queue-base/internal/modules/queue/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/usecase"
	serviceRepo "github.com/Roisfaozi/queue-base/internal/modules/service/repository"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type QueueModule struct {
	QueueController *queueHttp.QueueController
	QueueRepo       repository.QueueRepository
	QueueUseCase    usecase.QueueUseCase
}

type defaultRelationValidator struct {
	branchRepo  branchRepo.BranchRepository
	serviceRepo serviceRepo.ServiceRepository
	counterRepo counterRepo.CounterRepository
}

func NewDefaultRelationValidator(db *gorm.DB) usecase.RelationValidator {
	return &defaultRelationValidator{
		branchRepo:  branchRepo.NewBranchRepository(db),
		serviceRepo: serviceRepo.NewServiceRepository(db),
		counterRepo: counterRepo.NewCounterRepository(db),
	}
}

func (v *defaultRelationValidator) Validate(ctx context.Context, tenantID, branchID, serviceID, counterID string) error {
	if _, err := v.branchRepo.FindByID(ctx, tenantID, branchID); err != nil {
		return exception.ErrForbidden
	}
	if serviceID != "" {
		if _, err := v.serviceRepo.FindByID(ctx, tenantID, serviceID); err != nil {
			return exception.ErrForbidden
		}
	}
	if counterID != "" {
		counter, err := v.counterRepo.FindByID(ctx, tenantID, counterID)
		if err != nil {
			return exception.ErrForbidden
		}
		if counter.BranchID != branchID {
			return exception.ErrForbidden
		}
	}
	return nil
}

func NewQueueModule(db *gorm.DB, validate *validator.Validate, settingsResolver usecase.SettingsResolver) *QueueModule {
	repo := repository.NewQueueRepository(db)
	relationValidator := NewDefaultRelationValidator(db)
	uc := usecase.NewQueueUseCase(repo, settingsResolver, relationValidator)
	ctrl := queueHttp.NewQueueController(uc, validate)
	return &QueueModule{QueueController: ctrl, QueueRepo: repo, QueueUseCase: uc}
}
