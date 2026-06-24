package queue

import (
	"context"
	"testing"

	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	settingsModel "github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
)

type stubBranchRepo struct {
	branch *branchEntity.Branch
	err    error
}

func (s *stubBranchRepo) Create(ctx context.Context, branch *branchEntity.Branch) error { return nil }
func (s *stubBranchRepo) FindByID(ctx context.Context, tenantID, branchID string) (*branchEntity.Branch, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.branch, nil
}
func (s *stubBranchRepo) FindAll(ctx context.Context, tenantID string) ([]*branchEntity.Branch, error) {
	return nil, nil
}
func (s *stubBranchRepo) Update(ctx context.Context, branch *branchEntity.Branch) error { return nil }
func (s *stubBranchRepo) Delete(ctx context.Context, tenantID, branchID string) error   { return nil }

type stubServiceRepo struct {
	service *serviceEntity.Service
	err     error
}

func (s *stubServiceRepo) Create(ctx context.Context, service *serviceEntity.Service) error {
	return nil
}
func (s *stubServiceRepo) FindByID(ctx context.Context, tenantID, serviceID string) (*serviceEntity.Service, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.service, nil
}
func (s *stubServiceRepo) FindAll(ctx context.Context, tenantID string) ([]*serviceEntity.Service, error) {
	return nil, nil
}
func (s *stubServiceRepo) Update(ctx context.Context, service *serviceEntity.Service) error {
	return nil
}
func (s *stubServiceRepo) Delete(ctx context.Context, tenantID, serviceID string) error { return nil }

type stubCounterRepo struct {
	counter *counterEntity.Counter
	err     error
}

func (s *stubCounterRepo) Create(ctx context.Context, counter *counterEntity.Counter) error {
	return nil
}
func (s *stubCounterRepo) FindByID(ctx context.Context, tenantID, counterID string) (*counterEntity.Counter, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.counter, nil
}
func (s *stubCounterRepo) FindAll(ctx context.Context, tenantID string) ([]*counterEntity.Counter, error) {
	return nil, nil
}
func (s *stubCounterRepo) Update(ctx context.Context, counter *counterEntity.Counter) error {
	return nil
}
func (s *stubCounterRepo) Delete(ctx context.Context, tenantID, counterID string) error { return nil }

type stubQueueSettingsResolver struct {
	values map[string]string
}

func (s *stubQueueSettingsResolver) Resolve(ctx context.Context, key string, branchID string, serviceID string, counterID string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", exception.ErrNotFound
}

func TestDefaultRelationValidator_PositiveAllowsPharmacyFlowWhenEnabled(t *testing.T) {
	v := &defaultRelationValidator{
		branchRepo:  &stubBranchRepo{branch: &branchEntity.Branch{ID: "b-1", TenantID: "t-1"}},
		serviceRepo: &stubServiceRepo{service: &serviceEntity.Service{ID: "svc-1", TenantID: "t-1", IsPharmacy: true}},
		counterRepo: &stubCounterRepo{counter: &counterEntity.Counter{ID: "c-1", TenantID: "t-1", BranchID: "b-1"}},
		settings: &stubQueueSettingsResolver{values: map[string]string{
			settingsModel.SettingKeyPharmacyFlowEnabled:      "true",
			settingsModel.SettingKeyRequireCounterForService: "true",
		}},
	}

	err := v.Validate(context.Background(), "t-1", "b-1", "svc-1", "c-1")
	assert.NoError(t, err)
}

func TestDefaultRelationValidator_NegativeRejectsPharmacyFlowWhenDisabled(t *testing.T) {
	v := &defaultRelationValidator{
		branchRepo:  &stubBranchRepo{branch: &branchEntity.Branch{ID: "b-1", TenantID: "t-1"}},
		serviceRepo: &stubServiceRepo{service: &serviceEntity.Service{ID: "svc-1", TenantID: "t-1", IsPharmacy: true}},
		counterRepo: &stubCounterRepo{},
		settings: &stubQueueSettingsResolver{values: map[string]string{
			settingsModel.SettingKeyPharmacyFlowEnabled: "false",
		}},
	}

	err := v.Validate(context.Background(), "t-1", "b-1", "svc-1", "")
	assert.ErrorIs(t, err, exception.ErrForbidden)
}

func TestDefaultRelationValidator_EdgeRejectsRequiredCounterWhenMissing(t *testing.T) {
	v := &defaultRelationValidator{
		branchRepo:  &stubBranchRepo{branch: &branchEntity.Branch{ID: "b-1", TenantID: "t-1"}},
		serviceRepo: &stubServiceRepo{service: &serviceEntity.Service{ID: "svc-1", TenantID: "t-1"}},
		counterRepo: &stubCounterRepo{},
		settings: &stubQueueSettingsResolver{values: map[string]string{
			settingsModel.SettingKeyRequireCounterForService: "true",
		}},
	}

	err := v.Validate(context.Background(), "t-1", "b-1", "svc-1", "")
	assert.ErrorIs(t, err, exception.ErrForbidden)
}

func TestDefaultRelationValidator_SecurityRejectsForeignCounterBranch(t *testing.T) {
	v := &defaultRelationValidator{
		branchRepo:  &stubBranchRepo{branch: &branchEntity.Branch{ID: "b-1", TenantID: "t-1"}},
		serviceRepo: &stubServiceRepo{service: &serviceEntity.Service{ID: "svc-1", TenantID: "t-1"}},
		counterRepo: &stubCounterRepo{counter: &counterEntity.Counter{ID: "c-1", TenantID: "t-1", BranchID: "b-foreign"}},
	}

	err := v.Validate(context.Background(), "t-1", "b-1", "svc-1", "c-1")
	assert.ErrorIs(t, err, exception.ErrForbidden)
}
