package usecase

import (
	"context"

	counterRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/counter/repository"
	branchRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	serviceRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/service/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
)

type relationValidator struct {
	branchRepo  branchRepository.BranchRepository
	serviceRepo serviceRepository.ServiceRepository
	counterRepo counterRepository.CounterRepository
}

func NewRelationValidator(branchRepo branchRepository.BranchRepository, serviceRepo serviceRepository.ServiceRepository, counterRepo counterRepository.CounterRepository) RelationValidator {
	return &relationValidator{branchRepo: branchRepo, serviceRepo: serviceRepo, counterRepo: counterRepo}
}

func (v *relationValidator) Validate(ctx context.Context, tenantID, branchID, serviceID, counterID string) error {
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
