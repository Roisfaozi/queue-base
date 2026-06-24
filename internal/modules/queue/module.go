package queue

import (
	"context"

	counterRepo "github.com/Roisfaozi/queue-base/internal/modules/counter/repository"
	branchRepo "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	queueHttp "github.com/Roisfaozi/queue-base/internal/modules/queue/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/usecase"
	serviceRepo "github.com/Roisfaozi/queue-base/internal/modules/service/repository"
	settingsModel "github.com/Roisfaozi/queue-base/internal/modules/settings/model"
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
	settings    usecase.SettingsResolver
}

func NewDefaultRelationValidator(db *gorm.DB, settingsResolver usecase.SettingsResolver) usecase.RelationValidator {
	return &defaultRelationValidator{
		branchRepo:  branchRepo.NewBranchRepository(db),
		serviceRepo: serviceRepo.NewServiceRepository(db),
		counterRepo: counterRepo.NewCounterRepository(db),
		settings:    settingsResolver,
	}
}

func (v *defaultRelationValidator) Validate(ctx context.Context, tenantID, branchID, serviceID, counterID string) error {
	if _, err := v.branchRepo.FindByID(ctx, tenantID, branchID); err != nil {
		return exception.ErrForbidden
	}
	if serviceID != "" {
		service, err := v.serviceRepo.FindByID(ctx, tenantID, serviceID)
		if err != nil {
			return exception.ErrForbidden
		}
		requireCounter := service.IsPharmacy
		pharmacyFlowEnabled := service.IsPharmacy
		if v.settings != nil {
			if value, resolveErr := v.settings.Resolve(ctx, settingsModel.SettingKeyPharmacyFlowEnabled, branchID, serviceID, counterID); resolveErr == nil {
				pharmacyFlowEnabled = value == "true"
			}
			if value, resolveErr := v.settings.Resolve(ctx, settingsModel.SettingKeyRequireCounterForService, branchID, serviceID, counterID); resolveErr == nil {
				requireCounter = value == "true"
			}
		}
		if service.IsPharmacy && !pharmacyFlowEnabled {
			return exception.ErrForbidden
		}
		if requireCounter && counterID == "" {
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
	relationValidator := NewDefaultRelationValidator(db, settingsResolver)
	uc := usecase.NewQueueUseCase(repo, settingsResolver, relationValidator)
	ctrl := queueHttp.NewQueueController(uc, validate)
	return &QueueModule{QueueController: ctrl, QueueRepo: repo, QueueUseCase: uc}
}
