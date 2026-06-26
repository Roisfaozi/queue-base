package repository

import (
	"context"
	"testing"
	"time"

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

func TestSettingsRepository_CreateAndFindByID(t *testing.T) {
	db := newSettingsTestDB(t)
	repo := NewSettingsRepository(db)
	ctx := context.Background()

	setting := &entity.Setting{
		ID:        "s-1",
		TenantID:  "t-1",
		ScopeType: entity.ScopeTypeTenant,
		ScopeID:   "t-1",
		Key:       "theme",
		Value:     "dark",
		IsActive:  true,
	}

	err := repo.Create(ctx, setting)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, "t-1", "s-1")
	require.NoError(t, err)
	assert.Equal(t, "dark", found.Value)
	assert.Equal(t, "theme", found.Key)
}

func TestSettingsRepository_FindByScope(t *testing.T) {
	db := newSettingsTestDB(t)
	repo := NewSettingsRepository(db)
	ctx := context.Background()

	// Given active setting
	err := repo.Create(ctx, &entity.Setting{
		ID:        "s-1",
		TenantID:  "t-1",
		ScopeType: entity.ScopeTypeBranch,
		ScopeID:   "b-1",
		Key:       "pharma_flow",
		Value:     "true",
		IsActive:  true,
	})
	require.NoError(t, err)

	// Given inactive setting
	err = repo.Create(ctx, &entity.Setting{
		ID:        "s-2",
		TenantID:  "t-1",
		ScopeType: entity.ScopeTypeService,
		ScopeID:   "svc-1",
		Key:       "pharma_flow",
		Value:     "false",
		IsActive:  true, // create as true first
	})
	require.NoError(t, err)
	err = repo.Update(ctx, &entity.Setting{ID: "s-2", TenantID: "t-1", IsActive: false})
	require.NoError(t, err)

	// Test find active
	found, err := repo.FindByScope(ctx, "t-1", entity.ScopeTypeBranch, "b-1", "pharma_flow")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "true", found.Value)

	// Test find inactive (should return nil, nil)
	notFound, err := repo.FindByScope(ctx, "t-1", entity.ScopeTypeService, "svc-1", "pharma_flow")
	require.NoError(t, err)
	assert.Nil(t, notFound)

	// Test missing
	missing, err := repo.FindByScope(ctx, "t-1", entity.ScopeTypeBranch, "b-99", "pharma_flow")
	require.NoError(t, err)
	assert.Nil(t, missing)
}

func TestSettingsRepository_FindAllByKey(t *testing.T) {
	db := newSettingsTestDB(t)
	repo := NewSettingsRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &entity.Setting{ID: "s-1", TenantID: "t-1", Key: "theme", Value: "dark", IsActive: true})
	repo.Create(ctx, &entity.Setting{ID: "s-2", TenantID: "t-1", Key: "theme", Value: "light", IsActive: true})
	repo.Create(ctx, &entity.Setting{ID: "s-3", TenantID: "t-1", Key: "theme", Value: "blue", IsActive: true})
	repo.Update(ctx, &entity.Setting{ID: "s-3", TenantID: "t-1", IsActive: false})
	repo.Create(ctx, &entity.Setting{ID: "s-4", TenantID: "t-2", Key: "theme", Value: "red", IsActive: true})

	found, err := repo.FindAllByKey(ctx, "t-1", "theme")
	require.NoError(t, err)
	assert.Len(t, found, 2)
}

func TestSettingsRepository_UpdateAndDelete(t *testing.T) {
	db := newSettingsTestDB(t)
	repo := NewSettingsRepository(db)
	ctx := context.Background()

	setting := &entity.Setting{
		ID:        "s-1",
		TenantID:  "t-1",
		ScopeType: entity.ScopeTypeTenant,
		ScopeID:   "t-1",
		Key:       "theme",
		Value:     "dark",
		IsActive:  true,
	}
	repo.Create(ctx, setting)

	// Update
	now := time.Now().UnixMilli()
	err := repo.Update(ctx, &entity.Setting{
		ID:        "s-1",
		TenantID:  "t-1",
		Value:     "light",
		IsActive:  false,
		UpdatedAt: now,
	})
	require.NoError(t, err)

	updated, _ := repo.FindByID(ctx, "t-1", "s-1")
	assert.Equal(t, "light", updated.Value)
	assert.False(t, updated.IsActive)
	assert.Equal(t, now, updated.UpdatedAt)

	// Update missing
	err = repo.Update(ctx, &entity.Setting{ID: "s-99", TenantID: "t-1"})
	assert.ErrorIs(t, err, exception.ErrNotFound)

	// Delete
	err = repo.Delete(ctx, "t-1", "s-1")
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, "t-1", "s-1")
	assert.Error(t, err) // GORM record not found

	// Delete missing
	err = repo.Delete(ctx, "t-1", "s-99")
	assert.ErrorIs(t, err, exception.ErrNotFound)
}
