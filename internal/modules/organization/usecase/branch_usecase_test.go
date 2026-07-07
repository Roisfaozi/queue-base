package usecase

import (
	"context"
	"testing"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	testifyMock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResolveBranch(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		branchID  string
		tenantID  string
		mockSetup func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger)
		wantErr   error
		wantRes   func(t *testing.T, res *model.BranchResponse)
	}{
		{
			name:     "Positive_ResolveBranchUsesTenantScope",
			category: "positive",
			branchID: "branch-1",
			tenantID: "tenant-1",
			mockSetup: func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger) {
				repo := mocks.NewMockBranchRepository(t)
				var gotTenantID, gotBranchID string
				repo.EXPECT().FindByID(testifyMock.Anything, "tenant-1", "branch-1").Run(func(ctx context.Context, tenantID string, branchID string) {
					gotTenantID = tenantID
					gotBranchID = branchID
				}).Return(&entity.Branch{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN", Name: "Main", Status: entity.BranchStatusActive}, nil)
				t.Cleanup(func() {
					assert.Equal(t, "tenant-1", gotTenantID)
					assert.Equal(t, "branch-1", gotBranchID)
				})
				return repo, nil
			},
			wantErr: nil,
			wantRes: func(t *testing.T, res *model.BranchResponse) {
				assert.Equal(t, "tenant-1", res.TenantID)
			},
		},
		{
			name:     "Negative_RequiresTenantAndBranch",
			category: "negative",
			branchID: "",
			tenantID: "",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_RejectsCrossTenantLookup",
			category: "negative",
			branchID: "branch-2",
			tenantID: "tenant-1",
			mockSetup: func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger) {
				repo := mocks.NewMockBranchRepository(t)
				var gotTenantID, gotBranchID string
				repo.EXPECT().FindByID(testifyMock.Anything, "tenant-1", "branch-2").Run(func(ctx context.Context, tenantID string, branchID string) {
					gotTenantID = tenantID
					gotBranchID = branchID
				}).Return(nil, exception.ErrNotFound)
				t.Cleanup(func() {
					assert.Equal(t, "tenant-1", gotTenantID)
					assert.Equal(t, "branch-2", gotBranchID)
				})
				return repo, nil
			},
			wantErr: exception.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockBranchRepository(t)
			var audit *mocks.MockAuditLogger
			if tt.mockSetup != nil {
				repo, audit = tt.mockSetup(t)
			}
			var uc BranchUseCase
			if audit != nil {
				uc = NewBranchUseCase(repo, audit)
			} else {
				uc = NewBranchUseCase(repo)
			}

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.ResolveBranch(ctx, tt.branchID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, res)
			}
		})
	}
}

