package usecase_test

import (
	"context"
	"testing"

	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	counterMocks "github.com/Roisfaozi/queue-base/internal/modules/counter/test/mocks"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	orgMocks "github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	scannerMocks "github.com/Roisfaozi/queue-base/internal/modules/scanner/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/scanner/usecase"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	serviceMocks "github.com/Roisfaozi/queue-base/internal/modules/service/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRelationValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		tenantID  string
		branchID  string
		serviceID string
		counterID string
		setup     func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, bsRepo *serviceMocks.MockBranchServiceRepository, counterRepo *counterMocks.MockCounterRepository, settings *scannerMocks.MockSettingsResolver)
		wantErr   error
	}{
		{
			name:      "Positive_ValidateSuccess",
			category:  "positive",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, bsRepo *serviceMocks.MockBranchServiceRepository, counterRepo *counterMocks.MockCounterRepository, settings *scannerMocks.MockSettingsResolver) {
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(&branchEntity.Branch{ID: "b-1", TenantID: "t-1", Status: branchEntity.BranchStatusActive}, nil).Once()
				serviceRepo.On("FindByID", mock.Anything, "t-1", "s-1").Return(&serviceEntity.Service{ID: "s-1", TenantID: "t-1", Status: serviceEntity.ServiceStatusActive}, nil).Once()
				bsRepo.On("FindByService", mock.Anything, "t-1", "b-1", "s-1").Return(&serviceEntity.BranchService{ID: "bs-1", TenantID: "t-1", BranchID: "b-1", ServiceID: "s-1", IsActive: true}, nil).Once()
				counterRepo.On("FindByID", mock.Anything, "t-1", "c-1").Return(&counterEntity.Counter{ID: "c-1", TenantID: "t-1", BranchID: "b-1", Status: counterEntity.CounterStatusActive}, nil).Once()
			},
			wantErr: nil,
		},
		{
			name:      "Negative_ValidateMissingBranch",
			category:  "negative",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, bsRepo *serviceMocks.MockBranchServiceRepository, counterRepo *counterMocks.MockCounterRepository, settings *scannerMocks.MockSettingsResolver) {
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(nil, exception.ErrNotFound).Once()
			},
			wantErr: exception.ErrForbidden,
		},
		{
			name:      "Negative_RejectsInactiveBranch",
			category:  "negative",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, bsRepo *serviceMocks.MockBranchServiceRepository, counterRepo *counterMocks.MockCounterRepository, settings *scannerMocks.MockSettingsResolver) {
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(&branchEntity.Branch{ID: "b-1", TenantID: "t-1", Status: branchEntity.BranchStatusInactive}, nil).Once()
			},
			wantErr: exception.ErrForbidden,
		},
		{
			name:      "Negative_RejectsInactiveService",
			category:  "negative",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, bsRepo *serviceMocks.MockBranchServiceRepository, counterRepo *counterMocks.MockCounterRepository, settings *scannerMocks.MockSettingsResolver) {
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(&branchEntity.Branch{ID: "b-1", TenantID: "t-1", Status: branchEntity.BranchStatusActive}, nil).Once()
				serviceRepo.On("FindByID", mock.Anything, "t-1", "s-1").Return(&serviceEntity.Service{ID: "s-1", TenantID: "t-1", Status: serviceEntity.ServiceStatusInactive}, nil).Once()
			},
			wantErr: exception.ErrForbidden,
		},
		{
			name:      "Security_ValidateServiceLookupFailureMapsToForbidden",
			category:  "security",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, bsRepo *serviceMocks.MockBranchServiceRepository, counterRepo *counterMocks.MockCounterRepository, settings *scannerMocks.MockSettingsResolver) {
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(&branchEntity.Branch{ID: "b-1", TenantID: "t-1", Status: branchEntity.BranchStatusActive}, nil).Once()
				serviceRepo.On("FindByID", mock.Anything, "t-1", "s-1").Return(nil, exception.ErrNotFound).Once()
			},
			wantErr: exception.ErrForbidden,
		},
		{
			name:      "Security_ValidateCounterLookupFailureMapsToForbidden",
			category:  "security",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, bsRepo *serviceMocks.MockBranchServiceRepository, counterRepo *counterMocks.MockCounterRepository, settings *scannerMocks.MockSettingsResolver) {
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(&branchEntity.Branch{ID: "b-1", TenantID: "t-1", Status: branchEntity.BranchStatusActive}, nil).Once()
				serviceRepo.On("FindByID", mock.Anything, "t-1", "s-1").Return(&serviceEntity.Service{ID: "s-1", TenantID: "t-1", Status: serviceEntity.ServiceStatusActive}, nil).Once()
				bsRepo.On("FindByService", mock.Anything, "t-1", "b-1", "s-1").Return(&serviceEntity.BranchService{ID: "bs-1", TenantID: "t-1", BranchID: "b-1", ServiceID: "s-1", IsActive: true}, nil).Once()
				counterRepo.On("FindByID", mock.Anything, "t-1", "c-1").Return(nil, exception.ErrNotFound).Once()
			},
			wantErr: exception.ErrForbidden,
		},
		{
			name:      "Edge_ValidateNoCounter",
			category:  "edge",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, bsRepo *serviceMocks.MockBranchServiceRepository, counterRepo *counterMocks.MockCounterRepository, settings *scannerMocks.MockSettingsResolver) {
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(&branchEntity.Branch{ID: "b-1", TenantID: "t-1", Status: branchEntity.BranchStatusActive}, nil).Once()
				serviceRepo.On("FindByID", mock.Anything, "t-1", "s-1").Return(&serviceEntity.Service{ID: "s-1", TenantID: "t-1", Status: serviceEntity.ServiceStatusActive}, nil).Once()
				bsRepo.On("FindByService", mock.Anything, "t-1", "b-1", "s-1").Return(&serviceEntity.BranchService{ID: "bs-1", TenantID: "t-1", BranchID: "b-1", ServiceID: "s-1", IsActive: true}, nil).Once()
			},
			wantErr: nil,
		},
		{
			name:      "Security_ValidateCrossBranchCounter",
			category:  "vulnerability",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, bsRepo *serviceMocks.MockBranchServiceRepository, counterRepo *counterMocks.MockCounterRepository, settings *scannerMocks.MockSettingsResolver) {
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(&branchEntity.Branch{ID: "b-1", TenantID: "t-1", Status: branchEntity.BranchStatusActive}, nil).Once()
				serviceRepo.On("FindByID", mock.Anything, "t-1", "s-1").Return(&serviceEntity.Service{ID: "s-1", TenantID: "t-1", Status: serviceEntity.ServiceStatusActive}, nil).Once()
				bsRepo.On("FindByService", mock.Anything, "t-1", "b-1", "s-1").Return(&serviceEntity.BranchService{ID: "bs-1", TenantID: "t-1", BranchID: "b-1", ServiceID: "s-1", IsActive: true}, nil).Once()
				counterRepo.On("FindByID", mock.Anything, "t-1", "c-1").Return(&counterEntity.Counter{ID: "c-1", TenantID: "t-1", BranchID: "other-branch", Status: counterEntity.CounterStatusActive}, nil).Once()
			},
			wantErr: exception.ErrForbidden,
		},
		{
			name:      "Negative_ValidateRequireCounterSetting",
			category:  "negative",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, bsRepo *serviceMocks.MockBranchServiceRepository, counterRepo *counterMocks.MockCounterRepository, settings *scannerMocks.MockSettingsResolver) {
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(&branchEntity.Branch{ID: "b-1", TenantID: "t-1", Status: branchEntity.BranchStatusActive}, nil).Once()
				serviceRepo.On("FindByID", mock.Anything, "t-1", "s-1").Return(&serviceEntity.Service{ID: "s-1", TenantID: "t-1", Status: serviceEntity.ServiceStatusActive}, nil).Once()
				bsRepo.On("FindByService", mock.Anything, "t-1", "b-1", "s-1").Return(&serviceEntity.BranchService{ID: "bs-1", TenantID: "t-1", BranchID: "b-1", ServiceID: "s-1", IsActive: true}, nil).Once()
				settings.On("Resolve", mock.Anything, "pharmacy_flow_enabled", "b-1", "s-1", "").Return("true", nil).Maybe()
				settings.On("Resolve", mock.Anything, "require_counter_for_service", "b-1", "s-1", "").Return("true", nil).Once()
			},
			wantErr: exception.ErrForbidden,
		},
		{
			name:      "Negative_ValidatePharmacyFlowDisabledForForward",
			category:  "negative",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, bsRepo *serviceMocks.MockBranchServiceRepository, counterRepo *counterMocks.MockCounterRepository, settings *scannerMocks.MockSettingsResolver) {
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(&branchEntity.Branch{ID: "b-1", TenantID: "t-1", Status: branchEntity.BranchStatusActive}, nil).Once()
				serviceRepo.On("FindByID", mock.Anything, "t-1", "s-1").Return(&serviceEntity.Service{ID: "s-1", TenantID: "t-1", Status: serviceEntity.ServiceStatusActive, IsPharmacy: true}, nil).Once()
				bsRepo.On("FindByService", mock.Anything, "t-1", "b-1", "s-1").Return(&serviceEntity.BranchService{ID: "bs-1", TenantID: "t-1", BranchID: "b-1", ServiceID: "s-1", IsActive: true}, nil).Once()
				settings.On("Resolve", mock.Anything, "require_counter_for_service", "b-1", "s-1", "c-1").Return("false", nil).Maybe()
				settings.On("Resolve", mock.Anything, "pharmacy_flow_enabled", "b-1", "s-1", "c-1").Return("false", nil).Once()
			},
			wantErr: exception.ErrForbidden,
		},
		{
			name:      "Positive_ValidatePharmacyFlowEnabledAllowsForward",
			category:  "positive",
			tenantID:  "t-1",
			branchID:  "b-1",
			serviceID: "s-1",
			counterID: "c-1",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, bsRepo *serviceMocks.MockBranchServiceRepository, counterRepo *counterMocks.MockCounterRepository, settings *scannerMocks.MockSettingsResolver) {
				branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(&branchEntity.Branch{ID: "b-1", TenantID: "t-1", Status: branchEntity.BranchStatusActive}, nil).Once()
				serviceRepo.On("FindByID", mock.Anything, "t-1", "s-1").Return(&serviceEntity.Service{ID: "s-1", TenantID: "t-1", Status: serviceEntity.ServiceStatusActive, IsPharmacy: true}, nil).Once()
				bsRepo.On("FindByService", mock.Anything, "t-1", "b-1", "s-1").Return(&serviceEntity.BranchService{ID: "bs-1", TenantID: "t-1", BranchID: "b-1", ServiceID: "s-1", IsActive: true}, nil).Once()
				settings.On("Resolve", mock.Anything, "require_counter_for_service", "b-1", "s-1", "c-1").Return("false", nil).Maybe()
				settings.On("Resolve", mock.Anything, "pharmacy_flow_enabled", "b-1", "s-1", "c-1").Return("true", nil).Once()
				counterRepo.On("FindByID", mock.Anything, "t-1", "c-1").Return(&counterEntity.Counter{ID: "c-1", TenantID: "t-1", BranchID: "b-1", Status: counterEntity.CounterStatusActive}, nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branchRepo := orgMocks.NewMockBranchRepository(t)
			serviceRepo := serviceMocks.NewMockServiceRepository(t)
			bsRepo := serviceMocks.NewMockBranchServiceRepository(t)
			counterRepo := counterMocks.NewMockCounterRepository(t)
			settings := scannerMocks.NewMockSettingsResolver(t)

			var settingsResolver usecase.SettingsResolver
			if tt.name == "Negative_ValidateRequireCounterSetting" ||
				tt.name == "Negative_ValidatePharmacyFlowDisabledForForward" ||
				tt.name == "Positive_ValidatePharmacyFlowEnabledAllowsForward" {
				settingsResolver = settings
			}

			if tt.setup != nil {
				tt.setup(t, branchRepo, serviceRepo, bsRepo, counterRepo, settings)
			}

			v := usecase.NewRelationValidator(branchRepo, serviceRepo, bsRepo, counterRepo, settingsResolver)

			ctx := database.SetOrganizationContext(context.Background(), tt.tenantID)
			err := v.Validate(ctx, tt.tenantID, tt.branchID, tt.serviceID, tt.counterID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
