package usecase

import (
	"context"
	"fmt"

	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	counterRepository "github.com/Roisfaozi/queue-base/internal/modules/counter/repository"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	branchRepository "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	serviceRepository "github.com/Roisfaozi/queue-base/internal/modules/service/repository"
	"github.com/Roisfaozi/queue-base/pkg/exception"
)

const (
	queueConfigKeyPharmacyFlowEnabled      = "pharmacy_flow_enabled"
	queueConfigKeyRequireCounterForService = "require_counter_for_service"
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
	branch, err := v.branchRepo.FindByID(ctx, tenantID, branchID)
	if err != nil {
		return fmt.Errorf("branchRepo.FindByID failed (%v): %w", err, exception.ErrForbidden)
	}
	if branch.Status != branchEntity.BranchStatusActive {
		return fmt.Errorf("branch inactive: %w", exception.ErrForbidden)
	}
	if serviceID != "" {
		service, err := v.serviceRepo.FindByID(ctx, tenantID, serviceID)
		if err != nil {
			return fmt.Errorf("serviceRepo.FindByID failed (%v): %w", err, exception.ErrForbidden)
		}
		if service.Status != serviceEntity.ServiceStatusActive {
			return fmt.Errorf("service inactive: %w", exception.ErrForbidden)
		}
		if v.branchServiceRepo != nil {
			branchService, err := v.branchServiceRepo.FindByService(ctx, tenantID, branchID, serviceID)
			if err != nil {
				return fmt.Errorf("branchServiceRepo.FindByService failed (%v): %w", err, exception.ErrForbidden)
			}
			if !branchService.IsActive {
				return fmt.Errorf("branch service inactive: %w", exception.ErrForbidden)
			}
		} else {
			return fmt.Errorf("branchServiceRepo missing: %w", exception.ErrForbidden)
		}
		requireCounter := service.IsPharmacy
		pharmacyFlowEnabled := service.IsPharmacy
		if v.settings != nil {
			if value, resolveErr := v.settings.Resolve(ctx, queueConfigKeyPharmacyFlowEnabled, branchID, serviceID, counterID); resolveErr == nil {
				pharmacyFlowEnabled = value == "true"
			}
			if value, resolveErr := v.settings.Resolve(ctx, queueConfigKeyRequireCounterForService, branchID, serviceID, counterID); resolveErr == nil {
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
		if counter.Status != counterEntity.CounterStatusActive {
			return fmt.Errorf("counter inactive: %w", exception.ErrForbidden)
		}
	}
	return nil
}