func TestCreateBranch(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		req       model.CreateBranchRequest
		tenantID  string
		mockSetup func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger)
		wantErr   error
		wantRes   func(t *testing.T, res *model.BranchResponse)
	}{
		{
			name:     "Positive_CreateBranchUsesTenantContext",
			category: "positive",
			req:      model.CreateBranchRequest{Code: "main", Name: "Main Branch"},
			tenantID: "tenant-1",
			mockSetup: func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger) {
				repo := mocks.NewMockBranchRepository(t)
				var gotBranch *entity.Branch
				repo.EXPECT().Create(testifyMock.Anything, testifyMock.AnythingOfType("*entity.Branch")).Run(func(ctx context.Context, branch *entity.Branch) {
					gotBranch = branch
				}).Return(nil)
				t.Cleanup(func() {
					require.NotNil(t, gotBranch)
					assert.Equal(t, "tenant-1", gotBranch.TenantID)
					assert.Equal(t, "MAIN", gotBranch.Code)
					assert.Equal(t, "Main Branch", gotBranch.Name)
					assert.Equal(t, entity.BranchStatusDraft, gotBranch.Status)
				})
				return repo, nil
			},
			wantErr: nil,
			wantRes: func(t *testing.T, res *model.BranchResponse) {
				assert.Equal(t, "tenant-1", res.TenantID)
				assert.Equal(t, "MAIN", res.Code)
				assert.Equal(t, "Main Branch", res.Name)
				assert.Equal(t, entity.BranchStatusDraft, res.Status, "branch without required fields should be draft")
			},
		},
		{
			name:     "Negative_MissingTenantContext",
			category: "negative",
			req:      model.CreateBranchRequest{Code: "main", Name: "Main Branch"},
			tenantID: "",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Positive_CreateBranchWithFullProfile_SetsActive",
			category: "positive",
			req:      model.CreateBranchRequest{Code: "full", Name: "Full Branch", Address: "Jl. Raya", City: "Jakarta", Province: "DKI", Phone: "021123", Timezone: "Asia/Jakarta"},
			tenantID: "tenant-1",
			mockSetup: func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger) {
				repo := mocks.NewMockBranchRepository(t)
				var gotBranch *entity.Branch
				repo.EXPECT().Create(testifyMock.Anything, testifyMock.AnythingOfType("*entity.Branch")).Run(func(ctx context.Context, branch *entity.Branch) {
					gotBranch = branch
				}).Return(nil)
				t.Cleanup(func() {
					require.NotNil(t, gotBranch)
					assert.Equal(t, "tenant-1", gotBranch.TenantID)
					assert.Equal(t, "FULL", gotBranch.Code)
					assert.Equal(t, entity.BranchStatusActive, gotBranch.Status)
				})
				return repo, nil
			},
			wantErr: nil,
			wantRes: func(t *testing.T, res *model.BranchResponse) {
				assert.Equal(t, entity.BranchStatusActive, res.Status, "branch with all required fields should be active")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockBranchRepository(t)
			var audit *mocks.MockAuditLogger
			if tt.mockSetup != nil {
				repo, audit = tt.mockSetup(t)
			}
			var uc BranchUseCase
			if audit != nil {
				uc = NewBranchUseCase(repo, audit)
			} else {
				uc = NewBranchUseCase(repo)
			}

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.CreateBranch(ctx, &tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, res)
			}
		})
	}
}

func TestListBranches(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		tenantID  string
		mockSetup func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger)
		wantErr   error
		wantRes   func(t *testing.T, res []model.BranchResponse)
	}{
		{
			name:     "Positive_ListBranchesUsesTenantScope",
			category: "positive",
			tenantID: "tenant-1",
			mockSetup: func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger) {
				repo := mocks.NewMockBranchRepository(t)
				var gotTenantID string
				repo.EXPECT().FindAll(testifyMock.Anything, "tenant-1").Run(func(ctx context.Context, tenantID string) {
					gotTenantID = tenantID
				}).Return([]*entity.Branch{{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN", Name: "Main"}}, nil)
				t.Cleanup(func() {
					assert.Equal(t, "tenant-1", gotTenantID)
				})
				return repo, nil
			},
			wantErr: nil,
			wantRes: func(t *testing.T, res []model.BranchResponse) {
				require.Len(t, res, 1)
				assert.Equal(t, "tenant-1", res[0].TenantID)
			},
		},
		{
			name:     "Negative_MissingTenantContext",
			category: "negative",
			tenantID: "",
			wantErr:  exception.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockBranchRepository(t)
			var audit *mocks.MockAuditLogger
			if tt.mockSetup != nil {
				repo, audit = tt.mockSetup(t)
			}
			var uc BranchUseCase
			if audit != nil {
				uc = NewBranchUseCase(repo, audit)
			} else {
				uc = NewBranchUseCase(repo)
			}

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.ListBranches(ctx)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, res)
			}
		})
	}
}

