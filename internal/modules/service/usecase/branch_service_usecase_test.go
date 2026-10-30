package usecase_test

import (
	"context"
	"testing"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	orgMocks "github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/service/model"
	serviceMocks "github.com/Roisfaozi/queue-base/internal/modules/service/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/service/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBranchServiceDeactivationEdgeCase(t *testing.T) {
	branch := &branchEntity.Branch{ID: "b-1", TenantID: "t-1", Status: branchEntity.BranchStatusActive}
	service := &entity.Service{ID: "s-1", TenantID: "t-1", Status: entity.ServiceStatusActive}

	branchRepo := orgMocks.NewMockBranchRepository(t)
	serviceRepo := serviceMocks.NewMockServiceRepository(t)
	repo := serviceMocks.NewMockBranchServiceRepository(t)
	audit := serviceMocks.NewMockAuditLogger(t)

	branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(branch, nil).Once()
	serviceRepo.On("FindByID", mock.Anything, "t-1", "s-1").Return(service, nil).Once()

	var createdBS *entity.BranchService
	repo.On("Create", mock.Anything, mock.MatchedBy(func(bs *entity.BranchService) bool {
		createdBS = bs
		return bs.BranchID == "b-1" && bs.ServiceID == "s-1" && bs.IsActive
	})).Return(nil).Once()
	audit.On("LogActivity", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := usecase.NewBranchServiceUseCase(repo, serviceRepo, branchRepo, audit)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.CreateBranchService(ctx, "b-1", &model.CreateBranchServiceRequest{ServiceID: "s-1", CustomName: "Test"})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.IsActive)

	// Deactivate branch then check BS still exists in list
	branch.Status = branchEntity.BranchStatusInactive
	branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(branch, nil).Once()
	repo.On("FindAll", mock.Anything, "t-1", "b-1").Return([]*entity.BranchService{createdBS}, nil).Once()
	list, err := uc.ListBranchServices(ctx, "b-1")
	require.NoError(t, err)
	found := false
	for _, bs := range list {
		if bs.ID == res.ID {
			found = true
			assert.True(t, bs.IsActive, "branch-service remains active after branch deactivation (no cascade)")
		}
	}
	assert.True(t, found, "branch-service still listed after branch deactivation")

	// New BS creation should be rejected
	branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(branch, nil).Once()
	_, err = uc.CreateBranchService(ctx, "b-1", &model.CreateBranchServiceRequest{ServiceID: "s-1", CustomName: "Blocked"})
	assert.ErrorIs(t, err, exception.ErrForbidden, "create on inactive branch should fail")
}

func TestBranchServiceUseCase_AuditHooks(t *testing.T) {
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

	t.Run("Positive_LogsAuditActivitiesForCreateUpdateDelete", func(t *testing.T) {
		audit := serviceMocks.NewMockAuditLogger(t)
		branchRepo := orgMocks.NewMockBranchRepository(t)
		serviceRepo := serviceMocks.NewMockServiceRepository(t)
		repo := serviceMocks.NewMockBranchServiceRepository(t)

		branchRepo.On("FindByID", mock.Anything, "tenant-1", "branch-1").
			Return(&branchEntity.Branch{ID: "branch-1", TenantID: "tenant-1", Status: branchEntity.BranchStatusActive}, nil).Once()
		serviceRepo.On("FindByID", mock.Anything, "tenant-1", "svc-1").
			Return(&entity.Service{ID: "svc-1", TenantID: "tenant-1", Status: entity.ServiceStatusActive}, nil).Once()

		var createdBS *entity.BranchService
		repo.On("Create", mock.Anything, mock.MatchedBy(func(bs *entity.BranchService) bool {
			createdBS = bs
			return bs.BranchID == "branch-1" && bs.ServiceID == "svc-1"
		})).Return(nil).Once()
		audit.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.Action == "BRANCH_SERVICE_CREATE" && req.Entity == "branch_service"
		})).Return(nil).Once()

		uc := usecase.NewBranchServiceUseCase(repo, serviceRepo, branchRepo, audit)

		// Action 1: Create
		reqCreate := &model.CreateBranchServiceRequest{ServiceID: "svc-1"}
		created, err := uc.CreateBranchService(ctx, "branch-1", reqCreate)
		require.NoError(t, err)

		// Action 2: Update
		repo.On("FindByID", mock.Anything, "tenant-1", "branch-1", created.ID).Return(createdBS, nil).Once()
		repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
		audit.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.Action == "BRANCH_SERVICE_UPDATE" && req.Entity == "branch_service"
		})).Return(nil).Once()

		reqUpdate := &model.UpdateBranchServiceRequest{SortOrder: new(int)}
		_, err = uc.UpdateBranchService(ctx, "branch-1", created.ID, reqUpdate)
		require.NoError(t, err)

		// Action 3: Delete
		repo.On("Delete", mock.Anything, "tenant-1", "branch-1", created.ID).Return(nil).Once()
		audit.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.Action == "BRANCH_SERVICE_DELETE" && req.Entity == "branch_service"
		})).Return(nil).Once()

		err = uc.DeleteBranchService(ctx, "branch-1", created.ID)
		require.NoError(t, err)
	})
}

