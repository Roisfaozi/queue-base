package queue

import (
	"context"
	"testing"

	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	counterMocks "github.com/Roisfaozi/queue-base/internal/modules/counter/test/mocks"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	orgMocks "github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	queueMocks "github.com/Roisfaozi/queue-base/internal/modules/queue/test/mocks"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	serviceMocks "github.com/Roisfaozi/queue-base/internal/modules/service/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDefaultRelationValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		tenantID  string
		branchID  string
		serviceID string
		counterID string
		setup     func(t *testing.T) *defaultRelationValidator
		wantErr   error
	}{
		{
			name:      "Positive_AllowsPharmacyFlowWhenEnabled",
			category:  "positive",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "svc-1",
			counterID: "c-1",
			setup: func(t *testing.T) *defaultRelationValidator {
				branchRepo := orgMocks.NewMockBranchRepository(t)
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(&branchEntity.Branch{ID: "b-1", TenantID: "t-1"}, nil).Once()

				serviceRepo := serviceMocks.NewMockServiceRepository(t)
				serviceRepo.On("FindByID", mock.Anything, "t-1", "svc-1").Return(&serviceEntity.Service{ID: "svc-1", TenantID: "t-1", IsPharmacy: true}, nil).Once()

				bsRepo := serviceMocks.NewMockBranchServiceRepository(t)
				bsRepo.On("FindByService", mock.Anything, "t-1", "b-1", "svc-1").Return(&serviceEntity.BranchService{ServiceID: "svc-1", BranchID: "b-1", TenantID: "t-1", IsActive: true}, nil).Once()

				counterRepo := counterMocks.NewMockCounterRepository(t)
				counterRepo.On("FindByID", mock.Anything, "t-1", "c-1").Return(&counterEntity.Counter{ID: "c-1", TenantID: "t-1", BranchID: "b-1"}, nil).Once()

				settings := queueMocks.NewMockSettingsResolver(t)
				settings.On("Resolve", mock.Anything, queueConfigKeyPharmacyFlowEnabled, "b-1", "svc-1", "c-1").Return("true", nil).Once()
				settings.On("Resolve", mock.Anything, queueConfigKeyRequireCounterForService, "b-1", "svc-1", "c-1").Return("true", nil).Once()

				return &defaultRelationValidator{
					branchRepo:        branchRepo,
					serviceRepo:       serviceRepo,
					branchServiceRepo: bsRepo,
					counterRepo:       counterRepo,
					settings:          settings,
				}
			},
		},
		{
			name:      "Negative_RejectsPharmacyFlowWhenDisabled",
			category:  "negative",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "svc-1",
			counterID: "",
			setup: func(t *testing.T) *defaultRelationValidator {
				branchRepo := orgMocks.NewMockBranchRepository(t)
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(&branchEntity.Branch{ID: "b-1", TenantID: "t-1"}, nil).Once()

				serviceRepo := serviceMocks.NewMockServiceRepository(t)
				serviceRepo.On("FindByID", mock.Anything, "t-1", "svc-1").Return(&serviceEntity.Service{ID: "svc-1", TenantID: "t-1", IsPharmacy: true}, nil).Once()

				bsRepo := serviceMocks.NewMockBranchServiceRepository(t)
				bsRepo.On("FindByService", mock.Anything, "t-1", "b-1", "svc-1").Return(&serviceEntity.BranchService{ServiceID: "svc-1", BranchID: "b-1", TenantID: "t-1", IsActive: true}, nil).Once()

				counterRepo := counterMocks.NewMockCounterRepository(t)

				settings := queueMocks.NewMockSettingsResolver(t)
				settings.On("Resolve", mock.Anything, queueConfigKeyPharmacyFlowEnabled, "b-1", "svc-1", "").Return("false", nil).Once()
				settings.On("Resolve", mock.Anything, queueConfigKeyRequireCounterForService, "b-1", "svc-1", "").Return("false", nil).Maybe()

				return &defaultRelationValidator{
					branchRepo:        branchRepo,
					serviceRepo:       serviceRepo,
					branchServiceRepo: bsRepo,
					counterRepo:       counterRepo,
					settings:          settings,
				}
			},
			wantErr: exception.ErrForbidden,
		},
		{
			name:      "Edge_RejectsRequiredCounterWhenMissing",
			category:  "edge",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "svc-1",
			counterID: "",
			setup: func(t *testing.T) *defaultRelationValidator {
				branchRepo := orgMocks.NewMockBranchRepository(t)
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(&branchEntity.Branch{ID: "b-1", TenantID: "t-1"}, nil).Once()

				serviceRepo := serviceMocks.NewMockServiceRepository(t)
				serviceRepo.On("FindByID", mock.Anything, "t-1", "svc-1").Return(&serviceEntity.Service{ID: "svc-1", TenantID: "t-1"}, nil).Once()

				bsRepo := serviceMocks.NewMockBranchServiceRepository(t)
				bsRepo.On("FindByService", mock.Anything, "t-1", "b-1", "svc-1").Return(&serviceEntity.BranchService{ServiceID: "svc-1", BranchID: "b-1", TenantID: "t-1", IsActive: true}, nil).Once()

				counterRepo := counterMocks.NewMockCounterRepository(t)

				settings := queueMocks.NewMockSettingsResolver(t)
				settings.On("Resolve", mock.Anything, queueConfigKeyPharmacyFlowEnabled, "b-1", "svc-1", "").Return("false", nil).Maybe()
				settings.On("Resolve", mock.Anything, queueConfigKeyRequireCounterForService, "b-1", "svc-1", "").Return("true", nil).Once()

				return &defaultRelationValidator{
					branchRepo:        branchRepo,
					serviceRepo:       serviceRepo,
					branchServiceRepo: bsRepo,
					counterRepo:       counterRepo,
					settings:          settings,
				}
			},
			wantErr: exception.ErrForbidden,
		},
		{
			name:      "Security_RejectsForeignCounterBranch",
			category:  "vulnerability",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "svc-1",
			counterID: "c-1",
			setup: func(t *testing.T) *defaultRelationValidator {
				branchRepo := orgMocks.NewMockBranchRepository(t)
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(&branchEntity.Branch{ID: "b-1", TenantID: "t-1"}, nil).Once()

				serviceRepo := serviceMocks.NewMockServiceRepository(t)
				serviceRepo.On("FindByID", mock.Anything, "t-1", "svc-1").Return(&serviceEntity.Service{ID: "svc-1", TenantID: "t-1"}, nil).Once()

				bsRepo := serviceMocks.NewMockBranchServiceRepository(t)
				bsRepo.On("FindByService", mock.Anything, "t-1", "b-1", "svc-1").Return(&serviceEntity.BranchService{ServiceID: "svc-1", BranchID: "b-1", TenantID: "t-1", IsActive: true}, nil).Once()

				counterRepo := counterMocks.NewMockCounterRepository(t)
				counterRepo.On("FindByID", mock.Anything, "t-1", "c-1").Return(&counterEntity.Counter{ID: "c-1", TenantID: "t-1", BranchID: "b-foreign"}, nil).Once()

				return &defaultRelationValidator{
					branchRepo:        branchRepo,
					serviceRepo:       serviceRepo,
					branchServiceRepo: bsRepo,
					counterRepo:       counterRepo,
				}
			},
			wantErr: exception.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.setup(t)
			err := v.Validate(context.Background(), tt.tenantID, tt.branchID, tt.serviceID, tt.counterID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
