package usecase_test

import (
	"context"
	"testing"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/counter/model"
	counterMocks "github.com/Roisfaozi/queue-base/internal/modules/counter/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/counter/usecase"
	organizationEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	orgMocks "github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	serviceMocks "github.com/Roisfaozi/queue-base/internal/modules/service/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateCounter(t *testing.T) {
	tests := []struct {
		name     string
		category string
		req      model.CreateCounterRequest
		tenantID string
		setup    func(t *testing.T, repo *counterMocks.MockCounterRepository, branchRepo *orgMocks.MockBranchRepository, bsRepo *serviceMocks.MockBranchServiceRepository, createdCounter **entity.Counter)
		wantErr  error
		assert   func(t *testing.T, res *model.CounterResponse, createdCounter *entity.Counter)
	}{
		{
			name:     "Positive_CreatesCounterWithSanitization",
			category: "positive",
			req:      model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: " a1 ", Name: " Front Desk "},
			tenantID: "tenant-1",
			setup: func(t *testing.T, repo *counterMocks.MockCounterRepository, branchRepo *orgMocks.MockBranchRepository, bsRepo *serviceMocks.MockBranchServiceRepository, createdCounter **entity.Counter) {
				branchRepo.On("FindByID", mock.Anything, "tenant-1", "550e8400-e29b-41d4-a716-446655440000").
					Return(&organizationEntity.Branch{
						ID:       "550e8400-e29b-41d4-a716-446655440000",
						TenantID: "tenant-1",
						Status:   organizationEntity.BranchStatusActive,
					}, nil).Once()

				repo.On("Create", mock.Anything, mock.MatchedBy(func(c *entity.Counter) bool {
					*createdCounter = c
					return c.Code == "A1" && c.Name == "Front Desk" && c.TenantID == "tenant-1"
				})).Return(nil).Once()
			},
			wantErr: nil,
			assert: func(t *testing.T, res *model.CounterResponse, createdCounter *entity.Counter) {
				assert.Equal(t, "tenant-1", res.TenantID)
				require.NotNil(t, createdCounter)
				assert.Equal(t, "A1", createdCounter.Code)
				assert.Equal(t, "Front Desk", createdCounter.Name)
			},
		},
		{
			name:     "Negative_RequiresTenant",
			category: "negative",
			req:      model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Desk"},
			tenantID: "",
			setup: func(t *testing.T, repo *counterMocks.MockCounterRepository, branchRepo *orgMocks.MockBranchRepository, bsRepo *serviceMocks.MockBranchServiceRepository, createdCounter **entity.Counter) {
			},
			wantErr: exception.ErrBadRequest,
		},
		{
			name:     "Edge_RejectsInactiveBranch",
			category: "edge",
			req:      model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Desk"},
			tenantID: "tenant-1",
			setup: func(t *testing.T, repo *counterMocks.MockCounterRepository, branchRepo *orgMocks.MockBranchRepository, bsRepo *serviceMocks.MockBranchServiceRepository, createdCounter **entity.Counter) {
				branchRepo.On("FindByID", mock.Anything, "tenant-1", "550e8400-e29b-41d4-a716-446655440000").
					Return(&organizationEntity.Branch{
						ID:       "550e8400-e29b-41d4-a716-446655440000",
						TenantID: "tenant-1",
						Status:   organizationEntity.BranchStatusInactive,
					}, nil).Once()
			},
			wantErr: exception.ErrForbidden,
		},
		{
			name:     "Vulnerability_RejectsCrossTenantBranch",
			category: "vulnerability",
			req:      model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Desk"},
			tenantID: "tenant-1",
			setup: func(t *testing.T, repo *counterMocks.MockCounterRepository, branchRepo *orgMocks.MockBranchRepository, bsRepo *serviceMocks.MockBranchServiceRepository, createdCounter **entity.Counter) {
				branchRepo.On("FindByID", mock.Anything, "tenant-1", "550e8400-e29b-41d4-a716-446655440000").
					Return(nil, exception.ErrNotFound).Once()
			},
			wantErr: exception.ErrForbidden,
			assert: func(t *testing.T, res *model.CounterResponse, createdCounter *entity.Counter) {
				assert.Nil(t, createdCounter)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.category)
			repo := counterMocks.NewMockCounterRepository(t)
			branchRepo := orgMocks.NewMockBranchRepository(t)
			bsRepo := serviceMocks.NewMockBranchServiceRepository(t)

			var createdCounter *entity.Counter
			if tt.setup != nil {
				tt.setup(t, repo, branchRepo, bsRepo, &createdCounter)
			}

			uc := usecase.NewCounterUseCase(repo, branchRepo, bsRepo)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.CreateCounter(ctx, &tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				if tt.assert != nil {
					tt.assert(t, res, createdCounter)
				}
				return
			}
			require.NoError(t, err)
			if tt.assert != nil {
				tt.assert(t, res, createdCounter)
			}
		})
	}
}

