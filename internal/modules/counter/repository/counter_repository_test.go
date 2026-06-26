package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newCounterTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.Counter{}))
	return db
}

func TestCounterRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateAndFindByID", func(t *testing.T) {
		tests := []struct {
			name    string
			counter *entity.Counter
			assert  func(t *testing.T, repo CounterRepository)
		}{
			{
				name: "Positive_CreateSuccess",
				counter: &entity.Counter{
					ID:       "c-1",
					TenantID: "t-1",
					BranchID: "b-1",
					Code:     "C1",
					Name:     "Counter 1",
					Status:   entity.CounterStatusActive,
				},
				assert: func(t *testing.T, repo CounterRepository) {
					found, err := repo.FindByID(ctx, "t-1", "c-1")
					require.NoError(t, err)
					assert.Equal(t, "Counter 1", found.Name)
					assert.Equal(t, "C1", found.Code)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newCounterTestDB(t)
				repo := NewCounterRepository(db)

				err := repo.Create(ctx, tt.counter)
				require.NoError(t, err)
				tt.assert(t, repo)
			})
		}
	})

	t.Run("FindAll", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func(repo CounterRepository)
			tenantID string
			wantLen  int
		}{
			{
				name: "Positive_FindsAllByTenant",
				setup: func(repo CounterRepository) {
					_ = repo.Create(ctx, &entity.Counter{ID: "c-1", TenantID: "t-1", BranchID: "b-1", Code: "C1", Name: "C1"})
					_ = repo.Create(ctx, &entity.Counter{ID: "c-2", TenantID: "t-1", BranchID: "b-1", Code: "C2", Name: "C2"})
					_ = repo.Create(ctx, &entity.Counter{ID: "c-3", TenantID: "t-2", BranchID: "b-1", Code: "C3", Name: "C3"})
				},
				tenantID: "t-1",
				wantLen:  2,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newCounterTestDB(t)
				repo := NewCounterRepository(db)
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
			setup   func(repo CounterRepository)
			req     *entity.Counter
			wantErr error
			assert  func(t *testing.T, repo CounterRepository, now int64)
		}{
			{
				name: "Positive_UpdateSuccess",
				setup: func(repo CounterRepository) {
					_ = repo.Create(ctx, &entity.Counter{ID: "c-1", TenantID: "t-1", BranchID: "b-1", Code: "C1", Name: "Counter 1", Status: entity.CounterStatusActive})
				},
				req: &entity.Counter{ID: "c-1", TenantID: "t-1", Code: "C1-NEW", Name: "Counter One", Status: entity.CounterStatusInactive},
				assert: func(t *testing.T, repo CounterRepository, now int64) {
					updated, err := repo.FindByID(ctx, "t-1", "c-1")
					require.NoError(t, err)
					assert.Equal(t, "C1-NEW", updated.Code)
					assert.Equal(t, "Counter One", updated.Name)
					assert.Equal(t, entity.CounterStatusInactive, updated.Status)
					assert.Equal(t, now, updated.UpdatedAt)
				},
			},
			{
				name:    "Negative_UpdateMissing",
				setup:   func(repo CounterRepository) {},
				req:     &entity.Counter{ID: "c-99", TenantID: "t-1"},
				wantErr: exception.ErrNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newCounterTestDB(t)
				repo := NewCounterRepository(db)
				tt.setup(repo)

				now := time.Now().UnixMilli()
				if tt.req.UpdatedAt == 0 && tt.req.ID != "c-99" {
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
			name      string
			setup     func(repo CounterRepository)
			tenantID  string
			counterID string
			wantErr   error
			assert    func(t *testing.T, repo CounterRepository)
		}{
			{
				name: "Positive_DeleteSuccess",
				setup: func(repo CounterRepository) {
					_ = repo.Create(ctx, &entity.Counter{ID: "c-1", TenantID: "t-1", BranchID: "b-1", Code: "C1", Name: "C1"})
				},
				tenantID:  "t-1",
				counterID: "c-1",
				assert: func(t *testing.T, repo CounterRepository) {
					_, err := repo.FindByID(ctx, "t-1", "c-1")
					assert.Error(t, err)
				},
			},
			{
				name:      "Negative_DeleteMissing",
				setup:     func(repo CounterRepository) {},
				tenantID:  "t-1",
				counterID: "c-99",
				wantErr:   exception.ErrNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newCounterTestDB(t)
				repo := NewCounterRepository(db)
				tt.setup(repo)

				err := repo.Delete(ctx, tt.tenantID, tt.counterID)
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
