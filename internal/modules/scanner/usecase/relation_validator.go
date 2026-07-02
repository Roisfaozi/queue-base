package usecase

import (
	"context"
	"fmt"

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
	branchRepo        branchRepository.BranchRepository
	serviceRepo       serviceRepository.ServiceRepository
	branchServiceRepo serviceRepository.BranchServiceRepository
	counterRepo       counterRepository.CounterRepository
	settings          settingsResolver
}

func NewRelationValidator(branchRepo branchRepository.BranchRepository, serviceRepo serviceRepository.ServiceRepository, branchServiceRepo serviceRepository.BranchServiceRepository, counterRepo counterRepository.CounterRepository, settingsResolver settingsResolver) RelationValidator {
	return &relationValidator{branchRepo: branchRepo, serviceRepo: serviceRepo, branchServiceRepo: branchServiceRepo, counterRepo: counterRepo, settings: settingsResolver}
}

func (v *relationValidator) Validate(ctx context.Context, tenantID, branchID, serviceID, counterID string) error {
	if _, err := v.branchRepo.FindByID(ctx, tenantID, branchID); err != nil {
		return fmt.Errorf("branchRepo.FindByID failed (%v): %w", err, exception.ErrForbidden)
	}
	if serviceID != "" {
		service, err := v.serviceRepo.FindByID(ctx, tenantID, serviceID)
		if err != nil {
			return fmt.Errorf("serviceRepo.FindByID failed (%v): %w", err, exception.ErrForbidden)
		}
		if v.branchServiceRepo != nil {
			if _, err := v.branchServiceRepo.FindByService(ctx, tenantID, branchID, serviceID); err != nil {
				return fmt.Errorf("branchServiceRepo.FindByService failed (%v): %w", err, exception.ErrForbidden)
			}
		} else {
			return fmt.Errorf("branchServiceRepo missing: %w", exception.ErrForbidden)
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
			return fmt.Errorf("pharmacy flow disabled: %w", exception.ErrForbidden)
		}
		if requireCounter && counterID == "" {
			return fmt.Errorf("counter required: %w", exception.ErrForbidden)
		}
	}
	if counterID != "" {
		counter, err := v.counterRepo.FindByID(ctx, tenantID, counterID)
		if err != nil {
			return fmt.Errorf("counterRepo.FindByID failed (%v): %w", err, exception.ErrForbidden)
		}
		if counter.BranchID != branchID {
			return fmt.Errorf("counter branch mismatch: %w", exception.ErrForbidden)
		}
	}
	return nil
}