func TestBranchServiceUseCase_RelationGuard(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		setup     func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, repo *serviceMocks.MockBranchServiceRepository)
		wantError error
	}{
		{
			name:     "Positive_CreateBranchService",
			category: "positive",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, repo *serviceMocks.MockBranchServiceRepository) {
				branchRepo.On("FindByID", mock.Anything, "tenant-1", "b-1").
					Return(&branchEntity.Branch{ID: "b-1", TenantID: "tenant-1", Status: branchEntity.BranchStatusActive}, nil).Once()
				serviceRepo.On("FindByID", mock.Anything, "tenant-1", "svc-1").
					Return(&entity.Service{ID: "svc-1", TenantID: "tenant-1", Status: entity.ServiceStatusActive}, nil).Once()
				repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantError: nil,
		},
		{
			name:     "Negative_CreateBranchServiceForInactiveBranch",
			category: "negative",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, repo *serviceMocks.MockBranchServiceRepository) {
				branchRepo.On("FindByID", mock.Anything, "tenant-1", "b-1").
					Return(&branchEntity.Branch{ID: "b-1", TenantID: "tenant-1", Status: branchEntity.BranchStatusInactive}, nil).Once()
			},
			wantError: exception.ErrForbidden,
		},
		{
			name:     "Negative_CreateBranchServiceWithInactiveService",
			category: "negative",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, repo *serviceMocks.MockBranchServiceRepository) {
				branchRepo.On("FindByID", mock.Anything, "tenant-1", "b-1").
					Return(&branchEntity.Branch{ID: "b-1", TenantID: "tenant-1", Status: branchEntity.BranchStatusActive}, nil).Once()
				serviceRepo.On("FindByID", mock.Anything, "tenant-1", "svc-1").
					Return(&entity.Service{ID: "svc-1", TenantID: "tenant-1", Status: entity.ServiceStatusInactive}, nil).Once()
			},
			wantError: exception.ErrForbidden,
		},
		{
			name:     "Vulnerability_CrossTenantCreateRejected",
			category: "vulnerability",
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, serviceRepo *serviceMocks.MockServiceRepository, repo *serviceMocks.MockBranchServiceRepository) {
				branchRepo.On("FindByID", mock.Anything, "tenant-1", "b-1").
					Return(nil, exception.ErrNotFound).Once()
			},
			wantError: exception.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.category)
			ctx := database.SetOrganizationContext(context.Background(), "tenant-1")
			audit := serviceMocks.NewMockAuditLogger(t)
			audit.On("LogActivity", mock.Anything, mock.Anything).Return(nil).Maybe()
			branchRepo := orgMocks.NewMockBranchRepository(t)
			serviceRepo := serviceMocks.NewMockServiceRepository(t)
			repo := serviceMocks.NewMockBranchServiceRepository(t)

			tt.setup(t, branchRepo, serviceRepo, repo)

			uc := usecase.NewBranchServiceUseCase(repo, serviceRepo, branchRepo, audit)

			res, err := uc.CreateBranchService(ctx, "b-1", &model.CreateBranchServiceRequest{ServiceID: "svc-1"})

			if tt.wantError != nil {
				require.ErrorIs(t, err, tt.wantError)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.Equal(t, "b-1", res.BranchID)
				assert.Equal(t, "svc-1", res.ServiceID)
				assert.True(t, res.IsActive)
			}
		})
	}
}
