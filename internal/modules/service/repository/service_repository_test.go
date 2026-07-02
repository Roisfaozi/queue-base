package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.Service{}))
	return db
}

func TestServiceRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateAndFindByID", func(t *testing.T) {
		tests := []struct {
			name    string
			service *entity.Service
			assert  func(t *testing.T, repo ServiceRepository)
		}{
			{
				name: "Positive_CreateSuccess",
				service: &entity.Service{
					ID:                  "s-1",
					TenantID:            "t-1",
					Code:                "S1",
					Name:                "Service 1",
					Status:              entity.ServiceStatusActive,
					IsPharmacy:          true,
					IsPharmacyReception: false,
				},
				assert: func(t *testing.T, repo ServiceRepository) {
					found, err := repo.FindByID(ctx, "t-1", "s-1")
					require.NoError(t, err)
					assert.Equal(t, "Service 1", found.Name)
					assert.Equal(t, "S1", found.Code)
					assert.True(t, found.IsPharmacy)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newServiceTestDB(t)
				repo := NewServiceRepository(db)

				err := repo.Create(ctx, tt.service)
				require.NoError(t, err)
				tt.assert(t, repo)
			})
		}
	})

	t.Run("FindAll", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func(repo ServiceRepository)
			tenantID string
			wantLen  int
		}{
			{
				name: "Positive_FindsAllByTenant",
				setup: func(repo ServiceRepository) {
					_ = repo.Create(ctx, &entity.Service{ID: "s-1", TenantID: "t-1", Code: "S1", Name: "S1"})
					_ = repo.Create(ctx, &entity.Service{ID: "s-2", TenantID: "t-1", Code: "S2", Name: "S2"})
					_ = repo.Create(ctx, &entity.Service{ID: "s-3", TenantID: "t-2", Code: "S3", Name: "S3"})
				},
				tenantID: "t-1",
				wantLen:  2,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newServiceTestDB(t)
				repo := NewServiceRepository(db)
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
			setup   func(repo ServiceRepository)
			req     *entity.Service
			wantErr error
			assert  func(t *testing.T, repo ServiceRepository, now int64)
		}{
			{
				name: "Positive_UpdateSuccess",
				setup: func(repo ServiceRepository) {
					_ = repo.Create(ctx, &entity.Service{ID: "s-1", TenantID: "t-1", Code: "S1", Name: "Service 1", Status: entity.ServiceStatusActive})
				},
				req: &entity.Service{ID: "s-1", TenantID: "t-1", Code: "S1-NEW", Name: "Service One", Status: entity.ServiceStatusInactive, IsPharmacy: true, IsPharmacyReception: true},
				assert: func(t *testing.T, repo ServiceRepository, now int64) {
					updated, err := repo.FindByID(ctx, "t-1", "s-1")
					require.NoError(t, err)
					assert.Equal(t, "S1-NEW", updated.Code)
					assert.Equal(t, "Service One", updated.Name)
					assert.Equal(t, entity.ServiceStatusInactive, updated.Status)
					assert.True(t, updated.IsPharmacy)
					assert.True(t, updated.IsPharmacyReception)
					assert.InDelta(t, now, updated.UpdatedAt, 5)
				},
			},
			{
				name:    "Negative_UpdateMissing",
				setup:   func(repo ServiceRepository) {},
				req:     &entity.Service{ID: "s-99", TenantID: "t-1"},
				wantErr: exception.ErrNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newServiceTestDB(t)
				repo := NewServiceRepository(db)
				tt.setup(repo)

				now := time.Now().UnixMilli()
				if tt.req.UpdatedAt == 0 && tt.req.ID != "s-99" {
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
			setup     func(repo ServiceRepository)
			tenantID  string
			serviceID string
			wantErr   error
			assert    func(t *testing.T, repo ServiceRepository)
		}{
			{
				name: "Positive_DeleteSuccess",
				setup: func(repo ServiceRepository) {
					_ = repo.Create(ctx, &entity.Service{ID: "s-1", TenantID: "t-1", Code: "S1", Name: "S1"})
				},
				tenantID:  "t-1",
				serviceID: "s-1",
				assert: func(t *testing.T, repo ServiceRepository) {
					_, err := repo.FindByID(ctx, "t-1", "s-1")
					assert.Error(t, err)
				},
			},
			{
				name:      "Negative_DeleteMissing",
				setup:     func(repo ServiceRepository) {},
				tenantID:  "t-1",
				serviceID: "s-99",
				wantErr:   exception.ErrNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newServiceTestDB(t)
				repo := NewServiceRepository(db)
				tt.setup(repo)

				err := repo.Delete(ctx, tt.tenantID, tt.serviceID)
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
