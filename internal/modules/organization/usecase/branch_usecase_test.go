package usecase

import (
	"context"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubBranchRepo struct {
	branch *entity.Branch
	list   []*entity.Branch
	err    error
	seen   struct {
		tenantID string
		branchID string
	}
}

func (s *stubBranchRepo) Create(_ context.Context, branch *entity.Branch) error {
	s.branch = branch
	return s.err
}

func (s *stubBranchRepo) FindByID(_ context.Context, tenantID, branchID string) (*entity.Branch, error) {
	s.seen.tenantID = tenantID
	s.seen.branchID = branchID
	if s.err != nil {
		return nil, s.err
	}
	return s.branch, nil
}

func (s *stubBranchRepo) FindAll(_ context.Context, tenantID string) ([]*entity.Branch, error) {
	s.seen.tenantID = tenantID
	if s.err != nil {
		return nil, s.err
	}
	return s.list, nil
}

func (s *stubBranchRepo) Update(_ context.Context, branch *entity.Branch) error {
	s.branch = branch
	return s.err
}

func (s *stubBranchRepo) Delete(_ context.Context, tenantID, branchID string) error {
	s.seen.tenantID = tenantID
	s.seen.branchID = branchID
	return s.err
}

func TestResolveBranch(t *testing.T) {
	tests := []struct {
		name     string
		category string
		branchID string
		stubRepo struct {
			branch *entity.Branch
			err    error
		}
		tenantID string
		wantErr  error
		wantRes  func(t *testing.T, res *model.BranchResponse, repo *stubBranchRepo)
	}{
		{
			name:     "Positive_ResolveBranchUsesTenantScope",
			category: "positive",
			branchID: "branch-1",
			stubRepo: struct {
				branch *entity.Branch
				err    error
			}{
				branch: &entity.Branch{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN", Name: "Main", Status: entity.BranchStatusActive},
				err:    nil,
			},
			tenantID: "tenant-1",
			wantErr:  nil,
			wantRes: func(t *testing.T, res *model.BranchResponse, repo *stubBranchRepo) {
				assert.Equal(t, "tenant-1", res.TenantID)
				assert.Equal(t, "tenant-1", repo.seen.tenantID)
				assert.Equal(t, "branch-1", repo.seen.branchID)
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
			stubRepo: struct {
				branch *entity.Branch
				err    error
			}{
				err: exception.ErrNotFound,
			},
			tenantID: "tenant-1",
			wantErr:  exception.ErrNotFound,
			wantRes: func(t *testing.T, res *model.BranchResponse, repo *stubBranchRepo) {
				assert.Equal(t, "tenant-1", repo.seen.tenantID)
				assert.Equal(t, "branch-2", repo.seen.branchID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubBranchRepo{
				branch: tt.stubRepo.branch,
				err:    tt.stubRepo.err,
			}
			uc := NewBranchUseCase(repo)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.ResolveBranch(ctx, tt.branchID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				if tt.wantRes != nil {
					tt.wantRes(t, res, repo)
				}
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, res, repo)
			}
		})
	}
}

func TestCreateBranch(t *testing.T) {
	tests := []struct {
		name     string
		category string
		req      model.CreateBranchRequest
		tenantID string
		wantErr  error
		wantRes  func(t *testing.T, res *model.BranchResponse, repo *stubBranchRepo)
	}{
		{
			name:     "Positive_CreateBranchUsesTenantContext",
			category: "positive",
			req:      model.CreateBranchRequest{Code: "main", Name: "Main Branch"},
			tenantID: "tenant-1",
			wantErr:  nil,
			wantRes: func(t *testing.T, res *model.BranchResponse, repo *stubBranchRepo) {
				assert.Equal(t, "tenant-1", res.TenantID)
				require.NotNil(t, repo.branch)
				assert.Equal(t, "MAIN", repo.branch.Code)
				assert.Equal(t, "Main Branch", repo.branch.Name)
			},
		},
		{
			name:     "Negative_MissingTenantContext",
			category: "negative",
			req:      model.CreateBranchRequest{Code: "main", Name: "Main Branch"},
			tenantID: "",
			wantErr:  exception.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubBranchRepo{}
			uc := NewBranchUseCase(repo)

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
				tt.wantRes(t, res, repo)
			}
		})
	}
}

func TestListBranches(t *testing.T) {
	tests := []struct {
		name     string
		category string
		stubRepo struct {
			list []*entity.Branch
		}
		tenantID string
		wantErr  error
		wantRes  func(t *testing.T, res []model.BranchResponse, repo *stubBranchRepo)
	}{
		{
			name:     "Positive_ListBranchesUsesTenantScope",
			category: "positive",
			stubRepo: struct {
				list []*entity.Branch
			}{
				list: []*entity.Branch{{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN", Name: "Main"}},
			},
			tenantID: "tenant-1",
			wantErr:  nil,
			wantRes: func(t *testing.T, res []model.BranchResponse, repo *stubBranchRepo) {
				require.Len(t, res, 1)
				assert.Equal(t, "tenant-1", res[0].TenantID)
				assert.Equal(t, "tenant-1", repo.seen.tenantID)
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
			repo := &stubBranchRepo{
				list: tt.stubRepo.list,
			}
			uc := NewBranchUseCase(repo)

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
				tt.wantRes(t, res, repo)
			}
		})
	}
}

func TestUpdateBranch(t *testing.T) {
	code := " sub "
	name := " Branch Office "
	tests := []struct {
		name     string
		category string
		branchID string
		req      model.UpdateBranchRequest
		stubRepo struct {
			branch *entity.Branch
			err    error
		}
		tenantID string
		wantErr  error
		wantRes  func(t *testing.T, res *model.BranchResponse, repo *stubBranchRepo)
	}{
		{
			name:     "Positive_UpdateBranchSanitizesFields",
			category: "positive",
			branchID: "branch-1",
			req:      model.UpdateBranchRequest{Code: &code, Name: &name},
			stubRepo: struct {
				branch *entity.Branch
				err    error
			}{
				branch: &entity.Branch{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN", Name: "Main", Status: entity.BranchStatusActive},
				err:    nil,
			},
			tenantID: "tenant-1",
			wantErr:  nil,
			wantRes: func(t *testing.T, res *model.BranchResponse, repo *stubBranchRepo) {
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
			stubRepo: struct {
				branch *entity.Branch
				err    error
			}{
				err: exception.ErrNotFound,
			},
			tenantID: "tenant-1",
			wantErr:  exception.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubBranchRepo{
				branch: tt.stubRepo.branch,
				err:    tt.stubRepo.err,
			}
			uc := NewBranchUseCase(repo)

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
				tt.wantRes(t, res, repo)
			}
		})
	}
}

func TestDeleteBranch(t *testing.T) {
	tests := []struct {
		name     string
		category string
		branchID string
		tenantID string
		wantErr  error
	}{
		{
			name:     "Positive_DeleteBranch",
			category: "positive",
			branchID: "branch-1",
			tenantID: "tenant-1",
			wantErr:  nil,
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
			repo := &stubBranchRepo{}
			uc := NewBranchUseCase(repo)

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
