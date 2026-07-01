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

type stubServiceRepo struct {
	err     error
	service *serviceEntity.Service
}

func (s stubServiceRepo) Create(ctx context.Context, service *serviceEntity.Service) error {
	return nil
}
func (s stubServiceRepo) FindByID(ctx context.Context, tenantID, serviceID string) (*serviceEntity.Service, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.service != nil {
		return s.service, nil
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

func TestRelationValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		tenantID  string
		branchID  string
		serviceID string
		counterID string
		validator RelationValidator
		wantErr   error
	}{
		{
			name:      "Positive_ValidateSuccess",
			category:  "positive",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			validator: NewRelationValidator(stubBranchRepo{}, stubServiceRepo{}, stubCounterRepo{branchID: "b-1"}, nil),
			wantErr:   nil,
		},
		{
			name:      "Negative_ValidateMissingBranch",
			category:  "negative",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			validator: NewRelationValidator(stubBranchRepo{err: exception.ErrNotFound}, stubServiceRepo{}, stubCounterRepo{branchID: "b-1"}, nil),
			wantErr:   exception.ErrForbidden,
		},
		{
			name:      "Security_ValidateServiceLookupFailureMapsToForbidden",
			category:  "security",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			validator: NewRelationValidator(stubBranchRepo{}, stubServiceRepo{err: exception.ErrNotFound}, stubCounterRepo{branchID: "b-1"}, nil),
			wantErr:   exception.ErrForbidden,
		},
		{
			name:      "Security_ValidateCounterLookupFailureMapsToForbidden",
			category:  "security",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			validator: NewRelationValidator(stubBranchRepo{}, stubServiceRepo{}, stubCounterRepo{err: exception.ErrNotFound}, nil),
			wantErr:   exception.ErrForbidden,
		},
		{
			name:      "Edge_ValidateNoCounter",
			category:  "edge",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "",
			validator: NewRelationValidator(stubBranchRepo{}, stubServiceRepo{}, stubCounterRepo{branchID: "b-1"}, nil),
			wantErr:   nil,
		},
		{
			name:      "Security_ValidateCrossBranchCounter",
			category:  "vulnerability",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			validator: NewRelationValidator(stubBranchRepo{}, stubServiceRepo{}, stubCounterRepo{branchID: "other-branch"}, nil),
			wantErr:   exception.ErrForbidden,
		},
		{
			name:      "Negative_ValidateRequireCounterSetting",
			category:  "negative",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "",
			validator: NewRelationValidator(stubBranchRepo{}, stubServiceRepo{}, stubCounterRepo{branchID: "b-1"}, stubSettingsResolver{value: "true"}),
			wantErr:   exception.ErrForbidden,
		},
		{
			name:      "Negative_ValidatePharmacyFlowDisabledForForward",
			category:  "negative",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			validator: NewRelationValidator(stubBranchRepo{}, stubServiceRepo{service: &serviceEntity.Service{ID: "s-1", TenantID: "t-1", IsPharmacy: true}}, stubCounterRepo{branchID: "b-1"}, stubSettingsResolver{value: "false"}),
			wantErr:   exception.ErrForbidden,
		},
		{
			name:      "Positive_ValidatePharmacyFlowEnabledAllowsForward",
			category:  "positive",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			validator: NewRelationValidator(stubBranchRepo{}, stubServiceRepo{service: &serviceEntity.Service{ID: "s-1", TenantID: "t-1", IsPharmacy: true}}, stubCounterRepo{branchID: "b-1"}, stubSettingsResolver{value: "true"}),
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := database.SetOrganizationContext(context.Background(), tt.tenantID)
			err := tt.validator.Validate(ctx, tt.tenantID, tt.branchID, tt.serviceID, tt.counterID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
