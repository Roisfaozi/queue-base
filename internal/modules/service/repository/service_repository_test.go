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

func TestServiceRepository_CreateAndFindByID(t *testing.T) {
	db := newServiceTestDB(t)
	repo := NewServiceRepository(db)
	ctx := context.Background()

	service := &entity.Service{
		ID:                  "s-1",
		TenantID:            "t-1",
		Code:                "S1",
		Name:                "Service 1",
		Status:              entity.ServiceStatusActive,
		IsPharmacy:          true,
		IsPharmacyReception: false,
	}

	err := repo.Create(ctx, service)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, "t-1", "s-1")
	require.NoError(t, err)
	assert.Equal(t, "Service 1", found.Name)
	assert.Equal(t, "S1", found.Code)
	assert.True(t, found.IsPharmacy)
}

func TestServiceRepository_FindAll(t *testing.T) {
	db := newServiceTestDB(t)
	repo := NewServiceRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &entity.Service{ID: "s-1", TenantID: "t-1", Code: "S1", Name: "S1"})
	repo.Create(ctx, &entity.Service{ID: "s-2", TenantID: "t-1", Code: "S2", Name: "S2"})
	repo.Create(ctx, &entity.Service{ID: "s-3", TenantID: "t-2", Code: "S3", Name: "S3"})

	found, err := repo.FindAll(ctx, "t-1")
	require.NoError(t, err)
	assert.Len(t, found, 2)
}

func TestServiceRepository_UpdateAndDelete(t *testing.T) {
	db := newServiceTestDB(t)
	repo := NewServiceRepository(db)
	ctx := context.Background()

	service := &entity.Service{
		ID:       "s-1",
		TenantID: "t-1",
		Code:     "S1",
		Name:     "Service 1",
		Status:   entity.ServiceStatusActive,
	}
	repo.Create(ctx, service)

	// Update
	now := time.Now().UnixMilli()
	err := repo.Update(ctx, &entity.Service{
		ID:                  "s-1",
		TenantID:            "t-1",
		Code:                "S1-NEW",
		Name:                "Service One",
		Status:              entity.ServiceStatusInactive,
		IsPharmacy:          true,
		IsPharmacyReception: true,
		UpdatedAt:           now,
	})
	require.NoError(t, err)

	updated, _ := repo.FindByID(ctx, "t-1", "s-1")
	assert.Equal(t, "S1-NEW", updated.Code)
	assert.Equal(t, "Service One", updated.Name)
	assert.Equal(t, entity.ServiceStatusInactive, updated.Status)
	assert.True(t, updated.IsPharmacy)
	assert.True(t, updated.IsPharmacyReception)
	assert.Equal(t, now, updated.UpdatedAt)

	// Update missing
	err = repo.Update(ctx, &entity.Service{ID: "s-99", TenantID: "t-1"})
	assert.ErrorIs(t, err, exception.ErrNotFound)

	// Delete
	err = repo.Delete(ctx, "t-1", "s-1")
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, "t-1", "s-1")
	assert.Error(t, err)

	// Delete missing
	err = repo.Delete(ctx, "t-1", "s-99")
	assert.ErrorIs(t, err, exception.ErrNotFound)
}
