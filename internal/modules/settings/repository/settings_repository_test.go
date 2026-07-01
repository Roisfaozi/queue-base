package repository

import (
	"context"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/settings/entity"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newSettingsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.Setting{}))
	return db
}

func TestSettingsRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateAndFindByID", func(t *testing.T) {
		tests := []struct {
			name    string
			setting *entity.Setting
			assert  func(t *testing.T, repo SettingsRepository)
		}{
			{
				name: "Positive_CreateSuccess",
				setting: &entity.Setting{
					ID:        "s-1",
					TenantID:  "t-1",
					ScopeType: entity.ScopeTypeTenant,
					ScopeID:   "t-1",
					Key:       "theme",
					Value:     "dark",
					IsActive:  true,
				},
				assert: func(t *testing.T, repo SettingsRepository) {
					found, err := repo.FindByID(ctx, "t-1", "s-1")
					require.NoError(t, err)
					assert.Equal(t, "dark", found.Value)
					assert.Equal(t, "theme", found.Key)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newSettingsTestDB(t)
				repo := NewSettingsRepository(db)

				err := repo.Create(ctx, tt.setting)
				require.NoError(t, err)
				tt.assert(t, repo)
			})
		}
	})

	t.Run("FindByScope", func(t *testing.T) {
		tests := []struct {
			name      string
			setup     func(repo SettingsRepository)
			tenantID  string
			scopeType string
			scopeID   string
			key       string
			wantErr   error
			wantVal   *string
		}{
			{
				name: "Positive_FoundActive",
				setup: func(repo SettingsRepository) {
					_ = repo.Create(ctx, &entity.Setting{ID: "s-1", TenantID: "t-1", ScopeType: entity.ScopeTypeBranch, ScopeID: "b-1", Key: "pharma_flow", Value: "true", IsActive: true})
				},
				tenantID:  "t-1",
				scopeType: entity.ScopeTypeBranch,
				scopeID:   "b-1",
				key:       "pharma_flow",
				wantVal:   func() *string { s := "true"; return &s }(),
			},
			{
				name: "Negative_FoundInactiveReturnsNil",
				setup: func(repo SettingsRepository) {
					_ = repo.Create(ctx, &entity.Setting{ID: "s-2", TenantID: "t-1", ScopeType: entity.ScopeTypeService, ScopeID: "svc-1", Key: "pharma_flow", Value: "false", IsActive: true})
					_ = repo.Update(ctx, &entity.Setting{ID: "s-2", TenantID: "t-1", IsActive: false})
				},
				tenantID:  "t-1",
				scopeType: entity.ScopeTypeService,
				scopeID:   "svc-1",
				key:       "pharma_flow",
				wantVal:   nil,
			},
			{
				name:      "Negative_MissingReturnsNil",
				setup:     func(repo SettingsRepository) {},
				tenantID:  "t-1",
				scopeType: entity.ScopeTypeBranch,
				scopeID:   "b-99",
				key:       "pharma_flow",
				wantVal:   nil,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newSettingsTestDB(t)
				repo := NewSettingsRepository(db)
				tt.setup(repo)

				found, err := repo.FindByScope(ctx, tt.tenantID, tt.scopeType, tt.scopeID, tt.key)
				if tt.wantErr != nil {
					assert.ErrorIs(t, err, tt.wantErr)
					return
				}
				require.NoError(t, err)

				if tt.wantVal == nil {
					assert.Nil(t, found)
				} else {
					require.NotNil(t, found)
					assert.Equal(t, *tt.wantVal, found.Value)
				}
			})
		}
	})

	t.Run("FindAllByKey", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func(repo SettingsRepository)
			tenantID string
			key      string
			wantLen  int
		}{
			{
				name: "Positive_FindsActiveOnlyForTenant",
				setup: func(repo SettingsRepository) {
					_ = repo.Create(ctx, &entity.Setting{ID: "s-1", TenantID: "t-1", Key: "theme", Value: "dark", IsActive: true})
					_ = repo.Create(ctx, &entity.Setting{ID: "s-2", TenantID: "t-1", Key: "theme", Value: "light", IsActive: true})
					_ = repo.Create(ctx, &entity.Setting{ID: "s-3", TenantID: "t-1", Key: "theme", Value: "blue", IsActive: true})
					_ = repo.Update(ctx, &entity.Setting{ID: "s-3", TenantID: "t-1", IsActive: false})
					_ = repo.Create(ctx, &entity.Setting{ID: "s-4", TenantID: "t-2", Key: "theme", Value: "red", IsActive: true})
				},
				tenantID: "t-1",
				key:      "theme",
				wantLen:  2,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newSettingsTestDB(t)
				repo := NewSettingsRepository(db)
				tt.setup(repo)

				found, err := repo.FindAllByKey(ctx, tt.tenantID, tt.key)
				require.NoError(t, err)
				assert.Len(t, found, tt.wantLen)
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		tests := []struct {
			name    string
			setup   func(repo SettingsRepository)
			req     *entity.Setting
			wantErr error
			assert  func(t *testing.T, repo SettingsRepository)
		}{
			{
				name: "Positive_UpdateSuccess",
				setup: func(repo SettingsRepository) {
					_ = repo.Create(ctx, &entity.Setting{ID: "s-1", TenantID: "t-1", Value: "dark", IsActive: true})
				},
				req: &entity.Setting{ID: "s-1", TenantID: "t-1", Value: "light", IsActive: false},
				assert: func(t *testing.T, repo SettingsRepository) {
					updated, err := repo.FindByID(ctx, "t-1", "s-1")
					require.NoError(t, err)
					assert.Equal(t, "light", updated.Value)
					assert.False(t, updated.IsActive)
				},
			},
			{
				name:    "Negative_UpdateMissing",
				setup:   func(repo SettingsRepository) {},
				req:     &entity.Setting{ID: "s-99", TenantID: "t-1"},
				wantErr: exception.ErrNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newSettingsTestDB(t)
				repo := NewSettingsRepository(db)
				tt.setup(repo)

				err := repo.Update(ctx, tt.req)
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

	t.Run("Delete", func(t *testing.T) {
		tests := []struct {
			name      string
			setup     func(repo SettingsRepository)
			tenantID  string
			settingID string
			wantErr   error
			assert    func(t *testing.T, repo SettingsRepository)
		}{
			{
				name: "Positive_DeleteSuccess",
				setup: func(repo SettingsRepository) {
					_ = repo.Create(ctx, &entity.Setting{ID: "s-1", TenantID: "t-1"})
				},
				tenantID:  "t-1",
				settingID: "s-1",
				assert: func(t *testing.T, repo SettingsRepository) {
					_, err := repo.FindByID(ctx, "t-1", "s-1")
					assert.Error(t, err)
				},
			},
			{
				name:      "Negative_DeleteMissing",
				setup:     func(repo SettingsRepository) {},
				tenantID:  "t-1",
				settingID: "s-99",
				wantErr:   exception.ErrNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newSettingsTestDB(t)
				repo := NewSettingsRepository(db)
				tt.setup(repo)

				err := repo.Delete(ctx, tt.tenantID, tt.settingID)
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