func TestUpdateCounter(t *testing.T) {
	code := " b2 "
	name := " Front Office "

	tests := []struct {
		name      string
		category  string
		counterID string
		req       model.UpdateCounterRequest
		tenantID  string
		setup     func(t *testing.T, repo *counterMocks.MockCounterRepository, updatedCounter **entity.Counter)
		wantErr   error
		assert    func(t *testing.T, res *model.CounterResponse, updatedCounter *entity.Counter)
	}{
		{
			name:      "Positive_SanitizesFieldsOnUpdate",
			category:  "positive",
			counterID: "counter-1",
			req:       model.UpdateCounterRequest{Code: &code, Name: &name},
			tenantID:  "tenant-1",
			setup: func(t *testing.T, repo *counterMocks.MockCounterRepository, updatedCounter **entity.Counter) {
				repo.On("FindByID", mock.Anything, "tenant-1", "counter-1").
					Return(&entity.Counter{
						ID:       "counter-1",
						TenantID: "tenant-1",
						BranchID: "550e8400-e29b-41d4-a716-446655440000",
						Code:     "A1",
						Name:     "Desk",
						Status:   entity.CounterStatusActive,
					}, nil).Once()

				repo.On("Update", mock.Anything, mock.MatchedBy(func(c *entity.Counter) bool {
					*updatedCounter = c
					return c.Code == "B2" && c.Name == "Front Office"
				})).Return(nil).Once()
			},
			wantErr: nil,
			assert: func(t *testing.T, res *model.CounterResponse, updatedCounter *entity.Counter) {
				assert.Equal(t, "B2", res.Code)
				assert.Equal(t, "Front Office", res.Name)
			},
		},
		{
			name:      "Negative_NotFound",
			category:  "negative",
			counterID: "counter-missing",
			req:       model.UpdateCounterRequest{Code: &code, Name: &name},
			tenantID:  "tenant-1",
			setup: func(t *testing.T, repo *counterMocks.MockCounterRepository, updatedCounter **entity.Counter) {
				repo.On("FindByID", mock.Anything, "tenant-1", "counter-missing").
					Return(nil, exception.ErrNotFound).Once()
			},
			wantErr: exception.ErrNotFound,
		},
		{
			name:      "Vulnerability_EmptyTenant",
			category:  "vulnerability",
			counterID: "counter-1",
			req:       model.UpdateCounterRequest{Code: &code, Name: &name},
			tenantID:  "",
			setup: func(t *testing.T, repo *counterMocks.MockCounterRepository, updatedCounter **entity.Counter) {
			},
			wantErr: exception.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.category)
			repo := counterMocks.NewMockCounterRepository(t)
			branchRepo := orgMocks.NewMockBranchRepository(t)
			bsRepo := serviceMocks.NewMockBranchServiceRepository(t)

			var updatedCounter *entity.Counter
			if tt.setup != nil {
				tt.setup(t, repo, &updatedCounter)
			}

			uc := usecase.NewCounterUseCase(repo, branchRepo, bsRepo)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.UpdateCounter(ctx, tt.counterID, &tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				if tt.assert != nil {
					tt.assert(t, res, updatedCounter)
				}
				return
			}
			require.NoError(t, err)
			if tt.assert != nil {
				tt.assert(t, res, updatedCounter)
			}
		})
	}
}