func TestUpdateBranch(t *testing.T) {
	code := " sub "
	name := " Branch Office "
	active := entity.BranchStatusActive
	address := "Jl. Test"
	city := "Jakarta"
	province := "DKI Jakarta"
	phone := "021"
	timezone := "Asia/Jakarta"
	runningText := "Counter moved to floor 2"
	tests := []struct {
		name      string
		category  string
		branchID  string
		req       model.UpdateBranchRequest
		tenantID  string
		mockSetup func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger)
		wantErr   error
		wantRes   func(t *testing.T, res *model.BranchResponse)
	}{
		{
			name:     "Positive_UpdateBranchSanitizesFields",
			category: "positive",
			branchID: "branch-1",
			req:      model.UpdateBranchRequest{Code: &code, Name: &name},
			tenantID: "tenant-1",
			mockSetup: func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger) {
				repo := mocks.NewMockBranchRepository(t)
				var gotTenantID, gotBranchID string
				var gotBranch *entity.Branch
				repo.EXPECT().FindByID(testifyMock.Anything, "tenant-1", "branch-1").Run(func(ctx context.Context, tenantID string, branchID string) {
					gotTenantID = tenantID
					gotBranchID = branchID
				}).Return(&entity.Branch{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN", Name: "Main", Status: entity.BranchStatusActive}, nil)
				repo.EXPECT().Update(testifyMock.Anything, testifyMock.AnythingOfType("*entity.Branch")).Run(func(ctx context.Context, branch *entity.Branch) {
					gotBranch = branch
				}).Return(nil)
				t.Cleanup(func() {
					assert.Equal(t, "tenant-1", gotTenantID)
					assert.Equal(t, "branch-1", gotBranchID)
					require.NotNil(t, gotBranch)
					assert.Equal(t, "branch-1", gotBranch.ID)
					assert.Equal(t, "tenant-1", gotBranch.TenantID)
					assert.Equal(t, "SUB", gotBranch.Code)
					assert.Equal(t, "Branch Office", gotBranch.Name)
				})
				return repo, nil
			},
			wantErr: nil,
			wantRes: func(t *testing.T, res *model.BranchResponse) {
				assert.Equal(t, "SUB", res.Code)
				assert.Equal(t, "Branch Office", res.Name)
			},
		},
		{
			name:     "Negative_MissingTenantOrBranch",
			category: "negative",
			branchID: "",
			req:      model.UpdateBranchRequest{Code: &code, Name: &name},
			tenantID: "tenant-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_NotFound",
			category: "negative",
			branchID: "branch-1",
			req:      model.UpdateBranchRequest{Code: &code, Name: &name},
			tenantID: "tenant-1",
			mockSetup: func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger) {
				repo := mocks.NewMockBranchRepository(t)
				var gotTenantID, gotBranchID string
				repo.EXPECT().FindByID(testifyMock.Anything, "tenant-1", "branch-1").Run(func(ctx context.Context, tenantID string, branchID string) {
					gotTenantID = tenantID
					gotBranchID = branchID
				}).Return(nil, exception.ErrNotFound)
				t.Cleanup(func() {
					assert.Equal(t, "tenant-1", gotTenantID)
					assert.Equal(t, "branch-1", gotBranchID)
				})
				return repo, nil
			},
			wantErr: exception.ErrNotFound,
		},
		{
			name:     "Positive_RunningTextUpdateWritesAudit",
			category: "positive",
			branchID: "branch-1",
			req: model.UpdateBranchRequest{
				RunningText: &runningText,
			},
			tenantID: "tenant-1",
			mockSetup: func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger) {
				repo := mocks.NewMockBranchRepository(t)
				audit := mocks.NewMockAuditLogger(t)
				var gotTenantID, gotBranchID string
				var gotBranch *entity.Branch
				var gotAudit auditModel.CreateAuditLogRequest
				repo.EXPECT().FindByID(testifyMock.Anything, "tenant-1", "branch-1").Run(func(ctx context.Context, tenantID string, branchID string) {
					gotTenantID = tenantID
					gotBranchID = branchID
				}).Return(&entity.Branch{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN", Name: "Main", Status: entity.BranchStatusActive}, nil)
				repo.EXPECT().Update(testifyMock.Anything, testifyMock.AnythingOfType("*entity.Branch")).Run(func(ctx context.Context, branch *entity.Branch) {
					gotBranch = branch
				}).Return(nil)
				audit.EXPECT().LogActivity(testifyMock.Anything, testifyMock.AnythingOfType("model.CreateAuditLogRequest")).Run(func(ctx context.Context, req auditModel.CreateAuditLogRequest) {
					gotAudit = req
				}).Return(nil)
				t.Cleanup(func() {
					assert.Equal(t, "tenant-1", gotTenantID)
					assert.Equal(t, "branch-1", gotBranchID)
					require.NotNil(t, gotBranch)
					assert.Equal(t, "branch-1", gotBranch.ID)
					assert.Equal(t, runningText, gotBranch.RunningText)
					assert.Equal(t, "BRANCH_UPDATE", gotAudit.Action)
					assert.Equal(t, "branch", gotAudit.Entity)
					assert.Equal(t, "branch-1", gotAudit.EntityID)
				})
				return repo, audit
			},
			wantRes: func(t *testing.T, res *model.BranchResponse) {
				assert.Equal(t, runningText, res.RunningText)
			},
		},
		{
			name:     "Negative_ActivateMissingRequiredFields",
			category: "negative",
			branchID: "branch-1",
			req:      model.UpdateBranchRequest{Status: &active},
			tenantID: "tenant-1",
			mockSetup: func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger) {
				repo := mocks.NewMockBranchRepository(t)
				var gotTenantID, gotBranchID string
				repo.EXPECT().FindByID(testifyMock.Anything, "tenant-1", "branch-1").Run(func(ctx context.Context, tenantID string, branchID string) {
					gotTenantID = tenantID
					gotBranchID = branchID
				}).Return(&entity.Branch{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN", Name: "Main", Status: entity.BranchStatusInactive}, nil)
				t.Cleanup(func() {
					assert.Equal(t, "tenant-1", gotTenantID)
					assert.Equal(t, "branch-1", gotBranchID)
				})
				return repo, nil
			},
			wantErr: exception.ErrBadRequest,
		},
		{
			name:     "Positive_ActivateWithRequiredFieldsInRequest",
			category: "positive",
			branchID: "branch-1",
			req: model.UpdateBranchRequest{
				Address:  &address,
				City:     &city,
				Province: &province,
				Phone:    &phone,
				Timezone: &timezone,
				Status:   &active,
			},
			tenantID: "tenant-1",
			mockSetup: func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger) {
				repo := mocks.NewMockBranchRepository(t)
				var gotTenantID, gotBranchID string
				var gotBranch *entity.Branch
				repo.EXPECT().FindByID(testifyMock.Anything, "tenant-1", "branch-1").Run(func(ctx context.Context, tenantID string, branchID string) {
					gotTenantID = tenantID
					gotBranchID = branchID
				}).Return(&entity.Branch{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN", Name: "Main", Status: entity.BranchStatusInactive}, nil)
				repo.EXPECT().Update(testifyMock.Anything, testifyMock.AnythingOfType("*entity.Branch")).Run(func(ctx context.Context, branch *entity.Branch) {
					gotBranch = branch
				}).Return(nil)
				t.Cleanup(func() {
					assert.Equal(t, "tenant-1", gotTenantID)
					assert.Equal(t, "branch-1", gotBranchID)
					require.NotNil(t, gotBranch)
					assert.Equal(t, "branch-1", gotBranch.ID)
					assert.Equal(t, entity.BranchStatusActive, gotBranch.Status)
					assert.Equal(t, address, gotBranch.Address)
				})
				return repo, nil
			},
			wantRes: func(t *testing.T, res *model.BranchResponse) {
				assert.Equal(t, entity.BranchStatusActive, res.Status)
				assert.Equal(t, address, res.Address)
			},
		},
		{
			name:     "Edge_UpdateNonProfileFieldsWhenAlreadyActive",
			category: "edge",
			branchID: "branch-1",
			req: model.UpdateBranchRequest{
				Name: &name,
			},
			tenantID: "tenant-1",
			mockSetup: func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger) {
				repo := mocks.NewMockBranchRepository(t)
				var gotTenantID, gotBranchID string
				var gotBranch *entity.Branch
				repo.EXPECT().FindByID(testifyMock.Anything, "tenant-1", "branch-1").Run(func(ctx context.Context, tenantID string, branchID string) {
					gotTenantID = tenantID
					gotBranchID = branchID
				}).Return(&entity.Branch{
					ID:       "branch-1",
					TenantID: "tenant-1",
					Code:     "MAIN",
					Name:     "Main",
					Status:   entity.BranchStatusActive,
					Address:  "Jl. Test",
					City:     "Jakarta",
					Province: "DKI",
					Phone:    "123",
					Timezone: "Asia",
				}, nil)
				repo.EXPECT().Update(testifyMock.Anything, testifyMock.AnythingOfType("*entity.Branch")).Run(func(ctx context.Context, branch *entity.Branch) {
					gotBranch = branch
				}).Return(nil)
				t.Cleanup(func() {
					assert.Equal(t, "tenant-1", gotTenantID)
					assert.Equal(t, "branch-1", gotBranchID)
					require.NotNil(t, gotBranch)
					assert.Equal(t, "branch-1", gotBranch.ID)
					assert.Equal(t, entity.BranchStatusActive, gotBranch.Status)
					assert.Equal(t, "Branch Office", gotBranch.Name)
				})
				return repo, nil
			},
			wantRes: func(t *testing.T, res *model.BranchResponse) {
				assert.Equal(t, entity.BranchStatusActive, res.Status)
				assert.Equal(t, "Branch Office", res.Name)
			},
		},
		{
			name:     "Vulnerability_CrossTenantBranchUpdateRejected",
			category: "vulnerability",
			branchID: "branch-1",
			req:      model.UpdateBranchRequest{Name: &name},
			tenantID: "tenant-1",
			mockSetup: func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger) {
				repo := mocks.NewMockBranchRepository(t)
				var gotTenantID, gotBranchID string
				repo.EXPECT().FindByID(testifyMock.Anything, "tenant-1", "branch-1").Run(func(ctx context.Context, tenantID string, branchID string) {
					gotTenantID = tenantID
					gotBranchID = branchID
				}).Return(nil, exception.ErrNotFound)
				t.Cleanup(func() {
					assert.Equal(t, "tenant-1", gotTenantID)
					assert.Equal(t, "branch-1", gotBranchID)
				})
				return repo, nil
			},
			wantErr: exception.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockBranchRepository(t)
			var audit *mocks.MockAuditLogger
			if tt.mockSetup != nil {
				repo, audit = tt.mockSetup(t)
			}
			var uc BranchUseCase
			if audit != nil {
				uc = NewBranchUseCase(repo, audit)
			} else {
				uc = NewBranchUseCase(repo)
			}

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.UpdateBranch(ctx, tt.branchID, &tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, res)
			}
		})
	}
}

