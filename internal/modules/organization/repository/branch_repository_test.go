package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newBranchTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.Branch{}))
	return db
}

func TestBranchRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("InterfaceAndEntity", func(t *testing.T) {
		t.Run("InterfaceExists", func(t *testing.T) {
			var _ BranchRepository = (*branchRepository)(nil)
		})
		t.Run("TableName", func(t *testing.T) {
			if got := (entity.Branch{}).TableName(); got != "branches" {
				t.Fatalf("expected branches, got %s", got)
			}
		})
	})

	t.Run("CreateAndFindByID", func(t *testing.T) {
		tests := []struct {
			name   string
			branch *entity.Branch
			assert func(t *testing.T, repo BranchRepository)
		}{
			{
				name: "Positive_CreateSuccess",
				branch: &entity.Branch{
					ID:       "b-1",
					TenantID: "t-1",
					Code:     "B1",
					Name:     "Branch 1",
					Status:   entity.BranchStatusActive,
				},
				assert: func(t *testing.T, repo BranchRepository) {
					found, err := repo.FindByID(ctx, "t-1", "b-1")
					require.NoError(t, err)
					assert.Equal(t, "Branch 1", found.Name)
					assert.Equal(t, "B1", found.Code)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newBranchTestDB(t)
				repo := NewBranchRepository(db)

				err := repo.Create(ctx, tt.branch)
				require.NoError(t, err)
				tt.assert(t, repo)
			})
		}
	})

	t.Run("FindAll", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func(repo BranchRepository)
			tenantID string
			wantLen  int
		}{
			{
				name: "Positive_FindsAllByTenant",
				setup: func(repo BranchRepository) {
					_ = repo.Create(ctx, &entity.Branch{ID: "b-1", TenantID: "t-1", Code: "B1", Name: "B1"})
					_ = repo.Create(ctx, &entity.Branch{ID: "b-2", TenantID: "t-1", Code: "B2", Name: "B2"})
					_ = repo.Create(ctx, &entity.Branch{ID: "b-3", TenantID: "t-2", Code: "B3", Name: "B3"})
				},
				tenantID: "t-1",
				wantLen:  2,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newBranchTestDB(t)
				repo := NewBranchRepository(db)
				tt.setup(repo)

				found, err := repo.FindAll(ctx, tt.tenantID)
				require.NoError(t, err)
				assert.Len(t, found, tt.wantLen)
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		tests := []struct {
			name    string
			setup   func(repo BranchRepository)
			req     *entity.Branch
			wantErr error
			assert  func(t *testing.T, repo BranchRepository, now int64)
		}{
			{
				name: "Positive_UpdateSuccess",
				setup: func(repo BranchRepository) {
					_ = repo.Create(ctx, &entity.Branch{ID: "b-1", TenantID: "t-1", Code: "B1", Name: "Branch 1", Status: entity.BranchStatusActive})
				},
				req: &entity.Branch{ID: "b-1", TenantID: "t-1", Code: "B1-NEW", Name: "Branch One", Status: entity.BranchStatusInactive},
				assert: func(t *testing.T, repo BranchRepository, now int64) {
					updated, err := repo.FindByID(ctx, "t-1", "b-1")
					require.NoError(t, err)
					assert.Equal(t, "B1-NEW", updated.Code)
					assert.Equal(t, "Branch One", updated.Name)
					assert.Equal(t, entity.BranchStatusInactive, updated.Status)
					assert.Equal(t, now, updated.UpdatedAt)
				},
			},
			{
				name:    "Negative_UpdateMissing",
				setup:   func(repo BranchRepository) {},
				req:     &entity.Branch{ID: "b-99", TenantID: "t-1"},
				wantErr: exception.ErrNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newBranchTestDB(t)
				repo := NewBranchRepository(db)
				tt.setup(repo)

				now := time.Now().UnixMilli()
				if tt.req.UpdatedAt == 0 && tt.req.ID != "b-99" {
					tt.req.UpdatedAt = now
				}

				err := repo.Update(ctx, tt.req)
				if tt.wantErr != nil {
					assert.ErrorIs(t, err, tt.wantErr)
					return
				}
				require.NoError(t, err)
				if tt.assert != nil {
					tt.assert(t, repo, now)
				}
			})
		}
	})

	t.Run("Delete", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func(repo BranchRepository)
			tenantID string
			branchID string
			wantErr  error
			assert   func(t *testing.T, repo BranchRepository)
		}{
			{
				name: "Positive_DeleteSuccess",
				setup: func(repo BranchRepository) {
					_ = repo.Create(ctx, &entity.Branch{ID: "b-1", TenantID: "t-1", Code: "B1", Name: "B1"})
				},
				tenantID: "t-1",
				branchID: "b-1",
				assert: func(t *testing.T, repo BranchRepository) {
					_, err := repo.FindByID(ctx, "t-1", "b-1")
					assert.Error(t, err)
				},
			},
			{
				name:     "Negative_DeleteMissing",
				setup:    func(repo BranchRepository) {},
				tenantID: "t-1",
				branchID: "b-99",
				wantErr:  exception.ErrNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newBranchTestDB(t)
				repo := NewBranchRepository(db)
				tt.setup(repo)

				err := repo.Delete(ctx, tt.tenantID, tt.branchID)
				if tt.wantErr != nil {
					assert.ErrorIs(t, err, tt.wantErr)
					return
				}
				require.NoError(t, err)
				if tt.assert != nil {
					tt.assert(t, repo)
				}
			})
		}
	})
}
