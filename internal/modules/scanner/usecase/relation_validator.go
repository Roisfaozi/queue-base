package usecase

import (
	"context"

	counterRepository "github.com/Roisfaozi/queue-base/internal/modules/counter/repository"
	branchRepository "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	serviceRepository "github.com/Roisfaozi/queue-base/internal/modules/service/repository"
	settingsModel "github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	"github.com/Roisfaozi/queue-base/pkg/exception"
)

type settingsResolver interface {
	Resolve(ctx context.Context, key string, branchID string, serviceID string, counterID string) (string, error)
}

type relationValidator struct {
	branchRepo  branchRepository.BranchRepository
	serviceRepo serviceRepository.ServiceRepository
	counterRepo counterRepository.CounterRepository
	settings    settingsResolver
}

func NewRelationValidator(branchRepo branchRepository.BranchRepository, serviceRepo serviceRepository.ServiceRepository, counterRepo counterRepository.CounterRepository, settingsResolver settingsResolver) RelationValidator {
	return &relationValidator{branchRepo: branchRepo, serviceRepo: serviceRepo, counterRepo: counterRepo, settings: settingsResolver}
}

func (v *relationValidator) Validate(ctx context.Context, tenantID, branchID, serviceID, counterID string) error {
	if _, err := v.branchRepo.FindByID(ctx, tenantID, branchID); err != nil {
		return exception.ErrForbidden
	}
	if serviceID != "" {
		service, err := v.serviceRepo.FindByID(ctx, tenantID, serviceID)
		if err != nil {
			return exception.ErrForbidden
		}
		requireCounter := service.IsPharmacy
		if v.settings != nil {
			if value, resolveErr := v.settings.Resolve(ctx, settingsModel.SettingKeyRequireCounterForService, branchID, serviceID, counterID); resolveErr == nil {
				requireCounter = value == "true"
			}
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