func TestDeleteBranch(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		branchID  string
		tenantID  string
		mockSetup func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger)
		wantErr   error
	}{
		{
			name:     "Positive_DeleteBranch",
			category: "positive",
			branchID: "branch-1",
			tenantID: "tenant-1",
			mockSetup: func(t *testing.T) (*mocks.MockBranchRepository, *mocks.MockAuditLogger) {
				repo := mocks.NewMockBranchRepository(t)
				var gotTenantID, gotBranchID string
				repo.EXPECT().Delete(testifyMock.Anything, "tenant-1", "branch-1").Run(func(ctx context.Context, tenantID string, branchID string) {
					gotTenantID = tenantID
					gotBranchID = branchID
				}).Return(nil)
				t.Cleanup(func() {
					assert.Equal(t, "tenant-1", gotTenantID)
					assert.Equal(t, "branch-1", gotBranchID)
				})
				return repo, nil
			},
			wantErr: nil,
		},
		{
			name:     "Negative_DeleteBranchRequiresBranchID",
			category: "negative",
			branchID: "",
			tenantID: "tenant-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_MissingTenantContext",
			category: "negative",
			branchID: "branch-1",
			tenantID: "",
			wantErr:  exception.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockBranchRepository(t)
			var audit *mocks.MockAuditLogger
			if tt.mockSetup != nil {
				repo, audit = tt.mockSetup(t)
			}
			var uc BranchUseCase
			if audit != nil {
				uc = NewBranchUseCase(repo, audit)
			} else {
				uc = NewBranchUseCase(repo)
			}

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			err := uc.DeleteBranch(ctx, tt.branchID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
