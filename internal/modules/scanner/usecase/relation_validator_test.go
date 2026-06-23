package usecase

import (
	"context"
	"testing"

	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
)

type stubBranchRepo struct{ err error }

func (s stubBranchRepo) Create(ctx context.Context, branch *branchEntity.Branch) error { return nil }
func (s stubBranchRepo) FindByID(ctx context.Context, tenantID, branchID string) (*branchEntity.Branch, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &branchEntity.Branch{ID: branchID, TenantID: tenantID}, nil
}
func (s stubBranchRepo) FindAll(ctx context.Context, tenantID string) ([]*branchEntity.Branch, error) {
	return nil, nil
}
func (s stubBranchRepo) Update(ctx context.Context, branch *branchEntity.Branch) error { return nil }
func (s stubBranchRepo) Delete(ctx context.Context, tenantID, branchID string) error   { return nil }

type stubServiceRepo struct{ err error }

func (s stubServiceRepo) Create(ctx context.Context, service *serviceEntity.Service) error {
	return nil
}
func (s stubServiceRepo) FindByID(ctx context.Context, tenantID, serviceID string) (*serviceEntity.Service, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &serviceEntity.Service{ID: serviceID, TenantID: tenantID}, nil
}
func (s stubServiceRepo) FindAll(ctx context.Context, tenantID string) ([]*serviceEntity.Service, error) {
	return nil, nil
}
func (s stubServiceRepo) Update(ctx context.Context, service *serviceEntity.Service) error {
	return nil
}
func (s stubServiceRepo) Delete(ctx context.Context, tenantID, serviceID string) error { return nil }

type stubCounterRepo struct {
	err      error
	branchID string
}

type stubSettingsResolver struct {
	value string
	err   error
}

func (s stubSettingsResolver) Resolve(ctx context.Context, key string, branchID string, serviceID string, counterID string) (string, error) {
	return s.value, s.err
}

func (s stubCounterRepo) Create(ctx context.Context, counter *counterEntity.Counter) error {
	return nil
}
func (s stubCounterRepo) FindByID(ctx context.Context, tenantID, counterID string) (*counterEntity.Counter, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &counterEntity.Counter{ID: counterID, TenantID: tenantID, BranchID: s.branchID}, nil
}
func (s stubCounterRepo) FindAll(ctx context.Context, tenantID string) ([]*counterEntity.Counter, error) {
	return nil, nil
}
func (s stubCounterRepo) Update(ctx context.Context, counter *counterEntity.Counter) error {
	return nil
}
func (s stubCounterRepo) Delete(ctx context.Context, tenantID, counterID string) error { return nil }

func TestRelationValidator_ValidateSuccess(t *testing.T) {
	validator := NewRelationValidator(stubBranchRepo{}, stubServiceRepo{}, stubCounterRepo{branchID: "b-1"}, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	err := validator.Validate(ctx, "t-1", "b-1", "s-1", "c-1")
	assert.NoError(t, err)
}

func TestRelationValidator_ValidateNegativeMissingBranch(t *testing.T) {
	validator := NewRelationValidator(stubBranchRepo{err: exception.ErrNotFound}, stubServiceRepo{}, stubCounterRepo{branchID: "b-1"}, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	err := validator.Validate(ctx, "t-1", "b-1", "s-1", "c-1")
	assert.ErrorIs(t, err, exception.ErrForbidden)
}

func TestRelationValidator_ValidateEdgeNoCounter(t *testing.T) {
	validator := NewRelationValidator(stubBranchRepo{}, stubServiceRepo{}, stubCounterRepo{branchID: "b-1"}, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	err := validator.Validate(ctx, "t-1", "b-1", "s-1", "")
	assert.NoError(t, err)
}

func TestRelationValidator_ValidateSecurityCrossBranchCounter(t *testing.T) {
	validator := NewRelationValidator(stubBranchRepo{}, stubServiceRepo{}, stubCounterRepo{branchID: "other-branch"}, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	err := validator.Validate(ctx, "t-1", "b-1", "s-1", "c-1")
	assert.ErrorIs(t, err, exception.ErrForbidden)
}

func TestRelationValidator_ValidateNegativeRequireCounterSetting(t *testing.T) {
	validator := NewRelationValidator(stubBranchRepo{}, stubServiceRepo{}, stubCounterRepo{branchID: "b-1"}, stubSettingsResolver{value: "true"})
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	err := validator.Validate(ctx, "t-1", "b-1", "s-1", "")
	assert.ErrorIs(t, err, exception.ErrForbidden)
}