func TestCounterDeactivationEdgeCase(t *testing.T) {
	repo := counterMocks.NewMockCounterRepository(t)
	branchRepo := orgMocks.NewMockBranchRepository(t)
	bsRepo := serviceMocks.NewMockBranchServiceRepository(t)

	uc := usecase.NewCounterUseCase(repo, branchRepo, bsRepo, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	activeBranch := &organizationEntity.Branch{ID: "b-1", TenantID: "t-1", Status: organizationEntity.BranchStatusActive}
	inactiveBranch := &organizationEntity.Branch{ID: "b-1", TenantID: "t-1", Status: organizationEntity.BranchStatusInactive}

	branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(activeBranch, nil).Once()
	bsRepo.On("FindByID", mock.Anything, "t-1", "b-1", "bs-1").Return(&serviceEntity.BranchService{
		ID:       "bs-1",
		TenantID: "t-1",
		BranchID: "b-1",
		IsActive: true,
	}, nil).Once()

	var createdCounter *entity.Counter
	repo.On("Create", mock.Anything, mock.MatchedBy(func(c *entity.Counter) bool {
		createdCounter = c
		return c.Code == "C1" && c.Status == entity.CounterStatusActive
	})).Return(nil).Once()

	req := &model.CreateCounterRequest{
		BranchID:        "b-1",
		BranchServiceID: "bs-1",
		Code:            "C1",
		Name:            "Counter 1",
		DisplayName:     "Counter 1",
	}
	res, err := uc.CreateCounter(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, entity.CounterStatusActive, res.Status)

	// Counter still accessible via GetCounter
	repo.On("FindByID", mock.Anything, "t-1", res.ID).Return(createdCounter, nil).Once()
	getRes, err := uc.GetCounter(ctx, res.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.CounterStatusActive, getRes.Status, "counter remains active after branch deactivation (no cascade)")

	// New counter creation on inactive branch should be rejected
	branchRepo.On("FindByID", mock.Anything, "t-1", "b-1").Return(inactiveBranch, nil).Once()
	req2 := &model.CreateCounterRequest{
		BranchID:        "b-1",
		BranchServiceID: "bs-1",
		Code:            "C2",
		Name:            "Counter 2",
		DisplayName:     "Counter 2",
	}
	_, err = uc.CreateCounter(ctx, req2)
	assert.ErrorIs(t, err, exception.ErrForbidden, "create on inactive branch should fail")
}

func TestCounterServiceRelationGuard(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		req       model.CreateCounterRequest
		setup     func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, bsRepo *serviceMocks.MockBranchServiceRepository)
		wantError error
	}{
		{
			name:     "Vulnerability_RejectsInvalidBranchServiceRelation",
			category: "vulnerability",
			req: model.CreateCounterRequest{
				BranchID:        "branch-1",
				BranchServiceID: "service-unknown",
				Code:            "A1",
			},
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, bsRepo *serviceMocks.MockBranchServiceRepository) {
				branchRepo.On("FindByID", mock.Anything, "tenant-1", "branch-1").
					Return(&organizationEntity.Branch{ID: "branch-1", TenantID: "tenant-1", Status: organizationEntity.BranchStatusActive}, nil).Once()
				bsRepo.On("FindByID", mock.Anything, "tenant-1", "branch-1", "service-unknown").
					Return(nil, exception.ErrNotFound).Once()
			},
			wantError: exception.ErrForbidden,
		},
		{
			name:     "Vulnerability_RejectsInactiveBranchServiceRelation",
			category: "vulnerability",
			req: model.CreateCounterRequest{
				BranchID:        "branch-1",
				BranchServiceID: "service-inactive",
				Code:            "A1",
			},
			setup: func(t *testing.T, branchRepo *orgMocks.MockBranchRepository, bsRepo *serviceMocks.MockBranchServiceRepository) {
				branchRepo.On("FindByID", mock.Anything, "tenant-1", "branch-1").
					Return(&organizationEntity.Branch{ID: "branch-1", TenantID: "tenant-1", Status: organizationEntity.BranchStatusActive}, nil).Once()
				bsRepo.On("FindByID", mock.Anything, "tenant-1", "branch-1", "service-inactive").
					Return(&serviceEntity.BranchService{ID: "service-inactive", TenantID: "tenant-1", BranchID: "branch-1", IsActive: false}, nil).Once()
			},
			wantError: exception.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.category)
			repo := counterMocks.NewMockCounterRepository(t)
			branchRepo := orgMocks.NewMockBranchRepository(t)
			bsRepo := serviceMocks.NewMockBranchServiceRepository(t)

			tt.setup(t, branchRepo, bsRepo)

			uc := usecase.NewCounterUseCase(repo, branchRepo, bsRepo)
			ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

			_, err := uc.CreateCounter(ctx, &tt.req)
			assert.ErrorIs(t, err, tt.wantError)
		})
	}
}

func TestCounterAuditHooks(t *testing.T) {
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

	t.Run("Positive_LogsAuditActivitiesForCreateUpdateDelete", func(t *testing.T) {
		audit := counterMocks.NewMockAuditLogger(t)
		repo := counterMocks.NewMockCounterRepository(t)
		branchRepo := orgMocks.NewMockBranchRepository(t)
		bsRepo := serviceMocks.NewMockBranchServiceRepository(t)

		uc := usecase.NewCounterUseCase(repo, branchRepo, bsRepo, audit)

		branchRepo.On("FindByID", mock.Anything, "tenant-1", "550e8400-e29b-41d4-a716-446655440000").
			Return(&organizationEntity.Branch{ID: "550e8400-e29b-41d4-a716-446655440000", TenantID: "tenant-1", Status: organizationEntity.BranchStatusActive}, nil).Once()
		repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		audit.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.Action == "COUNTER_CREATE" && req.Entity == "counter"
		})).Return(nil).Once()

		_, err := uc.CreateCounter(ctx, &model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Counter A"})
		require.NoError(t, err)

		existingCounter := &entity.Counter{ID: "counter-1", TenantID: "tenant-1", BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Status: entity.CounterStatusActive}
		repo.On("FindByID", mock.Anything, "tenant-1", "counter-1").Return(existingCounter, nil).Once()
		repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
		audit.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.Action == "COUNTER_UPDATE" && req.Entity == "counter"
		})).Return(nil).Once()

		_, err = uc.UpdateCounter(ctx, "counter-1", &model.UpdateCounterRequest{})
		require.NoError(t, err)

		repo.On("Delete", mock.Anything, "tenant-1", "counter-1").Return(nil).Once()
		audit.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.Action == "COUNTER_DELETE" && req.Entity == "counter"
		})).Return(nil).Once()

		err = uc.DeleteCounter(ctx, "counter-1")
		require.NoError(t, err)
	})
}
